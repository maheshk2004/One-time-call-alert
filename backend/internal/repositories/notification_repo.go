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

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) Create(ctx context.Context, notif *models.Notification) error {
	notif.ID = primitive.NewObjectID()
	notif.CreatedAt = time.Now()
	notif.Read = false
	_, err := r.collection.InsertOne(ctx, notif)
	return err
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, unreadOnly bool, limit int) ([]*models.Notification, error) {
	query := bson.M{"userId": userID}
	if unreadOnly {
		query["read"] = false
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifs []*models.Notification
	if err := cursor.All(ctx, &notifs); err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "userId": userID},
		bson.M{"$set": bson.M{"read": true, "readAt": &now}},
	)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userID, "read": false},
		bson.M{"$set": bson.M{"read": true, "readAt": &now}},
	)
	return err
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"userId": userID, "read": false})
}
