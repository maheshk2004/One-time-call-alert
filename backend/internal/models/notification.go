package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotificationTypeOneCallFollowUp NotificationType = "ONE_CALL_FOLLOWUP_REQUIRED"
	NotificationTypeFollowUpOverdue NotificationType = "FOLLOWUP_OVERDUE"
	NotificationTypeAIReviewNeeded  NotificationType = "AI_REVIEW_NEEDED"
	NotificationTypeLeadAssigned    NotificationType = "LEAD_ASSIGNED"
	NotificationTypeDoNotCallAlert  NotificationType = "DO_NOT_CALL_ALERT"
)

type Notification struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID     `bson:"userId" json:"userId"`
	LeadID    *primitive.ObjectID    `bson:"leadId,omitempty" json:"leadId,omitempty"`
	Type      NotificationType       `bson:"type" json:"type"`
	Title     string                 `bson:"title" json:"title"`
	Message   string                 `bson:"message" json:"message"`
	Read      bool                   `bson:"read" json:"read"`
	ReadAt    *time.Time             `bson:"readAt,omitempty" json:"readAt,omitempty"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
}
