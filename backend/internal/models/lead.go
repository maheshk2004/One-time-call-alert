package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadStatus string

const (
	LeadStatusNew                LeadStatus = "NEW"
	LeadStatusAssigned           LeadStatus = "ASSIGNED"
	LeadStatusCallPending        LeadStatus = "CALL_PENDING"
	LeadStatusCalledOnce         LeadStatus = "CALLED_ONCE"
	LeadStatusFollowUpScheduled  LeadStatus = "FOLLOW_UP_SCHEDULED"
	LeadStatusFollowUpRequired   LeadStatus = "FOLLOW_UP_REQUIRED"
	LeadStatusContacted          LeadStatus = "CONTACTED"
	LeadStatusInterested         LeadStatus = "INTERESTED"
	LeadStatusUndecided          LeadStatus = "UNDECIDED"
	LeadStatusInfoRequested      LeadStatus = "INFO_REQUESTED"
	LeadStatusNotInterested      LeadStatus = "NOT_INTERESTED"
	LeadStatusDoNotCall          LeadStatus = "DO_NOT_CALL"
	LeadStatusConverted          LeadStatus = "CONVERTED"
	LeadStatusLost               LeadStatus = "LOST"
	LeadStatusDisqualified       LeadStatus = "DISQUALIFIED"
)

type Priority string

const (
	PriorityLow    Priority = "LOW"
	PriorityMedium Priority = "MEDIUM"
	PriorityHigh   Priority = "HIGH"
	PriorityUrgent Priority = "URGENT"
)

type FollowUpStatus string

const (
	FollowUpStatusNone      FollowUpStatus = "NONE"
	FollowUpStatusPending   FollowUpStatus = "PENDING"
	FollowUpStatusScheduled FollowUpStatus = "SCHEDULED"
	FollowUpStatusOverdue   FollowUpStatus = "OVERDUE"
	FollowUpStatusCompleted FollowUpStatus = "COMPLETED"
)

type Lead struct {
	ID                  primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name                string              `bson:"name" json:"name"`
	Email               string              `bson:"email" json:"email"`
	Phone               string              `bson:"phone" json:"phone"`
	AlternatePhone      string              `bson:"alternatePhone,omitempty" json:"alternatePhone,omitempty"`
	Source              string              `bson:"source" json:"source"`
	Campaign            string              `bson:"campaign,omitempty" json:"campaign,omitempty"`
	AdName              string              `bson:"adName,omitempty" json:"adName,omitempty"`
	Course              string              `bson:"course,omitempty" json:"course,omitempty"`
	Location            string              `bson:"location,omitempty" json:"location,omitempty"`
	AssignedTo          *primitive.ObjectID `bson:"assignedTo,omitempty" json:"assignedTo,omitempty"`
	AssignedUser        *UserSummary        `bson:"-" json:"assignedUser,omitempty"`
	Status              LeadStatus          `bson:"status" json:"status"`
	Priority            Priority            `bson:"priority" json:"priority"`
	CallAttemptCount    int                 `bson:"callAttemptCount" json:"callAttemptCount"`
	FirstCallAt         *time.Time          `bson:"firstCallAt,omitempty" json:"firstCallAt,omitempty"`
	LastCallAt          *time.Time          `bson:"lastCallAt,omitempty" json:"lastCallAt,omitempty"`
	NextFollowUpAt      *time.Time          `bson:"nextFollowUpAt,omitempty" json:"nextFollowUpAt,omitempty"`
	FollowUpStatus      FollowUpStatus      `bson:"followUpStatus" json:"followUpStatus"`
	DoNotCall           bool                `bson:"doNotCall" json:"doNotCall"`
	DoNotCallReason     string              `bson:"doNotCallReason,omitempty" json:"doNotCallReason,omitempty"`
	DoNotCallDetectedBy string              `bson:"doNotCallDetectedBy,omitempty" json:"doNotCallDetectedBy,omitempty"` // AI, Salesperson, Admin
	DoNotCallAt         *time.Time          `bson:"doNotCallAt,omitempty" json:"doNotCallAt,omitempty"`
	Notes               string              `bson:"notes,omitempty" json:"notes,omitempty"`
	LastNotificationAt  *time.Time          `bson:"lastNotificationAt,omitempty" json:"lastNotificationAt,omitempty"`
	NotificationCount   int                 `bson:"notificationCount" json:"notificationCount"`
	CreatedAt           time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time           `bson:"updatedAt" json:"updatedAt"`
}
