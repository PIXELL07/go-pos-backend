package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
)

// marks an item unavailable for a fixed duration.
// durationMinutes == 0 means permanently unavailable until manually re-enabled.
func (s *MenuService) SetAvailabilityWithDuration(itemID uuid.UUID, available bool, durationMinutes int) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := s.db.First(&item, "id = ?", itemID).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"is_available":  available,
		"offline_until": nil, // clear by default
	}

	if !available && durationMinutes > 0 {
		until := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		updates["offline_until"] = until
	}

	if err := s.db.Model(&item).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Category").First(&item, "id = ?", itemID)
	return &item, nil
}

// sets is_available = true for any items whose offline_until has passed.
// This is called by the background cron job.
func (s *MenuService) RestoreExpiredItems() (int64, error) {
	result := s.db.Model(&models.MenuItem{}).
		Where("is_available = false AND offline_until IS NOT NULL AND offline_until <= NOW()").
		Updates(map[string]interface{}{
			"is_available":  true,
			"offline_until": nil,
		})
	return result.RowsAffected, result.Error
}

// returns unavailable items with optional category/search filter.
func (s *MenuService) GetOutOfStockFiltered(outletID uuid.UUID, categoryID, search string) ([]models.MenuItem, error) {
	q := s.db.Preload("Category").
		Where("outlet_id = ? AND is_available = false AND deleted_at IS NULL", outletID)

	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}
	if search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}

	var items []models.MenuItem
	return items, q.Order("name ASC").Find(&items).Error
}

func (s *MenuService) UpdateCategory(id uuid.UUID, updates map[string]interface{}) (*models.Category, error) {
	var cat models.Category
	if err := s.db.First(&cat, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&cat).Updates(updates)
	return &cat, nil
}

func (s *MenuService) DeleteCategory(id uuid.UUID) error {
	return s.db.Delete(&models.Category{}, "id = ?", id).Error
}

func (s *MenuService) UpdateItem(id uuid.UUID, updates map[string]interface{}) (*models.MenuItem, error) {
	var item models.MenuItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&item).Updates(updates)
	s.db.Preload("Category").First(&item, "id = ?", id)
	return &item, nil
}

// soft-deletes a menu item.
func (s *MenuService) DeleteItem(id uuid.UUID) error {
	return s.db.Delete(&models.MenuItem{}, "id = ?", id).Error
}
