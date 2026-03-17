package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prayosha/go-pos-backend/internal/services"
)

// handles endpoints
type ReportsHandler struct {
	svc *services.ReportsService
}

func NewReportsHandler(svc *services.ReportsService) *ReportsHandler {
	return &ReportsHandler{svc: svc}
}

// extracts the common report query params.
func parseSalesFilter(c *gin.Context) services.SalesReportFilter {
	fromStr := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	return services.SalesReportFilter{
		OutletIDs: c.QueryArray("outlet_id"),
		From:      from,
		To:        to,
		Status:    c.DefaultQuery("status", ""),
	}
}

// GET /api/v1/reports/list
func (h *ReportsHandler) GetReportsList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": reportDefinitions()})
}

// GET /api/v1/reports/sales
func (h *ReportsHandler) GetSalesReport(c *gin.Context) {
	f := parseSalesFilter(c)
	rows, err := h.svc.GetSalesReport(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "from": f.From, "to": f.To})
}

// GET /api/v1/reports/item-wise
func (h *ReportsHandler) GetItemWiseReport(c *gin.Context) {
	rows, err := h.svc.GetItemWiseReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/category-wise
func (h *ReportsHandler) GetCategoryWiseReport(c *gin.Context) {
	rows, err := h.svc.GetCategoryWiseReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/invoices
func (h *ReportsHandler) GetInvoiceReport(c *gin.Context) {
	rows, err := h.svc.GetInvoiceReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/cancelled-orders
func (h *ReportsHandler) GetCancelledOrderReport(c *gin.Context) {
	rows, err := h.svc.GetCancelledOrderReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/discounts
func (h *ReportsHandler) GetDiscountReport(c *gin.Context) {
	rows, err := h.svc.GetDiscountReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/hourly
func (h *ReportsHandler) GetHourlyReport(c *gin.Context) {
	rows, err := h.svc.GetHourlyReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/pax-biller
func (h *ReportsHandler) GetPaxBillerReport(c *gin.Context) {
	rows, err := h.svc.GetPaxBillerReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/day-wise
func (h *ReportsHandler) GetDayWiseReport(c *gin.Context) {
	rows, err := h.svc.GetDayWiseReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/orders-master
func (h *ReportsHandler) GetOrderMasterReport(c *gin.Context) {
	rows, err := h.svc.GetOrderMasterReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/online-orders
func (h *ReportsHandler) GetOnlineOrderReport(c *gin.Context) {
	rows, err := h.svc.GetOnlineOrderReport(parseSalesFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}

// GET /api/v1/reports/chart-by-platform?date=2026-03-12&outlet_id=
func (h *ReportsHandler) GetChartByPlatform(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, _ := time.Parse("2006-01-02", dateStr)
	outletID := c.Query("outlet_id")
	rows, err := h.svc.GetDashboardChartByPlatform(outletID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "date": dateStr})
}

// the canonical list the Flutter reports_page.dart uses.
func reportDefinitions() []gin.H {
	return []gin.H{
		{"id": "all_restaurant_sales", "title": "All Restaurant Sales Report", "endpoint": "/reports/sales", "category": "All Restaurant Report"},
		{"id": "outlet_item_wise_row", "title": "Outlet-Item Wise Report (Row)", "endpoint": "/reports/item-wise", "category": "All Restaurant Report"},
		{"id": "invoice_report", "title": "Invoice Report: All Restaurants", "endpoint": "/reports/invoices", "category": "All Restaurant Report"},
		{"id": "pax_sales_biller", "title": "Pax Sales Report: Biller Wise", "endpoint": "/reports/pax-biller", "category": "All Restaurant Report"},
		{"id": "all_restaurant_day_wise", "title": "All Restaurant Report: Day Wise", "endpoint": "/reports/day-wise", "category": "All Restaurant Report"},
		{"id": "category_wise", "title": "Category Wise Report: All Restaurants", "endpoint": "/reports/category-wise", "category": "All Restaurant Report"},
		{"id": "orders_master", "title": "Orders Master Report: All Restaurants", "endpoint": "/reports/orders-master", "category": "All Restaurant Report"},
		{"id": "order_item_wise", "title": "Order Report: Item Wise All Restaurants", "endpoint": "/reports/item-wise", "category": "All Restaurant Report"},
		{"id": "cancel_order_report", "title": "Cancel Order Report: All Restaurants", "endpoint": "/reports/cancelled-orders", "category": "All Restaurant Report"},
		{"id": "discount_report", "title": "Discount Report", "endpoint": "/reports/discounts", "category": "All Restaurant Report"},
		{"id": "hourly_item_wise", "title": "All Restaurants Sales: Hourly Item Wise", "endpoint": "/reports/hourly", "category": "All Restaurant Report"},
		{"id": "online_order_report", "title": "Online Order Report: All Restaurants", "endpoint": "/reports/online-orders", "category": "All Restaurant Report"},
	}
}
