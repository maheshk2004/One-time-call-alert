package services

import (
	"context"
	"fmt"
	"time"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallService struct {
	callRepo     *repositories.CallRepository
	leadRepo     *repositories.LeadRepository
	userRepo     *repositories.UserRepository
	followUpRepo *repositories.FollowUpRepository
	auditService *AuditService
}

func NewCallService(
	callRepo *repositories.CallRepository,
	leadRepo *repositories.LeadRepository,
	userRepo *repositories.UserRepository,
	followUpRepo *repositories.FollowUpRepository,
	auditService *AuditService,
) *CallService {
	return &CallService{
		callRepo:     callRepo,
		leadRepo:     leadRepo,
		userRepo:     userRepo,
		followUpRepo: followUpRepo,
		auditService: auditService,
	}
}

type RecordCallInput struct {
	LeadID             primitive.ObjectID `json:"leadId"`
	CallStartedAt      *time.Time         `json:"callStartedAt"`
	CallEndedAt        *time.Time         `json:"callEndedAt"`
	DurationSeconds    int                `json:"durationSeconds"`
	CallStatus         models.CallStatus  `json:"callStatus" binding:"required"`
	CallOutcome        models.CallOutcome `json:"callOutcome" binding:"required"`
	RecordingConsent   bool               `json:"recordingConsent"`
	RecordingAvailable bool               `json:"recordingAvailable"`
	RecordingURL       string             `json:"recordingUrl"`
	Notes              string             `json:"notes"`
	NextFollowUpAt     *time.Time         `json:"nextFollowUpAt,omitempty"`
}

func (s *CallService) RecordCall(ctx context.Context, input RecordCallInput, actor *models.User) (*models.CallAttempt, error) {
	lead, err := s.leadRepo.FindByID(ctx, input.LeadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}

	callStart := time.Now()
	if input.CallStartedAt != nil {
		callStart = *input.CallStartedAt
	}

	// Calculate attempt number
	currentCount, _ := s.callRepo.CountByLeadID(ctx, lead.ID)
	attemptNumber := currentCount + 1

	call := &models.CallAttempt{
		LeadID:              lead.ID,
		SalespersonID:       actor.ID,
		AttemptNumber:       attemptNumber,
		CallStartedAt:       callStart,
		CallEndedAt:         input.CallEndedAt,
		DurationSeconds:     input.DurationSeconds,
		CallStatus:          input.CallStatus,
		CallOutcome:         input.CallOutcome,
		RecordingConsent:    input.RecordingConsent,
		RecordingAvailable:  input.RecordingAvailable,
		RecordingURL:        input.RecordingURL,
		TranscriptAvailable: false,
		ManuallyReviewed:    true, // salesperson entered it manually
		Notes:               input.Notes,
	}

	if err := s.callRepo.Create(ctx, call); err != nil {
		return nil, err
	}

	// Complete any pending follow-up with this call attempt
	pendingFollowUp, _ := s.followUpRepo.FindPendingByLeadID(ctx, lead.ID)
	if pendingFollowUp != nil {
		_ = s.followUpRepo.CompleteFollowUp(ctx, pendingFollowUp.ID, &call.ID)
	}

	// Update Lead state
	lead.CallAttemptCount = attemptNumber
	lead.LastCallAt = &callStart
	if lead.FirstCallAt == nil {
		lead.FirstCallAt = &callStart
	}

	// Update status based on call outcome
	switch input.CallOutcome {
	case models.CallOutcomeInterested:
		lead.Status = models.LeadStatusInterested
	case models.CallOutcomeNotInterested:
		lead.Status = models.LeadStatusNotInterested
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)
	case models.CallOutcomeDoNotCall:
		lead.Status = models.LeadStatusDoNotCall
		lead.DoNotCall = true
		lead.DoNotCallReason = "Customer requested during call"
		lead.DoNotCallDetectedBy = fmt.Sprintf("Salesperson (%s)", actor.Name)
		now := time.Now()
		lead.DoNotCallAt = &now
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)
	case models.CallOutcomeCallBackLater:
		if input.NextFollowUpAt != nil {
			lead.Status = models.LeadStatusFollowUpScheduled
			lead.FollowUpStatus = models.FollowUpStatusScheduled
			lead.NextFollowUpAt = input.NextFollowUpAt
			// Create scheduled follow-up
			_ = s.followUpRepo.Create(ctx, &models.FollowUp{
				LeadID:        lead.ID,
				SalespersonID: actor.ID,
				DueAt:         *input.NextFollowUpAt,
				ScheduledAt:   time.Now(),
				Status:        models.FollowUpTaskScheduled,
				Notes:         fmt.Sprintf("Call back scheduled after attempt #%d", attemptNumber),
			})
		} else {
			lead.Status = models.LeadStatusCalledOnce
			lead.FollowUpStatus = models.FollowUpStatusPending
		}
	case models.CallOutcomeInfoRequested:
		lead.Status = models.LeadStatusInfoRequested
		if input.NextFollowUpAt != nil {
			lead.FollowUpStatus = models.FollowUpStatusScheduled
			lead.NextFollowUpAt = input.NextFollowUpAt
			_ = s.followUpRepo.Create(ctx, &models.FollowUp{
				LeadID:        lead.ID,
				SalespersonID: actor.ID,
				DueAt:         *input.NextFollowUpAt,
				ScheduledAt:   time.Now(),
				Status:        models.FollowUpTaskScheduled,
				Notes:         "Follow-up for requested info",
			})
		}
	case models.CallOutcomeUndecided:
		lead.Status = models.LeadStatusUndecided
		if input.NextFollowUpAt != nil {
			lead.FollowUpStatus = models.FollowUpStatusScheduled
			lead.NextFollowUpAt = input.NextFollowUpAt
			_ = s.followUpRepo.Create(ctx, &models.FollowUp{
				LeadID:        lead.ID,
				SalespersonID: actor.ID,
				DueAt:         *input.NextFollowUpAt,
				ScheduledAt:   time.Now(),
				Status:        models.FollowUpTaskScheduled,
				Notes:         "Follow-up for undecided lead",
			})
		}
	case models.CallOutcomeConverted:
		lead.Status = models.LeadStatusConverted
		lead.FollowUpStatus = models.FollowUpStatusCompleted
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)
	case models.CallOutcomeLost:
		lead.Status = models.LeadStatusLost
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)
	default:
		// No answer, busy, switched off, unreachable, follow-up required
		if attemptNumber == 1 {
			lead.Status = models.LeadStatusCalledOnce
			lead.FollowUpStatus = models.FollowUpStatusPending
		} else if attemptNumber == 2 {
			lead.Status = models.LeadStatusFollowUpRequired
			lead.FollowUpStatus = models.FollowUpStatusPending
		} else {
			lead.Status = models.LeadStatusFollowUpRequired
			lead.FollowUpStatus = models.FollowUpStatusPending
		}
	}

	if err := s.leadRepo.Update(ctx, lead); err != nil {
		return nil, err
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "CALL_RECORDED", "call_attempt", call.ID.Hex(), nil, map[string]interface{}{
		"leadId":        lead.ID.Hex(),
		"attemptNumber": attemptNumber,
		"outcome":       input.CallOutcome,
		"status":        input.CallStatus,
	}, fmt.Sprintf("Recorded call attempt #%d", attemptNumber), nil)

	s.populateSalesperson(ctx, call)
	return call, nil
}

func (s *CallService) GetCallsForLead(ctx context.Context, leadID primitive.ObjectID) ([]*models.CallAttempt, error) {
	calls, err := s.callRepo.FindByLeadID(ctx, leadID)
	if err != nil {
		return nil, err
	}
	for _, call := range calls {
		s.populateSalesperson(ctx, call)
	}
	return calls, nil
}

func (s *CallService) GetCallByID(ctx context.Context, id primitive.ObjectID) (*models.CallAttempt, error) {
	call, err := s.callRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.populateSalesperson(ctx, call)
	return call, nil
}

func (s *CallService) populateSalesperson(ctx context.Context, call *models.CallAttempt) {
	user, err := s.userRepo.FindByID(ctx, call.SalespersonID)
	if err == nil && user != nil {
		call.SalespersonUser = &models.UserSummary{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		}
	}
}
