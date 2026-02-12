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

type ExpenseRepository struct {
	collection *mongo.Collection
}

func NewExpenseRepository(db *mongo.Database) Domain.ExpenseRepository {
	return &ExpenseRepository{
		collection: db.Collection("expenses"),
	}
}

func (r *ExpenseRepository) Create(expense *Domain.Expense) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}
	expense.Status = Domain.ExpenseStatusActive
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, expense)
	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	expense.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ExpenseRepository) FindByID(id string) (*Domain.Expense, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}

	var expense Domain.Expense
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&expense)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find expense: %w", err)
	}

	return &expense, nil
}

func (r *ExpenseRepository) FindByBusinessID(businessID string, filters Domain.ExpenseFilters) ([]Domain.Expense, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	query := bson.M{"business_id": objBusinessID}

	if filters.StartDate != nil && filters.EndDate != nil {
		query["date"] = bson.M{
			"$gte": *filters.StartDate,
			"$lte": *filters.EndDate,
		}
	} else if filters.StartDate != nil {
		query["date"] = bson.M{"$gte": *filters.StartDate}
	} else if filters.EndDate != nil {
		query["date"] = bson.M{"$lte": *filters.EndDate}
	}

	if filters.Category != nil {
		query["category"] = *filters.Category
	}

	if filters.Status != nil {
		query["status"] = *filters.Status
	}

	opts := options.Find().SetSort(bson.M{"date": -1})

	if filters.Limit > 0 {
		opts.SetLimit(int64(filters.Limit))
	}

	if filters.Offset > 0 {
		opts.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find expenses: %w", err)
	}
	defer cursor.Close(ctx)

	var expenses []Domain.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, fmt.Errorf("failed to decode expenses: %w", err)
	}

	return expenses, nil
}

func (r *ExpenseRepository) FindByLocalID(businessID, localID string) (*Domain.Expense, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	var expense Domain.Expense
	err = r.collection.FindOne(ctx, bson.M{
		"business_id": objBusinessID,
		"local_id":    localID,
	}).Decode(&expense)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find expense: %w", err)
	}

	return &expense, nil
}

func (r *ExpenseRepository) Update(expense *Domain.Expense) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	expense.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"category":    expense.Category,
			"amount":      expense.Amount,
			"description": expense.Description,
			"receipt_url": expense.ReceiptURL,
			"date":        expense.Date,
			"status":      expense.Status,
			"updated_at":  expense.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateByID(ctx, expense.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}

	return nil
}

func (r *ExpenseRepository) UpdateStatus(id string, status Domain.ExpenseStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid expense ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("failed to update expense status: %w", err)
	}

	return nil
}

func (r *ExpenseRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid expense ID: %w", err)
	}

	// Soft delete - mark as deleted
	update := bson.M{
		"$set": bson.M{
			"status":     Domain.ExpenseStatusDeleted,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	return err
}

func (r *ExpenseRepository) GetSummaryByCategory(businessID string, startDate, endDate time.Time) ([]Domain.ExpenseSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"business_id": objBusinessID,
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"status": Domain.ExpenseStatusActive,
			},
		},
		{
			"$group": bson.M{
				"_id":          "$category",
				"total_amount": bson.M{"$sum": "$amount"},
				"count":        bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"category":     "$_id",
				"total_amount": 1,
				"count":        1,
				"_id":          0,
			},
		},
		{
			"$sort": bson.M{"total_amount": -1},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate expenses: %w", err)
	}
	defer cursor.Close(ctx)

	var summaries []Domain.ExpenseSummary
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("failed to decode summaries: %w", err)
	}

	// Calculate total for percentages
	var totalAmount float64
	for _, summary := range summaries {
		totalAmount += summary.TotalAmount
	}

	// Add percentages
	for i := range summaries {
		if totalAmount > 0 {
			summaries[i].Percentage = (summaries[i].TotalAmount / totalAmount) * 100
		}
	}

	return summaries, nil
}

func (r *ExpenseRepository) GetTotal(businessID string, startDate, endDate time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return 0, fmt.Errorf("invalid business ID: %w", err)
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"business_id": objBusinessID,
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"status": Domain.ExpenseStatusActive,
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to aggregate total: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, fmt.Errorf("failed to decode result: %w", err)
		}
		return result.Total, nil
	}

	return 0, nil
}
