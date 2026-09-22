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

type LeadFilter struct {
	AssignedTo     *primitive.ObjectID
	Status         models.LeadStatus
	Priority       models.Priority
	Source         string
	Campaign       string
	DoNotCall      *bool
	FollowUpStatus models.FollowUpStatus
	Search         string
	OnlyOneCall    bool
}

type LeadRepository struct {
	collection *mongo.Collection
}

func NewLeadRepository(db *mongo.Database) *LeadRepository {
	return &LeadRepository{
		collection: db.Collection("leads"),
	}
}

func (r *LeadRepository) Create(ctx context.Context, lead *models.Lead) error {
	lead.ID = primitive.NewObjectID()
	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()
	if lead.Status == "" {
		lead.Status = models.LeadStatusNew
	}
	if lead.Priority == "" {
		lead.Priority = models.PriorityMedium
	}
	if lead.FollowUpStatus == "" {
		lead.FollowUpStatus = models.FollowUpStatusNone
	}
	_, err := r.collection.InsertOne(ctx, lead)
	return err
}

func (r *LeadRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Lead, error) {
	var lead models.Lead
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&lead)
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) FindByPhoneOrEmail(ctx context.Context, phone, email string) (*models.Lead, error) {
	orConditions := bson.A{}
	if phone != "" {
		orConditions = append(orConditions, bson.M{"phone": phone})
		orConditions = append(orConditions, bson.M{"alternatePhone": phone})
	}
	if email != "" {
		orConditions = append(orConditions, bson.M{"email": email})
	}

	if len(orConditions) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	var lead models.Lead
	err := r.collection.FindOne(ctx, bson.M{"$or": orConditions}).Decode(&lead)
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *LeadRepository) FindLeads(ctx context.Context, filter LeadFilter, page, pageSize int) ([]*models.Lead, int64, error) {
	query := bson.M{}

	if filter.AssignedTo != nil {
		query["assignedTo"] = filter.AssignedTo
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.Priority != "" {
		query["priority"] = filter.Priority
	}
	if filter.Source != "" {
		query["source"] = filter.Source
	}
	if filter.Campaign != "" {
		query["campaign"] = filter.Campaign
	}
	if filter.DoNotCall != nil {
		query["doNotCall"] = *filter.DoNotCall
	}
	if filter.FollowUpStatus != "" {
		query["followUpStatus"] = filter.FollowUpStatus
	}
	if filter.OnlyOneCall {
		query["callAttemptCount"] = 1
		query["doNotCall"] = false
		query["status"] = bson.M{
			"$nin": []models.LeadStatus{
				models.LeadStatusDoNotCall,
				models.LeadStatusNotInterested,
				models.LeadStatusConverted,
				models.LeadStatusLost,
				models.LeadStatusDisqualified,
			},
		}
	}

	if filter.Search != "" {
		searchRegex := bson.M{"$regex": filter.Search, "$options": "i"}
		query["$or"] = bson.A{
			bson.M{"name": searchRegex},
			bson.M{"phone": searchRegex},
			bson.M{"email": searchRegex},
			bson.M{"course": searchRegex},
		}
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

	var leads []*models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, err
	}

	return leads, totalCount, nil
}

// FindOneCallFollowUps finds leads called exactly once needing follow up
func (r *LeadRepository) FindOneCallFollowUps(ctx context.Context, page, pageSize int) ([]*models.Lead, int64, error) {
	query := bson.M{
		"callAttemptCount": 1,
		"doNotCall":        false,
		"status": bson.M{
			"$in": []models.LeadStatus{
				models.LeadStatusCalledOnce,
				models.LeadStatusFollowUpRequired,
				models.LeadStatusFollowUpScheduled,
			},
		},
	}

	totalCount, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var leads []*models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, err
	}

	return leads, totalCount, nil
}

// FindDoNotCallLeads returns all leads classified as DO_NOT_CALL or NOT_INTERESTED
func (r *LeadRepository) FindDoNotCallLeads(ctx context.Context, page, pageSize int) ([]*models.Lead, int64, error) {
	query := bson.M{
		"$or": bson.A{
			bson.M{"doNotCall": true},
			bson.M{"status": models.LeadStatusDoNotCall},
			bson.M{"status": models.LeadStatusNotInterested},
		},
	}

	totalCount, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var leads []*models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, err
	}

	return leads, totalCount, nil
}

// FindEligibleForOneCallAlert finds 1-call and 2-call unattended leads that have exceeded threshold or whose callback is overdue
func (r *LeadRepository) FindEligibleForOneCallAlert(ctx context.Context, cutoffTime time.Time, now time.Time) ([]*models.Lead, error) {
	query := bson.M{
		"callAttemptCount": bson.M{"$in": []int{1, 2}},
		"doNotCall":        false,
		"status": bson.M{
			"$nin": []models.LeadStatus{
				models.LeadStatusDoNotCall,
				models.LeadStatusNotInterested,
				models.LeadStatusConverted,
				models.LeadStatusLost,
				models.LeadStatusDisqualified,
			},
		},
		"$or": bson.A{
			// Situation A: 1st or 2nd call happened and threshold time has passed without scheduled follow-up
			bson.M{
				"$or": bson.A{
					bson.M{"lastCallAt": bson.M{"$lte": cutoffTime}},
					bson.M{"firstCallAt": bson.M{"$lte": cutoffTime}},
				},
				"nextFollowUpAt": bson.M{"$exists": false},
			},
			bson.M{
				"$or": bson.A{
					bson.M{"lastCallAt": bson.M{"$lte": cutoffTime}},
					bson.M{"firstCallAt": bson.M{"$lte": cutoffTime}},
				},
				"nextFollowUpAt": nil,
			},
			// Situation B: Follow-up/call-back was scheduled, but the scheduled time is now in the past
			bson.M{
				"nextFollowUpAt": bson.M{"$ne": nil, "$lte": now},
			},
		},
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var leads []*models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, err
	}
	return leads, nil
}

