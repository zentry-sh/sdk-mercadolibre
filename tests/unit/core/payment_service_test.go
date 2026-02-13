package core

import (
	"context"
	"testing"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/tests/mocks"
)

func TestPaymentService_CreatePayment(t *testing.T) {
	mockProvider := &mocks.MockPaymentProvider{
		CreatePaymentFn: func(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
			return &domain.Payment{
				ID:                "123456",
				ExternalReference: req.ExternalReference,
				Amount:            req.Amount,
				Status:            domain.PaymentStatusPending,
				CreatedAt:         time.Now(),
			}, nil
		},
	}

	service := usecases.NewPaymentService(mockProvider, nil)

	payment, err := service.CreatePayment(context.Background(), &domain.CreatePaymentRequest{
		ExternalReference: "order-test-001",
		Amount: domain.Money{
			Amount:   100.00,
			Currency: "PEN",
		},
		Payer: domain.Payer{
			Email: "test@example.com",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.ID != "123456" {
		t.Errorf("expected ID '123456', got '%s'", payment.ID)
	}

	if payment.ExternalReference != "order-test-001" {
		t.Errorf("expected ExternalReference 'order-test-001', got '%s'", payment.ExternalReference)
	}

	if payment.Status != domain.PaymentStatusPending {
		t.Errorf("expected status Pending, got %s", payment.Status.String())
	}
}

func TestPaymentService_CreatePayment_Validation(t *testing.T) {
	mockProvider := &mocks.MockPaymentProvider{
		CreatePaymentFn: func(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
			return &domain.Payment{
				ID:                "123456",
				ExternalReference: req.ExternalReference,
				Amount:            req.Amount,
				Status:            domain.PaymentStatusPending,
				CreatedAt:         time.Now(),
			}, nil
		},
	}
	service := usecases.NewPaymentService(mockProvider, nil)

	tests := []struct {
		name    string
		req     *domain.CreatePaymentRequest
		wantErr bool
	}{
		{
			name: "missing external reference",
			req: &domain.CreatePaymentRequest{
				Amount: domain.Money{Amount: 100, Currency: "PEN"},
				Payer:  domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name: "invalid amount",
			req: &domain.CreatePaymentRequest{
				ExternalReference: "order-001",
				Amount:            domain.Money{Amount: 0, Currency: "PEN"},
				Payer:             domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name: "missing currency",
			req: &domain.CreatePaymentRequest{
				ExternalReference: "order-001",
				Amount:            domain.Money{Amount: 100, Currency: ""},
				Payer:             domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name: "missing payer email",
			req: &domain.CreatePaymentRequest{
				ExternalReference: "order-001",
				Amount:            domain.Money{Amount: 100, Currency: "PEN"},
				Payer:             domain.Payer{},
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &domain.CreatePaymentRequest{
				ExternalReference: "order-001",
				Amount:            domain.Money{Amount: 100, Currency: "PEN"},
				Payer:             domain.Payer{Email: "test@example.com"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreatePayment(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentStatus_String(t *testing.T) {
	tests := []struct {
		status   domain.PaymentStatus
		expected string
	}{
		{domain.PaymentStatusPending, "pending"},
		{domain.PaymentStatusApproved, "approved"},
		{domain.PaymentStatusRejected, "rejected"},
		{domain.PaymentStatusCancelled, "cancelled"},
		{domain.PaymentStatusRefunded, "refunded"},
		{domain.PaymentStatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("PaymentStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPaymentMethod_IsValid(t *testing.T) {
	tests := []struct {
		method domain.PaymentMethod
		valid  bool
	}{
		{domain.PaymentMethodCard, true},
		{domain.PaymentMethodTransfer, true},
		{domain.PaymentMethodCash, true},
		{domain.PaymentMethodQR, true},
		{domain.PaymentMethodWallet, true},
		{domain.PaymentMethod("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := tt.method.IsValid(); got != tt.valid {
				t.Errorf("PaymentMethod.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
