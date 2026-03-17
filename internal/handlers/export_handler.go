package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prayosha/go-pos-backend/internal/services"
)

// ExportHandler serves CSV downloads for every report type.
type ExportHandler struct {
	svc *services.ExportService
}

func NewExportHandler(svc *services.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func writeCSV(c *gin.Context, result *services.ExportResult, err error) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Header("Content-Type", result.ContentType)
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

// GET /api/v1/export/sales?from=&to=&outlet_id=&status=
func (h *ExportHandler) SalesReport(c *gin.Context) {
	writeCSV(c, h.svc.SalesReportCSV(parseExportFilter(c)))
}

// GET /api/v1/export/item-wise
func (h *ExportHandler) ItemWise(c *gin.Context) {
	writeCSV(c, h.svc.ItemWiseCSV(parseExportFilter(c)))
}

// GET /api/v1/export/category-wise
func (h *ExportHandler) CategoryWise(c *gin.Context) {
	writeCSV(c, h.svc.CategoryWiseCSV(parseExportFilter(c)))
}

// GET /api/v1/export/invoices
func (h *ExportHandler) Invoices(c *gin.Context) {
	writeCSV(c, h.svc.InvoiceCSV(parseExportFilter(c)))
}

// GET /api/v1/export/orders-master
func (h *ExportHandler) OrdersMaster(c *gin.Context) {
	writeCSV(c, h.svc.OrderMasterCSV(parseExportFilter(c)))
}

// GET /api/v1/export/cancelled-orders
func (h *ExportHandler) CancelledOrders(c *gin.Context) {
	writeCSV(c, h.svc.CancelledOrdersCSV(parseExportFilter(c)))
}

// GET /api/v1/export/discounts
func (h *ExportHandler) Discounts(c *gin.Context) {
	writeCSV(c, h.svc.DiscountCSV(parseExportFilter(c)))
}

// GET /api/v1/export/hourly
func (h *ExportHandler) Hourly(c *gin.Context) {
	writeCSV(c, h.svc.HourlyCSV(parseExportFilter(c)))
}

// GET /api/v1/export/day-wise
func (h *ExportHandler) DayWise(c *gin.Context) {
	writeCSV(c, h.svc.DayWiseCSV(parseExportFilter(c)))
}

// GET /api/v1/export/pending-purchases?from=&to=&outlet_id=&type=
func (h *ExportHandler) PendingPurchases(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	filter := services.PurchaseFilter{
		OutletID: c.Query("outlet_id"),
		Type:     c.Query("type"),
		From:     from,
		To:       to,
		Page:     1,
		Limit:    10000,
	}
	writeCSV(c, h.svc.PendingPurchasesCSV(filter))
}

// GET /api/v1/export/store-status?outlet_id=&platform=
func (h *ExportHandler) StoreStatus(c *gin.Context) {
	filter := services.StoreStatusFilter{
		OutletID: c.Query("outlet_id"),
		Platform: c.Query("platform"),
		Page:     1,
		Limit:    10000,
	}
	writeCSV(c, h.svc.StoreStatusCSV(filter))
}

func parseExportFilter(c *gin.Context) services.SalesReportFilter {
	fromStr := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))
	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	return services.SalesReportFilter{
		OutletIDs: c.QueryArray("outlet_id"),
		From:      from,
		To:        to,
		Status:    c.Query("status"),
	}
}
