package Usecases

import (
	"fmt"

	Domain "ShopOps/Domain"
)

type InventoryUseCase interface {
	CreateProduct(businessID, userID string, req Domain.CreateProductRequest) (*Domain.Product, error)
	GetProductByID(id, businessID string) (*Domain.Product, error)
	GetProducts(businessID string, filters Domain.ProductFilters) ([]Domain.Product, error)
	UpdateProduct(id, businessID, userID string, req Domain.CreateProductRequest) (*Domain.Product, error)
	DeleteProduct(id, businessID, userID string) error
	AdjustStock(id, businessID, userID string, req Domain.AdjustStockRequest) error
	GetLowStock(businessID string, threshold float64) ([]Domain.Product, error)
	GetStockHistory(productID, businessID string, limit int) ([]Domain.StockMovement, error)
}

type inventoryUseCase struct {
	inventoryRepo Domain.ProductRepository
	businessRepo  Domain.BusinessRepository
}

func NewInventoryUseCase(
	inventoryRepo Domain.ProductRepository,
	businessRepo Domain.BusinessRepository,
) InventoryUseCase {
	return &inventoryUseCase{
		inventoryRepo: inventoryRepo,
		businessRepo:  businessRepo,
	}
}

func (uc *inventoryUseCase) CreateProduct(businessID, userID string, req Domain.CreateProductRequest) (*Domain.Product, error) {
	// Validate business exists
	business, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}
	if business == nil {
		return nil, fmt.Errorf("business not found")
	}

	// Validate selling price > cost price
	if req.SellingPrice <= req.CostPrice {
		return nil, fmt.Errorf("selling price must be greater than cost price")
	}

	// Validate min/max stock if provided
	if req.MinStock > 0 && req.MaxStock > 0 && req.MinStock >= req.MaxStock {
		return nil, fmt.Errorf("minimum stock must be less than maximum stock")
	}

	objBusinessID, err := Domain.PrimitiveObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	objUserID, err := Domain.PrimitiveObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	product := &Domain.Product{
		BusinessID:   objBusinessID,
		Name:         req.Name,
		Description:  req.Description,
		SKU:          req.SKU,
		Barcode:      req.Barcode,
		Category:     req.Category,
		Unit:         req.Unit,
		CostPrice:    req.CostPrice,
		SellingPrice: req.SellingPrice,
		Stock:        req.Stock,
		MinStock:     req.MinStock,
		MaxStock:     req.MaxStock,
		CreatedBy:    objUserID,
	}

	if err := uc.inventoryRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (uc *inventoryUseCase) GetProductByID(id, businessID string) (*Domain.Product, error) {
	product, err := uc.inventoryRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Verify product belongs to business
	if product.BusinessID.Hex() != businessID {
		return nil, fmt.Errorf("access denied: product does not belong to this business")
	}

	return product, nil
}

func (uc *inventoryUseCase) GetProducts(businessID string, filters Domain.ProductFilters) ([]Domain.Product, error) {
	return uc.inventoryRepo.FindByBusinessID(businessID, filters)
}

func (uc *inventoryUseCase) UpdateProduct(id, businessID, userID string, req Domain.CreateProductRequest) (*Domain.Product, error) {
	product, err := uc.GetProductByID(id, businessID)
	if err != nil {
		return nil, err
	}

	// Validate selling price > cost price
	if req.SellingPrice > 0 && req.CostPrice > 0 && req.SellingPrice <= req.CostPrice {
		return nil, fmt.Errorf("selling price must be greater than cost price")
	}

	// Validate min/max stock if provided
	if req.MinStock > 0 && req.MaxStock > 0 && req.MinStock >= req.MaxStock {
		return nil, fmt.Errorf("minimum stock must be less than maximum stock")
	}

	// Update product fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.SKU != "" {
		product.SKU = req.SKU
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Unit != "" {
		product.Unit = req.Unit
	}
	if req.CostPrice > 0 {
		product.CostPrice = req.CostPrice
	}
	if req.SellingPrice > 0 {
		product.SellingPrice = req.SellingPrice
	}
	if req.MinStock >= 0 {
		product.MinStock = req.MinStock
	}
	if req.MaxStock >= 0 {
		product.MaxStock = req.MaxStock
	}

	// Stock should only be updated via AdjustStock method
	// product.Stock = req.Stock

	if err := uc.inventoryRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (uc *inventoryUseCase) DeleteProduct(id, businessID, userID string) error {
	product, err := uc.GetProductByID(id, businessID)
	if err != nil {
		return err
	}

	// Check if product has stock
	if product.Stock > 0 {
		return fmt.Errorf("cannot delete product with remaining stock. Current stock: %.2f", product.Stock)
	}

	return uc.inventoryRepo.Delete(id)
}

func (uc *inventoryUseCase) AdjustStock(id, businessID, userID string, req Domain.AdjustStockRequest) error {
	// First, get the product to verify it belongs to business
	_, err := uc.GetProductByID(id, businessID)
	if err != nil {
		return err
	}

	// Validate movement type
	if !uc.isValidMovementType(req.Type) {
		return fmt.Errorf("invalid movement type: %s", req.Type)
	}

	// Validate reason
	if req.Reason == "" {
		return fmt.Errorf("reason is required for stock adjustment")
	}

	// Validate quantity
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Call repository method
	return uc.inventoryRepo.AdjustStock(
		id,
		req.Quantity,
		req.Type,
		req.Reason,
		nil, // referenceID
		"",  // referenceType
		userID,
	)
}

func (uc *inventoryUseCase) GetLowStock(businessID string, threshold float64) ([]Domain.Product, error) {
	// Use threshold if provided, otherwise use product's min_stock
	if threshold > 0 {
		// This would require a different implementation
		// For now, use repository's GetLowStock which uses min_stock
		return uc.inventoryRepo.GetLowStock(businessID, threshold)
	}

	return uc.inventoryRepo.GetLowStock(businessID, 0) // 0 means use product's min_stock
}

func (uc *inventoryUseCase) GetStockHistory(productID, businessID string, limit int) ([]Domain.StockMovement, error) {
	// Verify product belongs to business
	_, err := uc.GetProductByID(productID, businessID)
	if err != nil {
		return nil, err
	}

	return uc.inventoryRepo.GetStockHistory(productID, limit)
}

func (uc *inventoryUseCase) isValidMovementType(movementType Domain.MovementType) bool {
	validTypes := []Domain.MovementType{
		Domain.MovementTypePurchase,
		Domain.MovementTypeSale,
		Domain.MovementTypeAdjust,
		Domain.MovementTypeDamage,
		Domain.MovementTypeTheft,
		Domain.MovementTypeReturn,
	}

	for _, validType := range validTypes {
		if validType == movementType {
			return true
		}
	}
	return false
}
