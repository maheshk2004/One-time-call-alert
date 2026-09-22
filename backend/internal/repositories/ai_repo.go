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

type AIRepository struct {
	collection *mongo.Collection
}

func NewAIRepository(db *mongo.Database) *AIRepository {
	return &AIRepository{
		collection: db.Collection("ai_analyses"),
	}
}

func (r *AIRepository) Create(ctx context.Context, analysis *models.AIAnalysis) error {
	analysis.ID = primitive.NewObjectID()
	analysis.CreatedAt = time.Now()
	analysis.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, analysis)
	return err
}

func (r *AIRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.AIAnalysis, error) {
	var analysis models.AIAnalysis
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&analysis)
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *AIRepository) FindByCallAttemptID(ctx context.Context, callAttemptID primitive.ObjectID) (*models.AIAnalysis, error) {
	var analysis models.AIAnalysis
	err := r.collection.FindOne(ctx, bson.M{"callAttemptId": callAttemptID}).Decode(&analysis)
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *AIRepository) FindPendingReviews(ctx context.Context, page, pageSize int) ([]*models.AIAnalysis, int64, error) {
	query := bson.M{
		"status": models.AIReviewStatusPendingReview,
	}

	totalCount, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var analyses []*models.AIAnalysis
	if err := cursor.All(ctx, &analyses); err != nil {
		return nil, 0, err
	}
	return analyses, totalCount, nil
}

func (r *AIRepository) Update(ctx context.Context, analysis *models.AIAnalysis) error {
	analysis.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": analysis.ID},
		bson.M{"$set": analysis},
	)
	return err
}

func (r *AIRepository) GetAIClassificationStats(ctx context.Context) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$intent"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int64)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&item); err == nil {
			stats[item.ID] = item.Count
		}
	}
	return stats, nil
}
