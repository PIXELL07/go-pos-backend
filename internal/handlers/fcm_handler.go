package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/services"
)

// manages device token registration for push notifications.
type FCMHandler struct {
	svc *services.FCMService
}

func NewFCMHandler(svc *services.FCMService) *FCMHandler {
	return &FCMHandler{svc: svc}
}

// POST /api/v1/push/register
// Body: {"token":"fcm-token-string","platform":"android|ios|web"}
func (h *FCMHandler) RegisterToken(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required,oneof=android ios web"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := middleware.GetUserID(c)
	if err := h.svc.RegisterToken(userID, req.Token, req.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token registered"})
}

// DELETE /api/v1/push/register
// Body: {"token":"fcm-token-string"}
func (h *FCMHandler) RemoveToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RemoveToken(req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token removed"})
}
