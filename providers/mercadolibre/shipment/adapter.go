package shipment

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http   *httputil.Client
	logger logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &Adapter{
		http:   http,
		logger: log,
	}
}

func (a *Adapter) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	a.logger.Debug("creating shipment", "external_ref", req.ExternalReference)
	return nil, nil
}

func (a *Adapter) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	a.logger.Debug("getting shipment", "id", id)
	return nil, nil
}

func (a *Adapter) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	a.logger.Debug("getting shipment by order", "order_id", orderID)
	return nil, nil
}

func (a *Adapter) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	a.logger.Debug("listing shipments")
	return nil, nil
}

func (a *Adapter) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	a.logger.Debug("updating shipment", "id", id)
	return nil, nil
}

func (a *Adapter) CancelShipment(ctx context.Context, id string) error {
	a.logger.Debug("cancelling shipment", "id", id)
	return nil
}

func (a *Adapter) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	a.logger.Debug("getting tracking", "shipment_id", shipmentID)
	return nil, nil
}

func (a *Adapter) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	a.logger.Debug("getting label", "shipment_id", shipmentID)
	return nil, nil
}
