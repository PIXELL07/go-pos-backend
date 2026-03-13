package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prayosha/go-pos-backend/internal/models"
	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
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

type PaginatedOrders struct {
	Data       []models.Order `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

type CreateOrderRequest struct {
	Type            models.OrderType
	Source          models.OrderSource
	CustomerName    string
	CustomerPhone   string
	Pax             int
	Items           []OrderItemRequest
	Payments        []PaymentRequest
	DiscountPercent float64
	DiscountAmount  float64
	DeliveryCharge  float64
	ContainerCharge float64
	ServiceCharge   float64
	RoundOff        float64
	Notes           string
}

type OrderItemRequest struct {
	MenuItemID string
	Quantity   int
	Notes      string
}

type PaymentRequest struct {
	Method models.PaymentMethod
	Amount float64
	RefNo  string
}

func (s *OrderService) ListOrders(userID uuid.UUID, filter OrderFilter) (*PaginatedOrders, error) {
	query := s.db.Model(&models.Order{}).Preload("Items").Preload("Payments")

	if filter.OutletID != "" {
		query = query.Where("outlet_id = ?", filter.OutletID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.OrderNo != "" {
		query = query.Where("invoice_number ILIKE ?", "%"+filter.OrderNo+"%")
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", filter.From)
	}
	if filter.To != nil {
		to := filter.To.Add(24 * time.Hour)
		query = query.Where("created_at < ?", to)
	}

	var total int64
	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	var orders []models.Order
	query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&orders)

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit != 0 {
		totalPages++
	}

	return &PaginatedOrders{
		Data:       orders,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *OrderService) GetRunningOrders(userID uuid.UUID, outletID string) ([]models.Order, error) {
	runningStatuses := []models.OrderStatus{
		models.OrderStatusPending, models.OrderStatusAccepted,
		models.OrderStatusPreparing, models.OrderStatusReady, models.OrderStatusDispatched,
	}

	query := s.db.Model(&models.Order{}).
		Preload("Items").Preload("Payments").
		Where("status IN ?", runningStatuses).
		Order("created_at ASC")

	if outletID != "" {
		query = query.Where("outlet_id = ?", outletID)
	}

	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) GetOnlineOrders(userID uuid.UUID, filter OnlineOrderFilter) (*PaginatedOrders, error) {
	onlineSources := []models.OrderSource{
		models.OrderSourceZomato, models.OrderSourceSwiggy,
		models.OrderSourceFoodPanda, models.OrderSourceUberEats,
		models.OrderSourceDunzo, models.OrderSourceWebsite,
	}

	query := s.db.Model(&models.Order{}).Preload("Items").
		Where("source IN ?", onlineSources)

	if filter.OutletID != "" {
		query = query.Where("outlet_id = ?", filter.OutletID)
	}
	if filter.Platform != "" && filter.Platform != "all" {
		query = query.Where("source = ?", filter.Platform)
	}
	if filter.Status != "" && filter.Status != "all" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.OrderNo != "" {
		query = query.Where("invoice_number ILIKE ? OR external_order_id ILIKE ?",
			"%"+filter.OrderNo+"%", "%"+filter.OrderNo+"%")
	}

	// Handle record type
	now := time.Now()
	switch filter.RecordType {
	case "last_2_days":
		query = query.Where("created_at >= ?", now.AddDate(0, 0, -2))
	case "last_5_days":
		query = query.Where("created_at >= ?", now.AddDate(0, 0, -5))
	case "last_7_days":
		query = query.Where("created_at >= ?", now.AddDate(0, 0, -7))
	case "old_records":
		if filter.From != nil {
			query = query.Where("created_at >= ?", filter.From)
		}
		if filter.To != nil {
			to := filter.To.Add(24 * time.Hour)
			query = query.Where("created_at < ?", to)
		}
	}

	var total int64
	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	var orders []models.Order
	query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&orders)

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit != 0 {
		totalPages++
	}

	return &PaginatedOrders{
		Data: orders, Total: total,
		Page: filter.Page, Limit: filter.Limit, TotalPages: totalPages,
	}, nil
}

func (s *OrderService) GetByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Items.MenuItem").Preload("Payments").
		Preload("Outlet").Preload("Cashier").
		First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) CreateOrder(cashierID, outletID uuid.UUID, tableIDStr *string, req CreateOrderRequest) (*models.Order, error) {
	// Fetch menu items and calculate
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	var tableID *uuid.UUID
	if tableIDStr != nil && *tableIDStr != "" {
		id, err := uuid.Parse(*tableIDStr)
		if err == nil {
			tableID = &id
		}
	}

	// Build order items
	var orderItems []models.OrderItem
	var subTotal float64
	var totalTax float64

	for _, itemReq := range req.Items {
		menuItemID, err := uuid.Parse(itemReq.MenuItemID)
		if err != nil {
			return nil, fmt.Errorf("invalid menu item ID: %s", itemReq.MenuItemID)
		}

		var menuItem models.MenuItem
		if err := s.db.First(&menuItem, "id = ? AND outlet_id = ?", menuItemID, outletID).Error; err != nil {
			return nil, fmt.Errorf("menu item not found: %s", itemReq.MenuItemID)
		}

		lineTotal := menuItem.Price * float64(itemReq.Quantity)
		lineTax := lineTotal * menuItem.TaxRate / 100

		orderItems = append(orderItems, models.OrderItem{
			MenuItemID: menuItemID,
			Name:       menuItem.Name,
			Quantity:   itemReq.Quantity,
			UnitPrice:  menuItem.Price,
			TaxRate:    menuItem.TaxRate,
			TaxAmount:  lineTax,
			TotalPrice: lineTotal,
			Notes:      itemReq.Notes,
		})
		subTotal += lineTotal
		totalTax += lineTax
	}

	// Apply discount
	discountAmount := req.DiscountAmount
	if req.DiscountPercent > 0 {
		discountAmount = subTotal * req.DiscountPercent / 100
	}

	netSales := subTotal - discountAmount
	totalAmount := netSales + totalTax + req.DeliveryCharge + req.ContainerCharge +
		req.ServiceCharge + req.RoundOff

	// Generate invoice number
	invoiceNum := generateInvoiceNumber(outletID)

	source := req.Source
	if source == "" {
		source = models.OrderSourcePOS
	}
	orderType := req.Type
	if orderType == "" {
		orderType = models.OrderTypeDineIn
	}
	pax := req.Pax
	if pax == 0 {
		pax = 1
	}

	order := &models.Order{
		InvoiceNumber:   invoiceNum,
		OutletID:        outletID,
		TableID:         tableID,
		CashierID:       cashierID,
		Source:          source,
		Type:            orderType,
		Status:          models.OrderStatusPending,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		Pax:             pax,
		SubTotal:        subTotal,
		DiscountAmount:  discountAmount,
		DiscountPercent: req.DiscountPercent,
		TaxAmount:       totalTax,
		DeliveryCharge:  req.DeliveryCharge,
		ContainerCharge: req.ContainerCharge,
		ServiceCharge:   req.ServiceCharge,
		RoundOff:        req.RoundOff,
		TotalAmount:     totalAmount,
		NetSales:        netSales,
		Notes:           req.Notes,
	}

	// Transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}

		if len(req.Payments) > 0 {
			payments := make([]models.Payment, len(req.Payments))
			for i, p := range req.Payments {
				payments[i] = models.Payment{
					OrderID: order.ID,
					Method:  p.Method,
					Amount:  p.Amount,
					RefNo:   p.RefNo,
				}
			}
			if err := tx.Create(&payments).Error; err != nil {
				return err
			}
			// Mark as completed if fully paid
			var totalPaid float64
			for _, p := range req.Payments {
				totalPaid += p.Amount
			}
			if totalPaid >= totalAmount {
				tx.Model(order).Update("status", models.OrderStatusCompleted)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	order.Items = orderItems
	return order, nil
}

func (s *OrderService) UpdateStatus(id uuid.UUID, status models.OrderStatus) (*models.Order, error) {
	var order models.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
		return nil, errors.New("order not found")
	}
	if err := s.db.Model(&order).Update("status", status).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) CancelOrder(id uuid.UUID, reason string) (*models.Order, error) {
	return s.UpdateStatus(id, models.OrderStatusCancelled)
}

func (s *OrderService) MarkPrinted(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := s.db.First(&order, "id = ?", id).Error; err != nil {
		return nil, errors.New("order not found")
	}
	s.db.Model(&order).Updates(map[string]interface{}{
		"is_printed":  true,
		"print_count": gorm.Expr("print_count + 1"),
	})
	return &order, nil
}

func generateInvoiceNumber(outletID uuid.UUID) string {
	prefix := outletID.String()[:4]
	return fmt.Sprintf("INV-%s-%d", prefix, time.Now().UnixNano()/1e6)
}
