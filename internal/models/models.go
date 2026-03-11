package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleBiller UserRole = "biller"
	RoleOwner  UserRole = "owner"
)

type User struct {
	Base
	Name           string         `gorm:"not null" json:"name"`
	Email          string         `gorm:"uniqueIndex;not null" json:"email"`
	Mobile         string         `gorm:"uniqueIndex" json:"mobile"`
	PasswordHash   string         `gorm:"not null" json:"-"`
	Role           UserRole       `gorm:"type:varchar(20);default:'biller'" json:"role"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	GoogleID       string         `gorm:"uniqueIndex" json:"google_id,omitempty"`
	AvatarURL      string         `json:"avatar_url,omitempty"`
	LastLoginAt    *time.Time     `json:"last_login_at,omitempty"`
	OutletAccesses []OutletAccess `gorm:"foreignKey:UserID" json:"outlet_accesses,omitempty"`
}

type RefreshToken struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
}

type OutletType string

const (
	OutletTypeDineIn   OutletType = "dine_in"
	OutletTypeTakeaway OutletType = "takeaway"
	OutletTypeDelivery OutletType = "delivery"
	OutletTypeCloud    OutletType = "cloud"
)

type Outlet struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	RefID       string     `gorm:"uniqueIndex;not null" json:"ref_id"`
	Type        OutletType `gorm:"type:varchar(20);default:'dine_in'" json:"type"`
	Address     string     `json:"address"`
	City        string     `json:"city"`
	State       string     `json:"state"`
	PinCode     string     `json:"pin_code"`
	Phone       string     `json:"phone"`
	GSTNumber   string     `json:"gst_number"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	IsLocked    bool       `gorm:"default:false" json:"is_locked"`
	FranchiseID *uuid.UUID `gorm:"type:uuid" json:"franchise_id,omitempty"`
	Franchise   *Franchise `gorm:"foreignKey:FranchiseID" json:"franchise,omitempty"`
	Zones       []Zone     `gorm:"foreignKey:OutletID" json:"zones,omitempty"`
}

type OutletAccess struct {
	Base
	UserID   uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	OutletID uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	Role     UserRole  `gorm:"type:varchar(20)" json:"role"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Outlet   Outlet    `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
}

type Zone struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	OutletID uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	Outlet   Outlet    `gorm:"foreignKey:OutletID" json:"-"`
	Tables   []Table   `gorm:"foreignKey:ZoneID" json:"tables,omitempty"`
}

type Table struct {
	Base
	Name       string    `gorm:"not null" json:"name"`
	ZoneID     uuid.UUID `gorm:"type:uuid;index;not null" json:"zone_id"`
	Capacity   int       `gorm:"default:4" json:"capacity"`
	IsOccupied bool      `gorm:"default:false" json:"is_occupied"`
	Zone       Zone      `gorm:"foreignKey:ZoneID" json:"-"`
}

type Franchise struct {
	Base
	Name    string   `gorm:"not null" json:"name"`
	Outlets []Outlet `gorm:"foreignKey:FranchiseID" json:"outlets,omitempty"`
}

// Menu / Items

type Category struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	Description string     `json:"description"`
	OutletID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"outlet_id"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	Items       []MenuItem `gorm:"foreignKey:CategoryID" json:"items,omitempty"`
}

