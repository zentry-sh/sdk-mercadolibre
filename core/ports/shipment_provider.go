package ports

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type ShipmentProvider interface {
	CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error)
	GetShipment(ctx context.Context, id string) (*domain.Shipment, error)
	GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error)
	ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error)
	UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error)
	CancelShipment(ctx context.Context, id string) error
	GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error)
	GetLabel(ctx context.Context, shipmentID string) ([]byte, error)
}
