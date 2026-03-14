package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"

	//"github.com/prayosha/pos-backend/internal/models"
	"github.com/prayosha/go-pos-backend/pkg/logger"
	"gorm.io/gorm"
)

type FCMService struct {
	db         *gorm.DB
	serverKey  string
	projectID  string
	httpClient *http.Client
	enabled    bool
}

func NewFCMService(db *gorm.DB, serverKey, projectID string) *FCMService {
	return &FCMService{
		db:         db,
		serverKey:  serverKey,
		projectID:  projectID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		enabled:    serverKey != "" && projectID != "",
	}
}

func (s *FCMService) RegisterToken(userID uuid.UUID, token, platform string) error {
	dt := models.DeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	return s.db.
		Where(models.DeviceToken{Token: token}).
		Assign(models.DeviceToken{UserID: userID, Platform: platform}).
		FirstOrCreate(&dt).Error
}

// deletes a device token (on logout).
func (s *FCMService) RemoveToken(token string) error {
	return s.db.Where("token = ?", token).Delete(&models.DeviceToken{}).Error
}

// SendToUser sends a push notification to all registered devices of a user.
// If FCM credentials are not configured it stores the notification as
// undelivered and logs a warning — the app will see it on next poll.
func (s *FCMService) SendToUser(userID uuid.UUID, title, body, notifType string, data map[string]string) error {
	if !s.enabled {
		logger.Infof("FCM not configured — notification stored for user %s (pull only)", userID)
		return nil
	}

	var tokens []models.DeviceToken
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&tokens).Error; err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil // no devices registered yet
	}

	var tokenStrings []string
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}

	errs := s.sendMulticast(tokenStrings, title, body, data)
	if len(errs) > 0 {
		logger.Errorf("FCM send errors for user %s: %v", userID, errs)
	}
	return nil
}

func (s *FCMService) SendToOutletStaff(outletID uuid.UUID, title, body string, data map[string]string) error {
	var userIDs []uuid.UUID
	s.db.Model(&models.OutletAccess{}).
		Where("outlet_id = ? AND deleted_at IS NULL", outletID).
		Pluck("user_id", &userIDs)
	if len(userIDs) == 0 {
		return nil
	}

	var tokens []models.DeviceToken
	s.db.Where("user_id IN ? AND deleted_at IS NULL", userIDs).Find(&tokens)

	var tokenStrings []string
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}
	if len(tokenStrings) == 0 {
		return nil
	}

	s.sendMulticast(tokenStrings, title, body, data)
	return nil
}

type fcmLegacyPayload struct {
	RegistrationIDs []string          `json:"registration_ids"`
	Notification    fcmNotification   `json:"notification"`
	Data            map[string]string `json:"data,omitempty"`
	Priority        string            `json:"priority"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Sound string `json:"sound"`
}

func (s *FCMService) sendMulticast(tokens []string, title, body string, data map[string]string) []error {
	// FCM supports max 500 tokens per multicast send
	const batchSize = 500
	var allErrors []error

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		payload := fcmLegacyPayload{
			RegistrationIDs: batch,
			Notification: fcmNotification{
				Title: title,
				Body:  body,
				Sound: "default",
			},
			Data:     data,
			Priority: "high",
		}

		b, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(b))
		req.Header.Set("Authorization", "key="+s.serverKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("fcm request: %w", err))
			continue
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			allErrors = append(allErrors, fmt.Errorf("fcm http %d: %s", resp.StatusCode, string(respBody)))
		} else {
			logger.Debugf("FCM sent to %d tokens: %s", len(batch), string(respBody))
		}
	}
	return allErrors
}

// NotificationWithPush wraps NotificationService to add FCM.
type NotificationWithPush struct {
	*NotificationService
	fcm *FCMService
}

func NewNotificationWithPush(db *gorm.DB, fcm *FCMService) *NotificationWithPush {
	return &NotificationWithPush{
		NotificationService: NewNotificationService(db),
		fcm:                 fcm,
	}
}

func (s *NotificationWithPush) Push(userID uuid.UUID, title, body, notifType string, data map[string]string) error {
	n := &models.Notification{
		UserID: userID,
		Title:  title,
		Body:   body,
		Type:   notifType,
	}
	if err := s.Create(n); err != nil {
		return err
	}

	go func() {
		if err := s.fcm.SendToUser(userID, title, body, notifType, data); err != nil {
			logger.Errorf("FCM push failed for user %s: %v", userID, err)
			return
		}
		// Mark FCM as sent
		s.db.Model(&models.Notification{}).Where("id = ?", n.ID).Update("fcm_sent", true)
	}()

	return nil
}
