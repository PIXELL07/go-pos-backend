package services

import (
	"errors"
	"fmt"
	"time"

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

// added:
// Reports services

type ReportsService struct{ db *gorm.DB }

type SalesReportFilter struct {
	OutletIDs []string
	From      time.Time
	To        time.Time
	Status    string
	Page      int
	Limit     int
}

type SalesReportRow struct {
	RestaurantName   string  `json:"restaurant_name"`
	InvoiceNumbers   string  `json:"invoice_numbers"`
	TotalBills       int     `json:"total_bills"`
	MyAmount         float64 `json:"my_amount"`
	TotalDiscount    float64 `json:"total_discount"`
	NetSales         float64 `json:"net_sales"`
	DeliveryCharge   float64 `json:"delivery_charge"`
	ContainerCharge  float64 `json:"container_charge"`
	ServiceCharge    float64 `json:"service_charge"`
	AdditionalCharge float64 `json:"additional_charge"`
	TotalTax         float64 `json:"total_tax"`
	RoundOff         float64 `json:"round_off"`
	WaivedOff        float64 `json:"waived_off"`
	TotalSales       float64 `json:"total_sales"`
	Cash             float64 `json:"cash"`
	Card             float64 `json:"card"`
	DuePayment       float64 `json:"due_payment"`
	Online           float64 `json:"online"`
	Wallet           float64 `json:"wallet"`
	Pax              int     `json:"pax"`
	DataSynced       string  `json:"data_synced"`
}

type SalesReport struct {
	Data    []SalesReportRow `json:"data"`
	Summary SalesReportRow   `json:"summary"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

func NewReportsService(db *gorm.DB) *ReportsService { return &ReportsService{db: db} }

func (s *ReportsService) GetSalesReport(userID uuid.UUID, filter SalesReportFilter) (*SalesReport, error) {
	to := filter.To.Add(24 * time.Hour)
	query := s.db.Table("orders o").
		Joins("JOIN outlets ol ON ol.id = o.outlet_id").
		Where("o.deleted_at IS NULL AND o.created_at >= ? AND o.created_at < ?", filter.From, to)

	if len(filter.OutletIDs) > 0 {
		query = query.Where("o.outlet_id IN ?", filter.OutletIDs)
	}
	if filter.Status != "" && filter.Status != "all" {
		query = query.Where("o.status = ?", filter.Status)
	} else {
		query = query.Where("o.status != 'cancelled'")
	}

	var rows []SalesReportRow
	query.Select(`
		ol.name as restaurant_name,
		CONCAT(MIN(o.invoice_number), ' - ', MAX(o.invoice_number)) as invoice_numbers,
		COUNT(o.id) as total_bills,
		COALESCE(SUM(o.sub_total), 0) as my_amount,
		COALESCE(SUM(o.discount_amount), 0) as total_discount,
		COALESCE(SUM(o.net_sales), 0) as net_sales,
		COALESCE(SUM(o.delivery_charge), 0) as delivery_charge,
		COALESCE(SUM(o.container_charge), 0) as container_charge,
		COALESCE(SUM(o.service_charge), 0) as service_charge,
		COALESCE(SUM(o.additional_charge), 0) as additional_charge,
		COALESCE(SUM(o.tax_amount), 0) as total_tax,
		COALESCE(SUM(o.round_off), 0) as round_off,
		COALESCE(SUM(o.waived_off), 0) as waived_off,
		COALESCE(SUM(o.total_amount), 0) as total_sales,
		COALESCE(SUM(CASE WHEN p.method='cash' THEN p.amount ELSE 0 END), 0) as cash,
		COALESCE(SUM(CASE WHEN p.method='card' THEN p.amount ELSE 0 END), 0) as card,
		COALESCE(SUM(CASE WHEN p.method='due' THEN p.amount ELSE 0 END), 0) as due_payment,
		COALESCE(SUM(CASE WHEN p.method='online' THEN p.amount ELSE 0 END), 0) as online,
		COALESCE(SUM(CASE WHEN p.method='wallet' THEN p.amount ELSE 0 END), 0) as wallet,
		COALESCE(SUM(o.pax), 0) as pax,
		MAX(o.updated_at)::text as data_synced
	`).
		Joins("LEFT JOIN payments p ON p.order_id = o.id").
		Group("ol.id, ol.name").Scan(&rows)

	var summary SalesReportRow
	summary.RestaurantName = "Total"
	for _, r := range rows {
		summary.TotalBills += r.TotalBills
		summary.MyAmount += r.MyAmount
		summary.TotalDiscount += r.TotalDiscount
		summary.NetSales += r.NetSales
		summary.TotalTax += r.TotalTax
		summary.TotalSales += r.TotalSales
		summary.Cash += r.Cash
		summary.Card += r.Card
		summary.Online += r.Online
		summary.Wallet += r.Wallet
		summary.Pax += r.Pax
	}

	return &SalesReport{Data: rows, Summary: summary, Total: int64(len(rows)), Page: filter.Page, Limit: filter.Limit}, nil
}

// added:
// Inventory Service

type InventoryService struct{ db *gorm.DB }

type PurchaseFilter struct {
	OutletID string
	Type     string
	From     time.Time
	To       time.Time
}

func NewInventoryService(db *gorm.DB) *InventoryService { return &InventoryService{db: db} }

func (s *InventoryService) GetPendingPurchases(userID uuid.UUID, filter PurchaseFilter) ([]models.PendingPurchase, error) {
	to := filter.To.Add(24 * time.Hour)
	query := s.db.Model(&models.PendingPurchase{}).Preload("Outlet").
		Where("created_at >= ? AND created_at < ?", filter.From, to)
	if filter.OutletID != "" {
		query = query.Where("outlet_id = ?", filter.OutletID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	var purchases []models.PendingPurchase
	return purchases, query.Order("created_at DESC").Find(&purchases).Error
}

func (s *InventoryService) CreatePurchase(p *models.PendingPurchase) error {
	return s.db.Create(p).Error
}

func (s *InventoryService) UpdatePurchaseStatus(id uuid.UUID, status string) (*models.PendingPurchase, error) {
	var p models.PendingPurchase
	if err := s.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&p).Update("status", status)
	return &p, nil
}

// need to add:
// thirdparty
// logs
// franchise
// user sevices
// error handling