func (r *LeadRepository) Update(ctx context.Context, lead *models.Lead) error {
	lead.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": lead.ID},
		bson.M{"$set": lead},
	)
	return err
}

func (r *LeadRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.LeadStatus) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

func (r *LeadRepository) SetDoNotCall(ctx context.Context, id primitive.ObjectID, doNotCall bool, reason string, detectedBy string) error {
	now := time.Now()
	update := bson.M{
		"doNotCall":           doNotCall,
		"doNotCallReason":     reason,
		"doNotCallDetectedBy": detectedBy,
		"updatedAt":           now,
	}
	if doNotCall {
		update["doNotCallAt"] = &now
		update["status"] = models.LeadStatusDoNotCall
		update["followUpStatus"] = models.FollowUpStatusNone
		update["nextFollowUpAt"] = nil
	} else {
		update["status"] = models.LeadStatusContacted
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	return err
}

func (r *LeadRepository) CountByStatus(ctx context.Context, assignedTo *primitive.ObjectID) (map[string]int64, error) {
	match := bson.M{}
	if assignedTo != nil {
		match["assignedTo"] = assignedTo
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

	results := make(map[string]int64)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&item); err == nil {
			results[item.ID] = item.Count
		}
	}

	return results, nil
}

func (r *LeadRepository) TotalCount(ctx context.Context, assignedTo *primitive.ObjectID) (int64, error) {
	query := bson.M{}
	if assignedTo != nil {
		query["assignedTo"] = assignedTo
	}
	return r.collection.CountDocuments(ctx, query)
}

// FindBdaCallQueue returns leads for the BDA calling workstation
func (r *LeadRepository) FindBdaCallQueue(ctx context.Context, assignedTo *primitive.ObjectID, category string, search string, page, pageSize int) ([]*models.Lead, int64, error) {
	query := bson.M{
		"doNotCall": false,
		"status": bson.M{
			"$nin": []models.LeadStatus{
				models.LeadStatusDoNotCall,
				models.LeadStatusNotInterested,
				models.LeadStatusLost,
				models.LeadStatusDisqualified,
			},
		},
	}

	if assignedTo != nil {
		assignedCond := bson.M{
			"$or": bson.A{
				bson.M{"assignedTo": assignedTo},
				bson.M{"assignedTo": nil},
				bson.M{"assignedTo": bson.M{"$exists": false}},
			},
		}
		if search != "" {
			searchRegex := bson.M{"$regex": search, "$options": "i"}
			searchCond := bson.M{
				"$or": bson.A{
					bson.M{"name": searchRegex},
					bson.M{"phone": searchRegex},
					bson.M{"course": searchRegex},
				},
			}
			query["$and"] = bson.A{assignedCond, searchCond}
		} else {
			query["$or"] = assignedCond["$or"]
		}
	} else if search != "" {
		searchRegex := bson.M{"$regex": search, "$options": "i"}
		query["$or"] = bson.A{
			bson.M{"name": searchRegex},
			bson.M{"phone": searchRegex},
			bson.M{"course": searchRegex},
		}
	}


	switch category {
	case "unattended_1":
		query["callAttemptCount"] = 1
		query["status"] = bson.M{"$nin": []models.LeadStatus{
			models.LeadStatusInterested,
			models.LeadStatusConverted,
			models.LeadStatusFollowUpScheduled,
		}}
	case "unattended_2":
		query["callAttemptCount"] = 2
		query["status"] = bson.M{"$nin": []models.LeadStatus{
			models.LeadStatusInterested,
			models.LeadStatusConverted,
			models.LeadStatusFollowUpScheduled,
		}}
	case "callback_later":
		query["status"] = models.LeadStatusFollowUpScheduled

	case "fresh":
		query["callAttemptCount"] = 0
	case "interested":
		query["status"] = bson.M{"$in": []models.LeadStatus{
			models.LeadStatusInterested,
			models.LeadStatusConverted,
		}}
	default: // "all" action items
		query["status"] = bson.M{
			"$nin": []models.LeadStatus{
				models.LeadStatusDoNotCall,
				models.LeadStatusNotInterested,
				models.LeadStatusLost,
				models.LeadStatusDisqualified,
				models.LeadStatusConverted,
			},
		}
	}

	totalCount, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{
		{Key: "updatedAt", Value: -1},
	})

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var leads []*models.Lead
	if err := cursor.All(ctx, &leads); err != nil {
		return nil, 0, err
	}

	return leads, totalCount, nil
}

// CountBdaQueueCategories returns counts for each BDA queue category
func (r *LeadRepository) CountBdaQueueCategories(ctx context.Context, assignedTo *primitive.ObjectID) (map[string]int64, error) {
	counts := make(map[string]int64)
	categories := []string{"all", "unattended_1", "unattended_2", "callback_later", "fresh", "interested"}
	for _, cat := range categories {
		_, count, err := r.FindBdaCallQueue(ctx, assignedTo, cat, "", 1, 1)
		if err == nil {
			counts[cat] = count
		}
	}
	return counts, nil
}

