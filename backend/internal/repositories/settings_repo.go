package repositories

import (
	"context"
	"time"

	"lead-followup-system/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingsRepository struct {
	collection *mongo.Collection
}

func NewSettingsRepository(db *mongo.Database) *SettingsRepository {
	return &SettingsRepository{
		collection: db.Collection("system_settings"),
	}
}

func (r *SettingsRepository) GetSettings(ctx context.Context) (*models.SystemSettings, error) {
	var settings models.SystemSettings
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Return default settings
		defaultSettings := &models.SystemSettings{
			ID:                     primitive.NewObjectID(),
			FollowUpThresholdHours: 24,
			AIConfidenceThreshold:  0.90,
			AutoClassifyEnabled:    true,
			HumanReviewRequiredDNC: true,
			EscalationHours:        48,
			SupportedLanguages:     []string{"en", "hi", "ta", "te", "ml", "kn"},
			DuplicateCheckFields:   []string{"phone", "email"},
			UpdatedAt:              time.Now(),
		}
		_, _ = r.collection.InsertOne(ctx, defaultSettings)
		return defaultSettings, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *SettingsRepository) UpdateSettings(ctx context.Context, settings *models.SystemSettings) error {
	settings.UpdatedAt = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{},
		bson.M{"$set": settings},
		opts,
	)
	return err
}
