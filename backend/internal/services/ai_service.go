package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"lead-followup-system/internal/ai"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AIService struct {
	aiRepo       *repositories.AIRepository
	callRepo     *repositories.CallRepository
	leadRepo     *repositories.LeadRepository
	transRepo    *repositories.TranscriptRepository
	userRepo     *repositories.UserRepository
	settingsRepo *repositories.SettingsRepository
	followUpRepo *repositories.FollowUpRepository
	auditService *AuditService
	notifService *NotificationService
	aiProvider   ai.AIProvider
}

func NewAIService(
	aiRepo *repositories.AIRepository,
	callRepo *repositories.CallRepository,
	leadRepo *repositories.LeadRepository,
	transRepo *repositories.TranscriptRepository,
	userRepo *repositories.UserRepository,
	settingsRepo *repositories.SettingsRepository,
	followUpRepo *repositories.FollowUpRepository,
	auditService *AuditService,
	notifService *NotificationService,
	aiProvider ai.AIProvider,
) *AIService {
	return &AIService{
		aiRepo:       aiRepo,
		callRepo:     callRepo,
		leadRepo:     leadRepo,
		transRepo:    transRepo,
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		followUpRepo: followUpRepo,
		auditService: auditService,
		notifService: notifService,
		aiProvider:   aiProvider,
	}
}

func (s *AIService) AnalyzeCallConversation(ctx context.Context, callAttemptID primitive.ObjectID, actor *models.User) (*models.AIAnalysis, error) {
	call, err := s.callRepo.FindByID(ctx, callAttemptID)
	if err != nil {
		return nil, fmt.Errorf("call attempt not found: %w", err)
	}

	transcript, err := s.transRepo.FindByCallAttemptID(ctx, callAttemptID)
	if err != nil || transcript == nil {
		return nil, errors.New("no transcript available for this call attempt. Privacy rule: never pretend AI analyzed a conversation without a transcript.")
	}

	lead, err := s.leadRepo.FindByID(ctx, call.LeadID)
	if err != nil {
		return nil, fmt.Errorf("associated lead not found: %w", err)
	}

	result, err := s.aiProvider.AnalyzeConversation(ctx, transcript)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	settings, _ := s.settingsRepo.GetSettings(ctx)
	if settings == nil {
		settings = &models.SystemSettings{
			AIConfidenceThreshold:  0.90,
			AutoClassifyEnabled:    true,
			HumanReviewRequiredDNC: true,
		}
	}

	// Determine review status based on confidence and safety rules
	reviewStatus := models.AIReviewStatusAutoApplied
	if !settings.AutoClassifyEnabled || result.Confidence < settings.AIConfidenceThreshold {
		reviewStatus = models.AIReviewStatusPendingReview
	} else if settings.HumanReviewRequiredDNC && (result.Intent == models.AIIntentDoNotCall || result.Intent == models.AIIntentNotInterested) {
		// Strict safety rule: If configured, critical DNC / NOT_INTERESTED requires human confirmation
		reviewStatus = models.AIReviewStatusPendingReview
	}

	analysis := &models.AIAnalysis{
		CallAttemptID:          call.ID,
		LeadID:                 lead.ID,
		TranscriptID:           transcript.ID,
		Intent:                 result.Intent,
		Confidence:             result.Confidence,
		Sentiment:              result.Sentiment,
		FollowUpRequired:       result.FollowUpRequired,
		DoNotCall:              result.DoNotCall,
		SuggestedFollowUpHours: result.SuggestedFollowUpHours,
		Summary:                result.Summary,
		Reason:                 result.Reason,
		Evidence:               result.Evidence,
		Status:                 reviewStatus,
	}

	if err := s.aiRepo.Create(ctx, analysis); err != nil {
		return nil, err
	}

	// Link to call attempt
	call.AIAnalysisID = &analysis.ID
	_ = s.callRepo.Update(ctx, call)

	// If Auto-Applied, synchronize outcome and lead status immediately
	if reviewStatus == models.AIReviewStatusAutoApplied {
		_ = s.applyIntentToLeadAndCall(ctx, analysis.Intent, call, lead, "AI Auto-Applied Classification")
	} else {
		// Queue alert for review
		_ = s.notifService.Send(
			ctx,
			call.SalespersonID,
			&lead.ID,
			models.NotificationTypeAIReviewNeeded,
			"AI Review Required",
			fmt.Sprintf("Call with %s classified as %s (%.0f%% confidence) requires your review.", lead.Name, result.Intent, result.Confidence*100),
			map[string]interface{}{"callId": call.ID.Hex(), "analysisId": analysis.ID.Hex()},
		)
	}

	actorID := primitive.NilObjectID
	actorName := "System"
	if actor != nil {
		actorID = actor.ID
		actorName = actor.Name
	}

	s.auditService.Log(ctx, actorID, actorName, "AI_ANALYSIS_CREATED", "ai_analysis", analysis.ID.Hex(), nil, map[string]interface{}{
		"intent":     analysis.Intent,
		"confidence": analysis.Confidence,
		"status":     analysis.Status,
	}, "AI conversation analysis completed", nil)

	return analysis, nil
}

