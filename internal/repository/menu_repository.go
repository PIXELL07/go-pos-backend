package repository

import (
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	*Repository[models.Category]
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{Repository: NewRepository[models.Category](db)}
}

func (r *CategoryRepository) FindByOutlet(outletID uuid.UUID) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Where("outlet_id = ? AND deleted_at IS NULL", outletID).
		Order("sort_order ASC, name ASC").Find(&cats).Error
	return cats, err
}

func (r *CategoryRepository) FindByOutletWithItems(outletID uuid.UUID) ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Preload("Items", "deleted_at IS NULL AND is_available = true").
		Where("outlet_id = ? AND deleted_at IS NULL AND is_active = true", outletID).
		Order("sort_order ASC").Find(&cats).Error
	return cats, err
}

func (r *CategoryRepository) SetActive(id uuid.UUID, active bool) error {
	return r.db.Model(&models.Category{}).Where("id = ?", id).
		Update("is_active", active).Error
}

type MenuItemRepository struct {
	*Repository[models.MenuItem]
}

func NewMenuItemRepository(db *gorm.DB) *MenuItemRepository {
	return &MenuItemRepository{Repository: NewRepository[models.MenuItem](db)}
}

type MenuItemQueryFilter struct {
	CategoryID  string
	IsAvailable *bool
	IsOnline    *bool
	Search      string
}

func (r *MenuItemRepository) FindByOutlet(outletID uuid.UUID, f MenuItemQueryFilter) ([]models.MenuItem, error) {
	q := r.db.Preload("Category").
		Where("outlet_id = ? AND deleted_at IS NULL", outletID)

	if f.CategoryID != "" {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.IsAvailable != nil {
		q = q.Where("is_available = ?", *f.IsAvailable)
	}
	if f.IsOnline != nil {
		q = q.Where("is_online_active = ?", *f.IsOnline)
	}
	if f.Search != "" {
		q = q.Where("name ILIKE ?", "%"+f.Search+"%")
	}

	var items []models.MenuItem
	return items, q.Order("sort_order ASC, name ASC").Find(&items).Error
}

// returns items that are marked unavailable.
func (r *MenuItemRepository) FindOutOfStock(outletID uuid.UUID) ([]models.MenuItem, error) {
	var items []models.MenuItem
	err := r.db.Preload("Category").
		Where("outlet_id = ? AND is_available = false AND deleted_at IS NULL", outletID).
		Order("name ASC").Find(&items).Error
	return items, err
}

func (r *MenuItemRepository) SetAvailability(id uuid.UUID, available bool) error {
	return r.db.Model(&models.MenuItem{}).Where("id = ?", id).
		Update("is_available", available).Error
}

func (r *MenuItemRepository) SetOnlineStatus(id uuid.UUID, online bool) error {
	return r.db.Model(&models.MenuItem{}).Where("id = ?", id).
		Update("is_online_active", online).Error
}

// returns a batch of items by their UUIDs (used when building orders).
func (r *MenuItemRepository) FindByIDs(ids []uuid.UUID) ([]models.MenuItem, error) {
	var items []models.MenuItem
	return items, r.db.Where("id IN ? AND deleted_at IS NULL", ids).Find(&items).Error
}
