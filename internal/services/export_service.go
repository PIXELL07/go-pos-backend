package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
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
