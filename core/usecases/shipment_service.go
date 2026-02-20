package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/pkg/sanitize"
)

type ShipmentService struct {
	provider ports.ShipmentProvider
	log      logger.Logger
}

func NewShipmentService(provider ports.ShipmentProvider, log logger.Logger) *ShipmentService {
	if log == nil {
		log = logger.Nop()
	}
	return &ShipmentService{
		provider: provider,
		log:      log,
	}
}

func (s *ShipmentService) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	req.ExternalReference = sanitize.String(req.ExternalReference)
	req.OrderID = sanitize.ID(req.OrderID)
	req.CarrierID = sanitize.ID(req.CarrierID)
	req.ServiceType = sanitize.String(req.ServiceType)
	req.Origin = sanitizeAddress(req.Origin)
	req.Destination = sanitizeAddress(req.Destination)

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	s.log.Debug("create_shipment", "order_id", req.OrderID, "external_ref", req.ExternalReference)
	return s.provider.CreateShipment(ctx, req)
}

func (s *ShipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	id = sanitize.ID(id)
	if id == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}
	return s.provider.GetShipment(ctx, id)
}

func (s *ShipmentService) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	orderID = sanitize.ID(orderID)
	if orderID == "" {
		return nil, errors.InvalidRequest("order id is required")
	}
	return s.provider.GetShipmentByOrder(ctx, orderID)
}

func (s *ShipmentService) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	filters.OrderID = sanitize.ID(filters.OrderID)
	filters.ExternalReference = sanitize.String(filters.ExternalReference)
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	return s.provider.ListShipments(ctx, filters)
}

func (s *ShipmentService) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	id = sanitize.ID(id)
	if id == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}
	if req.Destination != nil {
		addr := sanitizeAddress(*req.Destination)
		req.Destination = &addr
	}
	return s.provider.UpdateShipment(ctx, id, req)
}

func (s *ShipmentService) CancelShipment(ctx context.Context, id string) error {
	id = sanitize.ID(id)
	if id == "" {
		return errors.InvalidRequest("shipment id is required")
	}
	return s.provider.CancelShipment(ctx, id)
}

func (s *ShipmentService) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	shipmentID = sanitize.ID(shipmentID)
	if shipmentID == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}
	return s.provider.GetTracking(ctx, shipmentID)
}

func (s *ShipmentService) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	shipmentID = sanitize.ID(shipmentID)
	if shipmentID == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}
	return s.provider.GetLabel(ctx, shipmentID)
}

func (s *ShipmentService) validateCreateRequest(req *domain.CreateShipmentRequest) error {
	if req.ExternalReference == "" {
		return errors.InvalidRequest("external_reference is required")
	}
	if req.Destination.IsEmpty() {
		return errors.InvalidRequest("destination address is required")
	}
	return nil
}

func sanitizeAddress(addr domain.Address) domain.Address {
	addr.Street = sanitize.String(addr.Street)
	addr.Number = sanitize.String(addr.Number)
	addr.Floor = sanitize.String(addr.Floor)
	addr.Apartment = sanitize.String(addr.Apartment)
	addr.City = sanitize.String(addr.City)
	addr.State = sanitize.String(addr.State)
	addr.ZipCode = sanitize.String(addr.ZipCode)
	addr.Country = sanitize.String(addr.Country)
	return addr
}
