package core

import (
	"context"
	"testing"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/usecases"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/tests/mocks"
)

func TestWebhookService_Process(t *testing.T) {
	mockHandler := &mocks.MockWebhookHandler{
		ValidateFn: func(req domain.WebhookRequest, secret string) error {
			return nil
		},
		ParseFn: func(payload []byte) (*domain.WebhookEvent, error) {
			return &domain.WebhookEvent{
				ID:     12345,
				Type:   domain.WebhookPaymentCreated,
				Action: "payment.created",
				DataID: "9876",
			}, nil
		},
	}

	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	event, err := service.Process(context.Background(), domain.WebhookRequest{
		Body:      []byte(`{"action":"payment.created","data":{"id":"9876"}}`),
		Signature: "ts=1234567890,v1=abc123",
		RequestID: "req-001",
		DataID:    "9876",
	}, "test-secret")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.ID != 12345 {
		t.Errorf("expected ID 12345, got %d", event.ID)
	}

	if event.Type != domain.WebhookPaymentCreated {
		t.Errorf("expected type %s, got %s", domain.WebhookPaymentCreated, event.Type)
	}

	if event.DataID != "9876" {
		t.Errorf("expected DataID '9876', got '%s'", event.DataID)
	}
}

func TestWebhookService_Process_EmptyBody(t *testing.T) {
	mockHandler := &mocks.MockWebhookHandler{}
	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	_, err := service.Process(context.Background(), domain.WebhookRequest{
		Body: nil,
	}, "test-secret")

	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}

	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		t.Fatalf("expected *errors.SDKError, got %T", err)
	}
	if sdkErr.Code != errors.ErrCodeInvalidRequest {
		t.Errorf("expected error code %s, got %s", errors.ErrCodeInvalidRequest, sdkErr.Code)
	}
}

func TestWebhookService_Process_ValidationFails(t *testing.T) {
	mockHandler := &mocks.MockWebhookHandler{
		ValidateFn: func(req domain.WebhookRequest, secret string) error {
			return errors.NewError(errors.ErrCodeInvalidWebhook, "signature mismatch")
		},
	}

	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	_, err := service.Process(context.Background(), domain.WebhookRequest{
		Body:      []byte(`{"action":"payment.created"}`),
		Signature: "ts=123,v1=bad",
		RequestID: "req-001",
		DataID:    "9876",
	}, "test-secret")

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		t.Fatalf("expected *errors.SDKError, got %T", err)
	}
	if sdkErr.Code != errors.ErrCodeInvalidWebhook {
		t.Errorf("expected error code %s, got %s", errors.ErrCodeInvalidWebhook, sdkErr.Code)
	}
}

func TestWebhookService_Process_ParseFails(t *testing.T) {
	mockHandler := &mocks.MockWebhookHandler{
		ValidateFn: func(req domain.WebhookRequest, secret string) error {
			return nil
		},
		ParseFn: func(payload []byte) (*domain.WebhookEvent, error) {
			return nil, errors.NewError(errors.ErrCodeInvalidWebhook, "bad json")
		},
	}

	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	_, err := service.Process(context.Background(), domain.WebhookRequest{
		Body:      []byte(`not-json`),
		Signature: "ts=123,v1=abc",
		RequestID: "req-001",
		DataID:    "9876",
	}, "test-secret")

	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestWebhookService_ValidateSignature(t *testing.T) {
	called := false
	mockHandler := &mocks.MockWebhookHandler{
		ValidateFn: func(req domain.WebhookRequest, secret string) error {
			called = true
			if secret != "my-secret" {
				t.Errorf("expected secret 'my-secret', got '%s'", secret)
			}
			return nil
		},
	}

	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	err := service.ValidateSignature(context.Background(), domain.WebhookRequest{
		Body:      []byte(`{}`),
		Signature: "ts=123,v1=abc",
		RequestID: "req-001",
		DataID:    "9876",
	}, "my-secret")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected ValidateFn to be called")
	}
}

func TestWebhookService_Parse(t *testing.T) {
	mockHandler := &mocks.MockWebhookHandler{
		ParseFn: func(payload []byte) (*domain.WebhookEvent, error) {
			return &domain.WebhookEvent{
				Type:   domain.WebhookShipmentUpdated,
				DataID: "ship-99",
			}, nil
		},
	}

	service := usecases.NewWebhookService(mockHandler, logger.Nop())

	event, err := service.Parse(context.Background(), []byte(`{"action":"shipment.updated"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != domain.WebhookShipmentUpdated {
		t.Errorf("expected type %s, got %s", domain.WebhookShipmentUpdated, event.Type)
	}
}

func TestWebhookEventType_String(t *testing.T) {
	tests := []struct {
		eventType domain.WebhookEventType
		expected  string
	}{
		{domain.WebhookPaymentCreated, "payment.created"},
		{domain.WebhookPaymentUpdated, "payment.updated"},
		{domain.WebhookRefundCreated, "refund.created"},
		{domain.WebhookChargebackCreated, "chargeback.created"},
		{domain.WebhookShipmentCreated, "shipment.created"},
		{domain.WebhookShipmentUpdated, "shipment.updated"},
		{domain.WebhookQRScanned, "qr.scanned"},
		{domain.WebhookQRPaid, "qr.paid"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("WebhookEventType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWebhookEvent_TypeChecks(t *testing.T) {
	tests := []struct {
		name       string
		event      domain.WebhookEvent
		isPayment  bool
		isShipment bool
		isQR       bool
		isRefund   bool
		isChgback  bool
	}{
		{
			name:      "payment created",
			event:     domain.WebhookEvent{Type: domain.WebhookPaymentCreated},
			isPayment: true,
		},
		{
			name:      "payment updated",
			event:     domain.WebhookEvent{Type: domain.WebhookPaymentUpdated},
			isPayment: true,
		},
		{
			name:       "shipment created",
			event:      domain.WebhookEvent{Type: domain.WebhookShipmentCreated},
			isShipment: true,
		},
		{
			name:  "qr scanned",
			event: domain.WebhookEvent{Type: domain.WebhookQRScanned},
			isQR:  true,
		},
		{
			name:     "refund created",
			event:    domain.WebhookEvent{Type: domain.WebhookRefundCreated},
			isRefund: true,
		},
		{
			name:      "chargeback created",
			event:     domain.WebhookEvent{Type: domain.WebhookChargebackCreated},
			isChgback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.IsPaymentEvent(); got != tt.isPayment {
				t.Errorf("IsPaymentEvent() = %v, want %v", got, tt.isPayment)
			}
			if got := tt.event.IsShipmentEvent(); got != tt.isShipment {
				t.Errorf("IsShipmentEvent() = %v, want %v", got, tt.isShipment)
			}
			if got := tt.event.IsQREvent(); got != tt.isQR {
				t.Errorf("IsQREvent() = %v, want %v", got, tt.isQR)
			}
			if got := tt.event.IsRefundEvent(); got != tt.isRefund {
				t.Errorf("IsRefundEvent() = %v, want %v", got, tt.isRefund)
			}
			if got := tt.event.IsChargebackEvent(); got != tt.isChgback {
				t.Errorf("IsChargebackEvent() = %v, want %v", got, tt.isChgback)
			}
		})
	}
}
