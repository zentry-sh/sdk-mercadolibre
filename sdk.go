package sdk

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/payment"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/qr"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/shipment"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/webhook"
)

type SDK struct {
	config       Config
	client       *mercadolibre.Client
	Payment      *PaymentAPI
	Shipment     *ShipmentAPI
	QR           *QRAPI
	Webhook      *WebhookAPI
	Capabilities *CapabilitiesAPI
}

func New(config Config) (*SDK, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if !mercadolibre.IsCountrySupported(config.Country) {
		return nil, errors.InvalidRequest(fmt.Sprintf("unsupported country: %s", config.Country))
	}

	log := config.Logger
	if log == nil {
		log = logger.Nop()
	}

	client := mercadolibre.NewClient(mercadolibre.Config{
		AccessToken:   config.AccessToken,
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		Country:       config.Country,
		Timeout:       config.Timeout,
		Logger:        log,
		WebhookSecret: config.WebhookSecret,
	})

	capabilitiesAdapter := mercadolibre.NewCapabilitiesAdapter()
	capabilitiesService := usecases.NewCapabilitiesService(capabilitiesAdapter)

	paymentAdapter := payment.NewAdapter(client.PaymentsHTTP(), log)
	paymentService := usecases.NewPaymentService(paymentAdapter, log)

	shipmentAdapter := shipment.NewAdapter(client.ShipmentsHTTP(), log)
	shipmentService := usecases.NewShipmentService(shipmentAdapter, log)

	qrAdapter := qr.NewAdapter(client.QRHTTP(), log)
	qrService := usecases.NewQRService(qrAdapter, log)

	webhookHandler := webhook.NewHandler(log)
	webhookService := usecases.NewWebhookService(webhookHandler, log)

	return &SDK{
		config: config,
		client: client,
		Payment: &PaymentAPI{
			service:      paymentService,
			capabilities: capabilitiesService,
			country:      config.Country,
		},
		Shipment: &ShipmentAPI{
			service:      shipmentService,
			capabilities: capabilitiesService,
			country:      config.Country,
		},
		QR: &QRAPI{
			service:      qrService,
			capabilities: capabilitiesService,
			country:      config.Country,
		},
		Webhook: &WebhookAPI{
			service: webhookService,
			secret:  config.WebhookSecret,
		},
		Capabilities: &CapabilitiesAPI{
			service: capabilitiesService,
			country: config.Country,
		},
	}, nil
}

func (s *SDK) SetAccessToken(token string) {
	s.config.AccessToken = token
	s.client.SetAccessToken(token)
}

func (s *SDK) Country() string {
	return s.config.Country
}

func (s *SDK) ForCountry(country string) (*SDK, error) {
	newConfig := s.config
	newConfig.Country = country
	return New(newConfig)
}

type PaymentAPI struct {
	service      *usecases.PaymentService
	capabilities *usecases.CapabilitiesService
	country      string
}

func (p *PaymentAPI) Create(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	if p.capabilities != nil {
		if err := p.capabilities.ValidatePaymentRequest(ctx, p.country, req); err != nil {
			return nil, err
		}
	}
	return p.service.CreatePayment(ctx, req)
}

func (p *PaymentAPI) Get(ctx context.Context, id string) (*domain.Payment, error) {
	return p.service.GetPayment(ctx, id)
}

func (p *PaymentAPI) List(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	return p.service.ListPayments(ctx, filters)
}

func (p *PaymentAPI) Refund(ctx context.Context, paymentID string, amount *domain.Money) (*domain.Refund, error) {
	return p.service.RefundPayment(ctx, paymentID, amount)
}

func (p *PaymentAPI) Cancel(ctx context.Context, paymentID string) error {
	return p.service.CancelPayment(ctx, paymentID)
}

func (p *PaymentAPI) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	return p.service.GetRefund(ctx, paymentID, refundID)
}

func (p *PaymentAPI) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
	return p.service.ListRefunds(ctx, paymentID)
}

type ShipmentAPI struct {
	service      *usecases.ShipmentService
	capabilities *usecases.CapabilitiesService
	country      string
}

func (s *ShipmentAPI) Create(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	if s.capabilities != nil {
		if err := s.capabilities.ValidateShipmentRequest(ctx, s.country, req); err != nil {
			return nil, err
		}
	}
	return s.service.CreateShipment(ctx, req)
}

func (s *ShipmentAPI) Get(ctx context.Context, id string) (*domain.Shipment, error) {
	return s.service.GetShipment(ctx, id)
}

func (s *ShipmentAPI) GetByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	return s.service.GetShipmentByOrder(ctx, orderID)
}

func (s *ShipmentAPI) List(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	return s.service.ListShipments(ctx, filters)
}

func (s *ShipmentAPI) Update(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	return s.service.UpdateShipment(ctx, id, req)
}

func (s *ShipmentAPI) Cancel(ctx context.Context, id string) error {
	return s.service.CancelShipment(ctx, id)
}

func (s *ShipmentAPI) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	return s.service.GetTracking(ctx, shipmentID)
}

func (s *ShipmentAPI) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	return s.service.GetLabel(ctx, shipmentID)
}

type QRAPI struct {
	service      *usecases.QRService
	capabilities *usecases.CapabilitiesService
	country      string
}

func (q *QRAPI) Create(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	if q.capabilities != nil {
		if err := q.capabilities.ValidateQRRequest(ctx, q.country, req); err != nil {
			return nil, err
		}
	}
	return q.service.CreateQR(ctx, req)
}

