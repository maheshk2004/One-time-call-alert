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

type FollowUpRepository struct {
	collection *mongo.Collection
}

func NewFollowUpRepository(db *mongo.Database) *FollowUpRepository {
	return &FollowUpRepository{
		collection: db.Collection("follow_ups"),
	}
}

func (r *FollowUpRepository) Create(ctx context.Context, followUp *models.FollowUp) error {
	followUp.ID = primitive.NewObjectID()
	followUp.CreatedAt = time.Now()
	followUp.UpdatedAt = time.Now()
	if followUp.Status == "" {
		followUp.Status = models.FollowUpTaskPending
	}
	_, err := r.collection.InsertOne(ctx, followUp)
	return err
}

func (r *FollowUpRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.FollowUp, error) {
	var followUp models.FollowUp
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&followUp)
	if err != nil {
		return nil, err
	}
	return &followUp, nil
}

func (r *FollowUpRepository) FindByLeadID(ctx context.Context, leadID primitive.ObjectID) ([]*models.FollowUp, error) {
	opts := options.Find().SetSort(bson.D{{Key: "dueAt", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"leadId": leadID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var followUps []*models.FollowUp
	if err := cursor.All(ctx, &followUps); err != nil {
		return nil, err
	}
	return followUps, nil
}

func (r *FollowUpRepository) FindPendingByLeadID(ctx context.Context, leadID primitive.ObjectID) (*models.FollowUp, error) {
	var followUp models.FollowUp
	err := r.collection.FindOne(ctx, bson.M{
		"leadId": leadID,
		"status": bson.M{"$in": []models.FollowUpTaskStatus{
			models.FollowUpTaskPending,
			models.FollowUpTaskScheduled,
			models.FollowUpTaskOverdue,
		}},
	}).Decode(&followUp)
	if err != nil {
		return nil, err
	}
	return &followUp, nil
}

func (r *FollowUpRepository) FindFollowUps(ctx context.Context, salespersonID *primitive.ObjectID, status models.FollowUpTaskStatus, timeRange string, page, pageSize int) ([]*models.FollowUp, int64, error) {
	query := bson.M{}
	if salespersonID != nil {
		query["salespersonId"] = salespersonID
	}
	if status != "" {
		query["status"] = status
	}

	now := time.Now()
	if timeRange == "today" {
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		query["dueAt"] = bson.M{"$gte": startOfDay, "$lt": endOfDay}
	} else if timeRange == "overdue" {
		query["dueAt"] = bson.M{"$lt": now}
		query["status"] = bson.M{"$in": []models.FollowUpTaskStatus{
			models.FollowUpTaskPending,
			models.FollowUpTaskScheduled,
			models.FollowUpTaskOverdue,
		}}
	}

	totalCount, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "dueAt", Value: 1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var followUps []*models.FollowUp
	if err := cursor.All(ctx, &followUps); err != nil {
		return nil, 0, err
	}
	return followUps, totalCount, nil
}

func (r *FollowUpRepository) FindOverdueFollowUps(ctx context.Context, now time.Time) ([]*models.FollowUp, error) {
	query := bson.M{
		"dueAt": bson.M{"$lt": now},
		"status": bson.M{"$in": []models.FollowUpTaskStatus{
			models.FollowUpTaskPending,
			models.FollowUpTaskScheduled,
		}},
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var followUps []*models.FollowUp
	if err := cursor.All(ctx, &followUps); err != nil {
		return nil, err
	}
	return followUps, nil
}

func (r *FollowUpRepository) Update(ctx context.Context, followUp *models.FollowUp) error {
	followUp.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": followUp.ID},
		bson.M{"$set": followUp},
	)
	return err
}

func (r *FollowUpRepository) CompleteFollowUp(ctx context.Context, id primitive.ObjectID, callAttemptID *primitive.ObjectID) error {
	now := time.Now()
	update := bson.M{
		"status":      models.FollowUpTaskCompleted,
		"completedAt": &now,
		"updatedAt":   now,
	}
	if callAttemptID != nil {
		update["completedByCallAttemptId"] = callAttemptID
	}
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *FollowUpRepository) CancelPendingByLeadID(ctx context.Context, leadID primitive.ObjectID) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"leadId": leadID,
			"status": bson.M{"$in": []models.FollowUpTaskStatus{
				models.FollowUpTaskPending,
				models.FollowUpTaskScheduled,
				models.FollowUpTaskOverdue,
			}},
		},
		bson.M{
			"$set": bson.M{
				"status":    models.FollowUpTaskCancelled,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

func (r *FollowUpRepository) CountStats(ctx context.Context, salespersonID *primitive.ObjectID) (map[string]int64, error) {
	match := bson.M{}
	if salespersonID != nil {
		match["salespersonId"] = salespersonID
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
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
