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

type CallRepository struct {
	collection *mongo.Collection
}

func NewCallRepository(db *mongo.Database) *CallRepository {
	return &CallRepository{
		collection: db.Collection("call_attempts"),
	}
}

func (r *CallRepository) Create(ctx context.Context, call *models.CallAttempt) error {
	call.ID = primitive.NewObjectID()
	call.CreatedAt = time.Now()
	call.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, call)
	return err
}

func (r *CallRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.CallAttempt, error) {
	var call models.CallAttempt
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func (r *CallRepository) FindByLeadID(ctx context.Context, leadID primitive.ObjectID) ([]*models.CallAttempt, error) {
	opts := options.Find().SetSort(bson.D{{Key: "attemptNumber", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"leadId": leadID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var calls []*models.CallAttempt
	if err := cursor.All(ctx, &calls); err != nil {
		return nil, err
	}
	return calls, nil
}

func (r *CallRepository) CountByLeadID(ctx context.Context, leadID primitive.ObjectID) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"leadId": leadID})
	return int(count), err
}

func (r *CallRepository) Update(ctx context.Context, call *models.CallAttempt) error {
	call.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": call.ID},
		bson.M{"$set": call},
	)
	return err
}

func (r *CallRepository) FindRecentCalls(ctx context.Context, salespersonID *primitive.ObjectID, limit int) ([]*models.CallAttempt, error) {
	filter := bson.M{}
	if salespersonID != nil {
		filter["salespersonId"] = salespersonID
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var calls []*models.CallAttempt
	if err := cursor.All(ctx, &calls); err != nil {
		return nil, err
	}
	return calls, nil
}

func (r *CallRepository) CountCallsToday(ctx context.Context, salespersonID *primitive.ObjectID) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	query := bson.M{
		"createdAt": bson.M{"$gte": startOfDay},
	}
	if salespersonID != nil {
		query["salespersonId"] = salespersonID
	}

	return r.collection.CountDocuments(ctx, query)
}

func (r *CallRepository) GetCallOutcomeStats(ctx context.Context, salespersonID *primitive.ObjectID) (map[string]int64, error) {
	match := bson.M{}
	if salespersonID != nil {
		match["salespersonId"] = salespersonID
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$callOutcome"},
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
