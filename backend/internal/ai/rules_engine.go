package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lead-followup-system/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RulesEngineProvider struct{}

func NewRulesEngineProvider() *RulesEngineProvider {
	return &RulesEngineProvider{}
}

func (p *RulesEngineProvider) Name() string {
	return "heuristic-nlp-engine"
}

func (p *RulesEngineProvider) TranscribeAudio(ctx context.Context, audioURL string, language string) (*models.Transcript, error) {
	// Simulated transcription for testing/mocking
	return &models.Transcript{
		ID:              primitive.NewObjectID(),
		Language:        language,
		Transcript:      "Mock transcription of audio at " + audioURL,
		SpeakerSegments: []models.SpeakerSegment{},
		DurationSeconds: 60,
		Provider:        "mock-transcriber",
		CreatedAt:       time.Now(),
	}, nil
}

func (p *RulesEngineProvider) AnalyzeConversation(ctx context.Context, transcript *models.Transcript) (*AnalysisResult, error) {
	if transcript == nil || len(transcript.SpeakerSegments) == 0 && transcript.Transcript == "" {
		return &AnalysisResult{
			Intent:           models.AIIntentUndecided,
			Confidence:       0.50,
			Sentiment:        models.SentimentNeutral,
			FollowUpRequired: true,
			DoNotCall:        false,
			Summary:          "No transcript available to analyze.",
			Reason:           "Empty conversation record",
			Evidence:         []models.EvidenceItem{},
		}, nil
	}

	// Extract lead dialogue lines
	var leadLines []string
	var leadSegments []models.SpeakerSegment

	for _, seg := range transcript.SpeakerSegments {
		if strings.ToLower(seg.Speaker) == "lead" || strings.ToLower(seg.Speaker) == "customer" {
			leadLines = append(leadLines, seg.Text)
			leadSegments = append(leadSegments, seg)
		}
	}

	// Fallback to full transcript if segments are not separated
	if len(leadLines) == 0 {
		leadLines = []string{transcript.Transcript}
		leadSegments = []models.SpeakerSegment{{Speaker: "lead", Text: transcript.Transcript}}
	}

	leadText := strings.ToLower(strings.Join(leadLines, " "))

	// --- 1. DO NOT CALL (Top priority refusal) ---
	dncPatterns := []string{
		"don't call me again", "dont call me again", "do not call me again", "do not call",
		"stop calling", "remove my number", "delete my number", "take me off your list",
		"stop disturbing", "don't call anymore", "never call me", "dobara call mat karna",
		"call mat karo", "call pannathinga", "malli call cheyoddu", "ini vilikkaruthu",
		"மீண்டும் அழைக்க வேண்டாம்", "கால் பண்ணாதீங்க", "மீண்டும் கால்", "அழைக்காதீர்கள்",
		"दोबारा कॉल मत करना", "कॉल मत करो", "फोन मत करना",
	}
	for _, pattern := range dncPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentDoNotCall,
				Confidence:             0.98,
				Sentiment:              models.SentimentNegative,
				FollowUpRequired:       false,
				DoNotCall:              true,
				SuggestedFollowUpHours: 0,
				Summary:                "Lead explicitly requested not to be contacted again.",
				Reason:                 "Explicit do-not-call request detected.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 2. NOT INTERESTED (Explicit Rejection) ---
	notInterestedPatterns := []string{
		"not interested", "i am not interested", "i'm not interested", "no i'm not interested",
		"don't want this", "dont want this", "not joining", "no interest",
		"i do not want", "please don't call", "mujhe nahi chahiye", "dilchaspi nahi hai",
		"interest illa", "thevai illa", "vendam", "interest ledu", "ishtam ledu", "tatsam illa",
	}
	for _, pattern := range notInterestedPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentNotInterested,
				Confidence:             0.96,
				Sentiment:              models.SentimentNegative,
				FollowUpRequired:       false,
				DoNotCall:              false,
				SuggestedFollowUpHours: 0,
				Summary:                "Lead explicitly stated they are not interested in the offering.",
				Reason:                 "Clear statement of disinterest from lead.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 3. WRONG NUMBER ---
	wrongNumberPatterns := []string{
		"wrong number", "galat number", "thappana number", "tappu number", "thettu number",
		"not rahul", "not the person", "wrong person",
	}
	for _, pattern := range wrongNumberPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentWrongNumber,
				Confidence:             0.95,
				Sentiment:              models.SentimentNeutral,
				FollowUpRequired:       false,
				DoNotCall:              true,
				SuggestedFollowUpHours: 0,
				Summary:                "Lead indicated phone number reached the wrong person.",
				Reason:                 "Wrong number stated by recipient.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 3.5 BROKEN CALL / DISCONNECTED ---
	brokenCallPatterns := []string{
		"call dropped", "call cut", "voice is breaking", "voice breaking", "line disconnected",
		"can you hear me", "hello hello", "call getting disconnected", "voice cut", "call disconnected",
		"awaz nahi aa rahi", "awaz cut rahi hai", "line cut gaya", "voice break aagudhu",
	}
	for _, pattern := range brokenCallPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentFollowUpRequired,
				Confidence:             0.94,
				Sentiment:              models.SentimentNeutral,
				FollowUpRequired:       true,
				DoNotCall:              false,
				SuggestedFollowUpHours: 1,
				Summary:                "Call dropped or got disconnected. Immediate re-dial required.",
				Reason:                 "Audio/call breakage detected in transcript.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 4. CALL BACK LATER (Busy / Request to reschedule) ---
	callLaterPatterns := []string{
		"call me tomorrow", "call me later", "call after", "call back later", "busy right now",
		"i am busy", "in a meeting", "driving", "cannot talk right now", "cant talk",
		"call me evening", "call in the evening", "call at 6", "baad me call karo", "kal call karna",
		"aprom call pannunga", "naalai pesalam", "repu call cheyandi", "pinne vilikkoo",
	}
	for _, pattern := range callLaterPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)

			// Smart Timing Calculation
			suggestedHours := 24
			if strings.Contains(leadText, "2 hours") || strings.Contains(leadText, "after 2 hours") {
				suggestedHours = 2
			} else if strings.Contains(leadText, "evening") || strings.Contains(leadText, "shaam") {
				suggestedHours = 6
			} else if strings.Contains(leadText, "tomorrow morning") || strings.Contains(leadText, "kal subah") {
				suggestedHours = 14
			} else if strings.Contains(leadText, "meeting") || strings.Contains(leadText, "driving") || strings.Contains(leadText, "busy right now") {
				suggestedHours = 3
			}

			return &AnalysisResult{
				Intent:                 models.AIIntentCallBackLater,
				Confidence:             0.92,
				Sentiment:              models.SentimentNeutral,
				FollowUpRequired:       true,
				DoNotCall:              false,
				SuggestedFollowUpHours: suggestedHours,
				Summary:                "Lead is temporarily occupied and requested a call back later.",
				Reason:                 "Call back or busy indication provided by lead.",
				Evidence:               evidence,
			}, nil
		}
	}


	// --- 5. INFO REQUESTED (WhatsApp / Email / Brochure) ---
	infoPatterns := []string{
		"send on whatsapp", "send details on whatsapp", "whatsapp details", "whatsapp me",
		"send me the details", "send brochure", "email me", "share the syllabus",
		"whatsapp pe bhejo", "details bhej do", "whatsapp la anupunga", "details share pannunga",
		"whatsapp lo pampandi", "whatsappil ayakkuka",
	}
	for _, pattern := range infoPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentInfoRequested,
				Confidence:             0.91,
				Sentiment:              models.SentimentNeutral,
				FollowUpRequired:       true,
				DoNotCall:              false,
				SuggestedFollowUpHours: 12,
				Summary:                "Lead requested product or course details via WhatsApp/email.",
				Reason:                 "Lead requested information channel before committing.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 6. CONVERTED / READY TO JOIN ---
	convertedPatterns := []string{
		"i want to enroll", "ready to pay", "where to pay", "send payment link",
		"i am joining", "take my admission", "admission confirm", "payment done",
		"fees pay karunga", "admission podren", "join chesthanu",
	}
	for _, pattern := range convertedPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentConverted,
				Confidence:             0.95,
				Sentiment:              models.SentimentPositive,
				FollowUpRequired:       false,
				DoNotCall:              false,
				SuggestedFollowUpHours: 0,
				Summary:                "Lead is ready to enroll or make payment.",
				Reason:                 "Conversion commitment expressed by lead.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 7. INTERESTED ---
	interestedPatterns := []string{
		"yes i am interested", "yes interested", "tell me more", "explain the course",
		"want to know more", "good course", "interested in learning", "haan mujhe chahiye",
		"aam enaku vendum", "avunu interest vundi",
	}
	for _, pattern := range interestedPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentInterested,
				Confidence:             0.94,
				Sentiment:              models.SentimentPositive,
				FollowUpRequired:       true,
				DoNotCall:              false,
				SuggestedFollowUpHours: 24,
				Summary:                "Lead expressed positive interest in the course or offering.",
				Reason:                 "Lead confirmed interest in learning more.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- 8. UNDECIDED / HESITANT (Requires follow-up, do NOT mark not-interested!) ---
	undecidedPatterns := []string{
		"i will think", "let me think", "discuss with my parents", "talk to my family",
		"not sure right now", "not suitable right now", "maybe later", "need some time",
		"soch ke bataunga", "yosithu solren", "alochan chesi chepthanu", "alochikkatte",
	}
	for _, pattern := range undecidedPatterns {
		if strings.Contains(leadText, pattern) {
			evidence := extractEvidence(leadSegments, pattern)
			return &AnalysisResult{
				Intent:                 models.AIIntentUndecided,
				Confidence:             0.88,
				Sentiment:              models.SentimentNeutral,
				FollowUpRequired:       true,
				DoNotCall:              false,
				SuggestedFollowUpHours: 48,
				Summary:                "Lead is undecided and needs time to evaluate or consult.",
				Reason:                 "Lead indicated need for deliberation before deciding.",
				Evidence:               evidence,
			}, nil
		}
	}

	// --- Default Fallback: Ambiguous / Low confidence conversation ---
	var fallbackEvidence []models.EvidenceItem
	if len(leadSegments) > 0 {
		fallbackEvidence = append(fallbackEvidence, models.EvidenceItem{
			Speaker: leadSegments[0].Speaker,
			Text:    leadSegments[0].Text,
		})
	}

	return &AnalysisResult{
		Intent:                 models.AIIntentFollowUpRequired,
		Confidence:             0.65, // Low confidence -> Will require human review queue!
		Sentiment:              models.SentimentNeutral,
		FollowUpRequired:       true,
		DoNotCall:              false,
		SuggestedFollowUpHours: 24,
		Summary:                "Conversation is ambiguous or inconclusive; human review recommended.",
		Reason:                 "No definitive intent pattern detected with high confidence.",
		Evidence:               fallbackEvidence,
	}, nil
}

func extractEvidence(segments []models.SpeakerSegment, pattern string) []models.EvidenceItem {
	var items []models.EvidenceItem
	for _, seg := range segments {
		if strings.Contains(strings.ToLower(seg.Text), pattern) {
			items = append(items, models.EvidenceItem{
				Speaker: seg.Speaker,
				Text:    seg.Text,
			})
		}
	}
	if len(items) == 0 && len(segments) > 0 {
		items = append(items, models.EvidenceItem{
			Speaker: segments[0].Speaker,
			Text:    fmt.Sprintf("Matched topic around '%s'", pattern),
		})
	}
	return items
}
