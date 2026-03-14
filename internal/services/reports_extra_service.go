package services

// Extended report types
// represents a single item's aggregated sales across all outlets.
type ItemWiseRow struct {
	ItemName   string  `json:"item_name"`
	Category   string  `json:"category"`
	OutletName string  `json:"outlet_name"`
	Quantity   int64   `json:"quantity"`
	Revenue    float64 `json:"revenue"`
	Tax        float64 `json:"tax"`
	Date       string  `json:"date,omitempty"`
}

type CategoryWiseRow struct {
	CategoryName string  `json:"category_name"`
	OutletName   string  `json:"outlet_name"`
	TotalItems   int64   `json:"total_items"`
	Revenue      float64 `json:"revenue"`
}

// one bill line.
type InvoiceRow struct {
	InvoiceNumber string  `json:"invoice_number"`
	OutletName    string  `json:"outlet_name"`
	Date          string  `json:"date"`
	CustomerName  string  `json:"customer_name"`
	Source        string  `json:"source"`
	Status        string  `json:"status"`
	TotalAmount   float64 `json:"total_amount"`
	CashierName   string  `json:"cashier_name"`
}

// contains cancelled-order info.
type CancelledOrderRow struct {
	InvoiceNumber string  `json:"invoice_number"`
	OutletName    string  `json:"outlet_name"`
	Date          string  `json:"date"`
	Reason        string  `json:"reason"`
	TotalAmount   float64 `json:"total_amount"`
}
type DiscountRow struct {
	InvoiceNumber   string  `json:"invoice_number"`
	OutletName      string  `json:"outlet_name"`
	Date            string  `json:"date"`
	DiscountAmount  float64 `json:"discount_amount"`
	DiscountPercent float64 `json:"discount_percent"`
	TotalAmount     float64 `json:"total_amount"`
	Reason          string  `json:"reason"`
}

// breaks sales down by hour-of-day.
type HourlyRow struct {
	Hour       int     `json:"hour"`
	OutletName string  `json:"outlet_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
}

// shows pax-per-biller stats.
type PaxRow struct {
	BillerName  string  `json:"biller_name"`
	OutletName  string  `json:"outlet_name"`
	TotalOrders int64   `json:"total_orders"`
	TotalPax    int64   `json:"total_pax"`
	Revenue     float64 `json:"revenue"`
}

type DayWiseRow struct {
	Date       string  `json:"date"`
	OutletName string  `json:"outlet_name"`
	Orders     int64   `json:"orders"`
	Revenue    float64 `json:"revenue"`
	NetSales   float64 `json:"net_sales"`
	Tax        float64 `json:"tax"`
}

// – one row per order in the master report.
type OrderMasterRow struct {
	InvoiceNumber    string  `json:"invoice_number"`
	Date             string  `json:"date"`
	OutletName       string  `json:"outlet_name"`
	Source           string  `json:"source"`
	Type             string  `json:"type"`
	Status           string  `json:"status"`
	CustomerName     string  `json:"customer_name"`
	CustomerPhone    string  `json:"customer_phone"`
	Pax              int     `json:"pax"`
	SubTotal         float64 `json:"sub_total"`
	DiscountAmount   float64 `json:"discount_amount"`
	TaxAmount        float64 `json:"tax_amount"`
	DeliveryCharge   float64 `json:"delivery_charge"`
	ContainerCharge  float64 `json:"container_charge"`
	ServiceCharge    float64 `json:"service_charge"`
	AdditionalCharge float64 `json:"additional_charge"`
	TotalAmount      float64 `json:"total_amount"`
	CashierName      string  `json:"cashier_name"`
}
