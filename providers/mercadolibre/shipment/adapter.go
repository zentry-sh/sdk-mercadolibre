package shipment

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http *httputil.Client
	log  logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.Nop()
	}
	return &Adapter{http: http, log: log}
}

func (a *Adapter) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	a.log.Debug("create_shipment", "external_ref", req.ExternalReference)
	return nil, nil
}

func (a *Adapter) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	return nil, nil
}

func (a *Adapter) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	return nil, nil
}

func (a *Adapter) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	return nil, nil
}

func (a *Adapter) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	return nil, nil
}

func (a *Adapter) CancelShipment(ctx context.Context, id string) error {
	return nil
}

func (a *Adapter) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	return nil, nil
}

func (a *Adapter) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	return nil, nil
}
