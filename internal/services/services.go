package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
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

// Notification Service

type NotificationService struct{ db *gorm.DB }

type NotificationFilter struct {
	IsRead string
	Page   int
	Limit  int
}

type NotifResult struct {
	Data  []models.Notification `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

func NewNotificationService(db *gorm.DB) *NotificationService { return &NotificationService{db: db} }

func (s *NotificationService) GetNotifications(userID uuid.UUID, filter NotificationFilter) (*NotifResult, error) {
	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if filter.IsRead == "true" {
		query = query.Where("is_read = true")
	} else if filter.IsRead == "false" {
		query = query.Where("is_read = false")
	}
	var total int64
	query.Count(&total)
	var notifications []models.Notification
	query.Order("created_at DESC").Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).Find(&notifications)
	return &NotifResult{Data: notifications, Total: total, Page: filter.Page, Limit: filter.Limit}, nil
}

func (s *NotificationService) MarkRead(userID, notifID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notifID, userID).Update("is_read", true).Error
}

func (s *NotificationService) MarkAllRead(userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error
}

// Thirdparty service

type ThirdPartyService struct{ db *gorm.DB }

func NewThirdPartyService(db *gorm.DB) *ThirdPartyService { return &ThirdPartyService{db: db} }

func (s *ThirdPartyService) GetConfigs(outletID string) ([]models.ThirdPartyConfig, error) {
	query := s.db.Model(&models.ThirdPartyConfig{}).Preload("Outlet")
	if outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}
	var configs []models.ThirdPartyConfig
	return configs, query.Find(&configs).Error
}

func (s *ThirdPartyService) UpdateConfig(id uuid.UUID, apiKey, storeID string, isActive bool, cfgJSON string) (*models.ThirdPartyConfig, error) {
	var cfg models.ThirdPartyConfig
	if err := s.db.First(&cfg, "id = ?", id).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"store_id": storeID, "is_active": isActive, "config": cfgJSON}
	if apiKey != "" {
		updates["api_key"] = apiKey
	}
	s.db.Model(&cfg).Updates(updates)
	return &cfg, nil
}

// added:
// Logs Service

type LogsService struct{ db *gorm.DB }

func NewLogsService(db *gorm.DB) *LogsService { return &LogsService{db: db} }

func (s *LogsService) GetMenuTriggerLogs(filter map[string]interface{}) (interface{}, error) {
	query := s.db.Model(&models.MenuTriggerLog{}).Preload("Outlet")
	if v, ok := filter["outlet_id"].(string); ok && v != "" {
		query = query.Where("outlet_id = ?", v)
	}
	applyDateFilter(query, filter)
	var logs []models.MenuTriggerLog
	return logs, query.Order("created_at DESC").Find(&logs).Error
}

func (s *LogsService) GetOnlineStoreLogs(filter map[string]interface{}) (interface{}, error) {
	query := s.db.Model(&models.OnlineStoreLog{}).Preload("Outlet")
	if v, ok := filter["outlet_id"].(string); ok && v != "" {
		query = query.Where("outlet_id = ?", v)
	}
	if v, ok := filter["platform"].(string); ok && v != "" {
		query = query.Where("platform = ?", v)
	}
	var logs []models.OnlineStoreLog
	return logs, query.Order("created_at DESC").Find(&logs).Error
}

func (s *LogsService) GetOnlineItemLogs(filter map[string]interface{}) (interface{}, error) {
	query := s.db.Model(&models.OnlineItemLog{})
	if v, ok := filter["outlet_id"].(string); ok && v != "" {
		query = query.Where("outlet_id = ?", v)
	}
	if v, ok := filter["platform"].(string); ok && v != "" {
		query = query.Where("platform = ?", v)
	}
	var logs []models.OnlineItemLog
	return logs, query.Order("created_at DESC").Find(&logs).Error
}

func applyDateFilter(query *gorm.DB, filter map[string]interface{}) {
	if from, ok := filter["from"].(string); ok && from != "" {
		t, _ := time.Parse("2006-01-02", from)
		query = query.Where("created_at >= ?", t)
	}
	if to, ok := filter["to"].(string); ok && to != "" {
		t, _ := time.Parse("2006-01-02", to)
		query = query.Where("created_at < ?", t.Add(24*time.Hour))
	}
}

// Franchise Service

type FranchiseService struct{ db *gorm.DB }

func NewFranchiseService(db *gorm.DB) *FranchiseService { return &FranchiseService{db: db} }

func (s *FranchiseService) List() ([]models.Franchise, error) {
	var franchises []models.Franchise
	return franchises, s.db.Preload("Outlets").Find(&franchises).Error
}

func (s *FranchiseService) Create(f *models.Franchise) error {
	return s.db.Create(f).Error
}

// User Service

type UserService struct{ db *gorm.DB }

func NewUserService(db *gorm.DB) *UserService { return &UserService{db: db} }

func (s *UserService) GetUsersByRole(role models.UserRole, outletID string) ([]models.User, error) {
	var users []models.User
	return users, s.db.Where("role = ? AND is_active = true", role).Find(&users).Error
}

func (s *UserService) InviteUser(name, email, mobile string, role models.UserRole, outletID string) (*models.User, error) {
	var count int64
	s.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return nil, errors.New("user with this email already exists")
	}
	tempPassword := fmt.Sprintf("Temp@%d", time.Now().Unix())
	hash, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	user := &models.User{
		Name: name, Email: email, Mobile: mobile,
		PasswordHash: string(hash), Role: role, IsActive: true,
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	if outletID != "" {
		if outID, err := uuid.Parse(outletID); err == nil {
			s.db.Create(&models.OutletAccess{UserID: user.ID, OutletID: outID, Role: role})
		}
	}
	return user, nil
}

func (s *UserService) UpdateUser(id uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	delete(updates, "id")
	delete(updates, "password_hash")
	s.db.Model(&user).Updates(updates)
	return &user, nil
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.db.Delete(&models.User{}, "id = ?", id).Error
}

// FranchiseService extras

func (s *FranchiseService) ListWithOutlets() ([]models.Franchise, error) {
	var franchises []models.Franchise
	return franchises, s.db.Preload("Outlets").Find(&franchises).Error
}

func (s *FranchiseService) AssignOutlet(franchiseID, outletID uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", outletID).Error; err != nil {
		return nil, err
	}
	outlet.FranchiseID = &franchiseID
	return &outlet, s.db.Save(&outlet).Error
}

// InventoryService extras
type PurchaseFilterExtra struct {
	OutletID string
	Type     string
	From     time.Time
	To       time.Time
	Page     int
	Limit    int
}

func (s *InventoryService) GetPendingPurchase(f PurchaseFilterExtra) ([]models.PendingPurchase, int64, error) {
	to := f.To.Add(24 * time.Hour)
	q := s.db.Model(&models.PendingPurchase{}).Preload("Outlet").
		Where("created_at >= ? AND created_at < ?", f.From, to)
	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	var total int64
	q.Count(&total)
	offset := (f.Page - 1) * f.Limit
	var purchases []models.PendingPurchase
	err := q.Order("created_at DESC").Offset(offset).Limit(f.Limit).Find(&purchases).Error
	return purchases, total, err
}

// added:
// NotificationService extras

func (s *NotificationService) GetByUser(userID uuid.UUID, page int) ([]models.Notification, int64, error) {
	q := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	var total int64
	q.Count(&total)
	var notifs []models.Notification
	err := q.Order("created_at DESC").Offset((page - 1) * 20).Limit(20).Find(&notifs).Error
	return notifs, total, err
}

func (s *NotificationService) MarkReads(userID, notifID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true).Error
}

func (s *NotificationService) MarkAllReads(userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

// Create inserts a new notification and optionally queues FCM push.
func (s *NotificationService) Create(n *models.Notification) error {
	return s.db.Create(n).Error
}

// ThirdPartyService extras

func (s *ThirdPartyService) Update(id uuid.UUID, updates map[string]interface{}) (*models.ThirdPartyConfig, error) {
	var cfg models.ThirdPartyConfig
	if err := s.db.First(&cfg, "id = ?", id).Error; err != nil {
		return nil, err
	}
	s.db.Model(&cfg).Updates(updates)
	s.db.Preload("Outlet").First(&cfg, "id = ?", id)
	return &cfg, nil
}

// need to add:
// UserService extras
// error handling
