package workers

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeDetectOneCallFollowUps = "lead:detect_one_call"
	TypeProcessCallTranscript  = "call:process_transcript"
	TypeAnalyzeCallConversation = "call:analyze_conversation"
	TypeSendFollowUpEmail      = "email:send_followup"
	TypeEscalateOverdueLead    = "lead:escalate_overdue"
)

type DetectOneCallPayload struct {
	TriggerSource string `json:"triggerSource"`
}

type AnalyzeCallPayload struct {
	CallAttemptID string `json:"callAttemptId"`
}

func NewDetectOneCallTask() (*asynq.Task, error) {
	payload, err := json.Marshal(DetectOneCallPayload{TriggerSource: "scheduler"})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeDetectOneCallFollowUps, payload), nil
}

func NewAnalyzeCallTask(callAttemptID string) (*asynq.Task, error) {
	payload, err := json.Marshal(AnalyzeCallPayload{CallAttemptID: callAttemptID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAnalyzeCallConversation, payload), nil
}