func (s *AIService) ConfirmAnalysis(ctx context.Context, analysisID primitive.ObjectID, actor *models.User) (*models.AIAnalysis, error) {
	analysis, err := s.aiRepo.FindByID(ctx, analysisID)
	if err != nil {
		return nil, errors.New("analysis record not found")
	}

	now := time.Now()
	analysis.Status = models.AIReviewStatusConfirmed
	analysis.ReviewedBy = &actor.ID
	analysis.ReviewedAt = &now

	if err := s.aiRepo.Update(ctx, analysis); err != nil {
		return nil, err
	}

	call, _ := s.callRepo.FindByID(ctx, analysis.CallAttemptID)
	lead, _ := s.leadRepo.FindByID(ctx, analysis.LeadID)

	if call != nil && lead != nil {
		_ = s.applyIntentToLeadAndCall(ctx, analysis.Intent, call, lead, fmt.Sprintf("Confirmed by %s", actor.Name))
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "AI_ANALYSIS_CONFIRMED", "ai_analysis", analysis.ID.Hex(),
		map[string]interface{}{"status": models.AIReviewStatusPendingReview},
		map[string]interface{}{"status": models.AIReviewStatusConfirmed, "intent": analysis.Intent},
		"Human reviewer confirmed AI intent", nil)

	s.populateReviewer(ctx, analysis)
	return analysis, nil
}

func (s *AIService) OverrideAnalysis(ctx context.Context, analysisID primitive.ObjectID, newIntent models.AIIntent, reason string, actor *models.User) (*models.AIAnalysis, error) {
	if reason == "" {
		return nil, errors.New("a reason is required when overriding AI classification")
	}

	analysis, err := s.aiRepo.FindByID(ctx, analysisID)
	if err != nil {
		return nil, errors.New("analysis record not found")
	}

	now := time.Now()
	originalIntent := analysis.Intent
	analysis.OriginalIntent = originalIntent
	analysis.Intent = newIntent
	analysis.Status = models.AIReviewStatusOverridden
	analysis.OverrideReason = reason
	analysis.ReviewedBy = &actor.ID
	analysis.ReviewedAt = &now

	// Update flags based on new intent
	if newIntent == models.AIIntentDoNotCall {
		analysis.DoNotCall = true
		analysis.FollowUpRequired = false
	} else if newIntent == models.AIIntentNotInterested {
		analysis.DoNotCall = false
		analysis.FollowUpRequired = false
	} else {
		analysis.DoNotCall = false
		analysis.FollowUpRequired = true
	}

	if err := s.aiRepo.Update(ctx, analysis); err != nil {
		return nil, err
	}

	call, _ := s.callRepo.FindByID(ctx, analysis.CallAttemptID)
	lead, _ := s.leadRepo.FindByID(ctx, analysis.LeadID)

	if call != nil && lead != nil {
		_ = s.applyIntentToLeadAndCall(ctx, newIntent, call, lead, fmt.Sprintf("Overridden by %s: %s", actor.Name, reason))
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "AI_ANALYSIS_OVERRIDDEN", "ai_analysis", analysis.ID.Hex(),
		map[string]interface{}{"intent": originalIntent},
		map[string]interface{}{"intent": newIntent, "reason": reason},
		reason, nil)

	s.populateReviewer(ctx, analysis)
	return analysis, nil
}

