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

type ReportRepository struct {
	db *mongo.Database
}

func NewReportRepository(db *mongo.Database) Domain.ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GenerateSalesReport(businessID string, startDate, endDate time.Time) (*Domain.SalesReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	salesCollection := r.db.Collection("sales")

	// Get total sales data
	totalPipeline := []bson.M{
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
				"_id":                nil,
				"total_sales":        bson.M{"$sum": "$quantity"},
				"total_amount":       bson.M{"$sum": "$final_amount"},
				"total_transactions": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := salesCollection.Aggregate(ctx, totalPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate sales: %w", err)
	}
	defer cursor.Close(ctx)

	var totalResult struct {
		TotalSales        float64 `bson:"total_sales"`
		TotalAmount       float64 `bson:"total_amount"`
		TotalTransactions int     `bson:"total_transactions"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&totalResult); err != nil {
			return nil, fmt.Errorf("failed to decode total result: %w", err)
		}
	}

	// Get top products
	productsPipeline := []bson.M{
		{
			"$match": bson.M{
				"business_id": objBusinessID,
				"created_at": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
				"status":     Domain.SaleStatusCompleted,
				"product_id": bson.M{"$ne": nil},
			},
		},
		{
			"$group": bson.M{
				"_id":          "$product_id",
				"quantity":     bson.M{"$sum": "$quantity"},
				"total_amount": bson.M{"$sum": "$final_amount"},
			},
		},
		{
			"$sort": bson.M{"total_amount": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err = salesCollection.Aggregate(ctx, productsPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate top products: %w", err)
	}
	defer cursor.Close(ctx)

	var topProducts []Domain.TopProduct
	for cursor.Next(ctx) {
		var result struct {
			ProductID   primitive.ObjectID `bson:"_id"`
			Quantity    float64            `bson:"quantity"`
			TotalAmount float64            `bson:"total_amount"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		// Get product name (simplified - in production, join with products collection)
		topProducts = append(topProducts, Domain.TopProduct{
			ProductID:   result.ProductID.Hex(),
			ProductName: "Product", // Would fetch from products collection
			Quantity:    result.Quantity,
			TotalAmount: result.TotalAmount,
		})
	}

	report := &Domain.SalesReport{
		Period:            fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalSales:        totalResult.TotalSales,
		TotalAmount:       totalResult.TotalAmount,
		TotalTransactions: totalResult.TotalTransactions,
		AverageSale:       0,
		TopProducts:       topProducts,
	}

	if totalResult.TotalTransactions > 0 {
		report.AverageSale = totalResult.TotalAmount / float64(totalResult.TotalTransactions)
	}

	return report, nil
}

func (r *ReportRepository) GenerateExpensesReport(businessID string, startDate, endDate time.Time) (*Domain.ExpensesReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	expensesCollection := r.db.Collection("expenses")

	// Get category breakdown
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
			"$sort": bson.M{"total_amount": -1},
		},
	}

	cursor, err := expensesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate expenses: %w", err)
	}
	defer cursor.Close(ctx)

	var categoryBreakdown []Domain.CategoryExpense
	var totalAmount float64

	for cursor.Next(ctx) {
		var result struct {
			Category    Domain.ExpenseCategory `bson:"_id"`
			TotalAmount float64                `bson:"total_amount"`
			Count       int                    `bson:"count"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		totalAmount += result.TotalAmount
		categoryBreakdown = append(categoryBreakdown, Domain.CategoryExpense{
			Category:    result.Category,
			TotalAmount: result.TotalAmount,
			Count:       result.Count,
			Percentage:  0, // Will calculate later
		})
	}

	// Calculate percentages
	for i := range categoryBreakdown {
		if totalAmount > 0 {
			categoryBreakdown[i].Percentage = (categoryBreakdown[i].TotalAmount / totalAmount) * 100
		}
	}

	report := &Domain.ExpensesReport{
		Period:            fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalExpenses:     totalAmount,
		CategoryBreakdown: categoryBreakdown,
	}

	return report, nil
}

func (r *ReportRepository) GenerateProfitReport(businessID string, startDate, endDate time.Time) (*Domain.ProfitReport, error) {
	// Get sales total
	salesReport, err := r.GenerateSalesReport(businessID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales data: %w", err)
	}

	// Get expenses total
	expensesReport, err := r.GenerateExpensesReport(businessID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses data: %w", err)
	}

	grossProfit := salesReport.TotalAmount - expensesReport.TotalExpenses
	profitMargin := 0.0
	if salesReport.TotalAmount > 0 {
		profitMargin = (grossProfit / salesReport.TotalAmount) * 100
	}

	report := &Domain.ProfitReport{
		Period:        fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalSales:    salesReport.TotalAmount,
		TotalExpenses: expensesReport.TotalExpenses,
		GrossProfit:   grossProfit,
		NetProfit:     grossProfit, // Would deduct taxes, fees, etc.
		ProfitMargin:  profitMargin,
	}

	return report, nil
}

func (r *ReportRepository) GenerateInventoryReport(businessID string) (*Domain.InventoryReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	productsCollection := r.db.Collection("products")

	// Get all active products
	cursor, err := productsCollection.Find(ctx, bson.M{
		"business_id": objBusinessID,
		"status":      Domain.ProductStatusActive,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []Domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	var totalProducts int
	var totalStock float64
	var totalValue float64
	var lowStockItems []Domain.LowStockItem

	for _, product := range products {
		totalProducts++
		totalStock += product.Stock
		totalValue += product.Stock * product.CostPrice

		if product.MinStock > 0 && product.Stock < product.MinStock {
			lowStockItems = append(lowStockItems, Domain.LowStockItem{
				ProductID:   product.ID.Hex(),
				ProductName: product.Name,
				Current:     product.Stock,
				Minimum:     product.MinStock,
				Difference:  product.MinStock - product.Stock,
			})
		}
	}

	report := &Domain.InventoryReport{
		TotalProducts: totalProducts,
		TotalStock:    totalStock,
		TotalValue:    totalValue,
		LowStockItems: lowStockItems,
	}

	return report, nil
}

func (r *ReportRepository) GetDashboardData(businessID string) (*Domain.DashboardData, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Today's sales
	todaySales, _ := r.getSalesTotal(businessID, today, today.Add(24*time.Hour))

	// Today's expenses
	todayExpenses, _ := r.getExpensesTotal(businessID, today, today.Add(24*time.Hour))

	// Week's data (last 7 days)
	weekStart := today.AddDate(0, 0, -7)
	weekSales, _ := r.getSalesTotal(businessID, weekStart, today)
	weekExpenses, _ := r.getExpensesTotal(businessID, weekStart, today)

	// Month's data (last 30 days)
	monthStart := today.AddDate(0, 0, -30)
	monthSales, _ := r.getSalesTotal(businessID, monthStart, today)
	monthExpenses, _ := r.getExpensesTotal(businessID, monthStart, today)

	// Low stock count
	lowStockCount, _ := r.getLowStockCount(businessID)

	data := &Domain.DashboardData{
		TodaySales:      todaySales,
		TodayExpenses:   todayExpenses,
		TodayProfit:     todaySales - todayExpenses,
		WeekSales:       weekSales,
		WeekExpenses:    weekExpenses,
		WeekProfit:      weekSales - weekExpenses,
		MonthSales:      monthSales,
		MonthExpenses:   monthExpenses,
		MonthProfit:     monthSales - monthExpenses,
		LowStockCount:   lowStockCount,
		PendingPayments: 0, // Would calculate from sales with pending status
	}

	return data, nil
}

func (r *ReportRepository) getSalesTotal(businessID string, startDate, endDate time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return 0, err
	}

	salesCollection := r.db.Collection("sales")

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
				"_id":   nil,
				"total": bson.M{"$sum": "$final_amount"},
			},
		},
	}

	cursor, err := salesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}

	return 0, nil
}

func (r *ReportRepository) getExpensesTotal(businessID string, startDate, endDate time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return 0, err
	}

	expensesCollection := r.db.Collection("expenses")

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

	cursor, err := expensesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}

	return 0, nil
}

func (r *ReportRepository) getLowStockCount(businessID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return 0, err
	}

	productsCollection := r.db.Collection("products")

	count, err := productsCollection.CountDocuments(ctx, bson.M{
		"business_id": objBusinessID,
		"status":      Domain.ProductStatusActive,
		"$expr":       bson.M{"$lt": []interface{}{"$stock", "$min_stock"}},
	})

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *ReportRepository) ExportCSV(report interface{}, reportType Domain.ReportType) ([]byte, error) {
	// Implementation would convert report to CSV
	// For now, return empty
	return []byte{}, nil
}
