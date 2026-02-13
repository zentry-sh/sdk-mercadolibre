package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type PaymentService struct {
	provider ports.PaymentProvider
	logger   logger.Logger
}

func NewPaymentService(provider ports.PaymentProvider, log logger.Logger) *PaymentService {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &PaymentService{
		provider: provider,
		logger:   log,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	s.logger.Info("creating payment",
		"external_ref", req.ExternalReference,
		"amount", req.Amount.Amount,
		"currency", req.Amount.Currency)

	payment, err := s.provider.CreatePayment(ctx, req)
	if err != nil {
		s.logger.Error("failed to create payment", "error", err)
		return nil, err
	}

	s.logger.Info("payment created",
		"id", payment.ID,
		"status", payment.Status.String())

	return payment, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	if id == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}

	s.logger.Debug("getting payment", "id", id)

	return s.provider.GetPayment(ctx, id)
}

func (s *PaymentService) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	s.logger.Debug("listing payments")

	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	return s.provider.ListPayments(ctx, filters)
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount *domain.Money) (*domain.Refund, error) {
	if paymentID == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}

	s.logger.Info("refunding payment",
		"payment_id", paymentID,
		"amount", amount)

	req := &domain.RefundRequest{
		PaymentID: paymentID,
		Amount:    amount,
	}

	refund, err := s.provider.RefundPayment(ctx, req)
	if err != nil {
		s.logger.Error("failed to refund payment", "error", err)
		return nil, err
	}

	s.logger.Info("payment refunded",
		"refund_id", refund.ID,
		"amount", refund.Amount.Amount)

	return refund, nil
}

func (s *PaymentService) CancelPayment(ctx context.Context, paymentID string) error {
	if paymentID == "" {
		return errors.InvalidRequest("payment id is required")
	}

	s.logger.Info("cancelling payment", "payment_id", paymentID)

	err := s.provider.CancelPayment(ctx, paymentID)
	if err != nil {
		s.logger.Error("failed to cancel payment", "error", err)
		return err
	}

	s.logger.Info("payment cancelled", "payment_id", paymentID)
	return nil
}

func (s *PaymentService) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	if paymentID == "" {
		return nil, errors.InvalidRequest("payment id is required")
	}
	if refundID == "" {
		return nil, errors.InvalidRequest("refund id is required")
	}

	return s.provider.GetRefund(ctx, paymentID, refundID)
}

func (s *PaymentService) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
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