func (q *QRAPI) Get(ctx context.Context, qrID string) (*domain.QRCode, error) {
	return q.service.GetQR(ctx, qrID)
}

func (q *QRAPI) GetByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	return q.service.GetQRByExternalReference(ctx, ref)
}

func (q *QRAPI) Delete(ctx context.Context, qrID string) error {
	return q.service.DeleteQR(ctx, qrID)
}

func (q *QRAPI) GetPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	return q.service.GetQRPayment(ctx, qrID)
}

func (q *QRAPI) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	return q.service.RegisterPOS(ctx, req)
}

func (q *QRAPI) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	return q.service.GetPOS(ctx, posID)
}

func (q *QRAPI) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	return q.service.ListPOS(ctx, storeID)
}

func (q *QRAPI) DeletePOS(ctx context.Context, posID string) error {
	return q.service.DeletePOS(ctx, posID)
}

func (q *QRAPI) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	return q.service.RegisterStore(ctx, req)
}

func (q *QRAPI) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	return q.service.GetStore(ctx, storeID)
}

func (q *QRAPI) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	return q.service.ListStores(ctx)
}

// WebhookAPI exposes webhook validation and parsing to SDK consumers.
type WebhookAPI struct {
	service *usecases.WebhookService
	secret  string
}

// WebhookHandlerFunc is the callback signature for processing validated webhook events.
type WebhookHandlerFunc func(ctx context.Context, event *domain.WebhookEvent) error

// Process validates the HMAC signature and parses the webhook payload in a single call.
func (w *WebhookAPI) Process(ctx context.Context, req domain.WebhookRequest) (*domain.WebhookEvent, error) {
	return w.service.Process(ctx, req, w.secret)
}

// Validate only checks the HMAC-SHA256 signature without parsing the body.
func (w *WebhookAPI) Validate(ctx context.Context, req domain.WebhookRequest) error {
	return w.service.ValidateSignature(ctx, req, w.secret)
}

// Parse deserializes the webhook payload without signature validation.
// Use this only if you have already validated the signature externally.
func (w *WebhookAPI) Parse(ctx context.Context, payload []byte) (*domain.WebhookEvent, error) {
	return w.service.Parse(ctx, payload)
}

// maxWebhookBody limits the webhook request body to 10 MiB (consistent with httputil.Client).
const maxWebhookBody = 10 << 20

// HTTPHandler returns a net/http Handler that extracts webhook headers,
// validates the HMAC-SHA256 signature, parses the event, and calls fn.
// The handler responds 200 on success, 400 on bad request, and 401 on
// signature verification failure — all within Mercado Pago's 22-second window.
//
// Usage:
//
//	http.Handle("/webhooks", client.Webhook.HTTPHandler(func(ctx context.Context, event *domain.WebhookEvent) error {
//	    log.Printf("event: %s data_id: %s", event.Type, event.DataID)
//	    return nil
//	}))
func (w *WebhookAPI) HTTPHandler(fn WebhookHandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
		defer r.Body.Close()
		if err != nil {
			http.Error(rw, "failed to read body", http.StatusBadRequest)
			return
		}

		req := domain.WebhookRequest{
			Body:      body,
			Signature: r.Header.Get("x-signature"),
			RequestID: r.Header.Get("x-request-id"),
			DataID:    r.URL.Query().Get("data.id"),
		}

		event, err := w.Process(r.Context(), req)
		if err != nil {
			if sdkErr, ok := err.(*errors.SDKError); ok && sdkErr.Code == errors.ErrCodeInvalidWebhook {
				http.Error(rw, sdkErr.Message, http.StatusUnauthorized)
				return
			}
			http.Error(rw, "bad request", http.StatusBadRequest)
			return
		}

		if err := fn(r.Context(), event); err != nil {
			http.Error(rw, "handler error", http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
	})
}

type CapabilitiesAPI struct {
	service *usecases.CapabilitiesService
	country string
}

func (c *CapabilitiesAPI) Get(ctx context.Context) (*domain.RegionCapabilities, error) {
	return c.service.GetCapabilities(ctx, c.country)
}

func (c *CapabilitiesAPI) GetForCountry(ctx context.Context, countryCode string) (*domain.RegionCapabilities, error) {
	return c.service.GetCapabilities(ctx, countryCode)
}

func (c *CapabilitiesAPI) ListRegions(ctx context.Context) ([]domain.Region, error) {
	return c.service.ListSupportedRegions(ctx)
}

func (c *CapabilitiesAPI) GetPaymentMethods(ctx context.Context) ([]domain.PaymentMethodInfo, error) {
	return c.service.GetPaymentMethods(ctx, c.country)
}

func (c *CapabilitiesAPI) GetCarriers(ctx context.Context) ([]domain.CarrierInfo, error) {
	return c.service.GetCarriers(ctx, c.country)
}

func (c *CapabilitiesAPI) IsQRSupported(ctx context.Context) (bool, error) {
	return c.service.IsQRSupported(ctx, c.country)
}

func (c *CapabilitiesAPI) GetCurrency(ctx context.Context) (string, error) {
	return c.service.GetCurrency(ctx, c.country)
}

func (c *CapabilitiesAPI) GetMaxInstallments(ctx context.Context) (int, error) {
	return c.service.GetMaxInstallments(ctx, c.country)
}
