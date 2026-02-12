package Infrastructure

import (
	"context"
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SyncService interface {
	ProcessBatch(batch Domain.SyncBatch) (*Domain.SyncResponse, error)
	GetSyncStatus(businessID string) (*Domain.SyncStatus, error)
}

type syncService struct {
	db          *mongo.Database
	salesRepo   Domain.SaleRepository
	expenseRepo Domain.ExpenseRepository
	productRepo Domain.ProductRepository
	syncRepo    Domain.SyncRepository
}

func NewSyncService(
	db *mongo.Database,
	salesRepo Domain.SaleRepository,
	expenseRepo Domain.ExpenseRepository,
	productRepo Domain.ProductRepository,
	syncRepo Domain.SyncRepository,
) SyncService {
	return &syncService{
		db:          db,
		salesRepo:   salesRepo,
		expenseRepo: expenseRepo,
		productRepo: productRepo,
		syncRepo:    syncRepo,
	}
}

func (s *syncService) ProcessBatch(batch Domain.SyncBatch) (*Domain.SyncResponse, error) {
	response := &Domain.SyncResponse{
		Success:    []Domain.SyncResult{},
		Failed:     []Domain.SyncResult{},
		ServerTime: time.Now(),
	}

	businessObjID, err := primitive.ObjectIDFromHex(batch.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	for _, item := range batch.Items {
		result := Domain.SyncResult{
			LocalID:   item.LocalID,
			Timestamp: time.Now(),
		}

		// Check if item already exists
		existing, err := s.findExistingItem(businessObjID, item.EntityType, item.LocalID)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("check failed: %v", err)
			response.Failed = append(response.Failed, result)
			continue
		}

		// Process based on operation
		switch item.Operation {
		case Domain.SyncOperationCreate:
			if existing != nil {
				// Already exists, skip
				result.Success = true
				result.ServerID = existing.(primitive.ObjectID).Hex()
				response.Success = append(response.Success, result)
				continue
			}

			serverID, err := s.createItem(batch.BusinessID, item.EntityType, item.Data)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("create failed: %v", err)
				response.Failed = append(response.Failed, result)
			} else {
				result.Success = true
				result.ServerID = serverID
				response.Success = append(response.Success, result)
			}

		case Domain.SyncOperationUpdate:
			if existing == nil {
				result.Success = false
				result.Error = "item not found for update"
				response.Failed = append(response.Failed, result)
				continue
			}

			err := s.updateItem(existing.(primitive.ObjectID).Hex(), item.EntityType, item.Data)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("update failed: %v", err)
				response.Failed = append(response.Failed, result)
			} else {
				result.Success = true
				result.ServerID = existing.(primitive.ObjectID).Hex()
				response.Success = append(response.Success, result)
			}

		case Domain.SyncOperationDelete:
			if existing == nil {
				// Already deleted, consider success
				result.Success = true
				response.Success = append(response.Success, result)
				continue
			}

			err := s.deleteItem(existing.(primitive.ObjectID).Hex(), item.EntityType)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("delete failed: %v", err)
				response.Failed = append(response.Failed, result)
			} else {
				result.Success = true
				response.Success = append(response.Success, result)
			}
		}
	}

	// Log sync result
	go s.syncRepo.LogSync(batch.BusinessID, batch.DeviceID, batch.Items, *response)

	return response, nil
}

func (s *syncService) findExistingItem(businessID primitive.ObjectID, entityType, localID string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.Collection(fmt.Sprintf("%ss", entityType))

	filter := bson.M{
		"business_id": businessID,
		"local_id":    localID,
	}

	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return result["_id"], nil
}

func (s *syncService) createItem(businessID, entityType string, data interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.Collection(fmt.Sprintf("%ss", entityType))

	// Convert data to BSON
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return "", err
	}

	var doc bson.M
	if err := bson.Unmarshal(bsonData, &doc); err != nil {
		return "", err
	}

	// Add business ID and timestamps
	businessObjID, _ := primitive.ObjectIDFromHex(businessID)
	doc["business_id"] = businessObjID
	doc["synced"] = true
	doc["synced_at"] = time.Now()
	doc["created_at"] = time.Now()
	doc["updated_at"] = time.Now()

	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (s *syncService) updateItem(id, entityType string, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.Collection(fmt.Sprintf("%ss", entityType))

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Convert data to BSON
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	var updateDoc bson.M
	if err := bson.Unmarshal(bsonData, &updateDoc); err != nil {
		return err
	}

	// Add update timestamp
	updateDoc["updated_at"] = time.Now()
	updateDoc["synced"] = true
	updateDoc["synced_at"] = time.Now()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updateDoc}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *syncService) deleteItem(id, entityType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.Collection(fmt.Sprintf("%ss", entityType))

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}

	// Soft delete by updating status based on entity type
	var status string
	switch entityType {
	case "sale":
		status = "voided"
	case "expense":
		status = "voided"
	case "product":
		status = "deleted"
	default:
		status = "deleted"
	}

	update := bson.M{"$set": bson.M{
		"status":     status,
		"synced":     true,
		"synced_at":  time.Now(),
		"updated_at": time.Now(),
	}}

	_, err = collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *syncService) GetSyncStatus(businessID string) (*Domain.SyncStatus, error) {
	// Delegate to sync repository
	return s.syncRepo.GetSyncStatus(businessID)
}
