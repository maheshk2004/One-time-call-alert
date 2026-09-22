package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowUpService struct {
	followUpRepo *repositories.FollowUpRepository
	leadRepo     *repositories.LeadRepository
	userRepo     *repositories.UserRepository
	settingsRepo *repositories.SettingsRepository
	notifService *NotificationService
	emailService *EmailService
	auditService *AuditService
}

func NewFollowUpService(
	followUpRepo *repositories.FollowUpRepository,
	leadRepo *repositories.LeadRepository,
	userRepo *repositories.UserRepository,
	settingsRepo *repositories.SettingsRepository,
	notifService *NotificationService,
	emailService *EmailService,
	auditService *AuditService,
) *FollowUpService {
	return &FollowUpService{
		followUpRepo: followUpRepo,
		leadRepo:     leadRepo,
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		notifService: notifService,
		emailService: emailService,
		auditService: auditService,
	}
}

type ScheduleFollowUpInput struct {
	LeadID primitive.ObjectID `json:"leadId" binding:"required"`
	DueAt  time.Time          `json:"dueAt" binding:"required"`
	Notes  string             `json:"notes"`
}

func (s *FollowUpService) ScheduleFollowUp(ctx context.Context, input ScheduleFollowUpInput, actor *models.User) (*models.FollowUp, error) {
	lead, err := s.leadRepo.FindByID(ctx, input.LeadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}

	if lead.DoNotCall || lead.Status == models.LeadStatusDoNotCall || lead.Status == models.LeadStatusNotInterested {
		return nil, fmt.Errorf("cannot schedule follow-up: lead is marked as %s or Do-Not-Call", lead.Status)
	}

	salespersonID := actor.ID
	if lead.AssignedTo != nil {
		salespersonID = *lead.AssignedTo
	}

	// Cancel any previous pending follow-ups for this lead
	_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	followUp := &models.FollowUp{
		LeadID:        lead.ID,
		SalespersonID: salespersonID,
		DueAt:         input.DueAt,
		ScheduledAt:   time.Now(),
		Status:        models.FollowUpTaskScheduled,
		Notes:         input.Notes,
	}

	if err := s.followUpRepo.Create(ctx, followUp); err != nil {
		return nil, err
	}

	// Update lead status
	lead.NextFollowUpAt = &input.DueAt
	lead.FollowUpStatus = models.FollowUpStatusScheduled
	lead.Status = models.LeadStatusFollowUpScheduled
	_ = s.leadRepo.Update(ctx, lead)

	s.auditService.Log(ctx, actor.ID, actor.Name, "FOLLOWUP_SCHEDULED", "follow_up", followUp.ID.Hex(), nil, map[string]interface{}{
		"leadId": lead.ID.Hex(),
		"dueAt":  input.DueAt,
	}, "Follow-up scheduled", nil)

	s.populateDetails(ctx, followUp)
	return followUp, nil
}

func (s *FollowUpService) CompleteFollowUp(ctx context.Context, followUpID primitive.ObjectID, callAttemptID *primitive.ObjectID, actor *models.User) error {
	followUp, err := s.followUpRepo.FindByID(ctx, followUpID)
	if err != nil {
		return fmt.Errorf("follow-up not found: %w", err)
	}

	if err := s.followUpRepo.CompleteFollowUp(ctx, followUpID, callAttemptID); err != nil {
		return err
	}

	lead, err := s.leadRepo.FindByID(ctx, followUp.LeadID)
	if err == nil && lead != nil {
		lead.FollowUpStatus = models.FollowUpStatusCompleted
		_ = s.leadRepo.Update(ctx, lead)
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "FOLLOWUP_COMPLETED", "follow_up", followUpID.Hex(), nil, map[string]interface{}{
		"completedAt": time.Now(),
	}, "Follow-up marked completed", nil)

	return nil
}

// DetectOneCallFollowUps implements the core detection engine
func (s *FollowUpService) DetectOneCallFollowUps(ctx context.Context) (int, error) {
	settings, err := s.settingsRepo.GetSettings(ctx)
	if err != nil || settings == nil {
		settings = &models.SystemSettings{
			FollowUpThresholdHours: 24,
			EscalationHours:        48,
		}
	}

	now := time.Now()
	cutoffTime := now.Add(-time.Duration(settings.FollowUpThresholdHours) * time.Hour)

	eligibleLeads, err := s.leadRepo.FindEligibleForOneCallAlert(ctx, cutoffTime, now)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, lead := range eligibleLeads {
		// Strict check of all business rules
		if lead.CallAttemptCount != 1 || lead.DoNotCall {
			continue
		}
		if lead.Status == models.LeadStatusDoNotCall ||
			lead.Status == models.LeadStatusNotInterested ||
			lead.Status == models.LeadStatusConverted ||
			lead.Status == models.LeadStatusLost ||
			lead.Status == models.LeadStatusDisqualified {
			continue
		}

		// Calculate overdue hours
		var overdueHours float64 = 0
		if lead.NextFollowUpAt != nil {
			overdueHours = now.Sub(*lead.NextFollowUpAt).Hours()
		} else if lead.FirstCallAt != nil {
			overdueHours = now.Sub(lead.FirstCallAt.Add(time.Duration(settings.FollowUpThresholdHours) * time.Hour)).Hours()
		}

		// Transition status to FOLLOW_UP_REQUIRED
		lead.Status = models.LeadStatusFollowUpRequired
		lead.FollowUpStatus = models.FollowUpStatusOverdue
		_ = s.leadRepo.Update(ctx, lead)

		// Create in-app notification if salesperson assigned
		if lead.AssignedTo != nil {
			_ = s.notifService.Send(
				ctx,
				*lead.AssignedTo,
				&lead.ID,
				models.NotificationTypeOneCallFollowUp,
				"One-Call Follow-Up Required",
				fmt.Sprintf("Lead %s was called once and follow-up is overdue by %.1f hours.", lead.Name, overdueHours),
				map[string]interface{}{"leadId": lead.ID.Hex(), "overdueHours": overdueHours},
			)

			// Enqueue / send email alert with deduplication
			salesperson, _ := s.userRepo.FindByID(ctx, *lead.AssignedTo)
			if salesperson != nil {
				_ = s.emailService.SendFollowUpRequiredAlert(ctx, lead, salesperson, overdueHours)
			}
		}

		count++
	}

	log.Printf("One-Call Follow-Up detector processed: %d leads flagged as FOLLOW_UP_REQUIRED", count)
	return count, nil
}

func (s *FollowUpService) ListFollowUps(ctx context.Context, salespersonID *primitive.ObjectID, status models.FollowUpTaskStatus, timeRange string, page, pageSize int) ([]*models.FollowUp, int64, error) {
	followUps, total, err := s.followUpRepo.FindFollowUps(ctx, salespersonID, status, timeRange, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, f := range followUps {
		s.populateDetails(ctx, f)
	}
	return followUps, total, nil
}

func (s *FollowUpService) populateDetails(ctx context.Context, followUp *models.FollowUp) {
	lead, err := s.leadRepo.FindByID(ctx, followUp.LeadID)
	if err == nil && lead != nil {
		followUp.LeadName = lead.Name
		followUp.LeadPhone = lead.Phone
	}

	user, err := s.userRepo.FindByID(ctx, followUp.SalespersonID)
	if err == nil && user != nil {
		followUp.SalespersonUser = &models.UserSummary{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		}
	}
}
