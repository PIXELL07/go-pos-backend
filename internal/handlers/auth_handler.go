package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/config"
	"github.com/prayosha/go-pos-backend/internal/auth"
	"github.com/prayosha/go-pos-backend/internal/middleware"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		EmailOrMobile string `json:"email_or_mobile" binding:"required"`
		Password      string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokenPair, err := h.authService.Login(req.EmailOrMobile, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Store refresh token
	if err := h.authService.StoreRefreshToken(user.ID, tokenPair.RefreshToken, h.cfg.JWT.RefreshExpiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          sanitizeUser(user),
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
	})
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required,min=2"`
		Email    string `json:"email" binding:"required,email"`
		Mobile   string `json:"mobile" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(req.Name, req.Email, req.Mobile, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	tokenPair, err := auth.GenerateTokenPair(user, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":          sanitizeUser(user),
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
	})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verify token is not revoked
	if err := h.authService.ValidateRefreshToken(req.RefreshToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked or expired"})
		return
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Rotate refresh token
	_ = h.authService.RevokeRefreshToken(req.RefreshToken)

	tokenPair, err := auth.GenerateTokenPair(user, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	_ = h.authService.StoreRefreshToken(user.ID, tokenPair.RefreshToken, h.cfg.JWT.RefreshExpiry)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
	})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.RefreshToken != "" {
		_ = h.authService.RevokeRefreshToken(req.RefreshToken)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, sanitizeUser(user))
}

// POST /api/v1/auth/google
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.LoginWithGoogle(req.IDToken, h.cfg)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Google authentication failed"})
		return
	}

	tokenPair, err := auth.GenerateTokenPair(user, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	_ = h.authService.StoreRefreshToken(user.ID, tokenPair.RefreshToken, h.cfg.JWT.RefreshExpiry)

	c.JSON(http.StatusOK, gin.H{
		"user":          sanitizeUser(user),
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt,
	})
}

// PUT /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// helper to sanitize user data before sending in response
func sanitizeUser(u *models.User) gin.H {
	return gin.H{
		"id":            u.ID,
		"name":          u.Name,
		"email":         u.Email,
		"mobile":        u.Mobile,
		"role":          u.Role,
		"is_active":     u.IsActive,
		"avatar_url":    u.AvatarURL,
		"last_login_at": u.LastLoginAt,
		"created_at":    u.CreatedAt,
	}
}

// dummy uuid usage suppression
var _ = uuid.Nil
var _ = time.Now