func (s *AIService) GetAnalysisByCallID(ctx context.Context, callAttemptID primitive.ObjectID) (*models.AIAnalysis, error) {
	analysis, err := s.aiRepo.FindByCallAttemptID(ctx, callAttemptID)
	if err != nil {
		return nil, err
	}
	s.populateReviewer(ctx, analysis)
	return analysis, nil
}

func (s *AIService) GetPendingReviews(ctx context.Context, page, pageSize int) ([]*models.AIAnalysis, int64, error) {
	analyses, total, err := s.aiRepo.FindPendingReviews(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, a := range analyses {
		s.populateReviewer(ctx, a)
	}
	return analyses, total, nil
}

func (s *AIService) applyIntentToLeadAndCall(ctx context.Context, intent models.AIIntent, call *models.CallAttempt, lead *models.Lead, note string) error {
	callOutcome := models.CallOutcome(intent)
	call.CallOutcome = callOutcome
	_ = s.callRepo.Update(ctx, call)

	switch intent {
	case models.AIIntentNotInterested:
		lead.Status = models.LeadStatusNotInterested
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	case models.AIIntentDoNotCall:
		lead.Status = models.LeadStatusDoNotCall
		lead.DoNotCall = true
		lead.DoNotCallReason = fmt.Sprintf("AI Detected: %s", note)
		lead.DoNotCallDetectedBy = "AI Engine"
		now := time.Now()
		lead.DoNotCallAt = &now
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	case models.AIIntentInterested:
		lead.Status = models.LeadStatusInterested

	case models.AIIntentCallBackLater:
		lead.Status = models.LeadStatusFollowUpScheduled
		lead.FollowUpStatus = models.FollowUpStatusScheduled
		nextDue := time.Now().Add(24 * time.Hour)
		lead.NextFollowUpAt = &nextDue
		_ = s.followUpRepo.Create(ctx, &models.FollowUp{
			LeadID:        lead.ID,
			SalespersonID: call.SalespersonID,
			DueAt:         nextDue,
			ScheduledAt:   time.Now(),
			Status:        models.FollowUpTaskScheduled,
			Notes:         "Follow-up scheduled per AI conversation detection",
		})

	case models.AIIntentInfoRequested:
		lead.Status = models.LeadStatusInfoRequested

	case models.AIIntentUndecided:
		lead.Status = models.LeadStatusUndecided

	case models.AIIntentConverted:
		lead.Status = models.LeadStatusConverted
		lead.FollowUpStatus = models.FollowUpStatusCompleted
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	case models.AIIntentLost:
		lead.Status = models.LeadStatusLost
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)
	}

	return s.leadRepo.Update(ctx, lead)
}

func (s *AIService) populateReviewer(ctx context.Context, analysis *models.AIAnalysis) {
	if analysis.ReviewedBy != nil {
		user, err := s.userRepo.FindByID(ctx, *analysis.ReviewedBy)
		if err == nil && user != nil {
			analysis.ReviewedByUser = &models.UserSummary{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			}
		}
	}
}

type AutoProcessInput struct {
	TranscriptText  string `json:"transcriptText" binding:"required"`
	DurationSeconds int    `json:"durationSeconds"`
}

