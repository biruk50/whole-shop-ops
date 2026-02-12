package Domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Expense struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID  primitive.ObjectID `bson:"business_id" json:"business_id"`
	LocalID     string             `bson:"local_id,omitempty" json:"local_id,omitempty"` // For offline sync
	Category    ExpenseCategory    `bson:"category" json:"category" validate:"required"`
	Amount      float64            `bson:"amount" json:"amount" validate:"required,gt=0"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	ReceiptURL  string             `bson:"receipt_url,omitempty" json:"receipt_url,omitempty"`
	Date        time.Time          `bson:"date" json:"date"`
	Status      ExpenseStatus      `bson:"status" json:"status"`
	Synced      bool               `bson:"synced" json:"synced"`
	SyncedAt    *time.Time         `bson:"synced_at,omitempty" json:"synced_at,omitempty"`
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type ExpenseCategory string

const (
	ExpenseCategoryRent          ExpenseCategory = "rent"
	ExpenseCategoryUtilities     ExpenseCategory = "utilities"
	ExpenseCategoryStockPurchase ExpenseCategory = "stock_purchase"
	ExpenseCategoryTransport     ExpenseCategory = "transport"
	ExpenseCategorySalaries      ExpenseCategory = "salaries"
	ExpenseCategoryMarketing     ExpenseCategory = "marketing"
	ExpenseCategoryMaintenance   ExpenseCategory = "maintenance"
	ExpenseCategoryOther         ExpenseCategory = "other"
)

type ExpenseStatus string

const (
	ExpenseStatusActive  ExpenseStatus = "active"
	ExpenseStatusVoided  ExpenseStatus = "voided"
	ExpenseStatusDeleted ExpenseStatus = "deleted"
)

type CreateExpenseRequest struct {
	Category    ExpenseCategory `json:"category" validate:"required"`
	Amount      float64         `json:"amount" validate:"required,gt=0"`
	Description string          `json:"description,omitempty"`
	Date        time.Time       `json:"date"`
	LocalID     string          `json:"local_id,omitempty"` // For offline sync
}

type ExpenseSummary struct {
	Category    ExpenseCategory `json:"category"`
	TotalAmount float64         `json:"total_amount"`
	Count       int             `json:"count"`
	Percentage  float64         `json:"percentage"`
}

type ExpenseRepository interface {
	Create(expense *Expense) error
	FindByID(id string) (*Expense, error)
	FindByBusinessID(businessID string, filters ExpenseFilters) ([]Expense, error)
	FindByLocalID(businessID, localID string) (*Expense, error)
	Update(expense *Expense) error
	UpdateStatus(id string, status ExpenseStatus) error
	Delete(id string) error
	GetSummaryByCategory(businessID string, startDate, endDate time.Time) ([]ExpenseSummary, error)
	GetTotal(businessID string, startDate, endDate time.Time) (float64, error)
}

type ExpenseFilters struct {
	StartDate *time.Time
	EndDate   *time.Time
	Category  *ExpenseCategory
	Status    *ExpenseStatus
	Limit     int
	Offset    int
}
