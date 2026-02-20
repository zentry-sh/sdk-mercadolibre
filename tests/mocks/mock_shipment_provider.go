package mocks

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type MockShipmentProvider struct {
	CreateShipmentFn     func(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error)
	GetShipmentFn        func(ctx context.Context, id string) (*domain.Shipment, error)
	GetShipmentByOrderFn func(ctx context.Context, orderID string) (*domain.Shipment, error)
	ListShipmentsFn      func(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error)
	UpdateShipmentFn     func(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error)
	CancelShipmentFn     func(ctx context.Context, id string) error
	GetTrackingFn        func(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error)
	GetLabelFn           func(ctx context.Context, shipmentID string) ([]byte, error)
}

func (m *MockShipmentProvider) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	if m.CreateShipmentFn != nil {
		return m.CreateShipmentFn(ctx, req)
	}
	return nil, nil
}

func (m *MockShipmentProvider) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	if m.GetShipmentFn != nil {
		return m.GetShipmentFn(ctx, id)
	}
	return nil, nil
}

func (m *MockShipmentProvider) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	if m.GetShipmentByOrderFn != nil {
		return m.GetShipmentByOrderFn(ctx, orderID)
	}
	return nil, nil
}

func (m *MockShipmentProvider) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	if m.ListShipmentsFn != nil {
		return m.ListShipmentsFn(ctx, filters)
	}
	return nil, nil
}

func (m *MockShipmentProvider) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	if m.UpdateShipmentFn != nil {
		return m.UpdateShipmentFn(ctx, id, req)
	}
	return nil, nil
}

func (m *MockShipmentProvider) CancelShipment(ctx context.Context, id string) error {
	if m.CancelShipmentFn != nil {
		return m.CancelShipmentFn(ctx, id)
	}
	return nil
}

func (m *MockShipmentProvider) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	if m.GetTrackingFn != nil {
		return m.GetTrackingFn(ctx, shipmentID)
	}
	return nil, nil
}

func (m *MockShipmentProvider) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	if m.GetLabelFn != nil {
		return m.GetLabelFn(ctx, shipmentID)
	}
	return nil, nil
}
