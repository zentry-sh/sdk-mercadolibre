package webhook

import (
	"encoding/json"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Handler struct {
	tolerance time.Duration
	log       logger.Logger
}

type Option func(*Handler)

func WithTimestampTolerance(d time.Duration) Option {
	return func(h *Handler) {
		if d > 0 {
			h.tolerance = d
		}
	}
}

func NewHandler(log logger.Logger, opts ...Option) *Handler {
	h := &Handler{
		tolerance: defaultTimestampTolerance,
		log:       log,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) Validate(req domain.WebhookRequest, secret string) error {
	if secret == "" {
		return errors.NewError(errors.ErrCodeInvalidWebhook, "webhook secret not configured")
	}

	ts, expectedHash, err := parseSignatureHeader(req.Signature)
	if err != nil {
		return err
	}

	if err := validateTimestamp(ts, h.tolerance); err != nil {
		return err
	}

	manifest := buildManifest(req.DataID, req.RequestID, ts)
	computed := computeHMAC(manifest, secret)

	if !verifyHMAC(computed, expectedHash) {
		return errors.NewError(errors.ErrCodeInvalidWebhook, "webhook signature verification failed")
	}

	h.log.Debug("webhook signature validated", "data_id", req.DataID)
	return nil
}

func (h *Handler) Parse(payload []byte) (*domain.WebhookEvent, error) {
	if len(payload) == 0 {
		return nil, errors.InvalidRequest("empty webhook payload")
	}

	var ml mlWebhookPayload
	if err := json.Unmarshal(payload, &ml); err != nil {
		return nil, errors.NewErrorWithCause(errors.ErrCodeInvalidWebhook, "failed to parse webhook payload", err)
	}

	return &domain.WebhookEvent{
		ID:          ml.ID,
		Type:        domain.WebhookEventType(ml.Action),
		Action:      ml.Action,
		LiveMode:    ml.LiveMode,
		APIVersion:  ml.APIVersion,
		UserID:      ml.UserID,
		DateCreated: ml.DateCreated,
		DataID:      ml.Data.ID,
	}, nil
}
