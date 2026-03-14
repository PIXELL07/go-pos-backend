package services

import (
	"time"

	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

// Extended report types
// represents a single item's aggregated sales across all outlets.
type ItemWiseRow struct {
	ItemName   string  `json:"item_name"`
	Category   string  `json:"category"`
	OutletName string  `json:"outlet_name"`
	Quantity   int64   `json:"quantity"`
	Revenue    float64 `json:"revenue"`
	Tax        float64 `json:"tax"`
	Date       string  `json:"date,omitempty"`
}

type CategoryWiseRow struct {
	CategoryName string  `json:"category_name"`
	OutletName   string  `json:"outlet_name"`
	TotalItems   int64   `json:"total_items"`
	Revenue      float64 `json:"revenue"`
}

// one bill line.
type InvoiceRow struct {
	InvoiceNumber string  `json:"invoice_number"`
	OutletName    string  `json:"outlet_name"`
	Date          string  `json:"date"`
	CustomerName  string  `json:"customer_name"`
	Source        string  `json:"source"`
	Status        string  `json:"status"`
	TotalAmount   float64 `json:"total_amount"`
	CashierName   string  `json:"cashier_name"`
}

// contains cancelled-order info.
type CancelledOrderRow struct {
	InvoiceNumber string  `json:"invoice_number"`
	OutletName    string  `json:"outlet_name"`
	Date          string  `json:"date"`
	Reason        string  `json:"reason"`
	TotalAmount   float64 `json:"total_amount"`
}
type DiscountRow struct {
	InvoiceNumber   string  `json:"invoice_number"`
	OutletName      string  `json:"outlet_name"`
	Date            string  `json:"date"`
	DiscountAmount  float64 `json:"discount_amount"`
	DiscountPercent float64 `json:"discount_percent"`
	TotalAmount     float64 `json:"total_amount"`
	Reason          string  `json:"reason"`
}

// breaks sales down by hour-of-day.
type HourlyRow struct {
	Hour       int     `json:"hour"`
	OutletName string  `json:"outlet_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
}

// shows pax-per-biller stats.
type PaxRow struct {
	BillerName  string  `json:"biller_name"`
	OutletName  string  `json:"outlet_name"`
	TotalOrders int64   `json:"total_orders"`
	TotalPax    int64   `json:"total_pax"`
	Revenue     float64 `json:"revenue"`
}

type DayWiseRow struct {
	Date       string  `json:"date"`
	OutletName string  `json:"outlet_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
	NetSales   float64 `json:"net_sales"`
	Tax        float64 `json:"tax"`
}

// – one row per order in the master report.
type OrderMasterRow struct {
	InvoiceNumber    string  `json:"invoice_number"`
	Date             string  `json:"date"`
	OutletName       string  `json:"outlet_name"`
	Source           string  `json:"source"`
	Type             string  `json:"type"`
	Status           string  `json:"status"`
	CustomerName     string  `json:"customer_name"`
	CustomerPhone    string  `json:"customer_phone"`
	Pax              int     `json:"pax"`
	SubTotal         float64 `json:"sub_total"`
	DiscountAmount   float64 `json:"discount_amount"`
	TaxAmount        float64 `json:"tax_amount"`
	DeliveryCharge   float64 `json:"delivery_charge"`
	ContainerCharge  float64 `json:"container_charge"`
	ServiceCharge    float64 `json:"service_charge"`
	AdditionalCharge float64 `json:"additional_charge"`
	TotalAmount      float64 `json:"total_amount"`
	CashierName      string  `json:"cashier_name"`
}

// Extended ReportsService methods

func (s *ReportsService) baseOrderQuery(outletIDs []string, from, to time.Time, excludeCancelled bool) *gorm.DB {
	toEnd := to.Add(24 * time.Hour)
	q := s.db.Table("orders o").
		Joins("JOIN outlets ol ON ol.id = o.outlet_id").
		Where("o.deleted_at IS NULL AND o.created_at >= ? AND o.created_at < ?", from, toEnd)
	if len(outletIDs) > 0 {
		q = q.Where("o.outlet_id IN ?", outletIDs)
	}
	if excludeCancelled {
		q = q.Where("o.status != 'cancelled'")
	}
	return q
}

// GetItemWiseReport – items sold per outlet.
func (s *ReportsService) GetItemWiseReport(f SalesReportFilter) ([]ItemWiseRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Joins("JOIN order_items oi ON oi.order_id = o.id AND oi.deleted_at IS NULL").
		Joins("JOIN categories cat ON cat.id = (SELECT category_id FROM menu_items WHERE id = oi.menu_item_id LIMIT 1)").
		Select(`
			oi.name      AS item_name,
			cat.name     AS category,
			ol.name      AS outlet_name,
			SUM(oi.quantity)              AS quantity,
			COALESCE(SUM(oi.total_price),0) AS revenue,
			COALESCE(SUM(oi.tax_amount),0)  AS tax
		`).
		Group("oi.name, cat.name, ol.name").
		Order("revenue DESC")

	var rows []ItemWiseRow
	return rows, q.Scan(&rows).Error
}

// GetCategoryWiseReport – revenue per category.
func (s *ReportsService) GetCategoryWiseReport(f SalesReportFilter) ([]CategoryWiseRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Joins("JOIN order_items oi ON oi.order_id = o.id AND oi.deleted_at IS NULL").
		Joins(`JOIN menu_items mi ON mi.id = oi.menu_item_id`).
		Joins(`JOIN categories cat ON cat.id = mi.category_id`).
		Select(`
			cat.name AS category_name,
			ol.name  AS outlet_name,
			COUNT(DISTINCT oi.id)           AS total_items,
			COALESCE(SUM(oi.total_price),0) AS revenue
		`).
		Group("cat.name, ol.name").
		Order("revenue DESC")

	var rows []CategoryWiseRow
	return rows, q.Scan(&rows).Error
}

// all invoices in range.
func (s *ReportsService) GetInvoiceReport(f SalesReportFilter) ([]InvoiceRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, false).
		Joins("LEFT JOIN users u ON u.id = o.cashier_id").
		Select(`
			o.invoice_number,
			ol.name          AS outlet_name,
			o.created_at::date::text AS date,
			o.customer_name,
			o.source,
			o.status,
			o.total_amount,
			u.name           AS cashier_name
		`).
		Order("o.created_at DESC")

	if f.Status != "" && f.Status != "all" {
		q = q.Where("o.status = ?", f.Status)
	}
	var rows []InvoiceRow
	return rows, q.Scan(&rows).Error
}