type MenuItem struct {
	Base
	Name           string    `gorm:"not null" json:"name"`
	Description    string    `json:"description"`
	CategoryID     uuid.UUID `gorm:"type:uuid;index;not null" json:"category_id"`
	OutletID       uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	Price          float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	TaxRate        float64   `gorm:"type:decimal(5,2);default:5.0" json:"tax_rate"`
	IsVeg          bool      `gorm:"default:true" json:"is_veg"`
	IsAvailable    bool      `gorm:"default:true" json:"is_available"`
	IsOnlineActive bool      `gorm:"default:false" json:"is_online_active"`
	ImageURL       string    `json:"image_url,omitempty"`
	SortOrder      int       `gorm:"default:0" json:"sort_order"`
	Category       Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// Orders

type OrderSource string
type OrderType string
type OrderStatus string
type PaymentMethod string

const (
	OrderSourcePOS       OrderSource = "pos"
	OrderSourceZomato    OrderSource = "zomato"
	OrderSourceSwiggy    OrderSource = "swiggy"
	OrderSourceFoodPanda OrderSource = "foodpanda"
	OrderSourceUberEats  OrderSource = "uber_eats"
	OrderSourceDunzo     OrderSource = "dunzo"
	OrderSourceWebsite   OrderSource = "website"

	OrderTypeDineIn   OrderType = "dine_in"
	OrderTypeTakeaway OrderType = "takeaway"
	OrderTypeDelivery OrderType = "delivery"

	OrderStatusPending       OrderStatus = "pending"
	OrderStatusAccepted      OrderStatus = "accepted"
	OrderStatusPreparing     OrderStatus = "preparing"
	OrderStatusReady         OrderStatus = "ready"
	OrderStatusDispatched    OrderStatus = "dispatched"
	OrderStatusDelivered     OrderStatus = "delivered"
	OrderStatusCancelled     OrderStatus = "cancelled"
	OrderStatusCompleted     OrderStatus = "completed"
	OrderStatusComplimentary OrderStatus = "complimentary"
	OrderStatusSalesReturn   OrderStatus = "sales_return"

	PaymentCash   PaymentMethod = "cash"
	PaymentCard   PaymentMethod = "card"
	PaymentOnline PaymentMethod = "online"
	PaymentWallet PaymentMethod = "wallet"
	PaymentDue    PaymentMethod = "due"
	PaymentOther  PaymentMethod = "other"
)

type Order struct {
	Base
	InvoiceNumber    string      `gorm:"uniqueIndex;not null" json:"invoice_number"`
	OutletID         uuid.UUID   `gorm:"type:uuid;index;not null" json:"outlet_id"`
	TableID          *uuid.UUID  `gorm:"type:uuid" json:"table_id,omitempty"`
	CashierID        uuid.UUID   `gorm:"type:uuid;index;not null" json:"cashier_id"`
	Source           OrderSource `gorm:"type:varchar(20);default:'pos'" json:"source"`
	Type             OrderType   `gorm:"type:varchar(20);default:'dine_in'" json:"type"`
	Status           OrderStatus `gorm:"type:varchar(30);default:'pending'" json:"status"`
	CustomerName     string      `json:"customer_name"`
	CustomerPhone    string      `json:"customer_phone"`
	CustomerOTP      string      `json:"-"`
	RiderDetails     string      `json:"rider_details,omitempty"`
	Pax              int         `gorm:"default:1" json:"pax"`
	SubTotal         float64     `gorm:"type:decimal(10,2)" json:"sub_total"`
	DiscountAmount   float64     `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	DiscountPercent  float64     `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	TaxAmount        float64     `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	DeliveryCharge   float64     `gorm:"type:decimal(10,2);default:0" json:"delivery_charge"`
	ContainerCharge  float64     `gorm:"type:decimal(10,2);default:0" json:"container_charge"`
	ServiceCharge    float64     `gorm:"type:decimal(10,2);default:0" json:"service_charge"`
	AdditionalCharge float64     `gorm:"type:decimal(10,2);default:0" json:"additional_charge"`
	RoundOff         float64     `gorm:"type:decimal(5,2);default:0" json:"round_off"`
	WaivedOff        float64     `gorm:"type:decimal(10,2);default:0" json:"waived_off"`
	TotalAmount      float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	NetSales         float64     `gorm:"type:decimal(10,2)" json:"net_sales"`
	OnlineTaxCalc    float64     `gorm:"type:decimal(10,2);default:0" json:"online_tax_calculated"`
	GSTByMerchant    float64     `gorm:"type:decimal(10,2);default:0" json:"gst_paid_by_merchant"`
	GSTByEcommerce   float64     `gorm:"type:decimal(10,2);default:0" json:"gst_paid_by_ecommerce"`
	IsModified       bool        `gorm:"default:false" json:"is_modified"`
	IsPrinted        bool        `gorm:"default:false" json:"is_printed"`
	PrintCount       int         `gorm:"default:0" json:"print_count"`
	ExternalOrderID  string      `json:"external_order_id,omitempty"`
	Notes            string      `json:"notes,omitempty"`
	Outlet           Outlet      `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
	Cashier          User        `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	Items            []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Payments         []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
}

type OrderItem struct {
	Base
	OrderID    uuid.UUID `gorm:"type:uuid;index;not null" json:"order_id"`
	MenuItemID uuid.UUID `gorm:"type:uuid;index;not null" json:"menu_item_id"`
	Name       string    `gorm:"not null" json:"name"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	UnitPrice  float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TaxRate    float64   `gorm:"type:decimal(5,2)" json:"tax_rate"`
	TaxAmount  float64   `gorm:"type:decimal(10,2)" json:"tax_amount"`
	TotalPrice float64   `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Notes      string    `json:"notes,omitempty"`
	MenuItem   MenuItem  `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

type Payment struct {
	Base
	OrderID uuid.UUID     `gorm:"type:uuid;index;not null" json:"order_id"`
	Method  PaymentMethod `gorm:"type:varchar(20);not null" json:"method"`
	Amount  float64       `gorm:"type:decimal(10,2);not null" json:"amount"`
	RefNo   string        `json:"ref_no,omitempty"`
}

// Inventory / Purchase

type PendingPurchase struct {
	Base
	OutletID    uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	ItemName    string    `gorm:"not null" json:"item_name"`
	Quantity    float64   `gorm:"type:decimal(10,3);not null" json:"quantity"`
	Unit        string    `gorm:"not null" json:"unit"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status      string    `gorm:"type:varchar(30);default:'pending'" json:"status"`
	Type        string    `gorm:"type:varchar(20);default:'purchase'" json:"type"`
	RequestedBy uuid.UUID `gorm:"type:uuid;index" json:"requested_by"`
	Outlet      Outlet    `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
}

// Logs

type MenuTriggerLog struct {
	Base
	OutletID    uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	ItemID      uuid.UUID `gorm:"type:uuid;index" json:"item_id"`
	Action      string    `gorm:"not null" json:"action"`
	Details     string    `json:"details"`
	TriggeredBy uuid.UUID `gorm:"type:uuid" json:"triggered_by"`
	Outlet      Outlet    `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
}

type OnlineStoreLog struct {
	Base
	OutletID    uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	Platform    string    `gorm:"not null" json:"platform"`
	Action      string    `gorm:"not null" json:"action"`
	Status      string    `gorm:"not null" json:"status"`
	Details     string    `json:"details"`
	TriggeredBy uuid.UUID `gorm:"type:uuid" json:"triggered_by"`
	Outlet      Outlet    `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
}

type OnlineItemLog struct {
	Base
	OutletID    uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	ItemID      uuid.UUID `gorm:"type:uuid;index" json:"item_id"`
	Platform    string    `gorm:"not null" json:"platform"`
	Action      string    `gorm:"not null" json:"action"`
	OldStatus   string    `json:"old_status"`
	NewStatus   string    `json:"new_status"`
	TriggeredBy uuid.UUID `gorm:"type:uuid" json:"triggered_by"`
}

// Third-Party Config

type ThirdPartyConfig struct {
	Base
	OutletID uuid.UUID `gorm:"type:uuid;index;not null" json:"outlet_id"`
	Platform string    `gorm:"not null" json:"platform"`
	APIKey   string    `json:"-"`
	StoreID  string    `json:"store_id"`
	IsActive bool      `gorm:"default:false" json:"is_active"`
	Config   string    `gorm:"type:jsonb" json:"config"`
	Outlet   Outlet    `gorm:"foreignKey:OutletID" json:"outlet,omitempty"`
}

// Notifications

type Notification struct {
	Base
	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Title  string    `gorm:"not null" json:"title"`
	Body   string    `gorm:"not null" json:"body"`
	Type   string    `gorm:"type:varchar(30)" json:"type"`
	IsRead bool      `gorm:"default:false" json:"is_read"`
	Data   string    `gorm:"type:jsonb" json:"data,omitempty"`
	User   User      `gorm:"foreignKey:UserID" json:"-"`
}
