package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sale struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	BusinessID    primitive.ObjectID  `bson:"business_id" json:"business_id"`
	LocalID       string              `bson:"local_id,omitempty" json:"local_id,omitempty"` // For offline sync
	ProductID     *primitive.ObjectID `bson:"product_id,omitempty" json:"product_id,omitempty"`
	CustomerName  string              `bson:"customer_name,omitempty" json:"customer_name,omitempty"`
	CustomerPhone string              `bson:"customer_phone,omitempty" json:"customer_phone,omitempty"`
	Quantity      float64             `bson:"quantity" json:"quantity" validate:"required,gt=0"`
	UnitPrice     float64             `bson:"unit_price" json:"unit_price" validate:"required,gt=0"`
	TotalAmount   float64             `bson:"total_amount" json:"total_amount"`
	Discount      float64             `bson:"discount,omitempty" json:"discount,omitempty"`
	Tax           float64             `bson:"tax,omitempty" json:"tax,omitempty"`
	FinalAmount   float64             `bson:"final_amount" json:"final_amount"`
	PaymentMethod PaymentMethod       `bson:"payment_method" json:"payment_method"`
	PaymentStatus PaymentStatus       `bson:"payment_status" json:"payment_status"`
	Notes         string              `bson:"notes,omitempty" json:"notes,omitempty"`
	Status        SaleStatus          `bson:"status" json:"status"`
	Synced        bool                `bson:"synced" json:"synced"`
	SyncedAt      *time.Time          `bson:"synced_at,omitempty" json:"synced_at,omitempty"`
	CreatedBy     primitive.ObjectID  `bson:"created_by" json:"created_by"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
}

type SaleStatus string

const (
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusVoided    SaleStatus = "voided"
	SaleStatusRefunded  SaleStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodMobile PaymentMethod = "mobile"
	PaymentMethodBank   PaymentMethod = "bank"
	PaymentMethodCredit PaymentMethod = "credit"
	PaymentMethodOther  PaymentMethod = "other"
)

type PaymentStatus string

const (
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type CreateSaleRequest struct {
	ProductID     *string       `json:"product_id,omitempty"`
	CustomerName  string        `json:"customer_name,omitempty"`
	CustomerPhone string        `json:"customer_phone,omitempty"`
	Quantity      float64       `json:"quantity" validate:"required,gt=0"`
	UnitPrice     float64       `json:"unit_price" validate:"required,gt=0"`
	Discount      float64       `json:"discount,omitempty"`
	Tax           float64       `json:"tax,omitempty"`
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required"`
	Notes         string        `json:"notes,omitempty"`
	LocalID       string        `json:"local_id,omitempty"` // For offline sync
}

type SaleSummary struct {
	Date             time.Time `json:"date"`
	TotalSales       float64   `json:"total_sales"`
	TotalAmount      float64   `json:"total_amount"`
	TotalDiscount    float64   `json:"total_discount"`
	TotalTax         float64   `json:"total_tax"`
	TransactionCount int       `json:"transaction_count"`
}

type SaleStats struct {
	DailyAverage   float64 `json:"daily_average"`
	WeeklyTotal    float64 `json:"weekly_total"`
	MonthlyTotal   float64 `json:"monthly_total"`
	BestSellingDay string  `json:"best_selling_day"`
	TopProduct     string  `json:"top_product,omitempty"`
}

type SaleRepository interface {
	Create(sale *Sale) error
	FindByID(id string) (*Sale, error)
	FindByBusinessID(businessID string, filters SaleFilters) ([]Sale, error)
	FindByLocalID(businessID, localID string) (*Sale, error)
	Update(sale *Sale) error
	UpdateStatus(id string, status SaleStatus) error
	Delete(id string) error
	GetSummary(businessID string, startDate, endDate time.Time) (*SaleSummary, error)
	GetStats(businessID string, period string) (*SaleStats, error)
	GetDailySales(businessID string, date time.Time) ([]Sale, error)
}

type SaleFilters struct {
	StartDate     *time.Time
	EndDate       *time.Time
	Status        *SaleStatus
	PaymentMethod *PaymentMethod
	PaymentStatus *PaymentStatus
	Limit         int
	Offset        int
}
