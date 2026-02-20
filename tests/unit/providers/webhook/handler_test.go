package webhook

import (
	"encoding/json"
	"testing"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	realWebhook "github.com/zentry/sdk-mercadolibre/providers/mercadolibre/webhook"
)

func newTestHandler() *realWebhook.Handler {
	return realWebhook.NewHandler(logger.Nop())
}

func validSignatureRequest(secret string) domain.WebhookRequest {
	// Build the request components
	dataID := "12345"
	requestID := "req-abc-123"

	// We will compute HMAC ourselves to generate a valid signature
	// Since the functions are internal, we use the handler to validate
	// Instead, we just test error paths here and test the handler via Process
	return domain.WebhookRequest{
		Body:      []byte(`{"id":99,"live_mode":false,"type":"payments","date_created":"2024-01-15T10:00:00Z","user_id":100,"api_version":"v1","action":"payment.created","data":{"id":"12345"}}`),
		Signature: "", // will be set per test
		RequestID: requestID,
		DataID:    dataID,
	}
}

func TestHandler_Validate_EmptySecret(t *testing.T) {
	h := newTestHandler()

	req := validSignatureRequest("")
	req.Signature = "ts=1234567890,v1=abc123"

	err := h.Validate(req, "")
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}

	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		t.Fatalf("expected *errors.SDKError, got %T", err)
	}
	if sdkErr.Code != errors.ErrCodeInvalidWebhook {
		t.Errorf("expected error code %s, got %s", errors.ErrCodeInvalidWebhook, sdkErr.Code)
	}
}

func TestHandler_Validate_MissingSignatureHeader(t *testing.T) {
	h := newTestHandler()

	req := domain.WebhookRequest{
		Body:      []byte(`{}`),
		Signature: "",
		RequestID: "req-001",
		DataID:    "123",
	}

	err := h.Validate(req, "my-secret")
	if err == nil {
		t.Fatal("expected error for missing signature header, got nil")
	}
}

func TestHandler_Validate_InvalidSignatureFormat(t *testing.T) {
	h := newTestHandler()

	tests := []struct {
		name      string
		signature string
	}{
		{"no equals", "garbage-data"},
		{"missing v1", "ts=1234567890"},
		{"missing ts", "v1=abcdef"},
		{"empty parts", ",,,"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := domain.WebhookRequest{
				Body:      []byte(`{}`),
				Signature: tt.signature,
				RequestID: "req-001",
				DataID:    "123",
			}

			err := h.Validate(req, "my-secret")
			if err == nil {
				t.Error("expected error for invalid signature format, got nil")
			}
		})
	}
}

func TestHandler_Validate_InvalidTimestamp(t *testing.T) {
	h := newTestHandler()

	req := domain.WebhookRequest{
		Body:      []byte(`{}`),
		Signature: "ts=not-a-number,v1=abc123",
		RequestID: "req-001",
		DataID:    "123",
	}

	err := h.Validate(req, "my-secret")
	if err == nil {
		t.Fatal("expected error for non-numeric timestamp, got nil")
	}
}

func TestHandler_Validate_ExpiredTimestamp(t *testing.T) {
	h := newTestHandler()

	// Timestamp from year 2000 — well outside the 5-minute tolerance
	req := domain.WebhookRequest{
		Body:      []byte(`{}`),
		Signature: "ts=946684800,v1=abc123",
		RequestID: "req-001",
		DataID:    "123",
	}

	err := h.Validate(req, "my-secret")
	if err == nil {
		t.Fatal("expected error for expired timestamp, got nil")
	}
}

func TestHandler_Validate_WrongSignature(t *testing.T) {
	h := realWebhook.NewHandler(logger.Nop(), realWebhook.WithTimestampTolerance(0))

	// Use a very large tolerance to skip timestamp check but provide wrong hash
	// Actually, tolerance 0 means any timestamp fails. Let's use a custom tolerance.
	h2 := realWebhook.NewHandler(logger.Nop(), realWebhook.WithTimestampTolerance(1<<63-1))

	req := domain.WebhookRequest{
		Body:      []byte(`{}`),
		Signature: "ts=1234567890,v1=definitely-wrong-hash",
		RequestID: "req-001",
		DataID:    "123",
	}

	err := h2.Validate(req, "my-secret")
	if err == nil {
		t.Fatal("expected error for wrong signature hash, got nil")
	}

	_ = h // suppress unused
}

