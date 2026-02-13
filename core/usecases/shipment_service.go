package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type ShipmentService struct {
	provider ports.ShipmentProvider
	logger   logger.Logger
}

func NewShipmentService(provider ports.ShipmentProvider, log logger.Logger) *ShipmentService {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &ShipmentService{
		provider: provider,
		logger:   log,
	}
}

func (s *ShipmentService) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	s.logger.Info("creating shipment",
		"order_id", req.OrderID,
		"external_ref", req.ExternalReference)

	shipment, err := s.provider.CreateShipment(ctx, req)
	if err != nil {
		s.logger.Error("failed to create shipment", "error", err)
		return nil, err
	}

	s.logger.Info("shipment created",
		"id", shipment.ID,
		"status", shipment.Status.String())

	return shipment, nil
}

func (s *ShipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	if id == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}

	return s.provider.GetShipment(ctx, id)
}

func (s *ShipmentService) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	if orderID == "" {
		return nil, errors.InvalidRequest("order id is required")
	}

	return s.provider.GetShipmentByOrder(ctx, orderID)
}

func (s *ShipmentService) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return s.provider.ListShipments(ctx, filters)
}

func (s *ShipmentService) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	if id == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}

	s.logger.Info("updating shipment", "id", id)

	shipment, err := s.provider.UpdateShipment(ctx, id, req)
	if err != nil {
		s.logger.Error("failed to update shipment", "error", err)
		return nil, err
	}

	s.logger.Info("shipment updated", "id", shipment.ID)
	return shipment, nil
}

func (s *ShipmentService) CancelShipment(ctx context.Context, id string) error {
	if id == "" {
		return errors.InvalidRequest("shipment id is required")
	}

	s.logger.Info("cancelling shipment", "id", id)

	err := s.provider.CancelShipment(ctx, id)
	if err != nil {
		s.logger.Error("failed to cancel shipment", "error", err)
		return err
	}

	s.logger.Info("shipment cancelled", "id", id)
	return nil
}

func (s *ShipmentService) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	if shipmentID == "" {
		return nil, errors.InvalidRequest("shipment id is required")
	}

	return s.provider.GetTracking(ctx, shipmentID)
}

func (s *ShipmentService) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
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
