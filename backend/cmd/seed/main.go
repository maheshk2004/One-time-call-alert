package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"lead-followup-system/internal/config"
	"lead-followup-system/internal/database"
	"lead-followup-system/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	log.Println("Starting database seeding for Lead Follow-Up Intelligence System...")

	cfg := config.LoadConfig()
	mongoDB, err := database.ConnectMongo(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	ctx := context.Background()

	// Clear existing collections for a clean, deterministic seed
	collections := []string{"users", "leads", "call_attempts", "transcripts", "ai_analyses", "follow_ups", "notifications", "audit_logs", "system_settings"}
	for _, collName := range collections {
		_ = mongoDB.Database.Collection(collName).Drop(ctx)
	}
	log.Println("Dropped previous collections.")

	// Recreate indexes
	_ = mongoDB.CreateIndexes(ctx)

	// Hash password "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	passStr := string(hashedPassword)

	now := time.Now()

	// 1. Seed Users
	usersColl := mongoDB.Database.Collection("users")

	adminUser := models.User{
		ID:           primitive.NewObjectID(),
		Name:         "Admin User",
		Email:        "admin@leadfollowup.com",
		PasswordHash: passStr,
		Role:         models.RoleAdmin,
		Phone:        "+91 98765 00001",
		IsActive:     true,
		CreatedAt:    now.Add(-30 * 24 * time.Hour),
		UpdatedAt:    now,
	}

	managerUser := models.User{
		ID:           primitive.NewObjectID(),
		Name:         "Rajesh Manager",
		Email:        "manager@leadfollowup.com",
		PasswordHash: passStr,
		Role:         models.RoleManager,
		TeamID:       "Enterprise Sales Team A",
		Phone:        "+91 98765 00002",
		IsActive:     true,
		CreatedAt:    now.Add(-30 * 24 * time.Hour),
		UpdatedAt:    now,
	}

	salesUser1 := models.User{
		ID:           primitive.NewObjectID(),
		Name:         "Vikram Sharma",
		Email:        "vikram@leadfollowup.com",
		PasswordHash: passStr,
		Role:         models.RoleSales,
		TeamID:       "Enterprise Sales Team A",
		Phone:        "+91 98765 11001",
		IsActive:     true,
		CreatedAt:    now.Add(-20 * 24 * time.Hour),
		UpdatedAt:    now,
	}

	salesUser2 := models.User{
		ID:           primitive.NewObjectID(),
		Name:         "Priya Nair",
		Email:        "priya@leadfollowup.com",
		PasswordHash: passStr,
		Role:         models.RoleSales,
		TeamID:       "Enterprise Sales Team A",
		Phone:        "+91 98765 11002",
		IsActive:     true,
		CreatedAt:    now.Add(-20 * 24 * time.Hour),
		UpdatedAt:    now,
	}

	salesUser3 := models.User{
		ID:           primitive.NewObjectID(),
		Name:         "Arun Patel",
		Email:        "arun@leadfollowup.com",
		PasswordHash: passStr,
		Role:         models.RoleSales,
		TeamID:       "Enterprise Sales Team A",
		Phone:        "+91 98765 11003",
		IsActive:     true,
		CreatedAt:    now.Add(-20 * 24 * time.Hour),
		UpdatedAt:    now,
	}

	_, _ = usersColl.InsertMany(ctx, []interface{}{adminUser, managerUser, salesUser1, salesUser2, salesUser3})
	log.Println("Seeded Users: Admin, Manager, and 3 Salespersons.")

	// 2. Seed System Settings
	settingsColl := mongoDB.Database.Collection("system_settings")
	settings := models.SystemSettings{
		ID:                     primitive.NewObjectID(),
		FollowUpThresholdHours: 24,
		AIConfidenceThreshold:  0.90,
		AutoClassifyEnabled:    true,
		HumanReviewRequiredDNC: true,
		EscalationHours:        48,
		SupportedLanguages:     []string{"en", "hi", "ta", "te", "ml", "kn"},
		DuplicateCheckFields:   []string{"phone", "email"},
		UpdatedAt:              now,
	}
	_, _ = settingsColl.InsertOne(ctx, settings)

	// 3. Seed Realistic Leads, Calls, Transcripts, AI Analyses, Follow-ups
	leadsColl := mongoDB.Database.Collection("leads")
	callsColl := mongoDB.Database.Collection("call_attempts")
	transcriptsColl := mongoDB.Database.Collection("transcripts")
	aiColl := mongoDB.Database.Collection("ai_analyses")
	followUpsColl := mongoDB.Database.Collection("follow_ups")
	notifsColl := mongoDB.Database.Collection("notifications")

	salesReps := []primitive.ObjectID{salesUser1.ID, salesUser2.ID, salesUser3.ID}

	courses := []string{
		"Full-Stack Web Development", "AI & Data Science Masterclass", "Cloud DevOps Engineering",
		"UI/UX Design Bootcamp", "Cybersecurity Specialist", "Product Management Executive",
	}

	sources := []string{
		"Facebook Ads", "Google Search Ads", "LinkedIn Ads", "Website Form", "WhatsApp Inbound", "Referral", "Meta Lead Ads",
	}

	// Helper to generate dates
	hoursAgo := func(h float64) time.Time {
		return now.Add(-time.Duration(h * float64(time.Hour)))
	}
	hoursFromNow := func(h float64) time.Time {
		return now.Add(time.Duration(h * float64(time.Hour)))
	}

	var leads []interface{}
	var calls []interface{}
	var transcripts []interface{}
	var aiAnalyses []interface{}
	var followUps []interface{}
	var notifs []interface{}

	// --- Category 1: Single Call, No Answer, Overdue (> 24 hours) -> FOLLOW_UP_REQUIRED (One-Call Alert!) ---
	for i := 1; i <= 10; i++ {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		firstCallTime := hoursAgo(float64(26 + i*4))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("Lead OneCall %d", i),
			Email:            fmt.Sprintf("onecall%d@example.com", i),
			Phone:            fmt.Sprintf("+91 91000 %05d", 100+i),
			Source:           sources[i%len(sources)],
			Campaign:         "Summer Accelerator 2026",
			Course:           courses[i%len(courses)],
			Location:         "Bengaluru",
			AssignedTo:       &repID,
			Status:           models.LeadStatusFollowUpRequired, // Triggered!
			Priority:         models.PriorityHigh,
			CallAttemptCount: 1,
			FirstCallAt:      &firstCallTime,
			LastCallAt:       &firstCallTime,
			FollowUpStatus:   models.FollowUpStatusOverdue,
			DoNotCall:        false,
			Notes:            "Attempted initial connect, phone was ringing with no answer.",
			CreatedAt:        firstCallTime.Add(-2 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       firstCallTime,
			DurationSeconds:     25,
			CallStatus:          models.CallStatusNoAnswer,
			CallOutcome:         models.CallOutcomeFollowUpRequired,
			RecordingConsent:    true,
			RecordingAvailable:  false,
			TranscriptAvailable: false,
			ManuallyReviewed:    true,
			Notes:               "No answer after 5 rings.",
			CreatedAt:           firstCallTime,
			UpdatedAt:           firstCallTime,
		}
		calls = append(calls, call)

		// Create in-app notification
		notifs = append(notifs, models.Notification{
			ID:        primitive.NewObjectID(),
			UserID:    repID,
			LeadID:    &leadID,
			Type:      models.NotificationTypeOneCallFollowUp,
			Title:     "One-Call Follow-Up Required",
			Message:   fmt.Sprintf("Lead %s was called once and follow-up is overdue by %.1f hours.", lead.Name, now.Sub(firstCallTime.Add(24*time.Hour)).Hours()),
			Read:      false,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		})
	}

	// --- Category 2: Single Call, "Not Interested" with transcript & AI analysis -> NOT_INTERESTED (No follow-up!) ---
	for i := 1; i <= 8; i++ {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		transID := primitive.NewObjectID()
		aiID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		callTime := hoursAgo(float64(15 + i*3))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("Declined Lead %d", i),
			Email:            fmt.Sprintf("declined%d@example.com", i),
			Phone:            fmt.Sprintf("+91 92000 %05d", 200+i),
			Source:           sources[(i+1)%len(sources)],
			Campaign:         "Skill Upgrade Q3",
			Course:           courses[(i+1)%len(courses)],
			Location:         "Mumbai",
			AssignedTo:       &repID,
			Status:           models.LeadStatusNotInterested,
			Priority:         models.PriorityLow,
			CallAttemptCount: 1,
			FirstCallAt:      &callTime,
			LastCallAt:       &callTime,
			FollowUpStatus:   models.FollowUpStatusNone,
			DoNotCall:        false,
			Notes:            "Lead stated they are already employed and have no interest in the program.",
			CreatedAt:        callTime.Add(-4 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       callTime,
			DurationSeconds:     85,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeNotInterested,
			RecordingConsent:    true,
			RecordingAvailable:  true,
			RecordingURL:        "https://storage.googleapis.com/lead-recordings/call_declined_" + fmt.Sprint(i) + ".mp3",
			TranscriptAvailable: true,
			TranscriptID:        &transID,
			AIAnalysisID:        &aiID,
			ManuallyReviewed:    true,
			Notes:               "Customer explicitly rejected the course offering.",
			CreatedAt:           callTime,
			UpdatedAt:           callTime,
		}
		calls = append(calls, call)

		trans := models.Transcript{
			ID:            transID,
			LeadID:        leadID,
			CallAttemptID: callID,
			Language:      "en",
			Transcript:    "Salesperson: Hello, am I speaking with the applicant? Lead: Yes, but I'm not interested in this course. Please do not follow up. Salesperson: Understood, thank you for your time.",
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Hello, am I speaking with the applicant?", TimestampMs: 1000},
				{Speaker: "lead", Text: "Yes, but I'm not interested in this course. Please do not follow up.", TimestampMs: 4000},
				{Speaker: "salesperson", Text: "Understood, thank you for your time.", TimestampMs: 9000},
			},
			DurationSeconds: 85,
			Provider:        "whisper-v3",
			CreatedAt:       callTime,
		}
		transcripts = append(transcripts, trans)

		aiAnalyses = append(aiAnalyses, models.AIAnalysis{
			ID:               aiID,
			CallAttemptID:    callID,
			LeadID:           leadID,
			TranscriptID:     transID,
			Intent:           models.AIIntentNotInterested,
			Confidence:       0.96,
			Sentiment:        models.SentimentNegative,
			FollowUpRequired: false,
			DoNotCall:        false,
			Summary:          "Lead explicitly stated they are not interested in the course offering.",
			Reason:           "Clear statement of disinterest from lead.",
			Evidence: []models.EvidenceItem{
				{Speaker: "lead", Text: "Yes, but I'm not interested in this course. Please do not follow up."},
			},
			Status:    models.AIReviewStatusAutoApplied,
			CreatedAt: callTime,
			UpdatedAt: callTime,
		})
	}

	// --- Category 3: Single Call, Explicit "DO NOT CALL" (Multilingual: EN, HI, TA) -> DO_NOT_CALL (Suppress All!) ---
	languages := []struct {
		lang string
		lead string
	}{
		{"en", "Stop calling me, remove my number from your database immediately!"},
		{"hi", "मुझे दोबारा कॉल मत करना, मेरा नंबर डिलीट कर दो!"},
		{"ta", "தயவுசெய்து எனக்கு மீண்டும் அழைக்க வேண்டாம், என் எண்ணை நீக்குங்கள்!"},
		{"en", "Never call this phone number again. I will report for harassment."},
		{"hi", "कॉल मत करो, परेशान करना बंद करो!"},
		{"ta", "மீண்டும் கால் பண்ணாதீங்க!"},
	}

	for i, item := range languages {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		transID := primitive.NewObjectID()
		aiID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		callTime := hoursAgo(float64(10 + i*5))

		dncAt := callTime
		lead := models.Lead{
			ID:                  leadID,
			Name:                fmt.Sprintf("DNC Contact %d", i+1),
			Email:               fmt.Sprintf("dnc%d@example.com", i+1),
			Phone:               fmt.Sprintf("+91 93000 %05d", 300+i+1),
			Source:              sources[i%len(sources)],
			Campaign:            "Enterprise Outreach",
			Course:              courses[i%len(courses)],
			Location:            "Chennai / Delhi",
			AssignedTo:          &repID,
			Status:              models.LeadStatusDoNotCall,
			Priority:            models.PriorityLow,
			CallAttemptCount:    1,
			FirstCallAt:         &callTime,
			LastCallAt:          &callTime,
			FollowUpStatus:      models.FollowUpStatusNone,
			DoNotCall:           true,
			DoNotCallReason:     "Explicit Do-Not-Call demand by lead during conversation",
			DoNotCallDetectedBy: "AI Engine",
			DoNotCallAt:         &dncAt,
			Notes:               "Lead demanded complete deletion of number from calling list.",
			CreatedAt:           callTime.Add(-5 * time.Hour),
			UpdatedAt:           now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       callTime,
			DurationSeconds:     45,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeDoNotCall,
			RecordingConsent:    true,
			RecordingAvailable:  true,
			RecordingURL:        "https://storage.googleapis.com/lead-recordings/dnc_call_" + fmt.Sprint(i+1) + ".mp3",
			TranscriptAvailable: true,
			TranscriptID:        &transID,
			AIAnalysisID:        &aiID,
			ManuallyReviewed:    true,
			Notes:               "Lead explicitly insisted on Do Not Call.",
			CreatedAt:           callTime,
			UpdatedAt:           callTime,
		}
		calls = append(calls, call)

		trans := models.Transcript{
			ID:            transID,
			LeadID:        leadID,
			CallAttemptID: callID,
			Language:      item.lang,
			Transcript:    fmt.Sprintf("Salesperson: Hello! Calling regarding your inquiry. Lead: %s Salesperson: We apologize, removing you now.", item.lead),
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Hello! Calling regarding your inquiry.", TimestampMs: 1000},
				{Speaker: "lead", Text: item.lead, TimestampMs: 3500},
				{Speaker: "salesperson", Text: "We apologize, removing you now.", TimestampMs: 7000},
			},
			DurationSeconds: 45,
			Provider:        "whisper-v3",
			CreatedAt:       callTime,
		}
		transcripts = append(transcripts, trans)

		aiAnalyses = append(aiAnalyses, models.AIAnalysis{
			ID:               aiID,
			CallAttemptID:    callID,
			LeadID:           leadID,
			TranscriptID:     transID,
			Intent:           models.AIIntentDoNotCall,
			Confidence:       0.98,
			Sentiment:        models.SentimentNegative,
			FollowUpRequired: false,
			DoNotCall:        true,
			Summary:          "Lead explicitly commanded to be placed on Do Not Call list.",
			Reason:           "Explicit do-not-call request detected.",
			Evidence: []models.EvidenceItem{
				{Speaker: "lead", Text: item.lead},
			},
			Status:    models.AIReviewStatusAutoApplied,
			CreatedAt: callTime,
			UpdatedAt: callTime,
		})
	}

	// --- Category 4: Ambiguous / Low Confidence Calls -> PENDING_REVIEW (AI Review Queue!) ---
	ambiguousDialogues := []string{
		"I am not entirely sure right now, I need some time to see if my work schedule allows this.",
		"Maybe next month, please hold on and I will check my finances.",
		"Send whatever you have, I might look into it when I'm free.",
		"I don't think this is suitable right now, but you can check back later.",
		"I will have to ask my family first before saying anything.",
	}

	for i, dialogue := range ambiguousDialogues {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		transID := primitive.NewObjectID()
		aiID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		callTime := hoursAgo(float64(4 + i*2))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("Review Case %d", i+1),
			Email:            fmt.Sprintf("review%d@example.com", i+1),
			Phone:            fmt.Sprintf("+91 94000 %05d", 400+i+1),
			Source:           sources[i%len(sources)],
			Campaign:         "Webinar Follow-up",
			Course:           courses[i%len(courses)],
			Location:         "Hyderabad",
			AssignedTo:       &repID,
			Status:           models.LeadStatusUndecided,
			Priority:         models.PriorityMedium,
			CallAttemptCount: 1,
			FirstCallAt:      &callTime,
			LastCallAt:       &callTime,
			FollowUpStatus:   models.FollowUpStatusPending,
			DoNotCall:        false,
			Notes:            "Hesitant tone, needs human review before final classification.",
			CreatedAt:        callTime.Add(-3 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       callTime,
			DurationSeconds:     110,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeUndecided,
			RecordingConsent:    true,
			RecordingAvailable:  true,
			RecordingURL:        "https://storage.googleapis.com/lead-recordings/ambiguous_call_" + fmt.Sprint(i+1) + ".mp3",
			TranscriptAvailable: true,
			TranscriptID:        &transID,
			AIAnalysisID:        &aiID,
			ManuallyReviewed:    false,
			Notes:               "Awaiting human validation of AI classification.",
			CreatedAt:           callTime,
			UpdatedAt:           callTime,
		}
		calls = append(calls, call)

		trans := models.Transcript{
			ID:            transID,
			LeadID:        leadID,
			CallAttemptID: callID,
			Language:      "en",
			Transcript:    fmt.Sprintf("Salesperson: Would you like to proceed with batch registration? Lead: %s Salesperson: Sure, we will keep in touch.", dialogue),
			SpeakerSegments: []models.SpeakerSegment{
				{Speaker: "salesperson", Text: "Would you like to proceed with batch registration?", TimestampMs: 1000},
				{Speaker: "lead", Text: dialogue, TimestampMs: 4000},
				{Speaker: "salesperson", Text: "Sure, we will keep in touch.", TimestampMs: 9000},
			},
			DurationSeconds: 110,
			Provider:        "whisper-v3",
			CreatedAt:       callTime,
		}
		transcripts = append(transcripts, trans)

		aiAnalyses = append(aiAnalyses, models.AIAnalysis{
			ID:               aiID,
			CallAttemptID:    callID,
			LeadID:           leadID,
			TranscriptID:     transID,
			Intent:           models.AIIntentUndecided,
			Confidence:       0.74, // Below 0.90 threshold -> PENDING_REVIEW!
			Sentiment:        models.SentimentNeutral,
			FollowUpRequired: true,
			DoNotCall:        false,
			Summary:          "Lead is undecided and expressed hesitation. Human review recommended.",
			Reason:           "Hesitation markers detected without explicit refusal.",
			Evidence: []models.EvidenceItem{
				{Speaker: "lead", Text: dialogue},
			},
			Status:    models.AIReviewStatusPendingReview,
			CreatedAt: callTime,
			UpdatedAt: callTime,
		})

		// Notification to review
		notifs = append(notifs, models.Notification{
			ID:        primitive.NewObjectID(),
			UserID:    repID,
			LeadID:    &leadID,
			Type:      models.NotificationTypeAIReviewNeeded,
			Title:     "AI Review Required",
			Message:   fmt.Sprintf("Call with %s classified with 74%% confidence. Please review outcome.", lead.Name),
			Read:      false,
			CreatedAt: callTime,
		})
	}

	// --- Category 5: Follow-Up Scheduled in Future -> FOLLOW_UP_SCHEDULED (No alert yet) ---
	for i := 1; i <= 8; i++ {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		callTime := hoursAgo(float64(5 + i))
		futureDue := hoursFromNow(float64(12 + i*6))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("Scheduled Prospect %d", i),
			Email:            fmt.Sprintf("scheduled%d@example.com", i),
			Phone:            fmt.Sprintf("+91 95000 %05d", 500+i),
			Source:           sources[(i+2)%len(sources)],
			Campaign:         "LinkedIn Executive Outreach",
			Course:           courses[(i+2)%len(courses)],
			Location:         "Pune",
			AssignedTo:       &repID,
			Status:           models.LeadStatusFollowUpScheduled,
			Priority:         models.PriorityHigh,
			CallAttemptCount: 1,
			FirstCallAt:      &callTime,
			LastCallAt:       &callTime,
			NextFollowUpAt:   &futureDue,
			FollowUpStatus:   models.FollowUpStatusScheduled,
			DoNotCall:        false,
			Notes:            "Lead asked to call back after office hours.",
			CreatedAt:        callTime.Add(-2 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       callTime,
			DurationSeconds:     135,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeCallBackLater,
			RecordingConsent:    true,
			RecordingAvailable:  true,
			RecordingURL:        "https://storage.googleapis.com/lead-recordings/scheduled_call_" + fmt.Sprint(i) + ".mp3",
			TranscriptAvailable: false,
			ManuallyReviewed:    true,
			Notes:               "Scheduled follow-up as requested by lead.",
			CreatedAt:           callTime,
			UpdatedAt:           callTime,
		}
		calls = append(calls, call)

		followUps = append(followUps, models.FollowUp{
			ID:            primitive.NewObjectID(),
			LeadID:        leadID,
			SalespersonID: repID,
			DueAt:         futureDue,
			ScheduledAt:   callTime,
			Status:        models.FollowUpTaskScheduled,
			Notes:         "Discuss syllabus breakdown and scholarship options.",
			CreatedAt:     callTime,
			UpdatedAt:     callTime,
		})
	}

	// --- Category 6: Converted & Highly Interested Leads ---
	for i := 1; i <= 6; i++ {
		leadID := primitive.NewObjectID()
		callID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		callTime := hoursAgo(float64(48 + i*12))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("Enrolled Student %d", i),
			Email:            fmt.Sprintf("student%d@example.com", i),
			Phone:            fmt.Sprintf("+91 96000 %05d", 600+i),
			Source:           sources[(i+3)%len(sources)],
			Campaign:         "Early Bird Discount",
			Course:           courses[i%len(courses)],
			Location:         "Kolkata",
			AssignedTo:       &repID,
			Status:           models.LeadStatusConverted,
			Priority:         models.PriorityUrgent,
			CallAttemptCount: 2,
			FirstCallAt:      &callTime,
			LastCallAt:       &callTime,
			FollowUpStatus:   models.FollowUpStatusCompleted,
			DoNotCall:        false,
			Notes:            "Successfully enrolled and fee payment received.",
			CreatedAt:        callTime.Add(-24 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call := models.CallAttempt{
			ID:                  callID,
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       2,
			CallStartedAt:       callTime,
			DurationSeconds:     320,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeConverted,
			RecordingConsent:    true,
			RecordingAvailable:  true,
			TranscriptAvailable: false,
			ManuallyReviewed:    true,
			Notes:               "Payment link shared and verified.",
			CreatedAt:           callTime,
			UpdatedAt:           callTime,
		}
		calls = append(calls, call)
	}

	// --- Category 7: Multi-Call Leads (Attempt #1 = No Answer, Attempt #2 = Answered) -> NOT a one-call lead! ---
	for i := 1; i <= 8; i++ {
		leadID := primitive.NewObjectID()
		repID := salesReps[i%len(salesReps)]
		firstCallTime := hoursAgo(float64(72 + i*5))
		secondCallTime := hoursAgo(float64(24 + i*2))

		lead := models.Lead{
			ID:               leadID,
			Name:             fmt.Sprintf("MultiCall Lead %d", i),
			Email:            fmt.Sprintf("multicall%d@example.com", i),
			Phone:            fmt.Sprintf("+91 97000 %05d", 700+i),
			Source:           sources[i%len(sources)],
			Campaign:         "Multi-Touch Campaign",
			Course:           courses[(i+1)%len(courses)],
			Location:         "Noida",
			AssignedTo:       &repID,
			Status:           models.LeadStatusContacted,
			Priority:         models.PriorityMedium,
			CallAttemptCount: 2, // 2 call attempts -> Not a one-call lead!
			FirstCallAt:      &firstCallTime,
			LastCallAt:       &secondCallTime,
			FollowUpStatus:   models.FollowUpStatusNone,
			DoNotCall:        false,
			Notes:            "Follow-up call completed after initial missed call.",
			CreatedAt:        firstCallTime.Add(-10 * time.Hour),
			UpdatedAt:        now,
		}
		leads = append(leads, lead)

		call1 := models.CallAttempt{
			ID:                  primitive.NewObjectID(),
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       1,
			CallStartedAt:       firstCallTime,
			DurationSeconds:     15,
			CallStatus:          models.CallStatusNoAnswer,
			CallOutcome:         models.CallOutcomeFollowUpRequired,
			RecordingConsent:    false,
			RecordingAvailable:  false,
			TranscriptAvailable: false,
			ManuallyReviewed:    true,
			Notes:               "First call - no response.",
			CreatedAt:           firstCallTime,
			UpdatedAt:           firstCallTime,
		}

		call2 := models.CallAttempt{
			ID:                  primitive.NewObjectID(),
			LeadID:              leadID,
			SalespersonID:       repID,
			AttemptNumber:       2,
			CallStartedAt:       secondCallTime,
			DurationSeconds:     180,
			CallStatus:          models.CallStatusAnswered,
			CallOutcome:         models.CallOutcomeInfoRequested,
			RecordingConsent:    true,
			RecordingAvailable:  false,
			TranscriptAvailable: false,
			ManuallyReviewed:    true,
			Notes:               "Second call - spoke with lead, shared details.",
			CreatedAt:           secondCallTime,
			UpdatedAt:           secondCallTime,
		}

		calls = append(calls, call1, call2)
	}

	// Insert all batches
	if len(leads) > 0 {
		_, err := leadsColl.InsertMany(ctx, leads)
		if err != nil {
			log.Fatalf("Failed to seed leads: %v", err)
		}
		log.Printf("Seeded %d Leads.", len(leads))
	}

	if len(calls) > 0 {
		_, err := callsColl.InsertMany(ctx, calls)
		if err != nil {
			log.Fatalf("Failed to seed calls: %v", err)
		}
		log.Printf("Seeded %d Call Attempts.", len(calls))
	}

	if len(transcripts) > 0 {
		_, err := transcriptsColl.InsertMany(ctx, transcripts)
		if err != nil {
			log.Fatalf("Failed to seed transcripts: %v", err)
		}
		log.Printf("Seeded %d Transcripts.", len(transcripts))
	}

	if len(aiAnalyses) > 0 {
		_, err := aiColl.InsertMany(ctx, aiAnalyses)
		if err != nil {
			log.Fatalf("Failed to seed AI analyses: %v", err)
		}
		log.Printf("Seeded %d AI Analyses.", len(aiAnalyses))
	}

	if len(followUps) > 0 {
		_, err := followUpsColl.InsertMany(ctx, followUps)
		if err != nil {
			log.Fatalf("Failed to seed follow-ups: %v", err)
		}
		log.Printf("Seeded %d Follow-Up Tasks.", len(followUps))
	}

	if len(notifs) > 0 {
		_, err := notifsColl.InsertMany(ctx, notifs)
		if err != nil {
			log.Fatalf("Failed to seed notifications: %v", err)
		}
		log.Printf("Seeded %d Notifications.", len(notifs))
	}

	// Add an initial audit log
	auditColl := mongoDB.Database.Collection("audit_logs")
	_, _ = auditColl.InsertOne(ctx, models.AuditLog{
		ID:         primitive.NewObjectID(),
		UserID:     adminUser.ID,
		UserName:   adminUser.Name,
		Action:     "DATABASE_SEEDED",
		EntityType: "system",
		EntityID:   "initial_seed",
		Reason:     "Development environment populated with realistic scenario dataset",
		CreatedAt:  now,
	})

	log.Println("=========================================================")
	log.Println("Database Seeding Completed Successfully!")
	log.Println("Demo Credentials:")
	log.Println("  Admin:       admin@leadfollowup.com   / password123")
	log.Println("  Manager:     manager@leadfollowup.com / password123")
	log.Println("  Salesperson: vikram@leadfollowup.com  / password123")
	log.Println("  Salesperson: priya@leadfollowup.com   / password123")
	log.Println("  Salesperson: arun@leadfollowup.com    / password123")
	log.Println("=========================================================")
}