func TestHandler_Parse_ValidPayload(t *testing.T) {
	h := newTestHandler()

	payload := map[string]any{
		"id":           99,
		"live_mode":    false,
		"type":         "payments",
		"date_created": "2024-01-15T10:00:00Z",
		"user_id":      100,
		"api_version":  "v1",
		"action":       "payment.created",
		"data":         map[string]any{"id": "12345"},
	}
	body, _ := json.Marshal(payload)

	event, err := h.Parse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.ID != 99 {
		t.Errorf("expected ID 99, got %d", event.ID)
	}
	if event.Action != "payment.created" {
		t.Errorf("expected action 'payment.created', got '%s'", event.Action)
	}
	if event.Type != domain.WebhookPaymentCreated {
		t.Errorf("expected type %s, got %s", domain.WebhookPaymentCreated, event.Type)
	}
	if event.LiveMode != false {
		t.Error("expected LiveMode false")
	}
	if event.UserID != 100 {
		t.Errorf("expected UserID 100, got %d", event.UserID)
	}
	if event.DataID != "12345" {
		t.Errorf("expected DataID '12345', got '%s'", event.DataID)
	}
	if event.APIVersion != "v1" {
		t.Errorf("expected APIVersion 'v1', got '%s'", event.APIVersion)
	}
	if event.DateCreated != "2024-01-15T10:00:00Z" {
		t.Errorf("expected DateCreated '2024-01-15T10:00:00Z', got '%s'", event.DateCreated)
	}
}

func TestHandler_Parse_EmptyPayload(t *testing.T) {
	h := newTestHandler()

	_, err := h.Parse(nil)
	if err == nil {
		t.Fatal("expected error for empty payload, got nil")
	}

	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		t.Fatalf("expected *errors.SDKError, got %T", err)
	}
	if sdkErr.Code != errors.ErrCodeInvalidRequest {
		t.Errorf("expected error code %s, got %s", errors.ErrCodeInvalidRequest, sdkErr.Code)
	}
}

func TestHandler_Parse_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	_, err := h.Parse([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		t.Fatalf("expected *errors.SDKError, got %T", err)
	}
	if sdkErr.Code != errors.ErrCodeInvalidWebhook {
		t.Errorf("expected error code %s, got %s", errors.ErrCodeInvalidWebhook, sdkErr.Code)
	}
}

func TestHandler_Parse_ShipmentEvent(t *testing.T) {
	h := newTestHandler()

	payload := map[string]any{
		"id":           200,
		"live_mode":    true,
		"type":         "shipments",
		"date_created": "2024-02-01T12:00:00Z",
		"user_id":      300,
		"api_version":  "v1",
		"action":       "shipment.updated",
		"data":         map[string]any{"id": "ship-456"},
	}
	body, _ := json.Marshal(payload)

	event, err := h.Parse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.Type != domain.WebhookShipmentUpdated {
		t.Errorf("expected type %s, got %s", domain.WebhookShipmentUpdated, event.Type)
	}
	if !event.IsShipmentEvent() {
		t.Error("expected IsShipmentEvent() to return true")
	}
	if event.LiveMode != true {
		t.Error("expected LiveMode true")
	}
}

func TestHandler_Parse_QREvent(t *testing.T) {
	h := newTestHandler()

	payload := map[string]any{
		"id":           500,
		"live_mode":    true,
		"type":         "qr",
		"date_created": "2024-03-01T08:00:00Z",
		"user_id":      600,
		"api_version":  "v1",
		"action":       "qr.paid",
		"data":         map[string]any{"id": "qr-789"},
	}
	body, _ := json.Marshal(payload)

	event, err := h.Parse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.Type != domain.WebhookQRPaid {
		t.Errorf("expected type %s, got %s", domain.WebhookQRPaid, event.Type)
	}
	if !event.IsQREvent() {
		t.Error("expected IsQREvent() to return true")
	}
}