func (s *AIService) AutoProcessConversation(ctx context.Context, leadID primitive.ObjectID, input AutoProcessInput, actor *models.User) (*models.CallAttempt, *models.AIAnalysis, *models.Lead, error) {
	lead, err := s.leadRepo.FindByID(ctx, leadID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lead not found: %w", err)
	}

	callCount, _ := s.callRepo.CountByLeadID(ctx, lead.ID)
	attemptNumber := int(callCount) + 1
	now := time.Now()

	// 1. Create Call Attempt
	call := &models.CallAttempt{
		LeadID:              lead.ID,
		SalespersonID:       actor.ID,
		AttemptNumber:       attemptNumber,
		CallStartedAt:       now,
		DurationSeconds:     input.DurationSeconds,
		CallStatus:          models.CallStatusAnswered,
		CallOutcome:         models.CallOutcomeFollowUpRequired,
		RecordingConsent:    true,
		TranscriptAvailable: true,
		ManuallyReviewed:    true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if call.DurationSeconds == 0 {
		call.DurationSeconds = 45
	}
	if err := s.callRepo.Create(ctx, call); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create call record: %w", err)
	}

	// 2. Parse Speaker Segments from Transcript
	lines := strings.Split(input.TranscriptText, "\n")
	var segments []models.SpeakerSegment
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(l), "lead:") || strings.HasPrefix(strings.ToLower(l), "student:") || strings.HasPrefix(strings.ToLower(l), "customer:") {
			parts := strings.SplitN(l, ":", 2)
			text := strings.TrimSpace(parts[1])
			segments = append(segments, models.SpeakerSegment{Speaker: "lead", Text: text})
		} else if strings.HasPrefix(strings.ToLower(l), "salesperson:") || strings.HasPrefix(strings.ToLower(l), "bda:") || strings.HasPrefix(strings.ToLower(l), "agent:") {
			parts := strings.SplitN(l, ":", 2)
			text := strings.TrimSpace(parts[1])
			segments = append(segments, models.SpeakerSegment{Speaker: "salesperson", Text: text})
		} else {
			segments = append(segments, models.SpeakerSegment{Speaker: "lead", Text: l})
		}
	}

	// 3. Save Transcript
	trans := &models.Transcript{
		CallAttemptID:   call.ID,
		Language:        "en",
		Transcript:      input.TranscriptText,
		SpeakerSegments: segments,
		DurationSeconds: call.DurationSeconds,
		Provider:        "telephony-stt",
		CreatedAt:       now,
	}
	_ = s.transRepo.Create(ctx, trans)
	call.TranscriptID = &trans.ID

	// 4. Run AI Analysis
	result, err := s.aiProvider.AnalyzeConversation(ctx, trans)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// 5. Save AI Analysis
	analysis := &models.AIAnalysis{
		CallAttemptID:    call.ID,
		TranscriptID:     trans.ID,
		Intent:           result.Intent,
		Confidence:       result.Confidence,
		Sentiment:        result.Sentiment,
		FollowUpRequired: result.FollowUpRequired,
		DoNotCall:        result.DoNotCall,
		Summary:          result.Summary,
		Reason:           result.Reason,
		Evidence:         result.Evidence,
		ReviewedBy:       &actor.ID,
		ReviewedAt:       &now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	_ = s.aiRepo.Create(ctx, analysis)
	call.AIAnalysisID = &analysis.ID

	// 6. Automatically Mutate Lead Status (AI Decisions Only - Zero BDA manual entry)
	evidenceQuote := ""
	if len(result.Evidence) > 0 {
		evidenceQuote = result.Evidence[0].Text
	} else if result.Summary != "" {
		evidenceQuote = result.Summary
	}

	lead.CallAttemptCount = attemptNumber
	lead.LastCallAt = &now
	if lead.FirstCallAt == nil {
		lead.FirstCallAt = &now
	}

	// Apply AI decision
	switch result.Intent {
	case models.AIIntentDoNotCall:
		lead.Status = models.LeadStatusDoNotCall
		lead.DoNotCall = true
		lead.DoNotCallReason = fmt.Sprintf("AI Detected from call: %s", evidenceQuote)
		lead.DoNotCallDetectedBy = "AI Transcript Engine"
		lead.DoNotCallAt = &now
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		lead.Notes = fmt.Sprintf("AI: Student requested Do Not Call: \"%s\"", evidenceQuote)
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	case models.AIIntentNotInterested:
		lead.Status = models.LeadStatusNotInterested
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		lead.Notes = fmt.Sprintf("AI: Student Not Interested: \"%s\"", evidenceQuote)
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, lead.ID)

	case models.AIIntentCallBackLater:
		lead.Status = models.LeadStatusFollowUpScheduled
		lead.FollowUpStatus = models.FollowUpStatusScheduled
		followUpHours := result.SuggestedFollowUpHours
		if followUpHours <= 0 {
			followUpHours = 24
		}
		dueAt := now.Add(time.Duration(followUpHours) * time.Hour)
		lead.NextFollowUpAt = &dueAt
		lead.Notes = fmt.Sprintf("AI Reminder: Student said \"%s\"", evidenceQuote)

		// Create automated scheduled reminder
		_ = s.followUpRepo.Create(ctx, &models.FollowUp{
			LeadID:        lead.ID,
			SalespersonID: actor.ID,
			DueAt:         dueAt,
			ScheduledAt:   now,
			Status:        models.FollowUpTaskScheduled,
			Notes:         fmt.Sprintf("AI Callback Reminder: \"%s\"", evidenceQuote),
		})

		// Send BDA Reminder Notification
		_ = s.notifService.Send(
			ctx,
			actor.ID,
			&lead.ID,
			models.NotificationTypeFollowUpOverdue,
			"AI Follow-Up Reminder Scheduled",
			fmt.Sprintf("Student %s asked to call later: \"%s\"", lead.Name, evidenceQuote),
			map[string]interface{}{"leadId": lead.ID.Hex(), "dueAt": dueAt},
		)

	case models.AIIntentInterested:
		lead.Status = models.LeadStatusInterested
		lead.Notes = fmt.Sprintf("AI: Student Interested: \"%s\"", evidenceQuote)
		_ = s.notifService.Send(
			ctx,
			actor.ID,
			&lead.ID,
			models.NotificationTypeLeadAssigned,
			"Positive Lead Response",
			fmt.Sprintf("Student %s is interested! Follow up with admissions syllabus.", lead.Name),
			map[string]interface{}{"leadId": lead.ID.Hex()},
		)

	case models.AIIntentFollowUpRequired:
		if attemptNumber == 1 {
			lead.Status = models.LeadStatusCalledOnce
		} else {
			lead.Status = models.LeadStatusFollowUpRequired
		}
		lead.FollowUpStatus = models.FollowUpStatusPending
		lead.Notes = fmt.Sprintf("AI: Call disconnected/broken voice: \"%s\"", evidenceQuote)

	default:
		// Default to follow up required
		if attemptNumber == 1 {
			lead.Status = models.LeadStatusCalledOnce
		} else {
			lead.Status = models.LeadStatusFollowUpRequired
		}
		lead.FollowUpStatus = models.FollowUpStatusPending
		lead.Notes = fmt.Sprintf("AI Transcript note: \"%s\"", evidenceQuote)
	}

	_ = s.leadRepo.Update(ctx, lead)
	call.CallOutcome = models.CallOutcome(result.Intent)
	_ = s.callRepo.Update(ctx, call)

	s.auditService.Log(ctx, actor.ID, actor.Name, "AI_CALL_ANALYZED", "lead", lead.ID.Hex(), nil, map[string]interface{}{
		"intent":   result.Intent,
		"quote":    evidenceQuote,
		"status":   lead.Status,
	}, fmt.Sprintf("AI processed call: %s", result.Intent), nil)

	return call, analysis, lead, nil
}

