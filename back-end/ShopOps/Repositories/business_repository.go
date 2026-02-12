package Repositories

import (
	"context"
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BusinessRepository struct {
	collection *mongo.Collection
}

func NewBusinessRepository(db *mongo.Database) Domain.BusinessRepository {
	return &BusinessRepository{
		collection: db.Collection("businesses"),
	}
}

func (r *BusinessRepository) Create(business *Domain.Business) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	business.CreatedAt = time.Now()
	business.UpdatedAt = time.Now()
	business.Status = Domain.BusinessStatusActive
	business.Timezone = "UTC"

	result, err := r.collection.InsertOne(ctx, business)
	if err != nil {
		return fmt.Errorf("failed to create business: %w", err)
	}

	business.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *BusinessRepository) FindByID(id string) (*Domain.Business, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	var business Domain.Business
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&business)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find business: %w", err)
	}

	return &business, nil
}

func (r *BusinessRepository) FindByUserID(userID string) ([]Domain.Business, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objUserID})
	if err != nil {
		return nil, fmt.Errorf("failed to find businesses: %w", err)
	}
	defer cursor.Close(ctx)

	var businesses []Domain.Business
	if err := cursor.All(ctx, &businesses); err != nil {
		return nil, fmt.Errorf("failed to decode businesses: %w", err)
	}

	return businesses, nil
}

func (r *BusinessRepository) Update(business *Domain.Business) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	business.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":          business.Name,
			"description":   business.Description,
			"business_type": business.BusinessType,
			"currency":      business.Currency,
			"timezone":      business.Timezone,
			"address":       business.Address,
			"city":          business.City,
			"country":       business.Country,
			"phone":         business.Phone,
			"email":         business.Email,
			"updated_at":    business.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateByID(ctx, business.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update business: %w", err)
	}

	return nil
}

func (r *BusinessRepository) UpdateStatus(id string, status Domain.BusinessStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid business ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("failed to update business status: %w", err)
	}

	return nil
}

func (r *BusinessRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid business ID: %w", err)
	}

	// Soft delete - update status to closed
	update := bson.M{
		"$set": bson.M{
			"status":     Domain.BusinessStatusClosed,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	return err
}

func (r *BusinessRepository) FindByPhone(phone string) (*Domain.Business, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var business Domain.Business
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&business)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find business: %w", err)
	}

	return &business, nil
}
