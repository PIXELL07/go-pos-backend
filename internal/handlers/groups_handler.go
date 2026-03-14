package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
)

type UserGroupHandler struct {
	svc *services.UserGroupService
}

func NewUserGroupHandler(svc *services.UserGroupService) *UserGroupHandler {
	return &UserGroupHandler{svc: svc}
}

// GET /api/v1/groups?type=admin|biller&outlet_id=&name=&page=&limit=
func (h *UserGroupHandler) List(c *gin.Context) {
	filter := services.GroupFilter{
		Type:     c.Query("type"),
		OutletID: c.Query("outlet_id"),
		Name:     c.Query("name"),
		Page:     parseIntDefault(c.DefaultQuery("page", "1"), 1),
		Limit:    parseIntDefault(c.DefaultQuery("limit", "20"), 20),
	}
	groups, total, err := h.svc.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  groups,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// GET /api/v1/groups/:id
func (h *UserGroupHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	group, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// POST /api/v1/groups
func (h *UserGroupHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required,oneof=admin biller"`
		OutletID string `json:"outlet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outletID, err := uuid.Parse(req.OutletID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet_id"})
		return
	}
	group := &models.UserGroup{
		Name:     req.Name,
		Type:     models.UserGroupType(req.Type),
		OutletID: outletID,
	}
	if err := h.svc.Create(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

// PUT /api/v1/groups/:id
func (h *UserGroupHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
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
	group, err := h.svc.Update(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// DELETE /api/v1/groups/:id
func (h *UserGroupHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// POST /api/v1/groups/:id/members
func (h *UserGroupHandler) AddMember(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		UserID     string `json:"user_id" binding:"required"`
		BillerType string `json:"biller_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	member, err := h.svc.AddMember(groupID, userID, models.UserBillerType(req.BillerType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

// DELETE /api/v1/groups/:id/members/:user_id
func (h *UserGroupHandler) RemoveMember(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	if err := h.svc.RemoveMember(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// POST /api/v1/groups/:id/bulk-status  body: {"is_active": true}
func (h *UserGroupHandler) BulkSetStatus(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BulkSetMemberStatus(groupID, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

type CloudAccessHandler struct {
	svc *services.UserService
}

func NewCloudAccessHandler(svc *services.UserService) *CloudAccessHandler {
	return &CloudAccessHandler{svc: svc}
}

// GET /api/v1/cloud-access?name=&email=&type=&status=active|inactive&page=&limit=
func (h *CloudAccessHandler) List(c *gin.Context) {
	filter := services.CloudUserFilter{
		Name:     c.Query("name"),
		Email:    c.Query("email"),
		UserType: c.Query("type"),
		Status:   c.DefaultQuery("status", "active"),
		Page:     parseIntDefault(c.DefaultQuery("page", "1"), 1),
		Limit:    parseIntDefault(c.DefaultQuery("limit", "20"), 20),
	}
	users, total, err := h.svc.ListCloudUsers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

// PATCH /api/v1/cloud-access/bulk-status   body: {"user_ids":["..."],"is_active":true}
func (h *CloudAccessHandler) BulkSetStatus(c *gin.Context) {
	var req struct {
		UserIDs  []string `json:"user_ids" binding:"required"`
		IsActive bool     `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var ids []uuid.UUID
	for _, s := range req.UserIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id: " + s})
			return
		}
		ids = append(ids, id)
	}
	if err := h.svc.BulkSetActiveStatus(ids, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated", "count": len(ids)})
}

type OutletTypeHandler struct {
	svc *services.OutletService
}

func NewOutletTypeHandler(svc *services.OutletService) *OutletTypeHandler {
	return &OutletTypeHandler{svc: svc}
}

// GET /api/v1/outlet-types   — list all franchise types + outlets with their current type
func (h *OutletTypeHandler) List(c *gin.Context) {
	outletID := c.Query("outlet_id")
	outlets, err := h.svc.ListWithFranchiseTypes(outletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch outlet types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"franchise_types": []gin.H{
			{"value": "COFO", "label": "COFO - Company Owned Franchisee Operated"},
			{"value": "FOFO", "label": "FOFO - Franchisee Owned Franchisee Operated"},
			{"value": "COCO", "label": "COCO - Company Owned Company Operated"},
			{"value": "FOCO", "label": "FOCO - Franchisee Owned Company Operated"},
		},
		"outlets": outlets,
	})
}

// PUT /api/v1/outlet-types/:id    body: {"franchise_type":"COFO"}
func (h *OutletTypeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outlet id"})
		return
	}
	var req struct {
		FranchiseType string `json:"franchise_type" binding:"required,oneof=COFO FOFO COCO FOCO"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outlet, err := h.svc.UpdateFranchiseType(id, models.FranchiseType(req.FranchiseType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, outlet)
}

type StoreStatusHandler struct {
	svc *services.StoreStatusService
}

func NewStoreStatusHandler(svc *services.StoreStatusService) *StoreStatusHandler {
	return &StoreStatusHandler{svc: svc}
}

// GET /api/v1/store-status?outlet_id=&platform=&offline_since_minutes=&page=&limit=
func (h *StoreStatusHandler) List(c *gin.Context) {
	filter := services.StoreStatusFilter{
		OutletID:            c.Query("outlet_id"),
		Platform:            c.Query("platform"),
		OfflineSinceMinutes: parseIntDefault(c.Query("offline_since_minutes"), 0),
		Page:                parseIntDefault(c.DefaultQuery("page", "1"), 1),
		Limit:               parseIntDefault(c.DefaultQuery("limit", "50"), 50),
	}
	results, total, err := h.svc.GetStatus(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": results, "total": total})
}

// POST /api/v1/store-status/refresh  — manually triggers a status re-check
func (h *StoreStatusHandler) Refresh(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	outletID := c.Query("outlet_id")
	count, err := h.svc.RefreshAll(outletID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "refreshed", "records_updated": count})
}

// GET /api/v1/store-status/history?outlet_id=&platform=&from=&to=
func (h *StoreStatusHandler) History(c *gin.Context) {
	outlet := c.Query("outlet_id")
	platform := c.Query("platform")
	from := c.DefaultQuery("from", "")
	to := c.DefaultQuery("to", "")
	logs, err := h.svc.GetHistory(outlet, platform, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

type KOTHandler struct {
	svc *services.KOTService
}

func NewKOTHandler(svc *services.KOTService) *KOTHandler {
	return &KOTHandler{svc: svc}
}

// GET /api/v1/kots?order_id=&outlet_id=&status=
func (h *KOTHandler) List(c *gin.Context) {
	filter := map[string]string{
		"order_id":  c.Query("order_id"),
		"outlet_id": c.Query("outlet_id"),
		"status":    c.Query("status"),
	}
	kots, err := h.svc.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch KOTs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": kots})
}

// POST /api/v1/kots
func (h *KOTHandler) Create(c *gin.Context) {
	var req services.CreateKOTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := middleware.GetUserID(c)
	kot, err := h.svc.Create(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, kot)
}

// PATCH /api/v1/kots/:id/status
func (h *KOTHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid KOT id"})
		return
	}
	var req struct {
		Status string `json:"status" binding:"required,oneof=pending preparing ready cancelled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	kot, err := h.svc.UpdateStatus(id, models.KOTStatus(req.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kot)
}

// POST /api/v1/kots/:id/print
func (h *KOTHandler) MarkPrinted(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid KOT id"})
		return
	}
	kot, err := h.svc.MarkPrinted(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kot)
}

type UploadHandler struct {
	svc *services.UploadService
}

func NewUploadHandler(svc *services.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// POST /api/v1/uploads?owner_type=menu_item&owner_id=...
func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	ownerType := c.DefaultQuery("owner_type", "")
	ownerIDStr := c.DefaultQuery("owner_id", "")

	var ownerID *uuid.UUID
	if ownerIDStr != "" {
		if id, err := uuid.Parse(ownerIDStr); err == nil {
			ownerID = &id
		}
	}

	upload, err := h.svc.Save(file, header, ownerType, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, upload)
}
