package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
)

// MenuHandler

type MenuHandler struct{ menuService *services.MenuService }

func NewMenuHandler(svc *services.MenuService) *MenuHandler {
	return &MenuHandler{menuService: svc}
}

// GET /api/v1/menu/categories?outlet_id=
func (h *MenuHandler) GetCategories(c *gin.Context) {
	outletID, err := uuid.Parse(c.Query("outlet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
		return
	}
	cats, err := h.menuService.GetCategories(outletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cats})
}

// POST /api/v1/menu/categories
func (h *MenuHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		OutletID    string `json:"outlet_id" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outletID, _ := uuid.Parse(req.OutletID)
	cat := &models.Category{
		Name: req.Name, OutletID: outletID,
		Description: req.Description, SortOrder: req.SortOrder, IsActive: true,
	}
	if err := h.menuService.CreateCategory(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// PUT /api/v1/menu/categories/:id
func (h *MenuHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name      string `json:"name"`
		IsActive  *bool  `json:"is_active"`
		SortOrder *int   `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	cat, err := h.menuService.UpdateCategory(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

// DELETE /api/v1/menu/categories/:id
func (h *MenuHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.menuService.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GET /api/v1/menu/items?outlet_id=&category_id=&is_available=&is_online=&search=
func (h *MenuHandler) GetItems(c *gin.Context) {
	outletID, err := uuid.Parse(c.Query("outlet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
		return
	}
	items, err := h.menuService.GetItems(outletID, services.MenuItemFilter{
		CategoryID:  c.Query("category_id"),
		IsAvailable: c.Query("is_available"),
		IsOnline:    c.Query("is_online"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// POST /api/v1/menu/items
func (h *MenuHandler) CreateItem(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		CategoryID  string  `json:"category_id" binding:"required"`
		OutletID    string  `json:"outlet_id" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		TaxRate     float64 `json:"tax_rate"`
		IsVeg       bool    `json:"is_veg"`
		Description string  `json:"description"`
		ImageURL    string  `json:"image_url"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catID, _ := uuid.Parse(req.CategoryID)
	outletID, _ := uuid.Parse(req.OutletID)
	item := &models.MenuItem{
		Name: req.Name, CategoryID: catID, OutletID: outletID,
		Price: req.Price, TaxRate: req.TaxRate, IsVeg: req.IsVeg,
		Description: req.Description, ImageURL: req.ImageURL,
		SortOrder: req.SortOrder, IsAvailable: true,
	}
	if err := h.menuService.CreateItem(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PUT /api/v1/menu/items/:id
func (h *MenuHandler) UpdateItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        string  `json:"name"`
		Price       float64 `json:"price"`
		TaxRate     float64 `json:"tax_rate"`
		IsVeg       *bool   `json:"is_veg"`
		Description string  `json:"description"`
		ImageURL    string  `json:"image_url"`
		SortOrder   *int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.TaxRate > 0 {
		updates["tax_rate"] = req.TaxRate
	}
	if req.IsVeg != nil {
		updates["is_veg"] = *req.IsVeg
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	item, err := h.menuService.UpdateItem(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DELETE /api/v1/menu/items/:id
func (h *MenuHandler) DeleteItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.menuService.DeleteItem(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// PATCH /api/v1/menu/items/:id/availability
// Body: {"is_available": false, "offline_duration_minutes": 60}
func (h *MenuHandler) ToggleAvailability(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	var req struct {
		IsAvailable            bool `json:"is_available"`
		OfflineDurationMinutes int  `json:"offline_duration_minutes"` // 0 = permanent
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.menuService.SetAvailabilityWithDuration(id, req.IsAvailable, req.OfflineDurationMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// PATCH /api/v1/menu/items/:id/online
func (h *MenuHandler) ToggleOnlineStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}
	var req struct {
		IsOnlineActive bool   `json:"is_online_active"`
		Platform       string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := middleware.GetUserID(c)
	item, err := h.menuService.ToggleOnlineStatus(id, req.IsOnlineActive, req.Platform, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// GET /api/v1/menu/out-of-stock?outlet_id=&category_id=&search=
func (h *MenuHandler) GetOutOfStockItems(c *gin.Context) {
	outletID, err := uuid.Parse(c.Query("outlet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
		return
	}
	items, err := h.menuService.GetOutOfStockFiltered(outletID, c.Query("category_id"), c.Query("search"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch out-of-stock items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
