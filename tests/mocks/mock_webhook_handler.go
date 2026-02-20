package mocks

import (
	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type MockWebhookHandler struct {
	ValidateFn func(req domain.WebhookRequest, secret string) error
	ParseFn    func(payload []byte) (*domain.WebhookEvent, error)
}

func (m *MockWebhookHandler) Validate(req domain.WebhookRequest, secret string) error {
	if m.ValidateFn != nil {
		return m.ValidateFn(req, secret)
	}
	return nil
}

func (m *MockWebhookHandler) Parse(payload []byte) (*domain.WebhookEvent, error) {
	if m.ParseFn != nil {
		return m.ParseFn(payload)
	}
	return &domain.WebhookEvent{}, nil
}
