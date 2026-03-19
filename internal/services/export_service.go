package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"
)

// ExportService

type ExportService struct {
	reports *ReportsService
}

func NewExportService(reports *ReportsService) *ExportService {
	return &ExportService{reports: reports}
}

// it carries the binary content and suggested filename.
type ExportResult struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Sales Report CSV

func (s *ExportService) SalesReportCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetSalesReport(f)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	w.Write([]string{
		"Restaurant", "Invoice Nos", "Total Bills",
		"My Amount", "Total Discount", "Net Sales",
		"Delivery Charge", "Container Charge", "Service Charge",
		"Additional Charge", "Total Tax", "Round Off", "Waived Off",
		"Total Sales", "Online Tax Calculated",
		"GST Paid by Merchant", "GST Paid by Ecommerce",
		"Cash", "Card", "Due Payment", "Other", "Wallet", "Online",
		"Pax", "Data Synced",
	})

	for _, r := range rows {
		w.Write([]string{
			r.RestaurantName, r.InvoiceNumbers,
			strconv.Itoa(r.TotalBills),
			fmtF(r.MyAmount), fmtF(r.TotalDiscount), fmtF(r.NetSales),
			fmtF(r.DeliveryCharge), fmtF(r.ContainerCharge), fmtF(r.ServiceCharge),
			fmtF(r.AdditionalCharge), fmtF(r.TotalTax), fmtF(r.RoundOff), fmtF(r.WaivedOff),
			fmtF(r.TotalSales), fmtF(r.OnlineTaxCalculated),
			fmtF(r.GSTByMerchant), fmtF(r.GSTByEcommerce),
			fmtF(r.Cash), fmtF(r.Card), fmtF(r.DuePayment),
			fmtF(r.Other), fmtF(r.Wallet), fmtF(r.Online),
			strconv.Itoa(r.Pax), r.DataSynced,
		})
	}
	w.Flush()

	return &ExportResult{
		Filename:    fmt.Sprintf("sales_report_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Item-Wise Report CSV

func (s *ExportService) ItemWiseCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetItemWiseReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Item Name", "Category", "Outlet", "Quantity", "Revenue", "Tax"})
	for _, r := range rows {
		w.Write([]string{
			r.ItemName, r.Category, r.OutletName,
			strconv.FormatInt(r.Quantity, 10),
			fmtF(r.Revenue), fmtF(r.Tax),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("item_wise_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Category-Wise CSV

func (s *ExportService) CategoryWiseCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetCategoryWiseReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Category", "Outlet", "Total Items", "Revenue"})
	for _, r := range rows {
		w.Write([]string{
			r.CategoryName, r.OutletName,
			strconv.FormatInt(r.TotalItems, 10), fmtF(r.Revenue),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("category_wise_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Invoice Report CSV

func (s *ExportService) InvoiceCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetInvoiceReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Invoice No", "Outlet", "Date", "Customer", "Source", "Status", "Total", "Cashier"})
	for _, r := range rows {
		w.Write([]string{
			r.InvoiceNumber, r.OutletName, r.Date,
			r.CustomerName, r.Source, r.Status,
			fmtF(r.TotalAmount), r.CashierName,
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("invoices_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Order Master CSV

func (s *ExportService) OrderMasterCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetOrderMasterReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{
		"Invoice No", "Date", "Outlet", "Source", "Type", "Status",
		"Customer", "Phone", "Pax",
		"Sub Total", "Discount", "Tax",
		"Delivery", "Container", "Service", "Additional",
		"Total", "Cashier",
	})
	for _, r := range rows {
		w.Write([]string{
			r.InvoiceNumber, r.Date, r.OutletName,
			r.Source, r.Type, r.Status,
			r.CustomerName, r.CustomerPhone, strconv.Itoa(r.Pax),
			fmtF(r.SubTotal), fmtF(r.DiscountAmount), fmtF(r.TaxAmount),
			fmtF(r.DeliveryCharge), fmtF(r.ContainerCharge),
			fmtF(r.ServiceCharge), fmtF(r.AdditionalCharge),
			fmtF(r.TotalAmount), r.CashierName,
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("orders_master_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Cancelled Orders CSV

func (s *ExportService) CancelledOrdersCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetCancelledOrderReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Invoice No", "Outlet", "Date", "Reason", "Total Amount"})
	for _, r := range rows {
		w.Write([]string{r.InvoiceNumber, r.OutletName, r.Date, r.Reason, fmtF(r.TotalAmount)})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("cancelled_orders_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Discount Report CSV

func (s *ExportService) DiscountCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetDiscountReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Invoice No", "Outlet", "Date", "Discount Amt", "Discount %", "Total", "Reason"})
	for _, r := range rows {
		w.Write([]string{
			r.InvoiceNumber, r.OutletName, r.Date,
			fmtF(r.DiscountAmount), fmtF(r.DiscountPercent),
			fmtF(r.TotalAmount), r.Reason,
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("discounts_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Hourly Report CSV

func (s *ExportService) HourlyCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetHourlyReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Hour", "Outlet", "Orders", "Revenue"})
	for _, r := range rows {
		w.Write([]string{
			strconv.Itoa(r.Hour), r.OutletName,
			strconv.FormatInt(r.Orders, 10), fmtF(r.Revenue),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("hourly_sales_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Day-Wise CSV

func (s *ExportService) DayWiseCSV(f SalesReportFilter) (*ExportResult, error) {
	rows, err := s.reports.GetDayWiseReport(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Date", "Outlet", "Orders", "Revenue", "Net Sales", "Tax"})
	for _, r := range rows {
		w.Write([]string{
			r.Date, r.OutletName,
			strconv.FormatInt(r.Orders, 10),
			fmtF(r.Revenue), fmtF(r.NetSales), fmtF(r.Tax),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("day_wise_%s_%s.csv", f.From.Format("20060102"), f.To.Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Pending Purchases CSV

func (s *ExportService) PendingPurchasesCSV(f PurchaseFilter) (*ExportResult, error) {
	invSvc := &InventoryService{db: s.reports.db}
	purchases2, _, err := invSvc.GetPendingPurchases(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"ID", "Outlet", "Item Name", "Quantity", "Unit", "Amount", "Status", "Type", "Date"})
	for _, r := range purchases2 {
		outletName := ""
		if r.Outlet.Name != "" {
			outletName = r.Outlet.Name
		}
		w.Write([]string{
			r.ID.String(), outletName, r.ItemName,
			strconv.FormatFloat(r.Quantity, 'f', 2, 64),
			r.Unit,
			fmtF(r.Amount),
			r.Status, r.Type,
			r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("pending_purchases_%s.csv", time.Now().Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// Store Status CSV

func (s *ExportService) StoreStatusCSV(f StoreStatusFilter) (*ExportResult, error) {
	ssSvc := &StoreStatusService{db: s.reports.db}
	rows, _, err := ssSvc.GetStatus(f)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"Outlet", "Platform", "Is Online", "Offline Since", "Last Checked"})
	for _, r := range rows {
		offlineSince := ""
		if r.OfflineSince != nil {
			offlineSince = r.OfflineSince.Format("2006-01-02 15:04")
		}
		outletName := ""
		if r.Outlet.Name != "" {
			outletName = r.Outlet.Name
		}
		online := "Yes"
		if !r.IsOnline {
			online = "No"
		}
		w.Write([]string{
			outletName, r.Platform, online,
			offlineSince,
			r.LastChecked.Format("2006-01-02 15:04"),
		})
	}
	w.Flush()
	return &ExportResult{
		Filename:    fmt.Sprintf("store_status_%s.csv", time.Now().Format("20060102")),
		ContentType: "text/csv",
		Data:        buf.Bytes(),
	}, nil
}

// fmtF formats a float to 2 decimal places.
func fmtF(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
