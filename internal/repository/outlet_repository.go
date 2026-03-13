package repository

import (
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type OutletRepository struct {
	*Repository[models.Outlet]
}

func NewOutletRepository(db *gorm.DB) *OutletRepository {
	return &OutletRepository{Repository: NewRepository[models.Outlet](db)}
}

// FindAll returns every active, non-deleted outlet.
func (r *OutletRepository) FindAll() ([]models.Outlet, error) {
	var outlets []models.Outlet
	err := r.db.Where("deleted_at IS NULL AND is_active = true").
		Order("name ASC").Find(&outlets).Error
	return outlets, err
}

// FindByIDWithZones returns an outlet and eagerly loads its zones and tables.
func (r *OutletRepository) FindByIDWithZones(id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	err := r.db.Preload("Zones.Tables").First(&outlet, "id = ?", id).Error
	return &outlet, err
}

// FindByRefID returns the outlet with the given ref_id (used for dedup checks).
func (r *OutletRepository) FindByRefID(refID string) (*models.Outlet, error) {
	var outlet models.Outlet
	err := r.db.Where("ref_id = ? AND deleted_at IS NULL", refID).First(&outlet).Error
	return &outlet, err
}

// ExistsByRefID returns true when an outlet with that ref_id already exists.
func (r *OutletRepository) ExistsByRefID(refID string) (bool, error) {
	return r.Exists("ref_id = ? AND deleted_at IS NULL", refID)
}

// FindByFranchise returns all outlets belonging to the given franchise.
func (r *OutletRepository) FindByFranchise(franchiseID uuid.UUID) ([]models.Outlet, error) {
	var outlets []models.Outlet
	err := r.db.Where("franchise_id = ? AND deleted_at IS NULL", franchiseID).
		Order("name ASC").Find(&outlets).Error
	return outlets, err
}

// SetLocked toggles the is_locked flag.
func (r *OutletRepository) SetLocked(id uuid.UUID, locked bool) error {
	return r.db.Model(&models.Outlet{}).Where("id = ?", id).
		Update("is_locked", locked).Error
}

// CountActive returns the number of active outlets.
func (r *OutletRepository) CountActive() (int64, error) {
	return r.Count("deleted_at IS NULL AND is_active = true")
}

// ─── Zone / Table helpers ────────────────────────────────────────────────────

type ZoneRepository struct {
	*Repository[models.Zone]
}

func NewZoneRepository(db *gorm.DB) *ZoneRepository {
	return &ZoneRepository{Repository: NewRepository[models.Zone](db)}
}

func (r *ZoneRepository) FindByOutlet(outletID uuid.UUID) ([]models.Zone, error) {
	var zones []models.Zone
	err := r.db.Preload("Tables").Where("outlet_id = ? AND deleted_at IS NULL", outletID).
		Order("name ASC").Find(&zones).Error
	return zones, err
}

type TableRepository struct {
	*Repository[models.Table]
}

func NewTableRepository(db *gorm.DB) *TableRepository {
	return &TableRepository{Repository: NewRepository[models.Table](db)}
}

func (r *TableRepository) FindByZone(zoneID uuid.UUID) ([]models.Table, error) {
	var tables []models.Table
	err := r.db.Where("zone_id = ? AND deleted_at IS NULL", zoneID).
		Order("name ASC").Find(&tables).Error
	return tables, err
}

func (r *TableRepository) SetOccupied(id uuid.UUID, occupied bool) error {
	return r.db.Model(&models.Table{}).Where("id = ?", id).
		Update("is_occupied", occupied).Error
}