// cancelled orders only.
func (s *ReportsService) GetCancelledOrderReport(f SalesReportFilter) ([]CancelledOrderRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, false).
		Where("o.status = 'cancelled'").
		Select(`
			o.invoice_number,
			ol.name                  AS outlet_name,
			o.created_at::date::text AS date,
			o.notes                  AS reason,
			o.total_amount
		`).
		Order("o.created_at DESC")

	var rows []CancelledOrderRow
	return rows, q.Scan(&rows).Error
}

// orders that had a discount applied.
func (s *ReportsService) GetDiscountReport(f SalesReportFilter) ([]DiscountRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Where("o.discount_amount > 0").
		Select(`
			o.invoice_number,
			ol.name                  AS outlet_name,
			o.created_at::date::text AS date,
			o.discount_amount,
			o.discount_percent,
			o.total_amount,
			o.notes                  AS reason
		`).
		Order("o.discount_amount DESC")

	var rows []DiscountRow
	return rows, q.Scan(&rows).Error
}

// order count + revenue per hour-of-day.
func (s *ReportsService) GetHourlyReport(f SalesReportFilter) ([]HourlyRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Select(`
			EXTRACT(HOUR FROM o.created_at)::int AS hour,
			ol.name                              AS outlet_name,
			COUNT(DISTINCT o.id)                 AS orders,
			COALESCE(SUM(o.total_amount),0)      AS revenue
		`).
		Group("EXTRACT(HOUR FROM o.created_at), ol.name").
		Order("hour ASC")

	var rows []HourlyRow
	return rows, q.Scan(&rows).Error
}

// pax and revenue per cashier.
func (s *ReportsService) GetPaxBillerReport(f SalesReportFilter) ([]PaxRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Joins("JOIN users u ON u.id = o.cashier_id").
		Select(`
			u.name  AS biller_name,
			ol.name AS outlet_name,
			COUNT(DISTINCT o.id)         AS total_orders,
			COALESCE(SUM(o.pax),0)       AS total_pax,
			COALESCE(SUM(o.total_amount),0) AS revenue
		`).
		Group("u.name, ol.name").
		Order("revenue DESC")

	var rows []PaxRow
	return rows, q.Scan(&rows).Error
}

