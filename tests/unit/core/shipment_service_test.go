package core

import (
	"context"
	"testing"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/tests/mocks"
)

func TestShipmentService_GetShipment(t *testing.T) {
	mockProvider := &mocks.MockShipmentProvider{
		GetShipmentFn: func(ctx context.Context, id string) (*domain.Shipment, error) {
			return &domain.Shipment{
				ID:             id,
				Status:         domain.ShipmentStatusInTransit,
				TrackingNumber: "PE123456789",
				Carrier:        domain.Carrier{ID: "olva", Name: "Olva Courier"},
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	service := usecases.NewShipmentService(mockProvider, nil)

	shipment, err := service.GetShipment(context.Background(), "12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if shipment.ID != "12345" {
		t.Errorf("expected ID '12345', got '%s'", shipment.ID)
	}

	if shipment.Status != domain.ShipmentStatusInTransit {
		t.Errorf("expected status InTransit, got %s", shipment.Status.String())
	}

	if shipment.TrackingNumber != "PE123456789" {
		t.Errorf("expected tracking 'PE123456789', got '%s'", shipment.TrackingNumber)
	}
}

func TestShipmentService_GetShipment_Validation(t *testing.T) {
	service := usecases.NewShipmentService(&mocks.MockShipmentProvider{}, nil)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"empty id", "", true},
		{"null bytes in id", "\x00\x00", true},
		{"valid id", "12345", false},
		{"id with special chars", "12345!@#$", false},
		{"id sanitized to valid", "abc-123_def", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetShipment(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetShipment(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestShipmentService_CreateShipment_Validation(t *testing.T) {
	mockProvider := &mocks.MockShipmentProvider{
		CreateShipmentFn: func(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
			return &domain.Shipment{
				ID:                "ship-001",
				ExternalReference: req.ExternalReference,
				Status:            domain.ShipmentStatusPending,
			}, nil
		},
	}
	service := usecases.NewShipmentService(mockProvider, nil)

	tests := []struct {
		name    string
		req     *domain.CreateShipmentRequest
		wantErr bool
	}{
		{
			name: "missing external reference",
			req: &domain.CreateShipmentRequest{
				Destination: domain.Address{Street: "Av Lima", City: "Lima", Country: "PE"},
			},
			wantErr: true,
		},
		{
			name: "missing destination",
			req: &domain.CreateShipmentRequest{
				ExternalReference: "order-001",
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &domain.CreateShipmentRequest{
				ExternalReference: "order-001",
				Destination:       domain.Address{Street: "Av Lima", City: "Lima", Country: "PE"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateShipment(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShipment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShipmentService_Sanitization(t *testing.T) {
	var capturedID string
	mockProvider := &mocks.MockShipmentProvider{
		GetShipmentFn: func(ctx context.Context, id string) (*domain.Shipment, error) {
			capturedID = id
			return &domain.Shipment{ID: id}, nil
		},
	}
	service := usecases.NewShipmentService(mockProvider, nil)

	_, _ = service.GetShipment(context.Background(), "  12345  ")
	if capturedID != "12345" {
		t.Errorf("expected sanitized ID '12345', got '%s'", capturedID)
	}
}

func TestShipmentService_ListShipments_LimitClamp(t *testing.T) {
	var capturedFilters domain.ShipmentFilters
	mockProvider := &mocks.MockShipmentProvider{
		ListShipmentsFn: func(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
			capturedFilters = filters
			return nil, nil
		},
	}
	service := usecases.NewShipmentService(mockProvider, nil)

	_, _ = service.ListShipments(context.Background(), domain.ShipmentFilters{Limit: 0})
	if capturedFilters.Limit != 50 {
		t.Errorf("expected default limit 50, got %d", capturedFilters.Limit)
	}

	_, _ = service.ListShipments(context.Background(), domain.ShipmentFilters{Limit: 999})
	if capturedFilters.Limit != 100 {
		t.Errorf("expected clamped limit 100, got %d", capturedFilters.Limit)
	}
}

func TestShipmentStatus_String(t *testing.T) {
	tests := []struct {
		status   domain.ShipmentStatus
		expected string
	}{
		{domain.ShipmentStatusPending, "pending"},
		{domain.ShipmentStatusReadyToShip, "ready_to_ship"},
		{domain.ShipmentStatusShipped, "shipped"},
		{domain.ShipmentStatusInTransit, "in_transit"},
		{domain.ShipmentStatusOutForDelivery, "out_for_delivery"},
		{domain.ShipmentStatusDelivered, "delivered"},
		{domain.ShipmentStatusCancelled, "cancelled"},
		{domain.ShipmentStatusReturned, "returned"},
		{domain.ShipmentStatusNotDelivered, "not_delivered"},
		{domain.ShipmentStatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("ShipmentStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShipmentStatus_IsFinal(t *testing.T) {
	finalStatuses := []domain.ShipmentStatus{
		domain.ShipmentStatusDelivered,
		domain.ShipmentStatusCancelled,
		domain.ShipmentStatusReturned,
	}
	nonFinalStatuses := []domain.ShipmentStatus{
		domain.ShipmentStatusPending,
		domain.ShipmentStatusReadyToShip,
		domain.ShipmentStatusInTransit,
		domain.ShipmentStatusShipped,
	}

	for _, s := range finalStatuses {
		if !s.IsFinal() {
			t.Errorf("expected %s to be final", s.String())
		}
	}
	for _, s := range nonFinalStatuses {
		if s.IsFinal() {
			t.Errorf("expected %s to not be final", s.String())
		}
	}
}

func TestShipmentStatus_CanCancel(t *testing.T) {
	cancellable := []domain.ShipmentStatus{
		domain.ShipmentStatusPending,
		domain.ShipmentStatusReadyToShip,
	}
	notCancellable := []domain.ShipmentStatus{
		domain.ShipmentStatusInTransit,
		domain.ShipmentStatusDelivered,
		domain.ShipmentStatusCancelled,
	}

	for _, s := range cancellable {
		if !s.CanCancel() {
			t.Errorf("expected %s to be cancellable", s.String())
		}
	}
	for _, s := range notCancellable {
		if s.CanCancel() {
			t.Errorf("expected %s to not be cancellable", s.String())
		}
	}
}
