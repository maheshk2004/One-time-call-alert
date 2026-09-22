package services

import (
	"context"
	"fmt"
	"time"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TranscriptService struct {
	transcriptRepo *repositories.TranscriptRepository
	callRepo       *repositories.CallRepository
	auditService   *AuditService
}

func NewTranscriptService(
	transcriptRepo *repositories.TranscriptRepository,
	callRepo *repositories.CallRepository,
	auditService *AuditService,
) *TranscriptService {
	return &TranscriptService{
		transcriptRepo: transcriptRepo,
		callRepo:       callRepo,
		auditService:   auditService,
	}
}

type SaveTranscriptInput struct {
	CallAttemptID   primitive.ObjectID      `json:"callAttemptId"`
	Language        string                  `json:"language"`
	Transcript      string                  `json:"transcript" binding:"required"`
	SpeakerSegments []models.SpeakerSegment `json:"speakerSegments"`
	DurationSeconds int                     `json:"durationSeconds"`
	Provider        string                  `json:"provider"`
}

func (s *TranscriptService) SaveTranscript(ctx context.Context, input SaveTranscriptInput, actor *models.User) (*models.Transcript, error) {
	call, err := s.callRepo.FindByID(ctx, input.CallAttemptID)
	if err != nil {
		return nil, fmt.Errorf("call attempt not found: %w", err)
	}

	lang := input.Language
	if lang == "" {
		lang = "en"
	}
	provider := input.Provider
	if provider == "" {
		provider = "whisper-speech-to-text"
	}

	transcript := &models.Transcript{
		LeadID:          call.LeadID,
		CallAttemptID:   call.ID,
		Language:        lang,
		Transcript:      input.Transcript,
		SpeakerSegments: input.SpeakerSegments,
		DurationSeconds: input.DurationSeconds,
		Provider:        provider,
		CreatedAt:       time.Now(),
	}

	if err := s.transcriptRepo.Create(ctx, transcript); err != nil {
		return nil, err
	}

	// Update call record
	call.TranscriptAvailable = true
	call.TranscriptID = &transcript.ID
	_ = s.callRepo.Update(ctx, call)

	s.auditService.Log(ctx, actor.ID, actor.Name, "TRANSCRIPT_CREATED", "transcript", transcript.ID.Hex(), nil, map[string]interface{}{
		"callAttemptId": call.ID.Hex(),
		"leadId":        call.LeadID.Hex(),
		"language":      lang,
	}, "Call conversation transcript uploaded", nil)

	return transcript, nil
}

func (s *TranscriptService) GetByCallAttemptID(ctx context.Context, callAttemptID primitive.ObjectID) (*models.Transcript, error) {
	return s.transcriptRepo.FindByCallAttemptID(ctx, callAttemptID)
}

func (s *TranscriptService) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Transcript, error) {
	return s.transcriptRepo.FindByID(ctx, id)
}