// day-by-day aggregation.
func (s *ReportsService) GetDayWiseReport(f SalesReportFilter) ([]DayWiseRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, true).
		Select(`
			o.created_at::date::text      AS date,
			ol.name                       AS outlet_name,
			COUNT(DISTINCT o.id)          AS orders,
			COALESCE(SUM(o.total_amount),0) AS revenue,
			COALESCE(SUM(o.net_sales),0)    AS net_sales,
			COALESCE(SUM(o.tax_amount),0)   AS tax
		`).
		Group("o.created_at::date, ol.name").
		Order("date ASC, ol.name")

	var rows []DayWiseRow
	return rows, q.Scan(&rows).Error
}

// one row per order with all financial fields.
func (s *ReportsService) GetOrderMasterReport(f SalesReportFilter) ([]OrderMasterRow, error) {
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, false).
		Joins("LEFT JOIN users u ON u.id = o.cashier_id").
		Select(`
			o.invoice_number,
			o.created_at::text           AS date,
			ol.name                      AS outlet_name,
			o.source, o.type, o.status,
			o.customer_name, o.customer_phone,
			o.pax,
			o.sub_total, o.discount_amount, o.tax_amount,
			o.delivery_charge, o.container_charge,
			o.service_charge, o.additional_charge,
			o.total_amount,
			u.name AS cashier_name
		`).
		Order("o.created_at DESC")

	if f.Status != "" && f.Status != "all" {
		q = q.Where("o.status = ?", f.Status)
	}
	var rows []OrderMasterRow
	return rows, q.Scan(&rows).Error
}

// online-platform orders with aggregator details.
func (s *ReportsService) GetOnlineOrderReport(f SalesReportFilter) ([]OrderMasterRow, error) {
	onlineSources := []models.OrderSource{
		models.OrderSourceZomato, models.OrderSourceSwiggy,
		models.OrderSourceFoodPanda, models.OrderSourceUberEats,
		models.OrderSourceDunzo, models.OrderSourceWebsite,
	}
	q := s.baseOrderQuery(f.OutletIDs, f.From, f.To, false).
		Where("o.source IN ?", onlineSources).
		Joins("LEFT JOIN users u ON u.id = o.cashier_id").
		Select(`
			o.invoice_number,
			o.created_at::text AS date,
			ol.name            AS outlet_name,
			o.source, o.type, o.status,
			o.customer_name, o.customer_phone,
			o.pax,
			o.sub_total, o.discount_amount, o.tax_amount,
			o.delivery_charge, o.container_charge,
			o.service_charge, o.additional_charge,
			o.total_amount,
			u.name AS cashier_name
		`).
		Order("o.created_at DESC")

	var rows []OrderMasterRow
	return rows, q.Scan(&rows).Error
}

// returns hourly order counts broken down by platform.
// Used by the ** Flutter orders chart which shows per-platform lines **
func (s *ReportsService) GetDashboardChartByPlatform(outletID string, date time.Time) ([]map[string]interface{}, error) {
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.Add(24 * time.Hour)

	q := s.db.Table("orders").
		Select(`
			EXTRACT(HOUR FROM created_at)::int AS hour,
			source                             AS platform,
			COUNT(*)                           AS count
		`).
		Where("deleted_at IS NULL AND created_at >= ? AND created_at < ?", from, to).
		Where("status != 'cancelled'").
		Group("EXTRACT(HOUR FROM created_at), source").
		Order("hour ASC")

	if outletID != "" {
		q = q.Where("outlet_id = ?", outletID)
	}

	var rows []struct {
		Hour     int    `json:"hour"`
		Platform string `json:"platform"`
		Count    int    `json:"count"`
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Pivot into [{hour:0, pos:3, zomato:1, ...}, ...]
	platforms := []string{"pos", "zomato", "swiggy", "foodpanda", "uber_eats", "dunzo", "website"}
	byHour := make(map[int]map[string]int)
	for h := 0; h < 24; h++ {
		byHour[h] = make(map[string]int)
	}
	for _, r := range rows {
		byHour[r.Hour][r.Platform] = r.Count
	}
	var result []map[string]interface{}
	for h := 0; h < 24; h++ {
		row := map[string]interface{}{"hour": h}
		for _, p := range platforms {
			row[p] = byHour[h][p]
		}
		result = append(result, row)
	}
	return result, nil
}

//add timeNow to filter
