package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

// Outlet Service

type OutletService struct{ db *gorm.DB }

func NewOutletService(db *gorm.DB) *OutletService { return &OutletService{db: db} }

func (s *OutletService) GetUserOutlets(userID uuid.UUID) ([]models.Outlet, error) {
	var outlets []models.Outlet
	return outlets, s.db.Where("deleted_at IS NULL AND is_active = true").Find(&outlets).Error
}

func (s *OutletService) GetByID(id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.Preload("Zones.Tables").First(&outlet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &outlet, nil
}

func (s *OutletService) Create(outlet *models.Outlet) error {
	var count int64
	s.db.Model(&models.Outlet{}).Where("ref_id = ?", outlet.RefID).Count(&count)
	if count > 0 {
		return errors.New("outlet with this ref_id already exists")
	}
	return s.db.Create(outlet).Error
}

func (s *OutletService) Update(id uuid.UUID, updates map[string]interface{}) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	delete(updates, "id")
	s.db.Model(&outlet).Updates(updates)
	return &outlet, nil
}

func (s *OutletService) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.Outlet{}, "id = ?", id).Error
}

func (s *OutletService) ToggleLock(id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&outlet).Update("is_locked", !outlet.IsLocked)
	return &outlet, nil
}

func (s *OutletService) GetZones(outletID uuid.UUID) ([]models.Zone, error) {
	var zones []models.Zone
	return zones, s.db.Preload("Tables").Where("outlet_id = ?", outletID).Find(&zones).Error
}

func (s *OutletService) CreateZone(outletID uuid.UUID, name string) (*models.Zone, error) {
	zone := &models.Zone{OutletID: outletID, Name: name}
	return zone, s.db.Create(zone).Error
}

// add:
// Menu services
// reports
// inventory
// thirdparty
// logs
// franchise
// user sevices
// error handling
