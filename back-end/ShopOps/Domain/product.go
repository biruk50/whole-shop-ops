package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID   primitive.ObjectID `bson:"business_id" json:"business_id"`
	Name         string             `bson:"name" json:"name" validate:"required"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	SKU          string             `bson:"sku,omitempty" json:"sku,omitempty"`
	Barcode      string             `bson:"barcode,omitempty" json:"barcode,omitempty"`
	Category     string             `bson:"category,omitempty" json:"category,omitempty"`
	Unit         string             `bson:"unit,omitempty" json:"unit,omitempty"`
	CostPrice    float64            `bson:"cost_price" json:"cost_price" validate:"required,gt=0"`
	SellingPrice float64            `bson:"selling_price" json:"selling_price" validate:"required,gt=0"`
	Stock        float64            `bson:"stock" json:"stock" validate:"gte=0"`
	MinStock     float64            `bson:"min_stock,omitempty" json:"min_stock,omitempty"`
	MaxStock     float64            `bson:"max_stock,omitempty" json:"max_stock,omitempty"`
	ImageURL     string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	Status       ProductStatus      `bson:"status" json:"status"`
	CreatedBy    primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

type StockMovement struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	BusinessID    primitive.ObjectID  `bson:"business_id" json:"business_id"`
	ProductID     primitive.ObjectID  `bson:"product_id" json:"product_id"`
	Type          MovementType        `bson:"type" json:"type"`
	Quantity      float64             `bson:"quantity" json:"quantity"`
	Previous      float64             `bson:"previous" json:"previous"`
	New           float64             `bson:"new" json:"new"`
	Reason        string              `bson:"reason" json:"reason"`
	ReferenceID   *primitive.ObjectID `bson:"reference_id,omitempty" json:"reference_id,omitempty"`
	ReferenceType string              `bson:"reference_type,omitempty" json:"reference_type,omitempty"`
	CreatedBy     primitive.ObjectID  `bson:"created_by" json:"created_by"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
}

type MovementType string

const (
	MovementTypePurchase MovementType = "purchase"
	MovementTypeSale     MovementType = "sale"
	MovementTypeAdjust   MovementType = "adjust"
	MovementTypeDamage   MovementType = "damage"
	MovementTypeTheft    MovementType = "theft"
	MovementTypeReturn   MovementType = "return"
)

type CreateProductRequest struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description,omitempty"`
	SKU          string  `json:"sku,omitempty"`
	Barcode      string  `json:"barcode,omitempty"`
	Category     string  `json:"category,omitempty"`
	Unit         string  `json:"unit,omitempty"`
	CostPrice    float64 `json:"cost_price" validate:"required,gt=0"`
	SellingPrice float64 `json:"selling_price" validate:"required,gt=0"`
	Stock        float64 `json:"stock" validate:"gte=0"`
	MinStock     float64 `json:"min_stock,omitempty"`
	MaxStock     float64 `json:"max_stock,omitempty"`
}

type AdjustStockRequest struct {
	Quantity float64      `json:"quantity" validate:"required"`
	Type     MovementType `json:"type" validate:"required"`
	Reason   string       `json:"reason" validate:"required"`
}

type ProductRepository interface {
	Create(product *Product) error
	FindByID(id string) (*Product, error)
	FindByBusinessID(businessID string, filters ProductFilters) ([]Product, error)
	Update(product *Product) error
	Delete(id string) error
	AdjustStock(productID string, quantity float64, movementType MovementType, reason string, referenceID *string, referenceType string, userID string) error
	GetLowStock(businessID string, threshold float64) ([]Product, error)
	GetStockHistory(productID string, limit int) ([]StockMovement, error)
}

type ProductFilters struct {
	Category *string
	Status   *ProductStatus
	LowStock *bool
	Search   *string
	Limit    int
	Offset   int
}
