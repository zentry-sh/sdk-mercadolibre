package core

import (
	"context"
	"testing"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/tests/mocks"
)

func TestQRService_CreateQR(t *testing.T) {
	mockProvider := &mocks.MockQRProvider{
		CreateQRFn: func(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
			return &domain.QRCode{
				ID:                "qr-001",
				ExternalReference: req.ExternalReference,
				Type:              req.Type,
				Status:            domain.QRStatusActive,
				QRData:            "00020101021226410014br.gov.bcb.pix",
			}, nil
		},
	}

	service := usecases.NewQRService(mockProvider, nil)

	qr, err := service.CreateQR(context.Background(), &domain.CreateQRRequest{
		ExternalReference: "order-qr-001",
		Type:              domain.QRTypeDynamic,
		Amount:            &domain.Money{Amount: 50.00, Currency: "PEN"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if qr.ID != "qr-001" {
		t.Errorf("expected ID 'qr-001', got '%s'", qr.ID)
	}
	if qr.Status != domain.QRStatusActive {
		t.Errorf("expected status Active, got %s", qr.Status.String())
	}
	if qr.QRData == "" {
		t.Error("expected QR data to be non-empty")
	}
}

func TestQRService_CreateQR_Validation(t *testing.T) {
	mockProvider := &mocks.MockQRProvider{
		CreateQRFn: func(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
			return &domain.QRCode{ID: "qr-001"}, nil
		},
	}
	service := usecases.NewQRService(mockProvider, nil)

	tests := []struct {
		name    string
		req     *domain.CreateQRRequest
		wantErr bool
	}{
		{
			name: "missing external reference",
			req: &domain.CreateQRRequest{
				Type:   domain.QRTypeDynamic,
				Amount: &domain.Money{Amount: 50, Currency: "PEN"},
			},
			wantErr: true,
		},
		{
			name: "invalid QR type",
			req: &domain.CreateQRRequest{
				ExternalReference: "order-001",
				Type:              domain.QRType("invalid"),
				Amount:            &domain.Money{Amount: 50, Currency: "PEN"},
			},
			wantErr: true,
		},
		{
			name: "dynamic QR without amount",
			req: &domain.CreateQRRequest{
				ExternalReference: "order-001",
				Type:              domain.QRTypeDynamic,
			},
			wantErr: true,
		},
		{
			name: "dynamic QR with zero amount",
			req: &domain.CreateQRRequest{
				ExternalReference: "order-001",
				Type:              domain.QRTypeDynamic,
				Amount:            &domain.Money{Amount: 0, Currency: "PEN"},
			},
			wantErr: true,
		},
		{
			name: "valid dynamic QR",
			req: &domain.CreateQRRequest{
				ExternalReference: "order-001",
				Type:              domain.QRTypeDynamic,
				Amount:            &domain.Money{Amount: 100, Currency: "PEN"},
			},
			wantErr: false,
		},
		{
			name: "valid static QR without amount",
			req: &domain.CreateQRRequest{
				ExternalReference: "pos-001",
				Type:              domain.QRTypeStatic,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateQR(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateQR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQRService_RegisterPOS_Validation(t *testing.T) {
	mockProvider := &mocks.MockQRProvider{
		RegisterPOSFn: func(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
			return &domain.POSInfo{ID: "pos-001", Name: req.Name}, nil
		},
	}
	service := usecases.NewQRService(mockProvider, nil)

	tests := []struct {
		name    string
		req     *domain.RegisterPOSRequest
		wantErr bool
	}{
		{
			name:    "missing name",
			req:     &domain.RegisterPOSRequest{ExternalID: "ext-001"},
			wantErr: true,
		},
		{
			name:    "missing external_id",
			req:     &domain.RegisterPOSRequest{Name: "Caja 1"},
			wantErr: true,
		},
		{
			name:    "valid request",
			req:     &domain.RegisterPOSRequest{Name: "Caja 1", ExternalID: "ext-001"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.RegisterPOS(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterPOS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQRService_Sanitization(t *testing.T) {
	var capturedID string
	mockProvider := &mocks.MockQRProvider{
		GetQRFn: func(ctx context.Context, qrID string) (*domain.QRCode, error) {
			capturedID = qrID
			return &domain.QRCode{ID: qrID}, nil
		},
	}
	service := usecases.NewQRService(mockProvider, nil)

	_, _ = service.GetQR(context.Background(), "  qr-001  ")
	if capturedID != "qr-001" {
		t.Errorf("expected sanitized ID 'qr-001', got '%s'", capturedID)
	}
}

func TestQRService_GetQR_EmptyID(t *testing.T) {
	service := usecases.NewQRService(&mocks.MockQRProvider{}, nil)

	_, err := service.GetQR(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty QR ID")
	}
}

func TestQRService_RegisterStore_EmptyName(t *testing.T) {
	service := usecases.NewQRService(&mocks.MockQRProvider{}, nil)

	_, err := service.RegisterStore(context.Background(), &domain.RegisterStoreRequest{})
	if err == nil {
		t.Error("expected error for empty store name")
	}
}

func TestQRType_IsValid(t *testing.T) {
	tests := []struct {
		qrType domain.QRType
		valid  bool
	}{
		{domain.QRTypeDynamic, true},
		{domain.QRTypeStatic, true},
		{domain.QRType("invalid"), false},
		{domain.QRType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.qrType), func(t *testing.T) {
			if got := tt.qrType.IsValid(); got != tt.valid {
				t.Errorf("QRType(%q).IsValid() = %v, want %v", tt.qrType, got, tt.valid)
			}
		})
	}
}

func TestQRStatus_String(t *testing.T) {
	tests := []struct {
		status   domain.QRStatus
		expected string
	}{
		{domain.QRStatusActive, "active"},
		{domain.QRStatusPending, "pending"},
		{domain.QRStatusApproved, "approved"},
		{domain.QRStatusRejected, "rejected"},
		{domain.QRStatusExpired, "expired"},
		{domain.QRStatusCancelled, "cancelled"},
		{domain.QRStatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("QRStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestQRStatus_IsFinal(t *testing.T) {
	finalStatuses := []domain.QRStatus{
		domain.QRStatusApproved,
		domain.QRStatusRejected,
		domain.QRStatusExpired,
		domain.QRStatusCancelled,
	}
	nonFinalStatuses := []domain.QRStatus{
		domain.QRStatusActive,
		domain.QRStatusPending,
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
