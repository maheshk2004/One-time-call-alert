package repositories

import (
	"context"
	"time"

	"lead-followup-system/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TranscriptRepository struct {
	collection *mongo.Collection
}

func NewTranscriptRepository(db *mongo.Database) *TranscriptRepository {
	return &TranscriptRepository{
		collection: db.Collection("transcripts"),
	}
}

func (r *TranscriptRepository) Create(ctx context.Context, transcript *models.Transcript) error {
	transcript.ID = primitive.NewObjectID()
	transcript.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, transcript)
	return err
}

func (r *TranscriptRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Transcript, error) {
	var transcript models.Transcript
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transcript)
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}

func (r *TranscriptRepository) FindByCallAttemptID(ctx context.Context, callAttemptID primitive.ObjectID) (*models.Transcript, error) {
	var transcript models.Transcript
	err := r.collection.FindOne(ctx, bson.M{"callAttemptId": callAttemptID}).Decode(&transcript)
	if err != nil {
		return nil, err
	}
	return &transcript, nil
}

func (r *TranscriptRepository) FindByLeadID(ctx context.Context, leadID primitive.ObjectID) ([]*models.Transcript, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"leadId": leadID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transcripts []*models.Transcript
	if err := cursor.All(ctx, &transcripts); err != nil {
		return nil, err
	}
	return transcripts, nil
}
