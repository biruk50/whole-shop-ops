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

type InventoryRepository struct {
	productsCollection  *mongo.Collection
	movementsCollection *mongo.Collection
}

func NewInventoryRepository(db *mongo.Database) Domain.ProductRepository {
	return &InventoryRepository{
		productsCollection:  db.Collection("products"),
		movementsCollection: db.Collection("stock_movements"),
	}
}

func (r *InventoryRepository) Create(product *Domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	product.Status = Domain.ProductStatusActive
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	result, err := r.productsCollection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	product.ID = result.InsertedID.(primitive.ObjectID)

	// Create initial stock movement if stock > 0
	if product.Stock > 0 {
		movement := Domain.StockMovement{
			BusinessID: product.BusinessID,
			ProductID:  product.ID,
			Type:       Domain.MovementTypePurchase,
			Quantity:   product.Stock,
			Previous:   0,
			New:        product.Stock,
			Reason:     "Initial stock",
			CreatedBy:  product.CreatedBy,
			CreatedAt:  time.Now(),
		}

		_, err = r.movementsCollection.InsertOne(ctx, movement)
		if err != nil {
			// Log error but don't fail product creation
			fmt.Printf("Failed to create stock movement: %v\n", err)
		}
	}

	return nil
}

func (r *InventoryRepository) FindByID(id string) (*Domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	var product Domain.Product
	err = r.productsCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	return &product, nil
}

func (r *InventoryRepository) FindByBusinessID(businessID string, filters Domain.ProductFilters) ([]Domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	query := bson.M{"business_id": objBusinessID}

	if filters.Category != nil {
		query["category"] = *filters.Category
	}

	if filters.Status != nil {
		query["status"] = *filters.Status
	}

	if filters.Search != nil {
		query["$or"] = []bson.M{
			{"name": bson.M{"$regex": *filters.Search, "$options": "i"}},
			{"sku": bson.M{"$regex": *filters.Search, "$options": "i"}},
			{"barcode": bson.M{"$regex": *filters.Search, "$options": "i"}},
		}
	}

	if filters.LowStock != nil && *filters.LowStock {
		query["$expr"] = bson.M{"$lt": []interface{}{"$stock", "$min_stock"}}
	}

	opts := options.Find().SetSort(bson.M{"name": 1})

	if filters.Limit > 0 {
		opts.SetLimit(int64(filters.Limit))
	}

	if filters.Offset > 0 {
		opts.SetSkip(int64(filters.Offset))
	}

	cursor, err := r.productsCollection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []Domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

func (r *InventoryRepository) Update(product *Domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	product.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":          product.Name,
			"description":   product.Description,
			"sku":           product.SKU,
			"barcode":       product.Barcode,
			"category":      product.Category,
			"unit":          product.Unit,
			"cost_price":    product.CostPrice,
			"selling_price": product.SellingPrice,
			"stock":         product.Stock,
			"min_stock":     product.MinStock,
			"max_stock":     product.MaxStock,
			"image_url":     product.ImageURL,
			"status":        product.Status,
			"updated_at":    product.UpdatedAt,
		},
	}

	_, err := r.productsCollection.UpdateByID(ctx, product.ID, update)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (r *InventoryRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	// Soft delete - mark as discontinued
	update := bson.M{
		"$set": bson.M{
			"status":     Domain.ProductStatusDiscontinued,
			"updated_at": time.Now(),
		},
	}

	_, err = r.productsCollection.UpdateByID(ctx, objID, update)
	return err
}

func (r *InventoryRepository) AdjustStock(productID string, quantity float64, movementType Domain.MovementType, reason string, referenceID *string, referenceType string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objProductID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get current product
	var product Domain.Product
	err = r.productsCollection.FindOne(ctx, bson.M{"_id": objProductID}).Decode(&product)
	if err != nil {
		return fmt.Errorf("failed to find product: %w", err)
	}

	// Calculate new stock based on movement type
	var newStock float64
	previousStock := product.Stock

	switch movementType {
	case Domain.MovementTypePurchase, Domain.MovementTypeReturn, Domain.MovementTypeAdjust:
		newStock = previousStock + quantity
	case Domain.MovementTypeSale, Domain.MovementTypeDamage, Domain.MovementTypeTheft:
		newStock = previousStock - quantity
		if newStock < 0 {
			return fmt.Errorf("insufficient stock. Available: %.2f, Required: %.2f", previousStock, quantity)
		}
	}

	// Update product stock
	update := bson.M{
		"$set": bson.M{
			"stock":      newStock,
			"updated_at": time.Now(),
		},
	}

	_, err = r.productsCollection.UpdateByID(ctx, objProductID, update)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	// Create stock movement record
	movement := Domain.StockMovement{
		BusinessID: product.BusinessID,
		ProductID:  objProductID,
		Type:       movementType,
		Quantity:   quantity,
		Previous:   previousStock,
		New:        newStock,
		Reason:     reason,
		CreatedBy:  objUserID,
		CreatedAt:  time.Now(),
	}

	if referenceID != nil {
		objReferenceID, err := primitive.ObjectIDFromHex(*referenceID)
		if err == nil {
			movement.ReferenceID = &objReferenceID
			movement.ReferenceType = referenceType
		}
	}

	_, err = r.movementsCollection.InsertOne(ctx, movement)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	return nil
}

func (r *InventoryRepository) GetLowStock(businessID string, threshold float64) ([]Domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	query := bson.M{
		"business_id": objBusinessID,
		"status":      Domain.ProductStatusActive,
		"$expr":       bson.M{"$lt": []interface{}{"$stock", "$min_stock"}},
	}

	cursor, err := r.productsCollection.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find low stock products: %w", err)
	}
	defer cursor.Close(ctx)

	var products []Domain.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, fmt.Errorf("failed to decode products: %w", err)
	}

	return products, nil
}

func (r *InventoryRepository) GetStockHistory(productID string, limit int) ([]Domain.StockMovement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objProductID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.movementsCollection.Find(ctx, bson.M{"product_id": objProductID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find stock history: %w", err)
	}
	defer cursor.Close(ctx)

	var movements []Domain.StockMovement
	if err := cursor.All(ctx, &movements); err != nil {
		return nil, fmt.Errorf("failed to decode movements: %w", err)
	}

	return movements, nil
}
