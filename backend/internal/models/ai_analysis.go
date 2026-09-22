package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AIIntent string

const (
	AIIntentInterested        AIIntent = "INTERESTED"
	AIIntentNotInterested     AIIntent = "NOT_INTERESTED"
	AIIntentCallBackLater     AIIntent = "CALL_BACK_LATER"
	AIIntentInfoRequested     AIIntent = "INFO_REQUESTED"
	AIIntentUndecided         AIIntent = "UNDECIDED"
	AIIntentFollowUpRequired  AIIntent = "FOLLOW_UP_REQUIRED"
	AIIntentConverted         AIIntent = "CONVERTED"
	AIIntentLost              AIIntent = "LOST"
	AIIntentDoNotCall         AIIntent = "DO_NOT_CALL"
	AIIntentWrongNumber       AIIntent = "WRONG_NUMBER"
)

type Sentiment string

const (
	SentimentPositive Sentiment = "POSITIVE"
	SentimentNeutral  Sentiment = "NEUTRAL"
	SentimentNegative Sentiment = "NEGATIVE"
)

type AIReviewStatus string

const (
	AIReviewStatusAutoApplied   AIReviewStatus = "AUTO_APPLIED"
	AIReviewStatusPendingReview AIReviewStatus = "PENDING_REVIEW"
	AIReviewStatusConfirmed     AIReviewStatus = "CONFIRMED"
	AIReviewStatusOverridden    AIReviewStatus = "OVERRIDDEN"
)

type EvidenceItem struct {
	Speaker string `bson:"speaker" json:"speaker"`
	Text    string `bson:"text" json:"text"`
}

type AIAnalysis struct {
	ID                     primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	CallAttemptID          primitive.ObjectID  `bson:"callAttemptId" json:"callAttemptId"`
	LeadID                 primitive.ObjectID  `bson:"leadId" json:"leadId"`
	TranscriptID           primitive.ObjectID  `bson:"transcriptId" json:"transcriptId"`
	Intent                 AIIntent            `bson:"intent" json:"intent"`
	Confidence             float64             `bson:"confidence" json:"confidence"`
	Sentiment              Sentiment           `bson:"sentiment" json:"sentiment"`
	FollowUpRequired       bool                `bson:"followUpRequired" json:"followUpRequired"`
	DoNotCall              bool                `bson:"doNotCall" json:"doNotCall"`
	SuggestedFollowUpHours int                 `bson:"suggestedFollowUpHours,omitempty" json:"suggestedFollowUpHours,omitempty"`
	Summary                string              `bson:"summary" json:"summary"`
	Reason                 string              `bson:"reason" json:"reason"`
	Evidence               []EvidenceItem      `bson:"evidence" json:"evidence"`
	Status                 AIReviewStatus      `bson:"status" json:"status"`
	OriginalIntent         AIIntent            `bson:"originalIntent,omitempty" json:"originalIntent,omitempty"`
	ReviewedBy             *primitive.ObjectID `bson:"reviewedBy,omitempty" json:"reviewedBy,omitempty"`
	ReviewedByUser         *UserSummary        `bson:"-" json:"reviewedByUser,omitempty"`
	ReviewedAt             *time.Time          `bson:"reviewedAt,omitempty" json:"reviewedAt,omitempty"`
	OverrideReason         string              `bson:"overrideReason,omitempty" json:"overrideReason,omitempty"`
	CreatedAt              time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time           `bson:"updatedAt" json:"updatedAt"`
}
