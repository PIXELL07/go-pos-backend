package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/config"
	"github.com/prayosha/go-pos-backend/internal/auth"
	"github.com/prayosha/go-pos-backend/internal/models"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db    *gorm.DB
	redis *redis.Client
	cfg   *config.Config
}

func NewAuthService(db *gorm.DB, redis *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{db: db, redis: redis, cfg: cfg}
}

func (s *AuthService) Register(name, email, mobile, password string) (*models.User, error) {
	// Check existing
	var count int64
	s.db.Model(&models.User{}).Where("email = ? OR mobile = ?", email, mobile).Count(&count)
	if count > 0 {
		return nil, errors.New("user with this email or mobile already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name: name, Email: email, Mobile: mobile,
		PasswordHash: string(hash), Role: models.RoleBiller, IsActive: true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *AuthService) Login(emailOrMobile, password string) (*models.User, *auth.TokenPair, error) {
	var user models.User
	if err := s.db.Where("email = ? OR mobile = ?", emailOrMobile, emailOrMobile).
		First(&user).Error; err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	tokenPair, err := auth.GenerateTokenPair(&user, s.cfg)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	return &user, tokenPair, nil
}

func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) StoreRefreshToken(userID uuid.UUID, token string, expiry time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.redis.Set(context.Background(), key, userID.String(), expiry).Err()
}

func (s *AuthService) ValidateRefreshToken(token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	_, err := s.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return errors.New("token not found or expired")
	}
	return err
}

func (s *AuthService) RevokeRefreshToken(token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return s.redis.Del(context.Background(), key).Err()
}

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPass, newPass string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPass)); err != nil {
		return errors.New("incorrect current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *AuthService) LoginWithGoogle(idToken string, cfg *config.Config) (*models.User, error) {
	// In production: validate Google ID token via Google API
	// For now, decode payload and upsert user
	// Use: https://oauth2.googleapis.com/tokeninfo?id_token=<TOKEN>
	// This is a placeholder for the real Google verification
	return nil, errors.New("google login requires valid Google client credentials")
}

// to delete tokens that have expired or been revoked.
func (s *AuthService) PurgeExpiredTokens() error {
	return s.db.Where("expires_at < NOW() OR is_revoked = true").
		Delete(&models.RefreshToken{}).Error
}
