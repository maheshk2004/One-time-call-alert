package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SpeakerSegment struct {
	Speaker     string `bson:"speaker" json:"speaker"` // "salesperson" or "lead"
	Text        string `bson:"text" json:"text"`
	TimestampMs int64  `bson:"timestampMs,omitempty" json:"timestampMs,omitempty"`
}

type Transcript struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LeadID          primitive.ObjectID `bson:"leadId" json:"leadId"`
	CallAttemptID   primitive.ObjectID `bson:"callAttemptId" json:"callAttemptId"`
	Language        string             `bson:"language" json:"language"` // "en", "hi", "ta", "te", "ml", "kn", etc.
	Transcript      string             `bson:"transcript" json:"transcript"`
	SpeakerSegments []SpeakerSegment   `bson:"speakerSegments" json:"speakerSegments"`
	DurationSeconds int                `bson:"durationSeconds" json:"durationSeconds"`
	Provider        string             `bson:"provider" json:"provider"` // "whisper", "deepgram", "manual", "mock"
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}
