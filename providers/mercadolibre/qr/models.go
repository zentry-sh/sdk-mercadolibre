package qr

import "time"

type MLCreateOrderRequest struct {
	Type              string          `json:"type"`
	ExternalReference string          `json:"external_reference"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	TotalAmount       float64         `json:"total_amount"`
	ExpirationDate    *time.Time      `json:"expiration_date,omitempty"`
	NotificationURL   string          `json:"notification_url,omitempty"`
	Items             []MLOrderItem   `json:"items"`
	CashOut           *MLCashOut      `json:"cash_out,omitempty"`
	Sponsor           *MLSponsor      `json:"sponsor,omitempty"`
}

type MLOrderItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalAmount float64 `json:"total_amount"`
	UnitMeasure string  `json:"unit_measure,omitempty"`
}

type MLCashOut struct {
	Amount float64 `json:"amount"`
}

type MLSponsor struct {
	ID int64 `json:"id"`
}

type MLOrderResponse struct {
	ID                string          `json:"id"`
	ExternalReference string          `json:"external_reference"`
	Type              string          `json:"type"`
	Status            string          `json:"status"`
	StatusDetail      string          `json:"status_detail"`
	TotalAmount       float64         `json:"total_amount"`
	PaidAmount        float64         `json:"paid_amount"`
	UserID            int64           `json:"user_id"`
	Items             []MLOrderItem   `json:"items"`
	Payments          []MLOrderPayment `json:"payments"`
	QRData            string          `json:"qr_data"`
	OrderStatus       string          `json:"order_status"`
	POSID             string          `json:"pos_id"`
	StoreID           string          `json:"store_id"`
	CollectorID       int64           `json:"collector_id"`
	CreatedDate       *time.Time      `json:"created_date"`
	LastUpdatedDate   *time.Time      `json:"last_updated_date"`
	ExpirationDate    *time.Time      `json:"expiration_date"`
}

type MLOrderPayment struct {
	ID              int64   `json:"id"`
	TransactionAmount float64 `json:"transaction_amount"`
	Status          string  `json:"status"`
	StatusDetail    string  `json:"status_detail"`
	PaymentMethodID string  `json:"payment_method_id"`
	PaymentTypeID   string  `json:"payment_type_id"`
	Installments    int     `json:"installments"`
}

type MLOrderSearchResponse struct {
	Elements []MLOrderResponse `json:"elements"`
	Total    int               `json:"total"`
}

type MLPOSRequest struct {
	Name        string `json:"name"`
	ExternalID  string `json:"external_id"`
	StoreID     string `json:"store_id,omitempty"`
	FixedAmount bool   `json:"fixed_amount"`
	Category    int    `json:"category,omitempty"`
}

type MLPOSResponse struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	ExternalID  string     `json:"external_id"`
	StoreID     string     `json:"store_id"`
	FixedAmount bool       `json:"fixed_amount"`
	Category    int        `json:"category"`
	QR          *MLPOSQR   `json:"qr"`
	URL         string     `json:"url"`
	DateCreated *time.Time `json:"date_created"`
	DateLastUpdated *time.Time `json:"date_last_updated"`
}

type MLPOSQR struct {
	Image            string `json:"image"`
	TemplateDocument string `json:"template_document"`
	TemplateImage    string `json:"template_image"`
}

type MLPOSSearchResponse struct {
	Paging   MLQRPaging      `json:"paging"`
	Results  []MLPOSResponse `json:"results"`
}

type MLQRPaging struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type MLStoreRequest struct {
	Name          string            `json:"name"`
	ExternalID    string            `json:"external_id,omitempty"`
	BusinessHours map[string]string `json:"business_hours,omitempty"`
	Location      *MLStoreLocation  `json:"location,omitempty"`
}

type MLStoreResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	ExternalID    string            `json:"external_id"`
	BusinessHours map[string]string `json:"business_hours"`
	Location      *MLStoreLocation  `json:"location"`
	DateCreated   *time.Time        `json:"date_creation"`
}

type MLStoreLocation struct {
	StreetNumber string  `json:"street_number"`
	StreetName   string  `json:"street_name"`
	CityName     string  `json:"city_name"`
	StateName    string  `json:"state_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Reference    string  `json:"reference"`
}

type MLStoreSearchResponse struct {
	Paging  MLQRPaging        `json:"paging"`
	Results []MLStoreResponse `json:"results"`
}

type MLUserResponse struct {
	ID int64 `json:"id"`
}
