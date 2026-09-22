package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID     `bson:"userId" json:"userId"`
	UserName   string                 `bson:"userName" json:"userName"`
	Action     string                 `bson:"action" json:"action"`
	EntityType string                 `bson:"entityType" json:"entityType"` // "lead", "call_attempt", "ai_analysis", "follow_up", "settings"
	EntityID   string                 `bson:"entityId" json:"entityId"`
	OldValue   map[string]interface{} `bson:"oldValue,omitempty" json:"oldValue,omitempty"`
	NewValue   map[string]interface{} `bson:"newValue,omitempty" json:"newValue,omitempty"`
	Reason     string                 `bson:"reason,omitempty" json:"reason,omitempty"`
	Metadata   map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt  time.Time              `bson:"createdAt" json:"createdAt"`
}
