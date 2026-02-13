package ports

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error)
	GetPayment(ctx context.Context, id string) (*domain.Payment, error)
	ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error)
	RefundPayment(ctx context.Context, req *domain.RefundRequest) (*domain.Refund, error)
	CancelPayment(ctx context.Context, paymentID string) error
	GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error)
	ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error)
}
