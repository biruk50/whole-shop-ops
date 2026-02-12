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

type SalesController struct {
	salesUC Usecases.SalesUseCase
}

func NewSalesController(salesUC Usecases.SalesUseCase) *SalesController {
	return &SalesController{salesUC: salesUC}
}

// CreateSale godoc
// @Summary      Record a new sale
// @Description  Record a sales transaction with optional product association
// @Tags         sales
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                    true  "Business ID"
// @Param        request     body  Domain.CreateSaleRequest  true  "Sale details"
// @Success      201  {object}  Domain.Sale
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales [post]
// @Security     BearerAuth
func (c *SalesController) CreateSale(ctx *gin.Context) {
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

	var req Domain.CreateSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	sale, err := c.salesUC.CreateSale(businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusCreated, sale)
}

// GetSales godoc
// @Summary      List all sales
// @Description  Get sales transactions with filtering and pagination
// @Tags         sales
// @Produce      json
// @Param        businessId      path      string  true   "Business ID"
// @Param        start_date      query     string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date        query     string  false  "End date (YYYY-MM-DD)"
// @Param        status          query     string  false  "Sale status"
// @Param        payment_method  query     string  false  "Payment method"
// @Param        payment_status  query     string  false  "Payment status"
// @Param        limit           query     int     false  "Limit results"
// @Param        offset          query     int     false  "Offset results"
// @Success      200  {array}   Domain.Sale
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales [get]
// @Security     BearerAuth
func (c *SalesController) GetSales(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	// Parse filters
	filters := Domain.SaleFilters{}

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

	// Status filter
	if status := ctx.Query("status"); status != "" {
		saleStatus := Domain.SaleStatus(status)
		filters.Status = &saleStatus
	}

	// Payment filters
	if paymentMethod := ctx.Query("payment_method"); paymentMethod != "" {
		pm := Domain.PaymentMethod(paymentMethod)
		filters.PaymentMethod = &pm
	}

	if paymentStatus := ctx.Query("payment_status"); paymentStatus != "" {
		ps := Domain.PaymentStatus(paymentStatus)
		filters.PaymentStatus = &ps
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

	sales, err := c.salesUC.GetSales(businessID, filters)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, sales)
}

// GetSale godoc
// @Summary      Get sale details
// @Description  Get detailed information about a specific sale
// @Tags         sales
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        saleId      path  string  true  "Sale ID"
// @Success      200  {object}  Domain.Sale
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales/{saleId} [get]
// @Security     BearerAuth
func (c *SalesController) GetSale(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	saleID := ctx.Param("saleId")
	if saleID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Sale ID is required")
		return
	}

	sale, err := c.salesUC.GetSaleByID(saleID, businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusNotFound, err, "")
		return
	}

	ctx.JSON(http.StatusOK, sale)
}

// UpdateSale godoc
// @Summary      Update sale
// @Description  Update sale details (before sync)
// @Tags         sales
// @Accept       json
// @Produce      json
// @Param        businessId  path  string                    true  "Business ID"
// @Param        saleId      path  string                    true  "Sale ID"
// @Param        request     body  Domain.CreateSaleRequest  true  "Sale update details"
// @Success      200  {object}  Domain.Sale
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales/{saleId} [patch]
// @Security     BearerAuth
func (c *SalesController) UpdateSale(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	saleID := ctx.Param("saleId")
	if saleID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Sale ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	var req Domain.CreateSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	sale, err := c.salesUC.UpdateSale(saleID, businessID, userID.(string), req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, sale)
}

// VoidSale godoc
// @Summary      Void/soft delete sale
// @Description  Void a completed sale transaction
// @Tags         sales
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Param        saleId      path  string  true  "Sale ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales/{saleId} [delete]
// @Security     BearerAuth
func (c *SalesController) VoidSale(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	saleID := ctx.Param("saleId")
	if saleID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Sale ID is required")
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		Infrastructure.JSONError(ctx, http.StatusUnauthorized, nil, "User not authenticated")
		return
	}

	if err := c.salesUC.VoidSale(saleID, businessID, userID.(string)); err != nil {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, err, "")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Sale voided successfully"})
}

// GetSalesSummary godoc
// @Summary      Get sales summary
// @Description  Get aggregated sales data for a period (daily/weekly/monthly)
// @Tags         sales
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: today, week, month"
// @Success      200  {object}  Domain.SaleSummary
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales/summary [get]
// @Security     BearerAuth
func (c *SalesController) GetSalesSummary(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	period := ctx.DefaultQuery("period", "month")

	summary, err := c.salesUC.GetSalesSummary(businessID, period)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, summary)
}

// GetSalesStats godoc
// @Summary      Get sales statistics
// @Description  Get sales statistics and analytics
// @Tags         sales
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: today, week, month"
// @Success      200  {object}  Domain.SaleStats
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/sales/stats [get]
// @Security     BearerAuth
func (c *SalesController) GetSalesStats(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	period := ctx.DefaultQuery("period", "month")

	stats, err := c.salesUC.GetSalesStats(businessID, period)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
