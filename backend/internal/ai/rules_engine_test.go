package ai

import (
	"context"
	"testing"

	"lead-followup-system/internal/models"
)

func TestAIAnalysisExplicitNotInterested(t *testing.T) {
	engine := NewRulesEngineProvider()
	transcript := &models.Transcript{
		Language: "en",
		SpeakerSegments: []models.SpeakerSegment{
			{Speaker: "salesperson", Text: "Hello, calling about the course."},
			{Speaker: "lead", Text: "No, I am not interested. Please do not call."},
		},
	}

	result, err := engine.AnalyzeConversation(context.Background(), transcript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Intent != models.AIIntentNotInterested && result.Intent != models.AIIntentDoNotCall {
		t.Errorf("expected NOT_INTERESTED or DO_NOT_CALL, got: %s", result.Intent)
	}
	if result.Confidence < 0.90 {
		t.Errorf("expected high confidence >= 0.90, got: %f", result.Confidence)
	}
	if result.FollowUpRequired != false {
		t.Errorf("expected FollowUpRequired to be false, got: true")
	}
}

func TestAIAnalysisDoNotCall(t *testing.T) {
	engine := NewRulesEngineProvider()
	transcript := &models.Transcript{
		Language: "en",
		SpeakerSegments: []models.SpeakerSegment{
			{Speaker: "salesperson", Text: "Hi Rahul, are you free to speak?"},
			{Speaker: "lead", Text: "Don't call me again! Remove my number immediately."},
		},
	}

	result, err := engine.AnalyzeConversation(context.Background(), transcript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Intent != models.AIIntentDoNotCall {
		t.Errorf("expected DO_NOT_CALL, got: %s", result.Intent)
	}
	if !result.DoNotCall {
		t.Errorf("expected DoNotCall flag to be true")
	}
	if result.FollowUpRequired {
		t.Errorf("expected FollowUpRequired to be false")
	}
}

func TestAIAnalysisCallBackLater(t *testing.T) {
	engine := NewRulesEngineProvider()
	transcript := &models.Transcript{
		Language: "en",
		SpeakerSegments: []models.SpeakerSegment{
			{Speaker: "salesperson", Text: "Hi, following up on your inquiry."},
			{Speaker: "lead", Text: "I am busy right now, in a meeting. Please call me tomorrow."},
		},
	}

	result, err := engine.AnalyzeConversation(context.Background(), transcript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Intent != models.AIIntentCallBackLater {
		t.Errorf("expected CALL_BACK_LATER, got: %s", result.Intent)
	}
	if !result.FollowUpRequired {
		t.Errorf("expected FollowUpRequired to be true")
	}
	if result.DoNotCall {
		t.Errorf("expected DoNotCall to be false")
	}
}

func TestAIAnalysisAmbiguousHesitant(t *testing.T) {
	engine := NewRulesEngineProvider()
	transcript := &models.Transcript{
		Language: "en",
		SpeakerSegments: []models.SpeakerSegment{
			{Speaker: "salesperson", Text: "Are you ready to join the cohort?"},
			{Speaker: "lead", Text: "I don't think this is suitable for me right now, let me discuss with my parents."},
		},
	}

	result, err := engine.AnalyzeConversation(context.Background(), transcript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must NOT classify as DO_NOT_CALL or NOT_INTERESTED per Section 12 & 39
	if result.Intent == models.AIIntentDoNotCall {
		t.Errorf("ambiguous hesitation should NOT be classified as DO_NOT_CALL")
	}
	if result.Intent != models.AIIntentUndecided {
		t.Errorf("expected UNDECIDED, got: %s", result.Intent)
	}
}

func TestAIAnalysisMultilingualTamilDNC(t *testing.T) {
	engine := NewRulesEngineProvider()
	transcript := &models.Transcript{
		Language: "ta",
		SpeakerSegments: []models.SpeakerSegment{
			{Speaker: "salesperson", Text: "வணக்கம், கோர்ஸ் பற்றி பேசலாமா?"},
			{Speaker: "lead", Text: "மீண்டும் கால் பண்ணாதீங்க"},
		},
	}

	result, err := engine.AnalyzeConversation(context.Background(), transcript)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Intent != models.AIIntentDoNotCall {
		t.Errorf("expected DO_NOT_CALL for Tamil refusal, got: %s", result.Intent)
	}
}
