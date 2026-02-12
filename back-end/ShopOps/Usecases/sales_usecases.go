package Usecases

import (
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesUseCase interface {
	CreateSale(businessID, userID string, req Domain.CreateSaleRequest) (*Domain.Sale, error)
	GetSaleByID(id, businessID string) (*Domain.Sale, error)
	GetSales(businessID string, filters Domain.SaleFilters) ([]Domain.Sale, error)
	UpdateSale(id, businessID, userID string, req Domain.CreateSaleRequest) (*Domain.Sale, error)
	VoidSale(id, businessID, userID string) error
	GetSalesSummary(businessID string, period string) (*Domain.SaleSummary, error)
	GetSalesStats(businessID string, period string) (*Domain.SaleStats, error)
	GetDailySales(businessID string, date time.Time) ([]Domain.Sale, error)
}

type salesUseCase struct {
	salesRepo     Domain.SaleRepository
	businessRepo  Domain.BusinessRepository
	inventoryRepo Domain.ProductRepository
}

func NewSalesUseCase(
	salesRepo Domain.SaleRepository,
	businessRepo Domain.BusinessRepository,
	inventoryRepo Domain.ProductRepository,
) SalesUseCase {
	return &salesUseCase{
		salesRepo:     salesRepo,
		businessRepo:  businessRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (uc *salesUseCase) CreateSale(businessID, userID string, req Domain.CreateSaleRequest) (*Domain.Sale, error) {
	// Validate business exists
	business, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}
	if business == nil {
		return nil, fmt.Errorf("business not found")
	}

	// Validate product if specified
	var productID *primitive.ObjectID
	if req.ProductID != nil {
		objProductID, err := primitive.ObjectIDFromHex(*req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}

		product, err := uc.inventoryRepo.FindByID(*req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to find product: %w", err)
		}
		if product == nil {
			return nil, fmt.Errorf("product not found")
		}

		// Check if sufficient stock
		if product.Stock < req.Quantity {
			return nil, fmt.Errorf("insufficient stock. Available: %.2f, Requested: %.2f",
				product.Stock, req.Quantity)
		}

		productID = &objProductID
	}

	objBusinessID, err := primitive.ObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	sale := &Domain.Sale{
		BusinessID:    objBusinessID,
		LocalID:       req.LocalID,
		ProductID:     productID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		Quantity:      req.Quantity,
		UnitPrice:     req.UnitPrice,
		Discount:      req.Discount,
		Tax:           req.Tax,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
		CreatedBy:     objUserID,
	}

	if err := uc.salesRepo.Create(sale); err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	// Update inventory if product was specified
	if productID != nil {
		referenceID := sale.ID.Hex()
		if err := uc.inventoryRepo.AdjustStock(
			productID.Hex(), // This should work now
			req.Quantity,
			Domain.MovementTypeSale,
			"Sale transaction",
			&referenceID,
			"sale",
			userID,
		); err != nil {
			// Rollback sale creation? For now, just log error
			fmt.Printf("Failed to update inventory for sale: %v\n", err)
		}
	}

	return sale, nil
}
func (uc *salesUseCase) GetSaleByID(id, businessID string) (*Domain.Sale, error) {
	sale, err := uc.salesRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find sale: %w", err)
	}
	if sale == nil {
		return nil, fmt.Errorf("sale not found")
	}

	// Verify sale belongs to business
	if sale.BusinessID.Hex() != businessID {
		return nil, fmt.Errorf("access denied: sale does not belong to this business")
	}

	return sale, nil
}

func (uc *salesUseCase) GetSales(businessID string, filters Domain.SaleFilters) ([]Domain.Sale, error) {
	return uc.salesRepo.FindByBusinessID(businessID, filters)
}

func (uc *salesUseCase) UpdateSale(id, businessID, userID string, req Domain.CreateSaleRequest) (*Domain.Sale, error) {
	sale, err := uc.GetSaleByID(id, businessID)
	if err != nil {
		return nil, err
	}

	// Check if sale can be updated (not voided/refunded)
	if sale.Status != Domain.SaleStatusCompleted {
		return nil, fmt.Errorf("cannot update sale with status: %s", sale.Status)
	}

	// Get previous product and quantity for inventory adjustment
	var previousProductID *string
	var previousQuantity float64
	if sale.ProductID != nil {
		// FIX: Dereference the pointer before calling Hex()
		productIDStr := sale.ProductID.Hex()
		previousProductID = &productIDStr
		previousQuantity = sale.Quantity
	}

	// Update sale fields
	if req.ProductID != nil {
		objProductID, err := Domain.PrimitiveObjectIDFromHex(*req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %w", err)
		}
		sale.ProductID = &objProductID
	}

	sale.CustomerName = req.CustomerName
	sale.CustomerPhone = req.CustomerPhone
	sale.Quantity = req.Quantity
	sale.UnitPrice = req.UnitPrice
	sale.Discount = req.Discount
	sale.Tax = req.Tax
	sale.PaymentMethod = req.PaymentMethod
	sale.Notes = req.Notes

	if err := uc.salesRepo.Update(sale); err != nil {
		return nil, fmt.Errorf("failed to update sale: %w", err)
	}

	// Handle inventory adjustments if product changed
	if previousProductID != nil {
		// Restore previous product stock
		referenceID := sale.ID.Hex()
		uc.inventoryRepo.AdjustStock(
			*previousProductID,
			previousQuantity,
			Domain.MovementTypeReturn,
			"Sale update - restoring stock",
			&referenceID,
			"sale",
			userID,
		)
	}

	if sale.ProductID != nil {
		// Deduct new product stock
		referenceID := sale.ID.Hex()
		if err := uc.inventoryRepo.AdjustStock(
			sale.ProductID.Hex(), // FIX: This should work now
			sale.Quantity,
			Domain.MovementTypeSale,
			"Sale update - new sale",
			&referenceID,
			"sale",
			userID,
		); err != nil {
			fmt.Printf("Failed to update inventory for sale update: %v\n", err)
		}
	}

	return sale, nil
}
func (uc *salesUseCase) VoidSale(id, businessID, userID string) error {
	sale, err := uc.GetSaleByID(id, businessID)
	if err != nil {
		return err
	}

	// Check if sale can be voided
	if sale.Status != Domain.SaleStatusCompleted {
		return fmt.Errorf("sale cannot be voided with status: %s", sale.Status)
	}

	// Update sale status
	if err := uc.salesRepo.UpdateStatus(id, Domain.SaleStatusVoided); err != nil {
		return fmt.Errorf("failed to void sale: %w", err)
	}

	// Restore inventory if product was sold
	if sale.ProductID != nil {
		referenceID := sale.ID.Hex()
		if err := uc.inventoryRepo.AdjustStock(
			sale.ProductID.Hex(),
			sale.Quantity,
			Domain.MovementTypeReturn,
			"Sale voided - restoring stock",
			&referenceID,
			"sale",
			userID,
		); err != nil {
			fmt.Printf("Failed to restore inventory for voided sale: %v\n", err)
		}
	}

	return nil
}

func (uc *salesUseCase) GetSalesSummary(businessID string, period string) (*Domain.SaleSummary, error) {
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "month":
		startDate = now.AddDate(0, 0, -30)
		endDate = now
	default:
		startDate = now.AddDate(0, 0, -30) // Default to last 30 days
		endDate = now
	}

	return uc.salesRepo.GetSummary(businessID, startDate, endDate)
}

func (uc *salesUseCase) GetSalesStats(businessID string, period string) (*Domain.SaleStats, error) {
	return uc.salesRepo.GetStats(businessID, period)
}

func (uc *salesUseCase) GetDailySales(businessID string, date time.Time) ([]Domain.Sale, error) {
	return uc.salesRepo.GetDailySales(businessID, date)
}
