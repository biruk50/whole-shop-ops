package controllers

import (
	"net/http"
	"strconv"
	"time"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
	Usecases "ShopOps/Usecases"

	"github.com/gin-gonic/gin"
)

type ExpenseController struct {
	expenseUC Usecases.ExpenseUseCase
}

func NewExpenseController(expenseUC Usecases.ExpenseUseCase) *ExpenseController {
	return &ExpenseController{expenseUC: expenseUC}
}

// CreateExpense godoc
// @Summary      Record a new expense
// @Description  Record a business expense with categorization
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                       true  "Business ID"
// @Param        request     body  Domain.CreateExpenseRequest  true  "Expense details"
// @Success      201  {object}  Domain.Expense
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses [post]
// @Security     BearerAuth
func (c *ExpenseController) CreateExpense(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	expense, err := c.expenseUC.CreateExpense(businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusCreated, expense)
}

// GetExpenses godoc
// @Summary      List all expenses
// @Description  Get expense transactions with filtering and pagination
// @Tags         expenses
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD)"
// @Param        category    query   string  false  "Expense category"
// @Param        status      query   string  false  "Expense status"
// @Param        limit       query   int     false  "Limit results"
// @Param        offset      query   int     false  "Offset results"
// @Success      200  {array}   Domain.Expense
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses [get]
// @Security     BearerAuth
func (c *ExpenseController) GetExpenses(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	// Parse filters
	filters := Domain.ExpenseFilters{}

	// Date filters
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	// Category filter
	if category := ctx.Query("category"); category != "" {
		expenseCategory := Domain.ExpenseCategory(category)
		filters.Category = &expenseCategory
	}

	// Status filter
	if status := ctx.Query("status"); status != "" {
		expenseStatus := Domain.ExpenseStatus(status)
		filters.Status = &expenseStatus
	}

	// Pagination
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	expenses, err := c.expenseUC.GetExpenses(businessID, filters)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, expenses)
}

// GetExpense godoc
// @Summary      Get expense details
// @Description  Get detailed information about a specific expense
// @Tags         expenses
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        expenseId   path  string  true  "Expense ID"
// @Success      200  {object}  Domain.Expense
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses/{expenseId} [get]
// @Security     BearerAuth
func (c *ExpenseController) GetExpense(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	expenseID := ctx.Param("expenseId")
	if expenseID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Expense ID is required")
		return
	}

	expense, err := c.expenseUC.GetExpenseByID(expenseID, businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusNotFound, err, "")
		return
	}

	ctx.JSON(http.StatusOK, expense)
}

// UpdateExpense godoc
// @Summary      Update expense
// @Description  Update expense details (before sync)
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                       true  "Business ID"
// @Param        expenseId   path  string                       true  "Expense ID"
// @Param        request     body  Domain.CreateExpenseRequest  true  "Expense update details"
// @Success      200  {object}  Domain.Expense
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses/{expenseId} [patch]
// @Security     BearerAuth
func (c *ExpenseController) UpdateExpense(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	expenseID := ctx.Param("expenseId")
	if expenseID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Expense ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateExpenseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	expense, err := c.expenseUC.UpdateExpense(expenseID, businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, expense)
}

// VoidExpense godoc
// @Summary      Void/soft delete expense
// @Description  Void an expense transaction
// @Tags         expenses
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        expenseId   path  string  true  "Expense ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses/{expenseId} [delete]
// @Security     BearerAuth
func (c *ExpenseController) VoidExpense(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	expenseID := ctx.Param("expenseId")
	if expenseID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Expense ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	if err := c.expenseUC.VoidExpense(expenseID, businessID, userID.(string)); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Expense voided successfully"})
}

// GetExpenseSummary godoc
// @Summary      Get expense summary by category
// @Description  Get aggregated expense data categorized
// @Tags         expenses
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: today, week, month"
// @Success      200  {array}   Domain.ExpenseSummary
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses/summary [get]
// @Security     BearerAuth
func (c *ExpenseController) GetExpenseSummary(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	period := ctx.DefaultQuery("period", "month")

	summary, err := c.expenseUC.GetExpenseSummary(businessID, period)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

// GetExpenseCategories godoc
// @Summary      List available expense categories
// @Description  Get all predefined expense categories
// @Tags         expenses
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Success      200  {array}   Domain.ExpenseCategory
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/expenses/categories [get]
// @Security     BearerAuth
func (c *ExpenseController) GetExpenseCategories(ctx *gin.Context) {
	categories := c.expenseUC.GetExpenseCategories()
	ctx.JSON(http.StatusOK, categories)
}
