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

type ReportController struct {
	reportUC Usecases.ReportUseCase
}

func NewReportController(reportUC Usecases.ReportUseCase) *ReportController {
	return &ReportController{reportUC: reportUC}
}

// GetDashboard godoc
// @Summary      Get dashboard overview
// @Description  Get key metrics for dashboard display (today's data)
// @Tags         reports
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Success      200  {object}  Domain.DashboardData
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/dashboard [get]
// @Security     BearerAuth
func (c *ReportController) GetDashboard(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	data, err := c.reportUC.GetDashboardData(businessID)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, data)
}

// GetSalesReport godoc
// @Summary      Get sales report
// @Description  Generate sales report with optional period filtering
// @Tags         reports
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: daily, weekly, monthly, yearly, custom"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD) for custom period"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD) for custom period"
// @Success      200  {object}  Domain.SalesReport
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/sales [get]
// @Security     BearerAuth
func (c *ReportController) GetSalesReport(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.ReportRequest
	req.BusinessID = businessID
	req.Type = Domain.ReportTypeSales

	// Parse period
	if period := ctx.Query("period"); period != "" {
		req.Period = Domain.PeriodType(period)
	} else {
		req.Period = Domain.PeriodTypeMonthly
	}

	// Parse dates
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	report, err := c.reportUC.GenerateReport(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetExpensesReport godoc
// @Summary      Get expenses report
// @Description  Generate expense report with optional period and category filtering
// @Tags         reports
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: daily, weekly, monthly, yearly, custom"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD) for custom period"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD) for custom period"
// @Param        category    query   string  false  "Filter by category"
// @Success      200  {object}  Domain.ExpensesReport
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/expenses [get]
// @Security     BearerAuth
func (c *ReportController) GetExpensesReport(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.ReportRequest
	req.BusinessID = businessID
	req.Type = Domain.ReportTypeExpenses

	// Parse period
	if period := ctx.Query("period"); period != "" {
		req.Period = Domain.PeriodType(period)
	} else {
		req.Period = Domain.PeriodTypeMonthly
	}

	// Parse dates
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	// Parse category
	if category := ctx.Query("category"); category != "" {
		req.Category = &category
	}

	report, err := c.reportUC.GenerateReport(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetProfitReport godoc
// @Summary      Get profit report
// @Description  Generate profit/loss report with optional period filtering
// @Tags         reports
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: daily, weekly, monthly, yearly, custom"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD) for custom period"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD) for custom period"
// @Success      200  {object}  Domain.ProfitReport
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/profit [get]
// @Security     BearerAuth
func (c *ReportController) GetProfitReport(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.ReportRequest
	req.BusinessID = businessID
	req.Type = Domain.ReportTypeProfit

	// Parse period
	if period := ctx.Query("period"); period != "" {
		req.Period = Domain.PeriodType(period)
	} else {
		req.Period = Domain.PeriodTypeMonthly
	}

	// Parse dates
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	report, err := c.reportUC.GenerateReport(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetInventoryReport godoc
// @Summary      Get inventory status report
// @Description  Generate inventory report with low stock alerts
// @Tags         reports
// @Produce      json
// @Param        businessId  path  string  true  "Business ID"
// @Success      200  {object}  Domain.InventoryReport
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/inventory [get]
// @Security     BearerAuth
func (c *ReportController) GetInventoryReport(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.ReportRequest
	req.BusinessID = businessID
	req.Type = Domain.ReportTypeInventory

	report, err := c.reportUC.GenerateReport(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// ExportReport godoc
// @Summary      Generate CSV export
// @Description  Export report data to CSV format
// @Tags         reports
// @Produce      text/csv
// @Param        businessId  path    string  true   "Business ID"
// @Param        type        query   string  true   "Report type: sales, expenses, profit, inventory"
// @Param        period      query   string  false  "Period: daily, weekly, monthly, yearly, custom"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD) for custom period"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD) for custom period"
// @Success      200  {string}  string  "CSV file"
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/export [get]
// @Security     BearerAuth
func (c *ReportController) ExportReport(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	var req Domain.ReportRequest
	req.BusinessID = businessID

	// Parse report type
	if reportType := ctx.Query("type"); reportType != "" {
		req.Type = Domain.ReportType(reportType)
	} else {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Report type is required")
		return
	}

	// Parse period
	if period := ctx.Query("period"); period != "" {
		req.Period = Domain.PeriodType(period)
	} else {
		req.Period = Domain.PeriodTypeMonthly
	}

	// Parse dates
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}

	// Set format to CSV
	format := "csv"
	req.Format = &format

	data, filename, err := c.reportUC.ExportReport(req)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(http.StatusOK, "text/csv", data)
}

// GetProfitSummary godoc
// @Summary      Get profit summary for period
// @Description  Get profit summary with custom date range
// @Tags         reports
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: daily, weekly, monthly, yearly, custom"
// @Param        start_date  query   string  false  "Start date (YYYY-MM-DD) for custom period"
// @Param        end_date    query   string  false  "End date (YYYY-MM-DD) for custom period"
// @Success      200  {object}  Domain.ProfitReport
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/profit/summary [get]
// @Security     BearerAuth
func (c *ReportController) GetProfitSummary(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	period := Domain.PeriodType(ctx.DefaultQuery("period", "monthly"))

	var startDate, endDate *time.Time
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if sd, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &sd
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if ed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &ed
		}
	}

	report, err := c.reportUC.GetProfitSummary(businessID, period, startDate, endDate)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetProfitTrends godoc
// @Summary      Get profit trends over time
// @Description  Get profit trends for multiple periods
// @Tags         reports
// @Produce      json
// @Param        businessId  path    string  true   "Business ID"
// @Param        period      query   string  false  "Period: daily, weekly, monthly"
// @Param        weeks       query   int     false  "Number of weeks to analyze (default 12)"
// @Success      200  {array}   Domain.ProfitTrend
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /api/v1/businesses/{businessId}/reports/profit/trends [get]
// @Security     BearerAuth
func (c *ReportController) GetProfitTrends(ctx *gin.Context) {
	businessID := ctx.Param("businessId")
	if businessID == "" {
		Infrastructure.JSONError(ctx, http.StatusBadRequest, nil, "Business ID is required")
		return
	}

	period := Domain.PeriodType(ctx.DefaultQuery("period", "weekly"))
	weeks := 12
	if weeksStr := ctx.Query("weeks"); weeksStr != "" {
		if w, err := strconv.Atoi(weeksStr); err == nil && w > 0 {
			weeks = w
		}
	}

	trends, err := c.reportUC.GetProfitTrends(businessID, period, weeks)
	if err != nil {
		Infrastructure.JSONError(ctx, http.StatusInternalServerError, err, "")
		return
	}

	ctx.JSON(http.StatusOK, trends)
}
