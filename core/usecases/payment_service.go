package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/pkg/sanitize"
)

type PaymentService struct {
	provider ports.PaymentProvider
	log      logger.Logger
}

func NewPaymentService(provider ports.PaymentProvider, log logger.Logger) *PaymentService {
	if log == nil {
		log = logger.Nop()
	}
	return &PaymentService{
		provider: provider,
		log:      log,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	req.ExternalReference = sanitize.String(req.ExternalReference)
	req.Payer.Email = sanitize.Email(req.Payer.Email)
	req.Payer.FirstName = sanitize.String(req.Payer.FirstName)
	req.Payer.LastName = sanitize.String(req.Payer.LastName)
	req.Amount.Currency = sanitize.CurrencyCode(req.Amount.Currency)
	req.Description = sanitize.String(req.Description)

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	s.log.Debug("create_payment", "external_ref", req.ExternalReference, "amount", req.Amount.Amount, "currency", req.Amount.Currency)

	return s.provider.CreatePayment(ctx, req)
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	id = sanitize.ID(id)
	if id == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}
	return s.provider.GetPayment(ctx, id)
}

func (s *PaymentService) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	filters.ExternalReference = sanitize.String(filters.ExternalReference)
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	return s.provider.ListPayments(ctx, filters)
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount *domain.Money) (*domain.Refund, error) {
	paymentID = sanitize.ID(paymentID)
	if paymentID == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}

	s.log.Debug("refund_payment", "payment_id", paymentID)

	return s.provider.RefundPayment(ctx, &domain.RefundRequest{
		PaymentID: paymentID,
		Amount:    amount,
	})
}

func (s *PaymentService) CancelPayment(ctx context.Context, paymentID string) error {
	paymentID = sanitize.ID(paymentID)
	if paymentID == "" {
		return errors.InvalidRequest("payment id is required")
	}
	return s.provider.CancelPayment(ctx, paymentID)
}

func (s *PaymentService) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	paymentID = sanitize.ID(paymentID)
	refundID = sanitize.ID(refundID)
	if paymentID == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}
	if refundID == "" {
		return nil, errors.InvalidRequest("refund id is required")
	}
	return s.provider.GetRefund(ctx, paymentID, refundID)
}

func (s *PaymentService) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
	paymentID = sanitize.ID(paymentID)
	if paymentID == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}
	return s.provider.ListRefunds(ctx, paymentID)
}

func (s *PaymentService) validateCreateRequest(req *domain.CreatePaymentRequest) error {
	if req.ExternalReference == "" {
		return errors.InvalidRequest("external_reference is required")
	}
	if req.Amount.Amount <= 0 {
		return errors.InvalidRequest("amount must be positive")
	}
	if req.Amount.Currency == "" {
		return errors.InvalidRequest("currency is required")
	}
	if req.Payer.Email == "" {
		return errors.InvalidRequest("payer email is required")
	}
	return nil
}
