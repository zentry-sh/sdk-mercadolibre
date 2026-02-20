package shipment

import (
	"testing"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	shipmentpkg "github.com/zentry/sdk-mercadolibre/providers/mercadolibre/shipment"
)

func TestMapper_ToDomainShipment(t *testing.T) {
	m := shipmentpkg.NewMapper()
	now := time.Now()

	mlResp := &shipmentpkg.MLShipmentResponse{
		ID:                12345,
		ExternalReference: "order-ext-001",
		Status:            "in_transit",
		SubStatus:         "en_camino",
		TrackingNumber:    "PE123456789",
		LogisticType:      "cross_docking",
		OrderID:           67890,
		DateCreated:       now,
		LastUpdated:       now,
		SenderAddress: &shipmentpkg.MLShippingAddress{
			StreetName:   "Av Javier Prado",
			StreetNumber: "1234",
			ZipCode:      "15024",
			City:         &shipmentpkg.MLGeo{Name: "San Isidro"},
			State:        &shipmentpkg.MLGeo{Name: "Lima"},
			Country:      &shipmentpkg.MLGeo{Name: "Peru"},
		},
		ReceiverAddress: &shipmentpkg.MLShippingAddress{
			StreetName:   "Calle Los Pinos",
			StreetNumber: "567",
			ZipCode:      "15036",
			City:         &shipmentpkg.MLGeo{Name: "Miraflores"},
			State:        &shipmentpkg.MLGeo{Name: "Lima"},
			Country:      &shipmentpkg.MLGeo{Name: "Peru"},
		},
		Carrier: &shipmentpkg.MLCarrierInfo{
			ID:   1,
			Name: "Olva Courier",
		},
		ShippingItems: []shipmentpkg.MLShippingItem{
			{
				Description: "Laptop",
				Weight:      2.5,
				Length:      40,
				Width:       30,
				Height:      5,
			},
		},
	}

	result := m.ToDomainShipment(mlResp)

	if result.ID != "12345" {
		t.Errorf("expected ID '12345', got '%s'", result.ID)
	}
	if result.OrderID != "67890" {
		t.Errorf("expected OrderID '67890', got '%s'", result.OrderID)
	}
	if result.Status != domain.ShipmentStatusInTransit {
		t.Errorf("expected status InTransit, got %s", result.Status.String())
	}
	if result.TrackingNumber != "PE123456789" {
		t.Errorf("expected tracking 'PE123456789', got '%s'", result.TrackingNumber)
	}
	if result.Carrier.Name != "Olva Courier" {
		t.Errorf("expected carrier 'Olva Courier', got '%s'", result.Carrier.Name)
	}
	if result.Origin.Street != "Av Javier Prado" {
		t.Errorf("expected origin street 'Av Javier Prado', got '%s'", result.Origin.Street)
	}
	if result.Destination.City != "Miraflores" {
		t.Errorf("expected destination city 'Miraflores', got '%s'", result.Destination.City)
	}
	if result.Package.Weight != 2.5 {
		t.Errorf("expected package weight 2.5, got %f", result.Package.Weight)
	}
}

func TestMapper_ToDomainShipment_Nil(t *testing.T) {
	m := shipmentpkg.NewMapper()

	if result := m.ToDomainShipment(nil); result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapper_MapShipmentStatus(t *testing.T) {
	m := shipmentpkg.NewMapper()

	tests := []struct {
		input    string
		expected domain.ShipmentStatus
	}{
		{"pending", domain.ShipmentStatusPending},
		{"handling", domain.ShipmentStatusReadyToShip},
		{"ready_to_ship", domain.ShipmentStatusReadyToShip},
		{"shipped", domain.ShipmentStatusShipped},
		{"in_transit", domain.ShipmentStatusInTransit},
		{"out_for_delivery", domain.ShipmentStatusOutForDelivery},
		{"delivered", domain.ShipmentStatusDelivered},
		{"cancelled", domain.ShipmentStatusCancelled},
		{"returned", domain.ShipmentStatusReturned},
		{"returning_to_sender", domain.ShipmentStatusReturned},
		{"not_delivered", domain.ShipmentStatusNotDelivered},
		{"unknown_status", domain.ShipmentStatusUnknown},
		{"", domain.ShipmentStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := m.MapShipmentStatus(tt.input); got != tt.expected {
				t.Errorf("MapShipmentStatus(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMapper_BuildShipmentSearchQuery(t *testing.T) {
	m := shipmentpkg.NewMapper()

	query := m.BuildShipmentSearchQuery(domain.ShipmentFilters{})
	if query != "" {
		t.Errorf("expected empty query for empty filters, got '%s'", query)
	}

	status := domain.ShipmentStatusInTransit
	query = m.BuildShipmentSearchQuery(domain.ShipmentFilters{
		OrderID:           "12345",
		ExternalReference: "ext-001",
		Status:            &status,
		Limit:             50,
		Offset:            10,
	})

	if query == "" {
		t.Error("expected non-empty query")
	}
	if query[0] != '?' {
		t.Errorf("expected query to start with '?', got '%c'", query[0])
	}
}

func TestMapper_ToDomainTrackingEvents(t *testing.T) {
	m := shipmentpkg.NewMapper()
	now := time.Now()

	entries := []shipmentpkg.MLShipmentHistoryEntry{
		{Status: "pending", Date: now, Description: "Envio creado"},
		{Status: "shipped", Date: now, Description: "En camino"},
		{Status: "delivered", Date: now, Description: "Entregado"},
	}

	events := m.ToDomainTrackingEvents(entries)

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Status != domain.ShipmentStatusPending {
		t.Errorf("expected first event to be Pending, got %s", events[0].Status.String())
	}
	if events[2].Status != domain.ShipmentStatusDelivered {
		t.Errorf("expected last event to be Delivered, got %s", events[2].Status.String())
	}
}
