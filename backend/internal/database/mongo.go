package database

import (
	"context"
	"log"
	"time"

	"lead-followup-system/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongo(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.MongoDB)
	log.Printf("Successfully connected to MongoDB database: %s", cfg.MongoDB)

	mongoDB := &MongoDB{
		Client:   client,
		Database: db,
	}

	if err := mongoDB.CreateIndexes(context.Background()); err != nil {
		log.Printf("Warning: Failed to create some MongoDB indexes: %v", err)
	}

	return mongoDB, nil
}

func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Users Indexes
	usersColl := m.Database.Collection("users")
	_, err := usersColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Index creation warning (users.email): %v", err)
	}

	// Leads Indexes
	leadsColl := m.Database.Collection("leads")
	leadIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "phone", Value: 1}}},
		{Keys: bson.D{{Key: "email", Value: 1}}},
		{Keys: bson.D{{Key: "assignedTo", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "callAttemptCount", Value: 1}, {Key: "doNotCall", Value: 1}, {Key: "status", Value: 1}, {Key: "nextFollowUpAt", Value: 1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "doNotCall", Value: 1}}},
	}
	_, err = leadsColl.Indexes().CreateMany(ctx, leadIndexes)
	if err != nil {
		log.Printf("Index creation warning (leads): %v", err)
	}

	// Call Attempts Indexes
	callsColl := m.Database.Collection("call_attempts")
	callIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "leadId", Value: 1}, {Key: "attemptNumber", Value: 1}}},
		{Keys: bson.D{{Key: "salespersonId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "callOutcome", Value: 1}}},
	}
	_, err = callsColl.Indexes().CreateMany(ctx, callIndexes)
	if err != nil {
		log.Printf("Index creation warning (call_attempts): %v", err)
	}

	// Transcripts Indexes
	transcriptsColl := m.Database.Collection("transcripts")
	transcriptIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "callAttemptId", Value: 1}}},
		{Keys: bson.D{{Key: "leadId", Value: 1}}},
	}
	_, err = transcriptsColl.Indexes().CreateMany(ctx, transcriptIndexes)
	if err != nil {
		log.Printf("Index creation warning (transcripts): %v", err)
	}

	// AI Analyses Indexes
	aiColl := m.Database.Collection("ai_analyses")
	aiIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "callAttemptId", Value: 1}}},
		{Keys: bson.D{{Key: "leadId", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "confidence", Value: 1}}},
		{Keys: bson.D{{Key: "intent", Value: 1}}},
	}
	_, err = aiColl.Indexes().CreateMany(ctx, aiIndexes)
	if err != nil {
		log.Printf("Index creation warning (ai_analyses): %v", err)
	}

	// Follow-ups Indexes
	followUpsColl := m.Database.Collection("follow_ups")
	followUpIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "leadId", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "salespersonId", Value: 1}, {Key: "dueAt", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "dueAt", Value: 1}}},
	}
	_, err = followUpsColl.Indexes().CreateMany(ctx, followUpIndexes)
	if err != nil {
		log.Printf("Index creation warning (follow_ups): %v", err)
	}

	// Notifications Indexes
	notifColl := m.Database.Collection("notifications")
	notifIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "read", Value: 1}, {Key: "createdAt", Value: -1}}},
	}
	_, err = notifColl.Indexes().CreateMany(ctx, notifIndexes)
	if err != nil {
		log.Printf("Index creation warning (notifications): %v", err)
	}

	// Audit Logs Indexes
	auditColl := m.Database.Collection("audit_logs")
	auditIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "entityType", Value: 1}, {Key: "entityId", Value: 1}}},
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "createdAt", Value: -1}}},
	}
	_, err = auditColl.Indexes().CreateMany(ctx, auditIndexes)
	if err != nil {
		log.Printf("Index creation warning (audit_logs): %v", err)
	}

	log.Println("MongoDB indexes successfully created.")
	return nil
}
