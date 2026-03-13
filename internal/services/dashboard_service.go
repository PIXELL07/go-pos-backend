package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

type DashboardStats struct {
	TotalSales         float64 `json:"total_sales"`
	TotalOutlets       int64   `json:"total_outlets"`
	TotalOrders        int64   `json:"total_orders"`
	OnlineSales        float64 `json:"online_sales"`
	OnlineSalesPercent string  `json:"online_sales_percent"`
	CashCollected      float64 `json:"cash_collected"`
	CashCollectedPct   string  `json:"cash_collected_percent"`
	NetSales           float64 `json:"net_sales"`
	NetSalesOutlets    int64   `json:"net_sales_outlets"`
	Expenses           float64 `json:"expenses"`
	Taxes              float64 `json:"taxes"`
	Discounts          float64 `json:"discounts"`
	DiscountsPercent   string  `json:"discounts_percent"`
}

type OutletStat struct {
	OutletID   string  `json:"outlet_id"`
	OutletName string  `json:"outlet_name"`
	IsTotal    bool    `json:"is_total"`
	Orders     int64   `json:"orders"`
	Sales      float64 `json:"sales"`
	NetSales   float64 `json:"net_sales"`
	Tax        float64 `json:"tax"`
	Discounts  float64 `json:"discounts"`
	Modified   int64   `json:"modified"`
	Reprint    int64   `json:"reprint"`
}

type ChartData struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
	Tab    string    `json:"tab"`
}

