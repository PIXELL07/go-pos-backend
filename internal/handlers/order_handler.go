package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// GET /api/v1/orders?outlet_id=&status=&source=&from=&to=&page=&limit=
func (h *OrderHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	filter := services.OrderFilter{
		OutletID: c.Query("outlet_id"),
		Status:   c.Query("status"),
		Source:   c.Query("source"),
		OrderNo:  c.Query("order_no"),
		Page:     parseIntDefault(c.Query("page"), 1),
		Limit:    parseIntDefault(c.Query("limit"), 20),
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr != "" {
		t, _ := time.Parse("2006-01-02", fromStr)
		filter.From = &t
	}
	if toStr != "" {
		t, _ := time.Parse("2006-01-02", toStr)
		filter.To = &t
	}

	result, err := h.orderService.ListOrders(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/v1/orders/running?outlet_id=
func (h *OrderHandler) GetRunningOrders(c *gin.Context) {
	outletID := c.Query("outlet_id")
	userID, _ := middleware.GetUserID(c)

	orders, err := h.orderService.GetRunningOrders(userID, outletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch running orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GET /api/v1/orders/online?outlet_id=&platform=&status=&from=&to=&record_type=
func (h *OrderHandler) GetOnlineOrders(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	filter := services.OnlineOrderFilter{
		OutletID:   c.Query("outlet_id"),
		Platform:   c.Query("platform"),
		Status:     c.Query("status"),
		OrderNo:    c.Query("order_no"),
		RecordType: c.DefaultQuery("record_type", "last_2_days"),
		Page:       parseIntDefault(c.Query("page"), 1),
		Limit:      parseIntDefault(c.Query("limit"), 20),
	}

	if fromStr := c.Query("from"); fromStr != "" {
		t, _ := time.Parse("2006-01-02", fromStr)
		filter.From = &t
	}
	if toStr := c.Query("to"); toStr != "" {
		t, _ := time.Parse("2006-01-02", toStr)
		filter.To = &t
	}

	result, err := h.orderService.GetOnlineOrders(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch online orders"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/v1/orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// POST /api/v1/orders
func (h *OrderHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		OutletID        string             `json:"outlet_id" binding:"required"`
		TableID         *string            `json:"table_id"`
		Type            models.OrderType   `json:"type"`
		Source          models.OrderSource `json:"source"`
		CustomerName    string             `json:"customer_name"`
		CustomerPhone   string             `json:"customer_phone"`
		Pax             int                `json:"pax"`
		Items           []CreateOrderItem  `json:"items" binding:"required,min=1"`
		Payments        []CreatePayment    `json:"payments"`
		DiscountPercent float64            `json:"discount_percent"`
		DiscountAmount  float64            `json:"discount_amount"`
		DeliveryCharge  float64            `json:"delivery_charge"`
		ContainerCharge float64            `json:"container_charge"`
		ServiceCharge   float64            `json:"service_charge"`
		RoundOff        float64            `json:"round_off"`
		Notes           string             `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outletID, err := uuid.Parse(req.OutletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	order, err := h.orderService.CreateOrder(userID, outletID, req.TableID, services.CreateOrderRequest{
		Type:            req.Type,
		Source:          req.Source,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		Pax:             req.Pax,
		Items:           toServiceItems(req.Items),
		Payments:        toServicePayments(req.Payments),
		DiscountPercent: req.DiscountPercent,
		DiscountAmount:  req.DiscountAmount,
		DeliveryCharge:  req.DeliveryCharge,
		ContainerCharge: req.ContainerCharge,
		ServiceCharge:   req.ServiceCharge,
		RoundOff:        req.RoundOff,
		Notes:           req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// PATCH /api/v1/orders/:id/status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateStatus(id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// PATCH /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	order, err := h.orderService.CancelOrder(id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// POST /api/v1/orders/:id/print
func (h *OrderHandler) MarkPrinted(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderService.MarkPrinted(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark printed"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GET /api/v1/orders/platforms
func (h *OrderHandler) GetPlatforms(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{"id": "all", "name": "All"},
		{"id": string(models.OrderSourceZomato), "name": "Zomato"},
		{"id": string(models.OrderSourceSwiggy), "name": "Swiggy"},
		{"id": string(models.OrderSourceFoodPanda), "name": "FoodPanda"},
		{"id": string(models.OrderSourceUberEats), "name": "Uber Eats"},
		{"id": string(models.OrderSourceDunzo), "name": "Dunzo"},
		{"id": string(models.OrderSourceWebsite), "name": "Home Website"},
	})
}

// Request DTOs
type CreateOrderItem struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Notes      string `json:"notes"`
}

type CreatePayment struct {
	Method models.PaymentMethod `json:"method" binding:"required"`
	Amount float64              `json:"amount" binding:"required"`
	RefNo  string               `json:"ref_no"`
}

func toServiceItems(items []CreateOrderItem) []services.OrderItemRequest {
	result := make([]services.OrderItemRequest, len(items))
	for i, it := range items {
		result[i] = services.OrderItemRequest{
			MenuItemID: it.MenuItemID,
			Quantity:   it.Quantity,
			Notes:      it.Notes,
		}
	}
	return result
}

func toServicePayments(payments []CreatePayment) []services.PaymentRequest {
	result := make([]services.PaymentRequest, len(payments))
	for i, p := range payments {
		result[i] = services.PaymentRequest{
			Method: p.Method,
			Amount: p.Amount,
			RefNo:  p.RefNo,
		}
	}
	return result
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var n int
	if _, err := parseIntStr(s, &n); err != nil {
		return def
	}
	return n
}

func parseIntStr(s string, n *int) (interface{}, error) {
	_, err := fmt.Sscanf(s, "%d", n)
	return nil, err
}

var _ = fmt.Sprintf
