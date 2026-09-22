package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FollowUpTaskStatus string

const (
	FollowUpTaskPending   FollowUpTaskStatus = "PENDING"
	FollowUpTaskScheduled FollowUpTaskStatus = "SCHEDULED"
	FollowUpTaskOverdue   FollowUpTaskStatus = "OVERDUE"
	FollowUpTaskCompleted FollowUpTaskStatus = "COMPLETED"
	FollowUpTaskCancelled FollowUpTaskStatus = "CANCELLED"
)

type FollowUp struct {
	ID                       primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	LeadID                   primitive.ObjectID  `bson:"leadId" json:"leadId"`
	LeadName                 string              `bson:"-" json:"leadName,omitempty"`
	LeadPhone                string              `bson:"-" json:"leadPhone,omitempty"`
	SalespersonID            primitive.ObjectID  `bson:"salespersonId" json:"salespersonId"`
	SalespersonUser          *UserSummary        `bson:"-" json:"salespersonUser,omitempty"`
	DueAt                    time.Time           `bson:"dueAt" json:"dueAt"`
	ScheduledAt              time.Time           `bson:"scheduledAt" json:"scheduledAt"`
	Status                   FollowUpTaskStatus  `bson:"status" json:"status"`
	Notes                    string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CompletedAt              *time.Time          `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	CompletedByCallAttemptID *primitive.ObjectID `bson:"completedByCallAttemptId,omitempty" json:"completedByCallAttemptId,omitempty"`
	CreatedAt                time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt                time.Time           `bson:"updatedAt" json:"updatedAt"`
}
