package ai

import (
	"context"
	"lead-followup-system/internal/models"
)

type AnalysisResult struct {
	Intent                 models.AIIntent       `json:"intent"`
	Confidence             float64               `json:"confidence"`
	Sentiment              models.Sentiment      `json:"sentiment"`
	FollowUpRequired       bool                  `json:"followUpRequired"`
	DoNotCall              bool                  `json:"doNotCall"`
	SuggestedFollowUpHours int                   `json:"suggestedFollowUpHours,omitempty"`
	Summary                string                `json:"summary"`
	Reason                 string                `json:"reason"`
	Evidence               []models.EvidenceItem `json:"evidence"`
}

type AIProvider interface {
	AnalyzeConversation(ctx context.Context, transcript *models.Transcript) (*AnalysisResult, error)
	TranscribeAudio(ctx context.Context, audioURL string, language string) (*models.Transcript, error)
	Name() string
}
