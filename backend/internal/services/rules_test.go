package services

import (
	"context"
	"testing"
	"time"

	"lead-followup-system/internal/ai"
	"lead-followup-system/internal/config"
	"lead-followup-system/internal/database"
	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupTestDB(t *testing.T) (*database.MongoDB, *config.Config) {
	cfg := config.LoadConfig()
	cfg.MongoDB = "lead_followup_test_db"
	mongoDB, err := database.ConnectMongo(cfg)
	if err != nil {
		t.Skipf("Skipping test, MongoDB not available: %v", err)
	}
	return mongoDB, cfg
}

func TestBusinessRulesSuite(t *testing.T) {
	mongoDB, cfg := setupTestDB(t)
	ctx := context.Background()
	defer func() {
		_ = mongoDB.Database.Drop(ctx)
	}()

	leadRepo := repositories.NewLeadRepository(mongoDB.Database)
	callRepo := repositories.NewCallRepository(mongoDB.Database)
	followUpRepo := repositories.NewFollowUpRepository(mongoDB.Database)
	aiRepo := repositories.NewAIRepository(mongoDB.Database)
	transRepo := repositories.NewTranscriptRepository(mongoDB.Database)
	userRepo := repositories.NewUserRepository(mongoDB.Database)
	notifRepo := repositories.NewNotificationRepository(mongoDB.Database)
	auditRepo := repositories.NewAuditRepository(mongoDB.Database)
	settingsRepo := repositories.NewSettingsRepository(mongoDB.Database)

	auditService := NewAuditService(auditRepo)
	notifService := NewNotificationService(notifRepo)
	emailService := NewEmailService(cfg, nil) // nil redis for unit test
	followUpService := NewFollowUpService(followUpRepo, leadRepo, userRepo, settingsRepo, notifService, emailService, auditService)
	aiProvider := ai.NewRulesEngineProvider()
	aiService := NewAIService(aiRepo, callRepo, leadRepo, transRepo, userRepo, settingsRepo, followUpRepo, auditService, notifService, aiProvider)

	now := time.Now()
	salesRepID := primitive.NewObjectID()
	actor := &models.User{
		ID:   salesRepID,
		Name: "Test Rep",
		Role: models.RoleSales,
	}

	// -------------------------------------------------------------
	// TEST 1: One call, No Answer, 24 hours passed -> FOLLOW_UP_REQUIRED
	// -------------------------------------------------------------
	t.Run("Test 1: One call, No Answer, 24h passed -> FOLLOW_UP_REQUIRED", func(t *testing.T) {
		firstCall := now.Add(-26 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 1",
			Phone:            "+91 99999 00001",
			Email:            "test1@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCall,
			LastCallAt:       &firstCall,
			Status:           models.LeadStatusCalledOnce,
			FollowUpStatus:   models.FollowUpStatusPending,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		count, err := followUpService.DetectOneCallFollowUps(ctx)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if count < 1 {
			t.Errorf("expected at least 1 lead flagged, got %d", count)
		}

		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status != models.LeadStatusFollowUpRequired {
			t.Errorf("expected status FOLLOW_UP_REQUIRED, got: %s", updated.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 2: One call, Call Back Later, follow-up not due -> FOLLOW_UP_SCHEDULED
	// -------------------------------------------------------------
	t.Run("Test 2: One call, Call Back Later, future due date -> FOLLOW_UP_SCHEDULED", func(t *testing.T) {
		futureDue := now.Add(24 * time.Hour)
		firstCall := now.Add(-2 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 2",
			Phone:            "+91 99999 00002",
			Email:            "test2@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCall,
			LastCallAt:       &firstCall,
			NextFollowUpAt:   &futureDue,
			Status:           models.LeadStatusFollowUpScheduled,
			FollowUpStatus:   models.FollowUpStatusScheduled,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		// Run detector - should NOT flag as FOLLOW_UP_REQUIRED
		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status != models.LeadStatusFollowUpScheduled {
			t.Errorf("expected status to remain FOLLOW_UP_SCHEDULED, got: %s", updated.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 3: One call, Call Back Later, follow-up overdue -> FOLLOW_UP_REQUIRED
	// -------------------------------------------------------------
	t.Run("Test 3: One call, Call Back Later, overdue due date -> FOLLOW_UP_REQUIRED", func(t *testing.T) {
		pastDue := now.Add(-2 * time.Hour)
		firstCall := now.Add(-10 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 3",
			Phone:            "+91 99999 00003",
			Email:            "test3@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCall,
			LastCallAt:       &firstCall,
			NextFollowUpAt:   &pastDue,
			Status:           models.LeadStatusFollowUpScheduled,
			FollowUpStatus:   models.FollowUpStatusScheduled,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status != models.LeadStatusFollowUpRequired {
			t.Errorf("expected status FOLLOW_UP_REQUIRED, got: %s", updated.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 4: One call, AI = NOT_INTERESTED (96% confidence) -> NOT_INTERESTED, No follow-up
	// -------------------------------------------------------------
	t.Run("Test 4: AI NOT_INTERESTED -> NOT_INTERESTED, No follow-up", func(t *testing.T) {
		lead := &models.Lead{
			Name:             "Test Lead 4",
			Phone:            "+91 99999 00004",
			Email:            "test4@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			Status:           models.LeadStatusCalledOnce,
		}
		_ = leadRepo.Create(ctx, lead)

		call := &models.CallAttempt{
			LeadID:        lead.ID,
			SalespersonID: salesRepID,
			AttemptNumber: 1,
			CallStartedAt: now,
			CallStatus:    models.CallStatusAnswered,
			CallOutcome:   models.CallOutcomeInterested,
		}
		_ = callRepo.Create(ctx, call)

		trans := &models.Transcript{
			LeadID:        lead.ID,
			CallAttemptID: call.ID,
			Language:      "en",
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Hello, calling about the course."},
				{Speaker: "lead", Text: "I am not interested in this course. I don't want this."},
			},
		}
		_ = transRepo.Create(ctx, trans)

		// Set humanReviewRequiredDNC = false for this test to verify auto-application
		settings, _ := settingsRepo.GetSettings(ctx)
		settings.HumanReviewRequiredDNC = false
		_ = settingsRepo.UpdateSettings(ctx, settings)

		analysis, err := aiService.AnalyzeCallConversation(ctx, call.ID, actor)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if analysis.Intent != models.AIIntentNotInterested {
			t.Errorf("expected NOT_INTERESTED intent, got: %s", analysis.Intent)
		}

		updatedLead, _ := leadRepo.FindByID(ctx, lead.ID)
		if updatedLead.Status != models.LeadStatusNotInterested {
			t.Errorf("expected lead status NOT_INTERESTED, got: %s", updatedLead.Status)
		}

		// Ensure detector never alerts on NOT_INTERESTED
		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		recheckedLead, _ := leadRepo.FindByID(ctx, lead.ID)
		if recheckedLead.Status == models.LeadStatusFollowUpRequired {
			t.Errorf("NOT_INTERESTED lead was improperly flagged as FOLLOW_UP_REQUIRED")
		}
	})

	// -------------------------------------------------------------
	// TEST 5: One call, AI = DO_NOT_CALL -> DO_NOT_CALL, No follow-up
	// -------------------------------------------------------------
	t.Run("Test 5: AI DO_NOT_CALL -> DO_NOT_CALL, No follow-up", func(t *testing.T) {
		lead := &models.Lead{
			Name:             "Test Lead 5",
			Phone:            "+91 99999 00005",
			Email:            "test5@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			Status:           models.LeadStatusCalledOnce,
		}
		_ = leadRepo.Create(ctx, lead)

		call := &models.CallAttempt{
			LeadID:        lead.ID,
			SalespersonID: salesRepID,
			AttemptNumber: 1,
			CallStartedAt: now,
			CallStatus:    models.CallStatusAnswered,
			CallOutcome:   models.CallOutcomeInterested,
		}
		_ = callRepo.Create(ctx, call)

		trans := &models.Transcript{
			LeadID:        lead.ID,
			CallAttemptID: call.ID,
			Language:      "en",
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Hello, calling about the course."},
				{Speaker: "lead", Text: "Stop calling me! Remove my number immediately."},
			},
		}
		_ = transRepo.Create(ctx, trans)

		analysis, err := aiService.AnalyzeCallConversation(ctx, call.ID, actor)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if analysis.Intent != models.AIIntentDoNotCall {
			t.Errorf("expected DO_NOT_CALL intent, got: %s", analysis.Intent)
		}

		updatedLead, _ := leadRepo.FindByID(ctx, lead.ID)
		if !updatedLead.DoNotCall {
			t.Errorf("expected lead.DoNotCall to be true")
		}

		// Ensure detector never alerts on DO_NOT_CALL
		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		recheckedLead, _ := leadRepo.FindByID(ctx, lead.ID)
		if recheckedLead.Status == models.LeadStatusFollowUpRequired {
			t.Errorf("DO_NOT_CALL lead was improperly flagged as FOLLOW_UP_REQUIRED")
		}
	})

	// -------------------------------------------------------------
	// TEST 6: Ambiguous Call -> Sent to Human Review Required (PENDING_REVIEW)
	// -------------------------------------------------------------
	t.Run("Test 6: Ambiguous Call -> Sent to PENDING_REVIEW", func(t *testing.T) {
		lead := &models.Lead{
			Name:             "Test Lead 6",
			Phone:            "+91 99999 00006",
			Email:            "test6@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			Status:           models.LeadStatusCalledOnce,
		}
		_ = leadRepo.Create(ctx, lead)

		call := &models.CallAttempt{
			LeadID:        lead.ID,
			SalespersonID: salesRepID,
			AttemptNumber: 1,
			CallStartedAt: now,
			CallStatus:    models.CallStatusAnswered,
		}
		_ = callRepo.Create(ctx, call)

		trans := &models.Transcript{
			LeadID:        lead.ID,
			CallAttemptID: call.ID,
			Language:      "en",
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Do you have any questions?"},
				{Speaker: "lead", Text: "Maybe, let me think."},
			},
		}
		_ = transRepo.Create(ctx, trans)

		// Reset settings to require 0.90 confidence
		settings, _ := settingsRepo.GetSettings(ctx)
		settings.AIConfidenceThreshold = 0.90
		_ = settingsRepo.UpdateSettings(ctx, settings)

		analysis, err := aiService.AnalyzeCallConversation(ctx, call.ID, actor)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if analysis.Status != models.AIReviewStatusPendingReview {
			t.Errorf("expected PENDING_REVIEW status, got: %s", analysis.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 7: Attempt 1 + Attempt 2 -> Not a one-call lead
	// -------------------------------------------------------------
	t.Run("Test 7: Attempt 1 + Attempt 2 -> Excluded from one-call alerts", func(t *testing.T) {
		firstCall := now.Add(-30 * time.Hour)
		secondCall := now.Add(-5 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 7",
			Phone:            "+91 99999 00007",
			Email:            "test7@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 2, // 2 calls!
			FirstCallAt:      &firstCall,
			LastCallAt:       &secondCall,
			Status:           models.LeadStatusContacted,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status == models.LeadStatusFollowUpRequired {
			t.Errorf("multi-call lead was improperly flagged as FOLLOW_UP_REQUIRED")
		}
	})

	// -------------------------------------------------------------
	// TEST 8: Lead marked LOST -> Alert suppressed
	// -------------------------------------------------------------
	t.Run("Test 8: Lead marked LOST -> Alert suppressed", func(t *testing.T) {
		firstCall := now.Add(-30 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 8",
			Phone:            "+91 99999 00008",
			Email:            "test8@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCall,
			LastCallAt:       &firstCall,
			Status:           models.LeadStatusLost,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status != models.LeadStatusLost {
			t.Errorf("expected status to remain LOST, got: %s", updated.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 9: Lead marked CONVERTED -> Alert suppressed
	// -------------------------------------------------------------
	t.Run("Test 9: Lead marked CONVERTED -> Alert suppressed", func(t *testing.T) {
		firstCall := now.Add(-30 * time.Hour)
		lead := &models.Lead{
			Name:             "Test Lead 9",
			Phone:            "+91 99999 00009",
			Email:            "test9@example.com",
			Source:           "Website",
			AssignedTo:       &salesRepID,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCall,
			LastCallAt:       &firstCall,
			Status:           models.LeadStatusConverted,
			DoNotCall:        false,
		}
		_ = leadRepo.Create(ctx, lead)

		_, _ = followUpService.DetectOneCallFollowUps(ctx)
		updated, _ := leadRepo.FindByID(ctx, lead.ID)
		if updated.Status != models.LeadStatusConverted {
			t.Errorf("expected status to remain CONVERTED, got: %s", updated.Status)
		}
	})

	// -------------------------------------------------------------
	// TEST 10: Lead explicitly requests no calls -> DO_NOT_CALL
	// -------------------------------------------------------------
	t.Run("Test 10: Explicit DO_NOT_CALL command recognized", func(t *testing.T) {
		transcript := &models.Transcript{
			Language: "en",
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Hello, regarding your inquiry..."},
				{Speaker: "lead", Text: "Do not call me again! Take me off your list."},
			},
		}

		result, err := aiProvider.AnalyzeConversation(ctx, transcript)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if result.Intent != models.AIIntentDoNotCall {
			t.Errorf("expected DO_NOT_CALL, got: %s", result.Intent)
		}
		if !result.DoNotCall {
			t.Errorf("expected DoNotCall == true")
		}
	})
}
