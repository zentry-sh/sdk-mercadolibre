package qr

import (
	"testing"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	qrpkg "github.com/zentry/sdk-mercadolibre/providers/mercadolibre/qr"
)

func TestMapper_ToDomainQR(t *testing.T) {
	m := qrpkg.NewMapper()
	now := time.Now()

	mlResp := &qrpkg.MLOrderResponse{
		ID:                "order-abc-123",
		ExternalReference: "ext-qr-001",
		Status:            "active",
		TotalAmount:       150.50,
		POSID:             "pos-001",
		StoreID:           "store-001",
		CollectorID:       99999,
		QRData:            "00020101021226410014br.gov.bcb.pix",
		CreatedDate:       &now,
		LastUpdatedDate:   &now,
	}

	result := m.ToDomainQR(mlResp)

	if result.ID != "order-abc-123" {
		t.Errorf("expected ID 'order-abc-123', got '%s'", result.ID)
	}
	if result.ExternalReference != "ext-qr-001" {
		t.Errorf("expected ext ref 'ext-qr-001', got '%s'", result.ExternalReference)
	}
	if result.Status != domain.QRStatusActive {
		t.Errorf("expected status Active, got %s", result.Status.String())
	}
	if result.Amount == nil || result.Amount.Amount != 150.50 {
		t.Errorf("expected amount 150.50, got %v", result.Amount)
	}
	if result.QRData != "00020101021226410014br.gov.bcb.pix" {
		t.Errorf("expected QR data, got '%s'", result.QRData)
	}
	if result.CollectorID != "99999" {
		t.Errorf("expected collector '99999', got '%s'", result.CollectorID)
	}
}

func TestMapper_ToDomainQR_Nil(t *testing.T) {
	m := qrpkg.NewMapper()

	if result := m.ToDomainQR(nil); result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestMapper_ToDomainQR_WithPayment(t *testing.T) {
	m := qrpkg.NewMapper()

	mlResp := &qrpkg.MLOrderResponse{
		ID:                "order-001",
		ExternalReference: "ext-001",
		Status:            "closed",
		TotalAmount:       100.0,
		Payments: []qrpkg.MLOrderPayment{
			{
				ID:                12345,
				TransactionAmount: 100.0,
				Status:            "approved",
			},
		},
	}

	result := m.ToDomainQR(mlResp)

	if result.Payment == nil {
		t.Fatal("expected payment to be present")
	}
	if result.Payment.ID != "12345" {
		t.Errorf("expected payment ID '12345', got '%s'", result.Payment.ID)
	}
	if result.Payment.Status != domain.PaymentStatusApproved {
		t.Errorf("expected payment status Approved, got %s", result.Payment.Status.String())
	}
}

func TestMapper_ToDomainPOS(t *testing.T) {
	m := qrpkg.NewMapper()
	now := time.Now()

	mlResp := &qrpkg.MLPOSResponse{
		ID:          123,
		Name:        "Caja 1",
		ExternalID:  "caja-001",
		StoreID:     "store-001",
		FixedAmount: false,
		Category:    621102,
		QR:          &qrpkg.MLPOSQR{Image: "https://example.com/qr.png"},
		DateCreated: &now,
	}

	result := m.ToDomainPOS(mlResp)

	if result.ID != "123" {
		t.Errorf("expected ID '123', got '%s'", result.ID)
	}
	if result.Name != "Caja 1" {
		t.Errorf("expected name 'Caja 1', got '%s'", result.Name)
	}
	if result.QRCode != "https://example.com/qr.png" {
		t.Errorf("expected QR image URL, got '%s'", result.QRCode)
	}
}

func TestMapper_ToDomainStore(t *testing.T) {
	m := qrpkg.NewMapper()
	now := time.Now()

	mlResp := &qrpkg.MLStoreResponse{
		ID:         "store-abc",
		Name:       "Tienda Centro",
		ExternalID: "tienda-001",
		Location: &qrpkg.MLStoreLocation{
			StreetName:   "Av Larco",
			StreetNumber: "345",
			CityName:     "Miraflores",
			StateName:    "Lima",
			Latitude:     -12.1191,
			Longitude:    -77.0300,
		},
		DateCreated: &now,
	}

	result := m.ToDomainStore(mlResp)

	if result.ID != "store-abc" {
		t.Errorf("expected ID 'store-abc', got '%s'", result.ID)
	}
	if result.Name != "Tienda Centro" {
		t.Errorf("expected name 'Tienda Centro', got '%s'", result.Name)
	}
	if result.Location.City != "Miraflores" {
		t.Errorf("expected city 'Miraflores', got '%s'", result.Location.City)
	}
}

func TestMapper_MapQRStatus(t *testing.T) {
	m := qrpkg.NewMapper()

	tests := []struct {
		input    string
		expected domain.QRStatus
	}{
		{"active", domain.QRStatusActive},
		{"opened", domain.QRStatusActive},
		{"pending", domain.QRStatusPending},
		{"approved", domain.QRStatusApproved},
		{"closed", domain.QRStatusApproved},
		{"rejected", domain.QRStatusRejected},
		{"expired", domain.QRStatusExpired},
		{"cancelled", domain.QRStatusCancelled},
		{"unknown_status", domain.QRStatusUnknown},
		{"", domain.QRStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := m.MapQRStatus(tt.input); got != tt.expected {
				t.Errorf("MapQRStatus(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMapper_ToMLCreateOrderRequest(t *testing.T) {
	m := qrpkg.NewMapper()

	req := &domain.CreateQRRequest{
		ExternalReference: "order-qr-001",
		Description:       "Pago de prueba",
		Amount:            &domain.Money{Amount: 100.50, Currency: "PEN"},
		ExpirationMinutes: 30,
		Items: []domain.QRItem{
			{
				Title:       "Producto 1",
				Quantity:    2,
				UnitPrice:   domain.Money{Amount: 25.25, Currency: "PEN"},
				TotalAmount: domain.Money{Amount: 50.50, Currency: "PEN"},
			},
			{
				Title:       "Producto 2",
				Quantity:    1,
				UnitPrice:   domain.Money{Amount: 50.00, Currency: "PEN"},
				TotalAmount: domain.Money{Amount: 50.00, Currency: "PEN"},
			},
		},
	}

	mlReq := m.ToMLCreateOrderRequest(req)

	if mlReq.ExternalReference != "order-qr-001" {
		t.Errorf("expected ext ref 'order-qr-001', got '%s'", mlReq.ExternalReference)
	}
	if mlReq.TotalAmount != 100.50 {
		t.Errorf("expected total 100.50, got %f", mlReq.TotalAmount)
	}
	if len(mlReq.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(mlReq.Items))
	}
	if mlReq.Items[0].Quantity != 2 {
		t.Errorf("expected first item quantity 2, got %d", mlReq.Items[0].Quantity)
	}
	if mlReq.ExpirationDate == nil {
		t.Error("expected expiration date to be set")
	}
}

func TestMapper_BuildExternalRefQuery(t *testing.T) {
	m := qrpkg.NewMapper()

	query := m.BuildExternalRefQuery("order-001")
	if query == "" {
		t.Error("expected non-empty query")
	}
	if query[0] != '?' {
		t.Errorf("expected query to start with '?', got '%c'", query[0])
	}
}
