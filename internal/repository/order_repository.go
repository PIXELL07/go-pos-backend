package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	*Repository[models.Order]
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{Repository: NewRepository[models.Order](db)}
}

type OrderFilter struct {
	OutletID string
	Status   string
	Source   string
	OrderNo  string
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// List returns a paginated, filtered list of orders.
func (r *OrderRepository) List(f OrderFilter) (*PaginatedResult[models.Order], error) {
	q := r.db.Model(&models.Order{}).
		Preload("Items").Preload("Payments").Preload("Outlet").Preload("Cashier")

	q = applyOrderFilters(q, f)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := q.Scopes(Paginate(f.Page, f.Limit)).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	tp := int(total) / f.Limit
	if int(total)%f.Limit != 0 {
		tp++
	}
	return &PaginatedResult[models.Order]{
		Data: orders, Total: total,
		Page: f.Page, Limit: f.Limit, TotalPages: tp,
	}, nil
}

// returns a fully-preloaded order.
func (r *OrderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.
		Preload("Items.MenuItem").
		Preload("Payments").
		Preload("Outlet").
		Preload("Cashier").
		First(&order, "id = ?", id).Error
	return &order, err
}

// returns the order with the given invoice number.
func (r *OrderRepository) FindByInvoice(invoiceNo string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("invoice_number = ?", invoiceNo).First(&order).Error
	return &order, err
}

// returns all live orders (pending → dispatched), optionally
// restricted to a single outlet.
func (r *OrderRepository) FindRunning(outletID string) ([]models.Order, error) {
	statuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusAccepted,
		models.OrderStatusPreparing,
		models.OrderStatusReady,
		models.OrderStatusDispatched,
	}
	q := r.db.Model(&models.Order{}).
		Preload("Items").Preload("Payments").
		Where("status IN ?", statuses)
	if outletID != "" {
		q = q.Where("outlet_id = ?", outletID)
	}
	var orders []models.Order
	return orders, q.Order("created_at ASC").Find(&orders).Error
}

// returns paginated online-platform orders.
func (r *OrderRepository) FindOnline(f OnlineOrderFilter) (*PaginatedResult[models.Order], error) {
	onlineSources := []models.OrderSource{
		models.OrderSourceZomato, models.OrderSourceSwiggy,
		models.OrderSourceFoodPanda, models.OrderSourceUberEats,
		models.OrderSourceDunzo, models.OrderSourceWebsite,
	}
	q := r.db.Model(&models.Order{}).
		Preload("Items").Preload("Outlet").
		Where("source IN ?", onlineSources)

	q = applyOnlineFilters(q, f)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var orders []models.Order
	if err := q.Scopes(Paginate(f.Page, f.Limit)).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	tp := int(total) / f.Limit
	if int(total)%f.Limit != 0 {
		tp++
	}
	return &PaginatedResult[models.Order]{
		Data: orders, Total: total,
		Page: f.Page, Limit: f.Limit, TotalPages: tp,
	}, nil
}

func (r *OrderRepository) UpdateStatus(id uuid.UUID, status models.OrderStatus) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *OrderRepository) IncrementPrintCount(id uuid.UUID) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_printed":  true,
		"print_count": gorm.Expr("print_count + 1"),
	}).Error
}

func (r *OrderRepository) DailySummary(outletID string, from, to time.Time) ([]DailySummaryRow, error) {
	q := r.db.Table("orders o").
		Joins("JOIN outlets ol ON ol.id = o.outlet_id").
		Where("o.deleted_at IS NULL AND o.status NOT IN ? AND o.created_at >= ? AND o.created_at < ?",
			[]string{"cancelled"}, from, to).
		Select(`
			ol.id   AS outlet_id,
			ol.name AS outlet_name,
			COUNT(o.id)                       AS orders,
			COALESCE(SUM(o.total_amount), 0)  AS total_sales,
			COALESCE(SUM(o.net_sales), 0)     AS net_sales,
			COALESCE(SUM(o.tax_amount), 0)    AS tax,
			COALESCE(SUM(o.discount_amount),0) AS discounts,
			SUM(CASE WHEN o.is_modified THEN 1 ELSE 0 END)    AS modified,
			COALESCE(SUM(o.print_count - 1), 0)               AS reprint
		`).
		Group("ol.id, ol.name")

	if outletID != "" {
		q = q.Where("o.outlet_id = ?", outletID)
	}

	var rows []DailySummaryRow
	return rows, q.Scan(&rows).Error
}

type OnlineOrderFilter struct {
	OutletID   string
	Platform   string
	Status     string
	OrderNo    string
	RecordType string
	From       *time.Time
	To         *time.Time
	Page       int
	Limit      int
}

// Aggregate result type
type DailySummaryRow struct {
	OutletID   string  `json:"outlet_id"`
	OutletName string  `json:"outlet_name"`
	Orders     int64   `json:"orders"`
	TotalSales float64 `json:"total_sales"`
	NetSales   float64 `json:"net_sales"`
	Tax        float64 `json:"tax"`
	Discounts  float64 `json:"discounts"`
	Modified   int64   `json:"modified"`
	Reprint    int64   `json:"reprint"`
}

// private filter helpers
func applyOrderFilters(q *gorm.DB, f OrderFilter) *gorm.DB {
	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Source != "" {
		q = q.Where("source = ?", f.Source)
	}
	if f.OrderNo != "" {
		q = q.Where("invoice_number ILIKE ?", "%"+f.OrderNo+"%")
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", f.From)
	}
	if f.To != nil {
		end := f.To.Add(24 * time.Hour)
		q = q.Where("created_at < ?", end)
	}
	return q
}

func applyOnlineFilters(q *gorm.DB, f OnlineOrderFilter) *gorm.DB {
	if f.OutletID != "" {
		q = q.Where("outlet_id = ?", f.OutletID)
	}
	if f.Platform != "" && f.Platform != "all" {
		q = q.Where("source = ?", f.Platform)
	}
	if f.Status != "" && f.Status != "all" {
		q = q.Where("status = ?", f.Status)
	}
	if f.OrderNo != "" {
		q = q.Where("invoice_number ILIKE ? OR external_order_id ILIKE ?",
			"%"+f.OrderNo+"%", "%"+f.OrderNo+"%")
	}

	now := time.Now()
	switch f.RecordType {
	case "last_2_days":
		q = q.Where("created_at >= ?", now.AddDate(0, 0, -2))
	case "last_5_days":
		q = q.Where("created_at >= ?", now.AddDate(0, 0, -5))
	case "last_7_days":
		q = q.Where("created_at >= ?", now.AddDate(0, 0, -7))
	default: // old_records / custom range
		if f.From != nil {
			q = q.Where("created_at >= ?", f.From)
		}
		if f.To != nil {
			end := f.To.Add(24 * time.Hour)
			q = q.Where("created_at < ?", end)
		}
	}
	return q
}
