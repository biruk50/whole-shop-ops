package Usecases

import (
	"fmt"
	"time"

	Domain "ShopOps/Domain"
)

type ExpenseUseCase interface {
	CreateExpense(businessID, userID string, req Domain.CreateExpenseRequest) (*Domain.Expense, error)
	GetExpenseByID(id, businessID string) (*Domain.Expense, error)
	GetExpenses(businessID string, filters Domain.ExpenseFilters) ([]Domain.Expense, error)
	UpdateExpense(id, businessID, userID string, req Domain.CreateExpenseRequest) (*Domain.Expense, error)
	VoidExpense(id, businessID, userID string) error
	GetExpenseSummary(businessID string, period string) ([]Domain.ExpenseSummary, error)
	GetExpenseTotal(businessID string, startDate, endDate time.Time) (float64, error)
	GetExpenseCategories() []Domain.ExpenseCategory
}

type expenseUseCase struct {
	expenseRepo  Domain.ExpenseRepository
	businessRepo Domain.BusinessRepository
}

func NewExpenseUseCase(
	expenseRepo Domain.ExpenseRepository,
	businessRepo Domain.BusinessRepository,
) ExpenseUseCase {
	return &expenseUseCase{
		expenseRepo:  expenseRepo,
		businessRepo: businessRepo,
	}
}

func (uc *expenseUseCase) CreateExpense(businessID, userID string, req Domain.CreateExpenseRequest) (*Domain.Expense, error) {
	// Validate business exists
	business, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}
	if business == nil {
		return nil, fmt.Errorf("business not found")
	}

	// Validate category
	if !uc.isValidCategory(req.Category) {
		return nil, fmt.Errorf("invalid expense category: %s", req.Category)
	}

	objBusinessID, err := Domain.PrimitiveObjectIDFromHex(businessID)
	if err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	objUserID, err := Domain.PrimitiveObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	expense := &Domain.Expense{
		BusinessID:  objBusinessID,
		LocalID:     req.LocalID,
		Category:    req.Category,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
		CreatedBy:   objUserID,
	}

	if expense.Date.IsZero() {
		expense.Date = time.Now()
	}

	if err := uc.expenseRepo.Create(expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return expense, nil
}

func (uc *expenseUseCase) GetExpenseByID(id, businessID string) (*Domain.Expense, error) {
	expense, err := uc.expenseRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find expense: %w", err)
	}
	if expense == nil {
		return nil, fmt.Errorf("expense not found")
	}

	// Verify expense belongs to business
	if expense.BusinessID.Hex() != businessID {
		return nil, fmt.Errorf("access denied: expense does not belong to this business")
	}

	return expense, nil
}

func (uc *expenseUseCase) GetExpenses(businessID string, filters Domain.ExpenseFilters) ([]Domain.Expense, error) {
	return uc.expenseRepo.FindByBusinessID(businessID, filters)
}

func (uc *expenseUseCase) UpdateExpense(id, businessID, userID string, req Domain.CreateExpenseRequest) (*Domain.Expense, error) {
	expense, err := uc.GetExpenseByID(id, businessID)
	if err != nil {
		return nil, err
	}

	// Check if expense can be updated (not voided/deleted)
	if expense.Status != Domain.ExpenseStatusActive {
		return nil, fmt.Errorf("cannot update expense with status: %s", expense.Status)
	}

	// Validate category if provided
	if req.Category != "" && !uc.isValidCategory(req.Category) {
		return nil, fmt.Errorf("invalid expense category: %s", req.Category)
	}

	// Update expense fields
	if req.Category != "" {
		expense.Category = req.Category
	}
	if req.Amount > 0 {
		expense.Amount = req.Amount
	}
	if req.Description != "" {
		expense.Description = req.Description
	}
	if !req.Date.IsZero() {
		expense.Date = req.Date
	}

	if err := uc.expenseRepo.Update(expense); err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	return expense, nil
}

func (uc *expenseUseCase) VoidExpense(id, businessID, userID string) error {
	expense, err := uc.GetExpenseByID(id, businessID)
	if err != nil {
		return err
	}

	// Check if expense can be voided
	if expense.Status != Domain.ExpenseStatusActive {
		return fmt.Errorf("expense cannot be voided with status: %s", expense.Status)
	}

	// Update expense status
	return uc.expenseRepo.UpdateStatus(id, Domain.ExpenseStatusVoided)
}

func (uc *expenseUseCase) GetExpenseSummary(businessID string, period string) ([]Domain.ExpenseSummary, error) {
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

	return uc.expenseRepo.GetSummaryByCategory(businessID, startDate, endDate)
}

func (uc *expenseUseCase) GetExpenseTotal(businessID string, startDate, endDate time.Time) (float64, error) {
	return uc.expenseRepo.GetTotal(businessID, startDate, endDate)
}

func (uc *expenseUseCase) GetExpenseCategories() []Domain.ExpenseCategory {
	return []Domain.ExpenseCategory{
		Domain.ExpenseCategoryRent,
		Domain.ExpenseCategoryUtilities,
		Domain.ExpenseCategoryStockPurchase,
		Domain.ExpenseCategoryTransport,
		Domain.ExpenseCategorySalaries,
		Domain.ExpenseCategoryMarketing,
		Domain.ExpenseCategoryMaintenance,
		Domain.ExpenseCategoryOther,
	}
}

func (uc *expenseUseCase) isValidCategory(category Domain.ExpenseCategory) bool {
	validCategories := uc.GetExpenseCategories()
	for _, validCat := range validCategories {
		if validCat == category {
			return true
		}
	}
	return false
}
