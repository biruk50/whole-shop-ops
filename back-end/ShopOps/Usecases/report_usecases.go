package Usecases

import (
	"fmt"
	"time"

	Domain "ShopOps/Domain"
	Infrastructure "ShopOps/Infrastructure"
)

type ReportUseCase interface {
	GenerateReport(req Domain.ReportRequest) (interface{}, error)
	GetDashboardData(businessID string) (*Domain.DashboardData, error)
	ExportReport(req Domain.ReportRequest) ([]byte, string, error)
	GetProfitSummary(businessID string, period Domain.PeriodType, startDate, endDate *time.Time) (*Domain.ProfitReport, error)
	GetProfitTrends(businessID string, period Domain.PeriodType, weeks int) ([]Domain.ProfitTrend, error)
	ComparePeriods(businessID string, period1, period2 Domain.ReportRequest) (interface{}, error)
}

type reportUseCase struct {
	reportRepo    Domain.ReportRepository
	businessRepo  Domain.BusinessRepository
	exportService Infrastructure.ExportService
}

func NewReportUseCase(
	reportRepo Domain.ReportRepository,
	businessRepo Domain.BusinessRepository,
	exportService Infrastructure.ExportService,
) ReportUseCase {
	return &reportUseCase{
		reportRepo:    reportRepo,
		businessRepo:  businessRepo,
		exportService: exportService,
	}
}

func (uc *reportUseCase) GenerateReport(req Domain.ReportRequest) (interface{}, error) {
	// Validate business exists
	_, err := uc.businessRepo.FindByID(req.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}

	// Set default dates based on period
	startDate, endDate := uc.getDateRange(req.Period, req.StartDate, req.EndDate)

	switch req.Type {
	case Domain.ReportTypeSales:
		return uc.reportRepo.GenerateSalesReport(req.BusinessID, startDate, endDate)
	case Domain.ReportTypeExpenses:
		return uc.reportRepo.GenerateExpensesReport(req.BusinessID, startDate, endDate)
	case Domain.ReportTypeProfit:
		return uc.reportRepo.GenerateProfitReport(req.BusinessID, startDate, endDate)
	case Domain.ReportTypeInventory:
		return uc.reportRepo.GenerateInventoryReport(req.BusinessID)
	default:
		return nil, fmt.Errorf("invalid report type: %s", req.Type)
	}
}

func (uc *reportUseCase) GetDashboardData(businessID string) (*Domain.DashboardData, error) {
	// Validate business exists
	_, err := uc.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to find business: %w", err)
	}

	return uc.reportRepo.GetDashboardData(businessID)
}

func (uc *reportUseCase) ExportReport(req Domain.ReportRequest) ([]byte, string, error) {
	// Generate report
	report, err := uc.GenerateReport(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate report: %w", err)
	}

	var data []byte
	var filename string

	if req.Format != nil && *req.Format == "csv" {
		// Export to CSV
		data, err = uc.exportService.ExportToCSV(report, req.Type)
		if err != nil {
			return nil, "", fmt.Errorf("failed to export to CSV: %w", err)
		}
		filename = Infrastructure.GenerateFilename(req.Type, time.Now())
	} else {
		// Default to JSON
		data, err = uc.exportService.ExportToJSON(report)
		if err != nil {
			return nil, "", fmt.Errorf("failed to export to JSON: %w", err)
		}
		filename = fmt.Sprintf("%s_%s.json",
			string(req.Type),
			time.Now().Format("20060102_150405"))
	}

	return data, filename, nil
}

func (uc *reportUseCase) GetProfitSummary(businessID string, period Domain.PeriodType, startDate, endDate *time.Time) (*Domain.ProfitReport, error) {
	if startDate == nil || endDate == nil {
		startDateVal, endDateVal := uc.getDateRange(period, nil, nil)
		startDate = &startDateVal
		endDate = &endDateVal
	}

	return uc.reportRepo.GenerateProfitReport(businessID, *startDate, *endDate)
}

func (uc *reportUseCase) GetProfitTrends(businessID string, period Domain.PeriodType, weeks int) ([]Domain.ProfitTrend, error) {
	if weeks <= 0 {
		weeks = 12 // Default to 12 weeks
	}

	var trends []Domain.ProfitTrend
	now := time.Now()

	for i := 0; i < weeks; i++ {
		endDate := now.AddDate(0, 0, -i*7)
		var startDate time.Time

		switch period {
		case Domain.PeriodTypeWeekly:
			startDate = endDate.AddDate(0, 0, -7)
		case Domain.PeriodTypeMonthly:
			startDate = endDate.AddDate(0, 0, -30)
		default:
			startDate = endDate.AddDate(0, 0, -7) // Default to weekly
		}

		report, err := uc.reportRepo.GenerateProfitReport(businessID, startDate, endDate)
		if err != nil {
			// Skip this period if there's an error
			continue
		}

		periodLabel := uc.getPeriodLabel(startDate, endDate, period)

		trends = append(trends, Domain.ProfitTrend{
			Period:   periodLabel,
			Sales:    report.TotalSales,
			Expenses: report.TotalExpenses,
			Profit:   report.NetProfit,
		})
	}

	return trends, nil
}

func (uc *reportUseCase) ComparePeriods(businessID string, period1, period2 Domain.ReportRequest) (interface{}, error) {
	// Generate reports for both periods
	report1, err := uc.GenerateReport(period1)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report for period 1: %w", err)
	}

	report2, err := uc.GenerateReport(period2)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report for period 2: %w", err)
	}

	// Create comparison result
	comparison := map[string]interface{}{
		"period1": report1,
		"period2": report2,
	}

	return comparison, nil
}

func (uc *reportUseCase) getDateRange(period Domain.PeriodType, customStart, customEnd *time.Time) (time.Time, time.Time) {
	now := time.Now()

	// Use custom dates if provided
	if customStart != nil && customEnd != nil {
		return *customStart, *customEnd
	}

	var startDate, endDate time.Time

	switch period {
	case Domain.PeriodTypeDaily:
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	case Domain.PeriodTypeWeekly:
		// Last 7 days
		endDate = now
		startDate = now.AddDate(0, 0, -7)
	case Domain.PeriodTypeMonthly:
		// Last 30 days
		endDate = now
		startDate = now.AddDate(0, 0, -30)
	case Domain.PeriodTypeYearly:
		// Last 365 days
		endDate = now
		startDate = now.AddDate(0, 0, -365)
	case Domain.PeriodTypeCustom:
		// Default to last 30 days
		endDate = now
		startDate = now.AddDate(0, 0, -30)
	default:
		// Default to last 30 days
		endDate = now
		startDate = now.AddDate(0, 0, -30)
	}

	return startDate, endDate
}

func (uc *reportUseCase) getPeriodLabel(startDate, endDate time.Time, period Domain.PeriodType) string {
	switch period {
	case Domain.PeriodTypeDaily:
		return startDate.Format("Jan 02")
	case Domain.PeriodTypeWeekly:
		_, week := startDate.ISOWeek()
		return fmt.Sprintf("Week %d", week)
	case Domain.PeriodTypeMonthly:
		return startDate.Format("Jan 2006")
	case Domain.PeriodTypeYearly:
		return startDate.Format("2006")
	default:
		return fmt.Sprintf("%s to %s",
			startDate.Format("Jan 02"),
			endDate.Format("Jan 02"))
	}
}
