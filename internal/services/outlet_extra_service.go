package services

import (
	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

// returns all outlets with their franchise_type populated.
func (s *OutletService) ListWithFranchiseTypes(outletID string) ([]models.Outlet, error) {
	q := s.db.Where("deleted_at IS NULL")
	if outletID != "" {
		q = q.Where("id = ?", outletID)
	}
	var outlets []models.Outlet
	return outlets, q.Order("name ASC").Find(&outlets).Error
}

// sets the franchise_type field on an outlet.
func (s *OutletService) UpdateFranchiseType(id uuid.UUID, ft models.FranchiseType) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	outlet.FranchiseType = ft
	if err := s.db.Save(&outlet).Error; err != nil {
		return nil, err
	}
	return &outlet, nil
}

// updates name/outlet association for a zone.
func (s *OutletService) UpdateZone(id uuid.UUID, updates map[string]interface{}) (*models.Zone, error) {
	var zone models.Zone
	if err := s.db.First(&zone, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&zone).Updates(updates)
	s.db.Preload("Tables").First(&zone, "id = ?", id)
	return &zone, nil
}

// soft-deletes a zone (and cascades to tables via DB).
func (s *OutletService) DeleteZone(id uuid.UUID) error {
	return s.db.Delete(&models.Zone{}, "id = ?", id).Error
}

func (s *OutletService) CreateTable(zoneID uuid.UUID, name string, capacity int) (*models.Table, error) {
	table := &models.Table{ZoneID: zoneID, Name: name, Capacity: capacity}
	return table, s.db.Create(table).Error
}

// updates table name/capacity.
func (s *OutletService) UpdateTable(id uuid.UUID, updates map[string]interface{}) (*models.Table, error) {
	var table models.Table
	if err := s.db.First(&table, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&table).Updates(updates)
	return &table, nil
}

// soft-deletes a table.
func (s *OutletService) DeleteTable(id uuid.UUID) error {
	return s.db.Delete(&models.Table{}, "id = ?", id).Error
}

// sets outlet.franchise_id.
func (s *OutletService) AssignOutletToFranchise(outletID, franchiseID uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", outletID).Error; err != nil {
		return nil, err
	}
	outlet.FranchiseID = &franchiseID
	if err := s.db.Save(&outlet).Error; err != nil {
		return nil, err
	}
	return &outlet, nil
}

type StoreStatusService struct{ db *gorm.DB }

func NewStoreStatusService(db *gorm.DB) *StoreStatusService {
	return &StoreStatusService{db: db}
}

type StoreStatusFilter struct {
	OutletID            string
	Platform            string
	OfflineSinceMinutes int
	Page                int
	Limit               int
}

func (s *StoreStatusService) GetStatus(f StoreStatusFilter) ([]models.StoreStatusSnapshot, int64, error) {
	q := s.db.Model(&models.StoreStatusSnapshot{}).Preload("Outlet")

	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Platform != "" && f.Platform != "all" {
		q = q.Where("platform = ?", f.Platform)
	}
	if f.OfflineSinceMinutes > 0 {
		q = q.Where("is_online = false AND offline_since <= NOW() - INTERVAL '? minutes'",
			f.OfflineSinceMinutes)
	}

	var total int64
	q.Count(&total)

	offset := (f.Page - 1) * f.Limit
	var results []models.StoreStatusSnapshot
	err := q.Order("is_online ASC, last_checked DESC").
		Offset(offset).Limit(f.Limit).Find(&results).Error
	return results, total, err
}

// FYI:
// RefreshAll re-checks connectivity for all configured third-party platforms
// and upserts StoreStatusSnapshot rows.  In production this would call real
// aggregator health-check APIs; here we just touch last_checked.

func (s *StoreStatusService) RefreshAll(outletID string, triggeredBy uuid.UUID) (int, error) {
	q := s.db.Model(&models.ThirdPartyConfig{}).Where("is_active = true")
	if outletID != "" {
		q = q.Where("outlet_id = ?", outletID)
	}
	var configs []models.ThirdPartyConfig
	if err := q.Find(&configs).Error; err != nil {
		return 0, err
	}
	for i := range configs {
		cfg := configs[i]
		snap := models.StoreStatusSnapshot{
			OutletID:    cfg.OutletID,
			Platform:    cfg.Platform,
			IsOnline:    true, // optimistic; replace with real API call
			LastChecked: timeNow(),
		}
		s.db.Where(models.StoreStatusSnapshot{OutletID: cfg.OutletID, Platform: cfg.Platform}).
			Assign(models.StoreStatusSnapshot{
				IsOnline:    snap.IsOnline,
				LastChecked: snap.LastChecked,
			}).
			FirstOrCreate(&snap)
	}
	return len(configs), nil
}

// returns the online_store_logs as a history proxy.
func (s *StoreStatusService) GetHistory(outletID, platform, from, to string) ([]models.OnlineStoreLog, error) {
	q := s.db.Model(&models.OnlineStoreLog{}).Preload("Outlet")
	if outletID != "" {
		q = q.Where("outlet_id = ?", outletID)
	}
	if platform != "" {
		q = q.Where("platform = ?", platform)
	}
	if from != "" {
		q = q.Where("created_at >= ?", from)
	}
	if to != "" {
		q = q.Where("created_at <= ?", to+" 23:59:59")
	}
	var logs []models.OnlineStoreLog
	return logs, q.Order("created_at DESC").Limit(200).Find(&logs).Error
}
