package Infrastructure

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	Domain "ShopOps/Domain"
)

type ExportService interface {
	ExportToCSV(data interface{}, reportType Domain.ReportType) ([]byte, error)
	ExportToJSON(data interface{}) ([]byte, error)
}

type exportService struct{}

func NewExportService() ExportService {
	return &exportService{}
}

func (s *exportService) ExportToCSV(data interface{}, reportType Domain.ReportType) ([]byte, error) {
	var records [][]string

	switch reportType {
	case Domain.ReportTypeSales:
		if sales, ok := data.([]Domain.Sale); ok {
			// Add header
			records = append(records, []string{
				"ID", "Date", "Customer", "Phone", "Product", "Quantity",
				"Unit Price", "Total", "Discount", "Tax", "Final Amount",
				"Payment Method", "Payment Status", "Notes",
			})

			// Add data rows
			for _, sale := range sales {
				productName := ""
				if sale.ProductID != nil {
					// In production, you'd fetch product name
					productName = "Product"
				}

				records = append(records, []string{
					sale.ID.Hex(),
					sale.CreatedAt.Format("2006-01-02 15:04:05"),
					sale.CustomerName,
					sale.CustomerPhone,
					productName,
					fmt.Sprintf("%.2f", sale.Quantity),
					fmt.Sprintf("%.2f", sale.UnitPrice),
					fmt.Sprintf("%.2f", sale.TotalAmount),
					fmt.Sprintf("%.2f", sale.Discount),
					fmt.Sprintf("%.2f", sale.Tax),
					fmt.Sprintf("%.2f", sale.FinalAmount),
					string(sale.PaymentMethod),
					string(sale.PaymentStatus),
					sale.Notes,
				})
			}
		}

	case Domain.ReportTypeExpenses:
		if expenses, ok := data.([]Domain.Expense); ok {
			// Add header
			records = append(records, []string{
				"ID", "Date", "Category", "Amount", "Description",
			})

			// Add data rows
			for _, expense := range expenses {
				records = append(records, []string{
					expense.ID.Hex(),
					expense.Date.Format("2006-01-02"),
					string(expense.Category),
					fmt.Sprintf("%.2f", expense.Amount),
					expense.Description,
				})
			}
		}

	case Domain.ReportTypeInventory:
		if products, ok := data.([]Domain.Product); ok {
			// Add header
			records = append(records, []string{
				"ID", "Name", "SKU", "Barcode", "Category", "Unit",
				"Cost Price", "Selling Price", "Stock", "Min Stock", "Max Stock",
				"Status",
			})

			// Add data rows
			for _, product := range products {
				records = append(records, []string{
					product.ID.Hex(),
					product.Name,
					product.SKU,
					product.Barcode,
					product.Category,
					product.Unit,
					fmt.Sprintf("%.2f", product.CostPrice),
					fmt.Sprintf("%.2f", product.SellingPrice),
					fmt.Sprintf("%.2f", product.Stock),
					fmt.Sprintf("%.2f", product.MinStock),
					fmt.Sprintf("%.2f", product.MaxStock),
					string(product.Status),
				})
			}
		}
	}

	// Write CSV
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	if err := writer.WriteAll(records); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %w", err)
	}

	return []byte(buf.String()), nil
}

func (s *exportService) ExportToJSON(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// GenerateFilename generates a filename for export
func GenerateFilename(reportType Domain.ReportType, timestamp time.Time) string {
	return fmt.Sprintf("%s_%s.csv",
		strings.ToLower(string(reportType)),
		timestamp.Format("20060102_150405"))
}
