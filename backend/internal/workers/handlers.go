package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lead-followup-system/internal/services"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	followUpService *services.FollowUpService
	aiService       *services.AIService
}

func NewTaskHandler(followUpService *services.FollowUpService, aiService *services.AIService) *TaskHandler {
	return &TaskHandler{
		followUpService: followUpService,
		aiService:       aiService,
	}
}

func (h *TaskHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	switch t.Type() {
	case TypeDetectOneCallFollowUps:
		return h.handleDetectOneCall(ctx, t)
	case TypeAnalyzeCallConversation:
		return h.handleAnalyzeCall(ctx, t)
	default:
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}
}

func (h *TaskHandler) handleDetectOneCall(ctx context.Context, t *asynq.Task) error {
	log.Println("[Worker] Running DETECT_ONE_CALL_FOLLOWUPS background job...")
	count, err := h.followUpService.DetectOneCallFollowUps(ctx)
	if err != nil {
		log.Printf("[Worker ERROR] DETECT_ONE_CALL_FOLLOWUPS failed: %v", err)
		return err
	}
	log.Printf("[Worker] DETECT_ONE_CALL_FOLLOWUPS completed successfully. %d leads updated.", count)
	return nil
}

func (h *TaskHandler) handleAnalyzeCall(ctx context.Context, t *asynq.Task) error {
	var payload AnalyzeCallPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	callID, err := primitive.ObjectIDFromHex(payload.CallAttemptID)
	if err != nil {
		return err
	}

	log.Printf("[Worker] Running ANALYZE_CALL_CONVERSATION for call %s", payload.CallAttemptID)
	_, err = h.aiService.AnalyzeCallConversation(ctx, callID, nil)
	return err
}