type SummaryData struct {
	TotalOrders   int64   `json:"total_orders"`
	TotalSales    float64 `json:"total_sales"`
	NetSales      float64 `json:"net_sales"`
	TotalTax      float64 `json:"total_tax"`
	TotalDiscount float64 `json:"total_discount"`
	Period        struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"period"`
}

func (s *DashboardService) GetDashboardStats(userID uuid.UUID, outletID string, date time.Time) (*DashboardStats, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := s.db.Table("orders").
		Where("deleted_at IS NULL AND status NOT IN ('cancelled') AND created_at >= ? AND created_at < ?", startOfDay, endOfDay)

	if outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}

	type Result struct {
		TotalOrders   int64
		TotalSales    float64
		TotalTax      float64
		TotalDiscount float64
		CashAmount    float64
		OnlineAmount  float64
		NetSales      float64
	}

	var r Result
	query.Select(`
		COUNT(*) as total_orders,
		COALESCE(SUM(total_amount), 0) as total_sales,
		COALESCE(SUM(tax_amount), 0) as total_tax,
		COALESCE(SUM(discount_amount), 0) as total_discount,
		COALESCE(SUM(net_sales), 0) as net_sales
	`).Scan(&r)

	// Cash from payments
	s.db.Table("payments p").
		Joins("JOIN orders o ON o.id = p.order_id").
		Where("p.method = 'cash' AND o.deleted_at IS NULL AND o.created_at >= ? AND o.created_at < ?", startOfDay, endOfDay).
		Select("COALESCE(SUM(p.amount), 0)").Scan(&r.CashAmount)

	// Online from payments
	s.db.Table("payments p").
		Joins("JOIN orders o ON o.id = p.order_id").
		Where("p.method IN ('online','wallet') AND o.deleted_at IS NULL AND o.created_at >= ? AND o.created_at < ?", startOfDay, endOfDay).
		Select("COALESCE(SUM(p.amount), 0)").Scan(&r.OnlineAmount)

	// Count active outlets
	var outletCount int64
	s.db.Table("outlets").Where("deleted_at IS NULL AND is_active = true").Count(&outletCount)

	cashPct := "0%"
	discPct := "0%"
	onlinePct := "0%"
	if r.TotalSales > 0 {
		cashPct = formatPercent(r.CashAmount / r.TotalSales * 100)
		discPct = formatPercent(r.TotalDiscount / r.TotalSales * 100)
		onlinePct = formatPercent(r.OnlineAmount / r.TotalSales * 100)
	}

	return &DashboardStats{
		TotalSales:         r.TotalSales,
		TotalOutlets:       outletCount,
		TotalOrders:        r.TotalOrders,
		OnlineSales:        r.OnlineAmount,
		OnlineSalesPercent: onlinePct,
		CashCollected:      r.CashAmount,
		CashCollectedPct:   cashPct,
		NetSales:           r.NetSales,
		NetSalesOutlets:    outletCount,
		Taxes:              r.TotalTax,
		Discounts:          r.TotalDiscount,
		DiscountsPercent:   discPct,
	}, nil
}

func (s *DashboardService) GetOutletStats(userID uuid.UUID, date time.Time) ([]OutletStat, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var rows []struct {
		OutletID   string
		OutletName string
		Orders     int64
		Sales      float64
		NetSales   float64
		Tax        float64
		Discounts  float64
		Modified   int64
		Reprint    int64
	}

	s.db.Table("orders o").
		Joins("JOIN outlets ol ON ol.id = o.outlet_id").
		Where("o.deleted_at IS NULL AND o.status != 'cancelled' AND o.created_at >= ? AND o.created_at < ?", startOfDay, endOfDay).
		Select(`
			ol.id as outlet_id, ol.name as outlet_name,
			COUNT(o.id) as orders,
			COALESCE(SUM(o.total_amount), 0) as sales,
			COALESCE(SUM(o.net_sales), 0) as net_sales,
			COALESCE(SUM(o.tax_amount), 0) as tax,
			COALESCE(SUM(o.discount_amount), 0) as discounts,
			SUM(CASE WHEN o.is_modified THEN 1 ELSE 0 END) as modified,
			COALESCE(SUM(o.print_count - 1), 0) as reprint
		`).
		Group("ol.id, ol.name").
		Scan(&rows)

	// Compute totals
	var totals OutletStat
	totals.IsTotal = true
	totals.OutletName = "Total"

	stats := make([]OutletStat, 0, len(rows)+1)
	for _, r := range rows {
		stat := OutletStat{
			OutletID: r.OutletID, OutletName: r.OutletName,
			Orders: r.Orders, Sales: r.Sales, NetSales: r.NetSales,
			Tax: r.Tax, Discounts: r.Discounts, Modified: r.Modified, Reprint: r.Reprint,
		}
		totals.Orders += r.Orders
		totals.Sales += r.Sales
		totals.NetSales += r.NetSales
		totals.Tax += r.Tax
		totals.Discounts += r.Discounts
		totals.Modified += r.Modified
		totals.Reprint += r.Reprint
		stats = append(stats, stat)
	}

	return append([]OutletStat{totals}, stats...), nil
}

func (s *DashboardService) GetOrdersChart(userID uuid.UUID, outletID, tab string, date time.Time) (*ChartData, error) {
	// Last 7 days hourly data for the selected day
	var labels []string
	var values []float64

	for hour := 0; hour < 24; hour++ {
		t := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, time.UTC)
		tEnd := t.Add(time.Hour)
		labels = append(labels, t.Format("15:00"))

		query := s.db.Table("orders").
			Where("deleted_at IS NULL AND status != 'cancelled' AND created_at >= ? AND created_at < ?", t, tEnd)
		if outletID != "" {
			query = query.Where("outlet_id = ?", outletID)
		}

		var val float64
		switch tab {
		case "orders":
			var cnt int64
			query.Count(&cnt)
			val = float64(cnt)
		case "sales":
			query.Select("COALESCE(SUM(total_amount), 0)").Scan(&val)
		case "net_sales":
			query.Select("COALESCE(SUM(net_sales), 0)").Scan(&val)
		case "tax":
			query.Select("COALESCE(SUM(tax_amount), 0)").Scan(&val)
		case "discounts":
			query.Select("COALESCE(SUM(discount_amount), 0)").Scan(&val)
		default:
			var cnt int64
			query.Count(&cnt)
			val = float64(cnt)
		}
		values = append(values, val)
	}

	return &ChartData{Labels: labels, Values: values, Tab: tab}, nil
}

func (s *DashboardService) GetSummary(userID uuid.UUID, from, to time.Time) (*SummaryData, error) {
	var result struct {
		TotalOrders   int64
		TotalSales    float64
		NetSales      float64
		TotalTax      float64
		TotalDiscount float64
	}

	s.db.Table("orders").
		Where("deleted_at IS NULL AND status != 'cancelled' AND created_at >= ? AND created_at <= ?", from, to.Add(24*time.Hour)).
		Select(`
			COUNT(*) as total_orders,
			COALESCE(SUM(total_amount), 0) as total_sales,
			COALESCE(SUM(net_sales), 0) as net_sales,
			COALESCE(SUM(tax_amount), 0) as total_tax,
			COALESCE(SUM(discount_amount), 0) as total_discount
		`).Scan(&result)

	summary := &SummaryData{
		TotalOrders:   result.TotalOrders,
		TotalSales:    result.TotalSales,
		NetSales:      result.NetSales,
		TotalTax:      result.TotalTax,
		TotalDiscount: result.TotalDiscount,
	}
	summary.Period.From = from
	summary.Period.To = to
	return summary, nil
}

func formatPercent(p float64) string {
	return fmt.Sprintf("%.2f%%", p)
}

var _ = fmt.Sprintf // fix import
