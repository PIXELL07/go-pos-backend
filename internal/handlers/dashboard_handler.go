package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/services"
)

type DashboardHandler struct {
	dashService *services.DashboardService
}

func NewDashboardHandler(dashService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashService: dashService}
}

// GET /api/v1/dashboard/stats?date=2026-03-13&outlet_id=...
func (h *DashboardHandler) GetStats(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	outletID := c.Query("outlet_id") // empty = all outlets

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	stats, err := h.dashService.GetDashboardStats(userID, outletID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/dashboard/outlet-stats?date=2026-03-13
func (h *DashboardHandler) GetOutletStats(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	stats, err := h.dashService.GetOutletStats(userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch outlet stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/dashboard/orders-chart?date=2026-03-13&outlet_id=...&tab=orders
func (h *DashboardHandler) GetOrdersChart(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	outletID := c.Query("outlet_id")
	tab := c.DefaultQuery("tab", "orders") // orders | sales | net_sales | tax | discounts

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	data, err := h.dashService.GetOrdersChart(userID, outletID, tab, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chart data"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GET /api/v1/dashboard/summary?from=2026-03-01&to=2026-03-13
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	fromStr := c.DefaultQuery("from", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date"})
		return
	}

	summary, err := h.dashService.GetSummary(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
