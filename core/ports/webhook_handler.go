package ports

type WebhookHandler interface {
	ValidateSignature(payload []byte, signature string, secret string) error
	ParsePaymentWebhook(payload []byte) (*PaymentWebhookEvent, error)
	ParseShipmentWebhook(payload []byte) (*ShipmentWebhookEvent, error)
	ParseQRWebhook(payload []byte) (*QRWebhookEvent, error)
}

type WebhookEventType string

const (
	WebhookEventPaymentCreated   WebhookEventType = "payment.created"
	WebhookEventPaymentUpdated   WebhookEventType = "payment.updated"
	WebhookEventRefundCreated    WebhookEventType = "refund.created"
	WebhookEventChargebackCreated WebhookEventType = "chargeback.created"
	WebhookEventShipmentCreated  WebhookEventType = "shipment.created"
	WebhookEventShipmentUpdated  WebhookEventType = "shipment.updated"
	WebhookEventQRScanned        WebhookEventType = "qr.scanned"
	WebhookEventQRPaid           WebhookEventType = "qr.paid"
)

type BaseWebhookEvent struct {
	ID        string
	Type      WebhookEventType
	Action    string
	LiveMode  bool
	Timestamp int64
	Data      map[string]interface{}
}

type PaymentWebhookEvent struct {
	BaseWebhookEvent
	PaymentID    string
	Status       string
	StatusDetail string
}

type ShipmentWebhookEvent struct {
	BaseWebhookEvent
	ShipmentID string
	Status     string
	SubStatus  string
}

type QRWebhookEvent struct {
	BaseWebhookEvent
	QRID      string
	PaymentID string
	Status    string
}
