package domain

type WebhookEventType string

const (
	WebhookPaymentCreated    WebhookEventType = "payment.created"
	WebhookPaymentUpdated    WebhookEventType = "payment.updated"
	WebhookRefundCreated     WebhookEventType = "refund.created"
	WebhookChargebackCreated WebhookEventType = "chargeback.created"
	WebhookShipmentCreated   WebhookEventType = "shipment.created"
	WebhookShipmentUpdated   WebhookEventType = "shipment.updated"
	WebhookQRScanned         WebhookEventType = "qr.scanned"
	WebhookQRPaid            WebhookEventType = "qr.paid"
)

func (t WebhookEventType) String() string {
	return string(t)
}

type WebhookRequest struct {
	Body      []byte
	Signature string
	RequestID string
	DataID    string
}

type WebhookEvent struct {
	ID          int64
	Type        WebhookEventType
	Action      string
	LiveMode    bool
	APIVersion  string
	UserID      int64
	DateCreated string
	DataID      string
}

func (e *WebhookEvent) IsPaymentEvent() bool {
	return e.Type == WebhookPaymentCreated || e.Type == WebhookPaymentUpdated
}

func (e *WebhookEvent) IsShipmentEvent() bool {
	return e.Type == WebhookShipmentCreated || e.Type == WebhookShipmentUpdated
}

func (e *WebhookEvent) IsQREvent() bool {
	return e.Type == WebhookQRScanned || e.Type == WebhookQRPaid
}

func (e *WebhookEvent) IsRefundEvent() bool {
	return e.Type == WebhookRefundCreated
}

func (e *WebhookEvent) IsChargebackEvent() bool {
	return e.Type == WebhookChargebackCreated
}
