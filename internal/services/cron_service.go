package services

import (
	"time"

	"github.com/prayosha/go-pos-backend/pkg/logger"
)

// this holds references to all services that need periodic tasks.
type CronService struct {
	menu         *MenuService
	refreshToken *RefreshTokenService
	storeStatus  *StoreStatusService
}

type RefreshTokenService struct {
	authSvc *AuthService
}

func NewCronService(menu *MenuService, auth *AuthService, store *StoreStatusService) *CronService {
	return &CronService{
		menu:         menu,
		refreshToken: &RefreshTokenService{authSvc: auth},
		storeStatus:  store,
	}
}

// Start launches all background goroutines. Call once after server starts.
func (c *CronService) Start() {
	go c.runEvery(5*time.Minute, "restore-expired-items", func() {
		n, err := c.menu.RestoreExpiredItems()
		if err != nil {
			logger.Errorf("restore expired items: %v", err)
		} else if n > 0 {
			logger.Infof("restored %d offline items", n)
		}
	})

	go c.runEvery(1*time.Hour, "purge-refresh-tokens", func() {
		if err := c.refreshToken.authSvc.PurgeExpiredTokens(); err != nil {
			logger.Errorf("purge refresh tokens: %v", err)
		}
	})
}

func (c *CronService) runEvery(d time.Duration, name string, fn func()) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for range ticker.C {
		logger.Debugf("cron: %s running", name)
		fn()
	}
}
