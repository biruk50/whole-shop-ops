package Domain

import "time"

type ReportType string

const (
	ReportTypeSales     ReportType = "sales"
	ReportTypeExpenses  ReportType = "expenses"
	ReportTypeProfit    ReportType = "profit"
	ReportTypeInventory ReportType = "inventory"
)

type PeriodType string

const (
	PeriodTypeDaily   PeriodType = "daily"
	PeriodTypeWeekly  PeriodType = "weekly"
	PeriodTypeMonthly PeriodType = "monthly"
	PeriodTypeYearly  PeriodType = "yearly"
	PeriodTypeCustom  PeriodType = "custom"
)

type ReportRequest struct {
	BusinessID string     `json:"business_id" validate:"required"`
	Type       ReportType `json:"type" validate:"required"`
	Period     PeriodType `json:"period" validate:"required"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Category   *string    `json:"category,omitempty"`
	Format     *string    `json:"format,omitempty"` // json, csv
}

type SalesReport struct {
	Period            string       `json:"period"`
	TotalSales        float64      `json:"total_sales"`
	TotalAmount       float64      `json:"total_amount"`
	TotalTransactions int          `json:"total_transactions"`
	AverageSale       float64      `json:"average_sale"`
	TopProducts       []TopProduct `json:"top_products,omitempty"`
	DailyBreakdown    []DailySales `json:"daily_breakdown,omitempty"`
}

type TopProduct struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

type DailySales struct {
	Date         string  `json:"date"`
	Sales        float64 `json:"sales"`
	Amount       float64 `json:"amount"`
	Transactions int     `json:"transactions"`
}

type ExpensesReport struct {
	Period            string            `json:"period"`
	TotalExpenses     float64           `json:"total_expenses"`
	CategoryBreakdown []CategoryExpense `json:"category_breakdown"`
	DailyExpenses     []DailyExpense    `json:"daily_expenses,omitempty"`
}

type CategoryExpense struct {
	Category    ExpenseCategory `json:"category"`
	TotalAmount float64         `json:"total_amount"`
	Count       int             `json:"count"`
	Percentage  float64         `json:"percentage"`
}

type DailyExpense struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type ProfitReport struct {
	Period        string        `json:"period"`
	TotalSales    float64       `json:"total_sales"`
	TotalExpenses float64       `json:"total_expenses"`
	GrossProfit   float64       `json:"gross_profit"`
	NetProfit     float64       `json:"net_profit"`
	ProfitMargin  float64       `json:"profit_margin"`
	Trends        []ProfitTrend `json:"trends,omitempty"`
}

type ProfitTrend struct {
	Period   string  `json:"period"`
	Sales    float64 `json:"sales"`
	Expenses float64 `json:"expenses"`
	Profit   float64 `json:"profit"`
}

type InventoryReport struct {
	TotalProducts int             `json:"total_products"`
	TotalStock    float64         `json:"total_stock"`
	TotalValue    float64         `json:"total_value"`
	LowStockItems []LowStockItem  `json:"low_stock_items"`
	StockMovement []StockMovement `json:"stock_movement,omitempty"`
}

type LowStockItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Current     float64 `json:"current"`
	Minimum     float64 `json:"minimum"`
	Difference  float64 `json:"difference"`
}

type DashboardData struct {
	TodaySales      float64 `json:"today_sales"`
	TodayExpenses   float64 `json:"today_expenses"`
	TodayProfit     float64 `json:"today_profit"`
	WeekSales       float64 `json:"week_sales"`
	WeekExpenses    float64 `json:"week_expenses"`
	WeekProfit      float64 `json:"week_profit"`
	MonthSales      float64 `json:"month_sales"`
	MonthExpenses   float64 `json:"month_expenses"`
	MonthProfit     float64 `json:"month_profit"`
	LowStockCount   int     `json:"low_stock_count"`
	PendingPayments float64 `json:"pending_payments"`
}

type ReportRepository interface {
	GenerateSalesReport(businessID string, startDate, endDate time.Time) (*SalesReport, error)
	GenerateExpensesReport(businessID string, startDate, endDate time.Time) (*ExpensesReport, error)
	GenerateProfitReport(businessID string, startDate, endDate time.Time) (*ProfitReport, error)
	GenerateInventoryReport(businessID string) (*InventoryReport, error)
	GetDashboardData(businessID string) (*DashboardData, error)
	ExportCSV(report interface{}, reportType ReportType) ([]byte, error)
}
