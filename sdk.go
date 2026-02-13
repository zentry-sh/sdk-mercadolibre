package sdk

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/payment"
)

type SDK struct {
	config       Config
	client       *mercadolibre.Client
	logger       logger.Logger
	Payment      *PaymentAPI
	Shipment     *ShipmentAPI
	QR           *QRAPI
	Capabilities *CapabilitiesAPI
}

func New(config Config) (*SDK, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if !mercadolibre.IsCountrySupported(config.Country) {
		return nil, errors.InvalidRequest("unsupported country: " + config.Country)
	}

	log := config.Logger
	if log == nil {
		log = logger.NewNopLogger()
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

	sdk := &SDK{
		config: config,
		client: client,
		logger: log,
		Payment: &PaymentAPI{
			service:      paymentService,
			capabilities: capabilitiesService,
			country:      config.Country,
		},
		Shipment: &ShipmentAPI{},
		QR:       &QRAPI{},
		Capabilities: &CapabilitiesAPI{
			service: capabilitiesService,
			country: config.Country,
		},
	}

	return sdk, nil
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
	service *usecases.ShipmentService
}

func (s *ShipmentAPI) Create(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.CreateShipment(ctx, req)
}

func (s *ShipmentAPI) Get(ctx context.Context, id string) (*domain.Shipment, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.GetShipment(ctx, id)
}

func (s *ShipmentAPI) GetByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.GetShipmentByOrder(ctx, orderID)
}

func (s *ShipmentAPI) List(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.ListShipments(ctx, filters)
}

func (s *ShipmentAPI) Update(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.UpdateShipment(ctx, id, req)
}

func (s *ShipmentAPI) Cancel(ctx context.Context, id string) error {
	if s.service == nil {
		return errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.CancelShipment(ctx, id)
}

func (s *ShipmentAPI) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.GetTracking(ctx, shipmentID)
}

func (s *ShipmentAPI) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	if s.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "shipment service not initialized")
	}
	return s.service.GetLabel(ctx, shipmentID)
}

type QRAPI struct {
	service *usecases.QRService
}

func (q *QRAPI) Create(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.CreateQR(ctx, req)
}

func (q *QRAPI) Get(ctx context.Context, qrID string) (*domain.QRCode, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.GetQR(ctx, qrID)
}

func (q *QRAPI) GetByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.GetQRByExternalReference(ctx, ref)
}

func (q *QRAPI) Delete(ctx context.Context, qrID string) error {
	if q.service == nil {
		return errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.DeleteQR(ctx, qrID)
}

func (q *QRAPI) GetPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.GetQRPayment(ctx, qrID)
}

func (q *QRAPI) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.RegisterPOS(ctx, req)
}

func (q *QRAPI) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.GetPOS(ctx, posID)
}

func (q *QRAPI) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.ListPOS(ctx, storeID)
}

func (q *QRAPI) DeletePOS(ctx context.Context, posID string) error {
	if q.service == nil {
		return errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.DeletePOS(ctx, posID)
}

func (q *QRAPI) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.RegisterStore(ctx, req)
}

func (q *QRAPI) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.GetStore(ctx, storeID)
}

func (q *QRAPI) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	if q.service == nil {
		return nil, errors.NewError(errors.ErrCodeInternal, "QR service not initialized")
	}
	return q.service.ListStores(ctx)
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
