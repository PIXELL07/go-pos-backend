package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
)

type OutletHandler struct {
	outletService *services.OutletService
}

func NewOutletHandler(outletService *services.OutletService) *OutletHandler {
	return &OutletHandler{outletService: outletService}
}

// GET /api/v1/outlets
func (h *OutletHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	outlets, err := h.outletService.GetUserOutlets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch outlets"})
		return
	}
	c.JSON(http.StatusOK, outlets)
}

// GET /api/v1/outlets/:id
func (h *OutletHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	outlet, err := h.outletService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Outlet not found"})
		return
	}
	c.JSON(http.StatusOK, outlet)
}

// POST /api/v1/outlets
func (h *OutletHandler) Create(c *gin.Context) {
	var req struct {
		Name      string            `json:"name" binding:"required"`
		RefID     string            `json:"ref_id" binding:"required"`
		Type      models.OutletType `json:"type"`
		Address   string            `json:"address"`
		City      string            `json:"city"`
		State     string            `json:"state"`
		PinCode   string            `json:"pin_code"`
		Phone     string            `json:"phone"`
		GSTNumber string            `json:"gst_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outlet := &models.Outlet{
		Name: req.Name, RefID: req.RefID, Type: req.Type,
		Address: req.Address, City: req.City, State: req.State,
		PinCode: req.PinCode, Phone: req.Phone, GSTNumber: req.GSTNumber,
	}
	if err := h.outletService.Create(outlet); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, outlet)
}

// PUT /api/v1/outlets/:id
func (h *OutletHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outlet, err := h.outletService.Update(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update outlet"})
		return
	}
	c.JSON(http.StatusOK, outlet)
}

// DELETE /api/v1/outlets/:id
func (h *OutletHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	if err := h.outletService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete outlet"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Outlet deleted successfully"})
}

// PATCH /api/v1/outlets/:id/lock
func (h *OutletHandler) ToggleLock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	outlet, err := h.outletService.ToggleLock(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle lock"})
		return
	}
	c.JSON(http.StatusOK, outlet)
}

// GET /api/v1/outlets/:id/zones
func (h *OutletHandler) GetZones(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	zones, err := h.outletService.GetZones(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch zones"})
		return
	}
	c.JSON(http.StatusOK, zones)
}

// POST /api/v1/outlets/:id/zones
func (h *OutletHandler) CreateZone(c *gin.Context) {
	outletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid outlet ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.outletService.CreateZone(outletID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zone"})
		return
	}
	c.JSON(http.StatusCreated, zone)
}

// GET /api/v1/outlets/types
func (h *OutletHandler) GetTypes(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{"value": models.OutletTypeDineIn, "label": "Dine In"},
		{"value": models.OutletTypeTakeaway, "label": "Takeaway"},
		{"value": models.OutletTypeDelivery, "label": "Delivery"},
		{"value": models.OutletTypeCloud, "label": "Cloud Kitchen"},
	})
}

// PUT /api/v1/outlets/zones/:zone_id
func (h *OutletHandler) UpdateZone(c *gin.Context) {
	id, err := uuid.Parse(c.Param("zone_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	zone, err := h.outletService.UpdateZone(id, map[string]interface{}{"name": req.Name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zone)
}

// DELETE /api/v1/outlets/zones/:zone_id
func (h *OutletHandler) DeleteZone(c *gin.Context) {
	id, err := uuid.Parse(c.Param("zone_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	if err := h.outletService.DeleteZone(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "zone deleted"})
}

// POST /api/v1/outlets/zones/:zone_id/tables
func (h *OutletHandler) CreateTable(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("zone_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}
	var req struct {
		Name     string `json:"name" binding:"required"`
		Capacity int    `json:"capacity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Capacity == 0 {
		req.Capacity = 4
	}
	table, err := h.outletService.CreateTable(zoneID, req.Name, req.Capacity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, table)
}

// PUT /api/v1/outlets/tables/:table_id
func (h *OutletHandler) UpdateTable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("table_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table id"})
		return
	}
	var req struct {
		Name       string `json:"name"`
		Capacity   *int   `json:"capacity"`
		IsOccupied *bool  `json:"is_occupied"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Capacity != nil {
		updates["capacity"] = *req.Capacity
	}
	if req.IsOccupied != nil {
		updates["is_occupied"] = *req.IsOccupied
	}
	table, err := h.outletService.UpdateTable(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, table)
}

// DELETE /api/v1/outlets/tables/:table_id
func (h *OutletHandler) DeleteTable(c *gin.Context) {
	id, err := uuid.Parse(c.Param("table_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table id"})
		return
	}
	if err := h.outletService.DeleteTable(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "table deleted"})
}
