package services

import (
	"context"
	"log"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	notifRepo *repositories.NotificationRepository
}

func NewNotificationService(notifRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
	}
}

func (s *NotificationService) Send(ctx context.Context, userID primitive.ObjectID, leadID *primitive.ObjectID, nType models.NotificationType, title string, message string, metadata map[string]interface{}) error {
	notif := &models.Notification{
		UserID:   userID,
		LeadID:   leadID,
		Type:     nType,
		Title:    title,
		Message:  message,
		Metadata: metadata,
	}

	err := s.notifRepo.Create(ctx, notif)
	if err != nil {
		log.Printf("Failed to create notification for user %s: %v", userID.Hex(), err)
		return err
	}
	return nil
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID primitive.ObjectID, unreadOnly bool, limit int) ([]*models.Notification, error) {
	return s.notifRepo.FindByUserID(ctx, userID, unreadOnly, limit)
}

func (s *NotificationService) MarkRead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}
