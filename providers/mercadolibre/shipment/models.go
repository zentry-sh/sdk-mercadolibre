package shipment

import "time"

type MLShipmentResponse struct {
	ID                int64                     `json:"id"`
	ExternalReference string                    `json:"external_reference"`
	Status            string                    `json:"status"`
	SubStatus         string                    `json:"substatus"`
	StatusHistory     *MLStatusHistory          `json:"status_history"`
	TrackingNumber    string                    `json:"tracking_number"`
	TrackingMethod    string                    `json:"tracking_method"`
	ServiceID         int64                     `json:"service_id"`
	SenderID          int64                     `json:"sender_id"`
	ReceiverID        int64                     `json:"receiver_id"`
	SenderAddress     *MLShippingAddress        `json:"sender_address"`
	ReceiverAddress   *MLShippingAddress        `json:"receiver_address"`
	ShippingItems     []MLShippingItem          `json:"shipping_items"`
	ShippingOption    *MLShippingOption         `json:"shipping_option"`
	LogisticType      string                    `json:"logistic_type"`
	OrderID           int64                     `json:"order_id"`
	DateCreated       time.Time                 `json:"date_created"`
	LastUpdated       time.Time                 `json:"last_updated"`
	DateFirstPrinted  *time.Time                `json:"date_first_printed"`
	Carrier           *MLCarrierInfo            `json:"carrier"`
}

type MLShippingAddress struct {
	AddressID    int64   `json:"id"`
	AddressLine  string  `json:"address_line"`
	StreetName   string  `json:"street_name"`
	StreetNumber string  `json:"street_number"`
	ZipCode      string  `json:"zip_code"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	City         *MLGeo  `json:"city"`
	State        *MLGeo  `json:"state"`
	Country      *MLGeo  `json:"country"`
	Comment      string  `json:"comment"`
}

type MLGeo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MLShippingItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"dimensions_source_weight"`
	Length      float64 `json:"dimensions_source_length"`
	Width       float64 `json:"dimensions_source_width"`
	Height      float64 `json:"dimensions_source_height"`
}

type MLShippingOption struct {
	ID                    int64              `json:"id"`
	Name                  string             `json:"name"`
	ShippingMethodID      int64              `json:"shipping_method_id"`
	EstimatedDeliveryTime *MLEstimatedTime   `json:"estimated_delivery_time"`
	ListCost              float64            `json:"list_cost"`
	Cost                  float64            `json:"cost"`
	CurrencyID            string             `json:"currency_id"`
}

type MLEstimatedTime struct {
	Type    string     `json:"type"`
	Date    *time.Time `json:"date"`
	Unit    string     `json:"unit"`
	Offset  *MLOffset  `json:"offset"`
}

type MLOffset struct {
	Date *time.Time `json:"date"`
	From int        `json:"from"`
	To   int        `json:"to"`
}

type MLStatusHistory struct {
	DateCancelled     *time.Time `json:"date_cancelled"`
	DateDelivered     *time.Time `json:"date_delivered"`
	DateHandling      *time.Time `json:"date_handling"`
	DateNotDelivered  *time.Time `json:"date_not_delivered"`
	DateReadyToShip   *time.Time `json:"date_ready_to_ship"`
	DateShipped       *time.Time `json:"date_shipped"`
	DateReturned      *time.Time `json:"date_returned"`
}

type MLCarrierInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type MLShipmentHistoryEntry struct {
	Status      string    `json:"status"`
	SubStatus   string    `json:"substatus"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type MLShipmentSearchResponse struct {
	Paging  MLPaging             `json:"paging"`
	Results []MLShipmentResponse `json:"results"`
}

type MLPaging struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type MLUpdateShipmentRequest struct {
	ReceiverAddress *MLUpdateAddress `json:"receiver_address,omitempty"`
	TrackingNumber  string           `json:"tracking_number,omitempty"`
}

type MLUpdateAddress struct {
	StreetName   string `json:"street_name,omitempty"`
	StreetNumber string `json:"street_number,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
	Comment      string `json:"comment,omitempty"`
}
