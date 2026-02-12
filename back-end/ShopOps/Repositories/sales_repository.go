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

type SalesRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewSalesRepository(db *mongo.Database) Domain.SaleRepository {
	return &SalesRepository{
		collection: db.Collection("sales"),
		db:         db,
	}
}

func (r *SalesRepository) Create(sale *Domain.Sale) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Calculate totals
	sale.TotalAmount = sale.Quantity * sale.UnitPrice
	sale.FinalAmount = sale.TotalAmount - sale.Discount + sale.Tax

	sale.Status = Domain.SaleStatusCompleted
	sale.PaymentStatus = Domain.PaymentStatusPaid
	sale.CreatedAt = time.Now()
	sale.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, sale)
	if err != nil {
		return fmt.Errorf("failed to create sale: %w", err)
	}

	sale.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *SalesRepository) FindByID(id string) (*Domain.Sale, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid sale ID: %w", err)
	}

	var sale Domain.Sale
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&sale)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find sale: %w", err)
	}

	return &sale, nil
}

func (r *SalesRepository) FindByBusinessID(businessID string, filters Domain.SaleFilters) ([]Domain.Sale, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	query := bson.M{"business_id": objBusinessID}

	// Apply filters
	if filters.StartDate != nil && filters.EndDate != nil {
		query["created_at"] = bson.M{
			"$gte": *filters.StartDate,
			"$lte": *filters.EndDate,
		}
	} else if filters.StartDate != nil {
		query["created_at"] = bson.M{"$gte": *filters.StartDate}
	} else if filters.EndDate != nil {
		query["created_at"] = bson.M{"$lte": *filters.EndDate}
	}

	if filters.Status != nil {
		query["status"] = *filters.Status
	}

	if filters.PaymentMethod != nil {
		query["payment_method"] = *filters.PaymentMethod
	}

	if filters.PaymentStatus != nil {
		query["payment_status"] = *filters.PaymentStatus
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1})

	if filters.Limit > 0 {
		opts.SetLimit(int64(filters.Limit))
	}

	if filters.Offset > 0 {
		opts.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find sales: %w", err)
	}
	defer cursor.Close(ctx)

	var sales []Domain.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, fmt.Errorf("failed to decode sales: %w", err)
	}

	return sales, nil
}

func (r *SalesRepository) FindByLocalID(businessID, localID string) (*Domain.Sale, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	var sale Domain.Sale
	err = r.collection.FindOne(ctx, bson.M{
		"business_id": objBusinessID,
		"local_id":    localID,
	}).Decode(&sale)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find sale: %w", err)
	}

	return &sale, nil
}

func (r *SalesRepository) Update(sale *Domain.Sale) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sale.UpdatedAt = time.Now()
	sale.TotalAmount = sale.Quantity * sale.UnitPrice
	sale.FinalAmount = sale.TotalAmount - sale.Discount + sale.Tax

	update := bson.M{
		"$set": bson.M{
			"product_id":     sale.ProductID,
			"customer_name":  sale.CustomerName,
			"customer_phone": sale.CustomerPhone,
			"quantity":       sale.Quantity,
			"unit_price":     sale.UnitPrice,
			"total_amount":   sale.TotalAmount,
			"discount":       sale.Discount,
			"tax":            sale.Tax,
			"final_amount":   sale.FinalAmount,
			"payment_method": sale.PaymentMethod,
			"payment_status": sale.PaymentStatus,
			"notes":          sale.Notes,
			"status":         sale.Status,
			"updated_at":     sale.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateByID(ctx, sale.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}

	return nil
}

func (r *SalesRepository) UpdateStatus(id string, status Domain.SaleStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid sale ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("failed to update sale status: %w", err)
	}

	return nil
}

func (r *SalesRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid sale ID: %w", err)
	}

	// Soft delete - mark as voided
	update := bson.M{
		"$set": bson.M{
			"status":     Domain.SaleStatusVoided,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateByID(ctx, objID, update)
	return err
}

func (r *SalesRepository) GetSummary(businessID string, startDate, endDate time.Time) (*Domain.SaleSummary, error) {
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
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"status": Domain.SaleStatusCompleted,
			},
		},
		{
			"$group": bson.M{
				"_id":               nil,
				"total_sales":       bson.M{"$sum": "$quantity"},
				"total_amount":      bson.M{"$sum": "$final_amount"},
				"total_discount":    bson.M{"$sum": "$discount"},
				"total_tax":         bson.M{"$sum": "$tax"},
				"transaction_count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate sales: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalSales       float64 `bson:"total_sales"`
		TotalAmount      float64 `bson:"total_amount"`
		TotalDiscount    float64 `bson:"total_discount"`
		TotalTax         float64 `bson:"total_tax"`
		TransactionCount int     `bson:"transaction_count"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}
	}

	summary := &Domain.SaleSummary{
		Date:             startDate,
		TotalSales:       result.TotalSales,
		TotalAmount:      result.TotalAmount,
		TotalDiscount:    result.TotalDiscount,
		TotalTax:         result.TotalTax,
		TransactionCount: result.TransactionCount,
	}

	return summary, nil
}

func (r *SalesRepository) GetStats(businessID string, period string) (*Domain.SaleStats, error) {
	// Implementation for sales statistics
	// This would include daily averages, weekly/monthly totals, etc.
	return &Domain.SaleStats{
		DailyAverage:   0,
		WeeklyTotal:    0,
		MonthlyTotal:   0,
		BestSellingDay: "Monday",
	}, nil
}

func (r *SalesRepository) GetDailySales(businessID string, date time.Time) ([]Domain.Sale, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := bson.M{
		"business_id": objBusinessID,
		"created_at": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
		"status": Domain.SaleStatusCompleted,
	}

	cursor, err := r.collection.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find daily sales: %w", err)
	}
	defer cursor.Close(ctx)

	var sales []Domain.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, fmt.Errorf("failed to decode sales: %w", err)
	}

	return sales, nil
}
