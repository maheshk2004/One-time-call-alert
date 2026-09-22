package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SystemSettings struct {
	ID                     primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	FollowUpThresholdHours int                 `bson:"followUpThresholdHours" json:"followUpThresholdHours"`
	AIConfidenceThreshold  float64             `bson:"aiConfidenceThreshold" json:"aiConfidenceThreshold"`
	AutoClassifyEnabled    bool                `bson:"autoClassifyEnabled" json:"autoClassifyEnabled"`
	HumanReviewRequiredDNC bool                `bson:"humanReviewRequiredDNC" json:"humanReviewRequiredDNC"` // Require review for DNC / NOT_INTERESTED
	EscalationHours        int                 `bson:"escalationHours" json:"escalationHours"`
	SupportedLanguages     []string            `bson:"supportedLanguages" json:"supportedLanguages"`
	DuplicateCheckFields   []string            `bson:"duplicateCheckFields" json:"duplicateCheckFields"` // "phone", "email"
	UpdatedBy              *primitive.ObjectID `bson:"updatedBy,omitempty" json:"updatedBy,omitempty"`
	UpdatedAt              time.Time           `bson:"updatedAt" json:"updatedAt"`
}
