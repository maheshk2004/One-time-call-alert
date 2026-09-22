package services

import (
	"context"
	"log"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditService struct {
	auditRepo *repositories.AuditRepository
}

func NewAuditService(auditRepo *repositories.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

func (s *AuditService) Log(ctx context.Context, userID primitive.ObjectID, userName string, action string, entityType string, entityID string, oldValue map[string]interface{}, newValue map[string]interface{}, reason string, metadata map[string]interface{}) {
	entry := &models.AuditLog{
		UserID:     userID,
		UserName:   userName,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Reason:     reason,
		Metadata:   metadata,
	}

	if err := s.auditRepo.Create(ctx, entry); err != nil {
		log.Printf("Failed to record audit log: %v", err)
	}
}

func (s *AuditService) GetLogsForEntity(ctx context.Context, entityType, entityID string) ([]*models.AuditLog, error) {
	return s.auditRepo.FindByEntity(ctx, entityType, entityID)
}

func (s *AuditService) GetRecentLogs(ctx context.Context, limit int) ([]*models.AuditLog, error) {
	return s.auditRepo.FindRecent(ctx, limit)
}
