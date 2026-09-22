package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CallStatus string

const (
	CallStatusInitiated   CallStatus = "INITIATED"
	CallStatusRinging     CallStatus = "RINGING"
	CallStatusAnswered    CallStatus = "ANSWERED"
	CallStatusNoAnswer    CallStatus = "NO_ANSWER"
	CallStatusBusy        CallStatus = "BUSY"
	CallStatusSwitchedOff CallStatus = "SWITCHED_OFF"
	CallStatusUnreachable CallStatus = "UNREACHABLE"
	CallStatusFailed      CallStatus = "FAILED"
)

type CallOutcome string

const (
	CallOutcomeInterested         CallOutcome = "INTERESTED"
	CallOutcomeNotInterested      CallOutcome = "NOT_INTERESTED"
	CallOutcomeCallBackLater      CallOutcome = "CALL_BACK_LATER"
	CallOutcomeInfoRequested      CallOutcome = "INFO_REQUESTED"
	CallOutcomeUndecided          CallOutcome = "UNDECIDED"
	CallOutcomeFollowUpRequired   CallOutcome = "FOLLOW_UP_REQUIRED"
	CallOutcomeFollowUpScheduled  CallOutcome = "FOLLOW_UP_SCHEDULED"
	CallOutcomeConverted          CallOutcome = "CONVERTED"
	CallOutcomeLost               CallOutcome = "LOST"
	CallOutcomeWrongNumber        CallOutcome = "WRONG_NUMBER"
	CallOutcomeDoNotCall          CallOutcome = "DO_NOT_CALL"
)

type CallAttempt struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	LeadID              primitive.ObjectID  `bson:"leadId" json:"leadId"`
	SalespersonID       primitive.ObjectID  `bson:"salespersonId" json:"salespersonId"`
	SalespersonUser     *UserSummary        `bson:"-" json:"salespersonUser,omitempty"`
	AttemptNumber       int                 `bson:"attemptNumber" json:"attemptNumber"`
	CallStartedAt       time.Time           `bson:"callStartedAt" json:"callStartedAt"`
	CallEndedAt         *time.Time          `bson:"callEndedAt,omitempty" json:"callEndedAt,omitempty"`
	DurationSeconds     int                 `bson:"durationSeconds" json:"durationSeconds"`
	CallStatus          CallStatus          `bson:"callStatus" json:"callStatus"`
	CallOutcome         CallOutcome         `bson:"callOutcome" json:"callOutcome"`
	RecordingConsent    bool                `bson:"recordingConsent" json:"recordingConsent"`
	RecordingAvailable  bool                `bson:"recordingAvailable" json:"recordingAvailable"`
	RecordingURL        string              `bson:"recordingUrl,omitempty" json:"recordingUrl,omitempty"`
	RecordingProvider   string              `bson:"recordingProvider,omitempty" json:"recordingProvider,omitempty"`
	TranscriptAvailable bool                `bson:"transcriptAvailable" json:"transcriptAvailable"`
	TranscriptID        *primitive.ObjectID `bson:"transcriptId,omitempty" json:"transcriptId,omitempty"`
	AIAnalysisID        *primitive.ObjectID `bson:"aiAnalysisId,omitempty" json:"aiAnalysisId,omitempty"`
	ManuallyReviewed    bool                `bson:"manuallyReviewed" json:"manuallyReviewed"`
	Notes               string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt           time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time           `bson:"updatedAt" json:"updatedAt"`
}
