package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

type SalesReportFilter struct {
	OutletIDs []string
	From      time.Time
	To        time.Time
	Status    string
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
	OnlineTaxCalc    float64 `json:"online_tax_calculated"`
	GSTByMerchant    float64 `json:"gst_paid_by_merchant"`
	GSTByEcommerce   float64 `json:"gst_paid_by_ecommerce"`
	Cash             float64 `json:"cash"`
	Card             float64 `json:"card"`
	DuePayment       float64 `json:"due_payment"`
	Other            float64 `json:"other"`
	Wallet           float64 `json:"wallet"`
	Online           float64 `json:"online"`
	Pax              int     `json:"pax"`
	DataSynced       string  `json:"data_synced"`
}

func (r *ReportRepository) GetSalesReport(f SalesReportFilter) ([]SalesReportRow, error) {
	to := f.To.Add(24 * time.Hour)

	q := r.db.Table("orders o").
		Joins("JOIN outlets ol ON ol.id = o.outlet_id").
		Joins("LEFT JOIN payments p ON p.order_id = o.id").
		Where("o.deleted_at IS NULL AND o.created_at >= ? AND o.created_at < ?", f.From, to)

	if len(f.OutletIDs) > 0 {
		q = q.Where("o.outlet_id IN ?", f.OutletIDs)
	}
	if f.Status != "" && f.Status != "all" {
		q = q.Where("o.status = ?", f.Status)
	} else {
		q = q.Where("o.status != 'cancelled'")
	}

	var rows []SalesReportRow
	err := q.Select(`
		ol.name AS restaurant_name,
		CONCAT(MIN(o.invoice_number),' – ',MAX(o.invoice_number)) AS invoice_numbers,
		COUNT(DISTINCT o.id)                                      AS total_bills,
		COALESCE(SUM(o.sub_total),0)                             AS my_amount,
		COALESCE(SUM(o.discount_amount),0)                       AS total_discount,
		COALESCE(SUM(o.net_sales),0)                             AS net_sales,
		COALESCE(SUM(o.delivery_charge),0)                       AS delivery_charge,
		COALESCE(SUM(o.container_charge),0)                      AS container_charge,
		COALESCE(SUM(o.service_charge),0)                        AS service_charge,
		COALESCE(SUM(o.additional_charge),0)                     AS additional_charge,
		COALESCE(SUM(o.tax_amount),0)                            AS total_tax,
		COALESCE(SUM(o.round_off),0)                             AS round_off,
		COALESCE(SUM(o.waived_off),0)                            AS waived_off,
		COALESCE(SUM(o.total_amount),0)                          AS total_sales,
		COALESCE(SUM(o.online_tax_calc),0)                       AS online_tax_calculated,
		COALESCE(SUM(o.gst_by_merchant),0)                       AS gst_paid_by_merchant,
		COALESCE(SUM(o.gst_by_ecommerce),0)                      AS gst_paid_by_ecommerce,
		COALESCE(SUM(CASE WHEN p.method='cash'   THEN p.amount ELSE 0 END),0) AS cash,
		COALESCE(SUM(CASE WHEN p.method='card'   THEN p.amount ELSE 0 END),0) AS card,
		COALESCE(SUM(CASE WHEN p.method='due'    THEN p.amount ELSE 0 END),0) AS due_payment,
		COALESCE(SUM(CASE WHEN p.method='other'  THEN p.amount ELSE 0 END),0) AS other,
		COALESCE(SUM(CASE WHEN p.method='wallet' THEN p.amount ELSE 0 END),0) AS wallet,
		COALESCE(SUM(CASE WHEN p.method='online' THEN p.amount ELSE 0 END),0) AS online,
		COALESCE(SUM(o.pax),0)                                   AS pax,
		MAX(o.updated_at)::text                                  AS data_synced
	`).Group("ol.id, ol.name").Scan(&rows).Error

	return rows, err
}

// PendingPurchase

type PurchaseRepository struct {
	*Repository[models.PendingPurchase]
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{Repository: NewRepository[models.PendingPurchase](db)}
}

type PurchaseFilter struct {
	OutletID string
	Type     string
	From     time.Time
	To       time.Time
	Page     int
	Limit    int
}

func (r *PurchaseRepository) List(f PurchaseFilter) ([]models.PendingPurchase, error) {
	to := f.To.Add(24 * time.Hour)
	q := r.db.Model(&models.PendingPurchase{}).Preload("Outlet").
		Where("created_at >= ? AND created_at < ?", f.From, to)
	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	var purchases []models.PendingPurchase
	return purchases, q.Order("created_at DESC").
		Scopes(Paginate(f.Page, f.Limit)).Find(&purchases).Error
}

func (r *PurchaseRepository) SetStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.PendingPurchase{}).Where("id = ?", id).
		Update("status", status).Error
}

// Notification

type NotificationRepository struct {
	*Repository[models.Notification]
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{Repository: NewRepository[models.Notification](db)}
}

type NotifFilter struct {
	IsRead *bool
	Page   int
	Limit  int
}

func (r *NotificationRepository) FindByUser(userID uuid.UUID, f NotifFilter) ([]models.Notification, int64, error) {
	q := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	if f.IsRead != nil {
		q = q.Where("is_read = ?", *f.IsRead)
	}
	var total int64
	q.Count(&total)
	var notifs []models.Notification
	err := q.Order("created_at DESC").Scopes(Paginate(f.Page, f.Limit)).Find(&notifs).Error
	return notifs, total, err
}

func (r *NotificationRepository) MarkRead(userID, notifID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) MarkAllRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *NotificationRepository) UnreadCount(userID uuid.UUID) (int64, error) {
	return r.Count("user_id = ? AND is_read = false", userID)
}

// Thirdparty Config
type ThirdPartyRepository struct {
	*Repository[models.ThirdPartyConfig]
}

func NewThirdPartyRepository(db *gorm.DB) *ThirdPartyRepository {
	return &ThirdPartyRepository{Repository: NewRepository[models.ThirdPartyConfig](db)}
}

func (r *ThirdPartyRepository) FindByOutlet(outletID string) ([]models.ThirdPartyConfig, error) {
	q := r.db.Preload("Outlet")
	if outletID != "" {
		q = q.Where("outlet_id = ?", outletID)
	}
	var configs []models.ThirdPartyConfig
	return configs, q.Find(&configs).Error
}

func (r *ThirdPartyRepository) UpsertForOutlet(outletID uuid.UUID, platform string) (*models.ThirdPartyConfig, error) {
	cfg := &models.ThirdPartyConfig{OutletID: outletID, Platform: platform}
	err := r.db.
		Where(models.ThirdPartyConfig{OutletID: outletID, Platform: platform}).
		FirstOrCreate(cfg).Error
	return cfg, err
}

// Logs
type LogFilter struct {
	OutletID string
	Platform string
	From     time.Time
	To       time.Time
	Page     int
	Limit    int
}

// add:
// MenuTriggerLogRepository
// OnlineStoreLogRepository
// frachise log
// refresh tokens
// helpers
