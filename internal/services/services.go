package services

import (
	"errors"
	"fmt"

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

// added:
// Menu services

type MenuService struct{ db *gorm.DB }
type MenuItemFilter struct {
	CategoryID  string
	IsAvailable string
	IsOnline    string
}

func NewMenuService(db *gorm.DB) *MenuService { return &MenuService{db: db} }

func (s *MenuService) GetCategories(outletID uuid.UUID) ([]models.Category, error) {
	var cats []models.Category
	return cats, s.db.Where("outlet_id = ? AND deleted_at IS NULL", outletID).Order("sort_order").Find(&cats).Error
}

func (s *MenuService) CreateCategory(cat *models.Category) error {
	return s.db.Create(cat).Error
}

func (s *MenuService) GetItems(outletID uuid.UUID, filter MenuItemFilter) ([]models.MenuItem, error) {
	query := s.db.Where("outlet_id = ? AND deleted_at IS NULL", outletID)
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.IsAvailable == "true" {
		query = query.Where("is_available = true")
	} else if filter.IsAvailable == "false" {
		query = query.Where("is_available = false")
	}
	if filter.IsOnline == "true" {
		query = query.Where("is_online_active = true")
	}
	var items []models.MenuItem
	return items, query.Preload("Category").Order("sort_order").Find(&items).Error
}

func (s *MenuService) CreateItem(item *models.MenuItem) error {
	return s.db.Create(item).Error
}

func (s *MenuService) ToggleAvailability(id uuid.UUID, available bool, triggeredBy uuid.UUID) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&item).Update("is_available", available)
	s.db.Create(&models.MenuTriggerLog{
		OutletID: item.OutletID, ItemID: id,
		Action:      fmt.Sprintf("availability_set_%v", available),
		TriggeredBy: triggeredBy,
	})
	return &item, nil
}

func (s *MenuService) ToggleOnlineStatus(id uuid.UUID, online bool, platform string, triggeredBy uuid.UUID) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	oldStatus := fmt.Sprintf("%v", item.IsOnlineActive)
	s.db.Model(&item).Update("is_online_active", online)
	s.db.Create(&models.OnlineItemLog{
		OutletID: item.OutletID, ItemID: id,
		Platform:    platform,
		Action:      "online_status_change",
		OldStatus:   oldStatus,
		NewStatus:   fmt.Sprintf("%v", online),
		TriggeredBy: triggeredBy,
	})
	return &item, nil
}

func (s *MenuService) GetOutOfStockItems(outletID uuid.UUID) ([]models.MenuItem, error) {
	var items []models.MenuItem
	return items, s.db.Where("outlet_id = ? AND is_available = false AND deleted_at IS NULL", outletID).
		Preload("Category").Find(&items).Error
}

// add:
// reports
// inventory
// thirdparty
// logs
// franchise
// user sevices
// error handling
