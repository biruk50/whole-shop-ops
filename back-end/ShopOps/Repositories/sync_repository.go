package Repositories

import (
	"context"
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SyncRepository struct {
	collection *mongo.Collection
}

func NewSyncRepository(db *mongo.Database) Domain.SyncRepository {
	return &SyncRepository{
		collection: db.Collection("sync_logs"),
	}
}

func (r *SyncRepository) LogSync(businessID, deviceID string, items []Domain.SyncItem, result Domain.SyncResponse) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return fmt.Errorf("invalid business ID: %w", err)
	}

	logEntry := bson.M{
		"business_id":   objBusinessID,
		"device_id":     deviceID,
		"items_count":   len(items),
		"success_count": len(result.Success),
		"failed_count":  len(result.Failed),
		"server_time":   result.ServerTime,
		"created_at":    time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, logEntry)
	return err
}

func (r *SyncRepository) GetSyncStatus(businessID string) (*Domain.SyncStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	// Get last sync time
	var lastSyncLog struct {
		ServerTime time.Time `bson:"server_time"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"business_id": objBusinessID,
	}, options.FindOne().SetSort(bson.M{"created_at": -1})).Decode(&lastSyncLog)

	var lastSync time.Time
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to get last sync: %w", err)
	}
	if err == nil {
		lastSync = lastSyncLog.ServerTime
	}

	// Get pending sync count (this would require checking other collections)
	// For simplicity, we'll return a placeholder status
	status := &Domain.SyncStatus{
		LastSync:    lastSync,
		Pending:     0, // Would be calculated from unsynced items
		SyncedToday: 0, // Would be calculated from today's syncs
		Total:       0, // Would be calculated from total syncs
	}

	return status, nil
}

func (r *SyncRepository) GetLastSync(businessID, deviceID string) (*time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	var lastSyncLog struct {
		ServerTime time.Time `bson:"server_time"`
	}

	err = r.collection.FindOne(ctx, bson.M{
		"business_id": objBusinessID,
		"device_id":   deviceID,
	}, options.FindOne().SetSort(bson.M{"created_at": -1})).Decode(&lastSyncLog)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last sync: %w", err)
	}

	return &lastSyncLog.ServerTime, nil
}
