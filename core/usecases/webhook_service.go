package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/pkg/sanitize"
)

type WebhookService struct {
	handler ports.WebhookHandler
	log     logger.Logger
}

func NewWebhookService(handler ports.WebhookHandler, log logger.Logger) *WebhookService {
	return &WebhookService{
		handler: handler,
		log:     log,
	}
}

func (s *WebhookService) Process(ctx context.Context, req domain.WebhookRequest, secret string) (*domain.WebhookEvent, error) {
	if len(req.Body) == 0 {
		return nil, errors.InvalidRequest("webhook body is empty")
	}

	sanitized := domain.WebhookRequest{
		Body:      req.Body,
		Signature: sanitize.String(req.Signature),
		RequestID: sanitize.String(req.RequestID),
		DataID:    sanitize.String(req.DataID),
	}

	if err := s.handler.Validate(sanitized, secret); err != nil {
		return nil, err
	}

	event, err := s.handler.Parse(sanitized.Body)
	if err != nil {
		return nil, err
	}

	s.log.Debug("webhook event processed", "type", string(event.Type), "data_id", event.DataID)
	return event, nil
}

func (s *WebhookService) ValidateSignature(ctx context.Context, req domain.WebhookRequest, secret string) error {
	sanitized := domain.WebhookRequest{
		Body:      req.Body,
		Signature: sanitize.String(req.Signature),
		RequestID: sanitize.String(req.RequestID),
		DataID:    sanitize.String(req.DataID),
	}
	return s.handler.Validate(sanitized, secret)
}

func (s *WebhookService) Parse(ctx context.Context, payload []byte) (*domain.WebhookEvent, error) {
	return s.handler.Parse(payload)
}
