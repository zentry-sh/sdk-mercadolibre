package ports

import "github.com/zentry/sdk-mercadolibre/core/domain"

type WebhookHandler interface {
	Validate(req domain.WebhookRequest, secret string) error
	Parse(payload []byte) (*domain.WebhookEvent, error)
}
