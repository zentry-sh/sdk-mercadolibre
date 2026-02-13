package mocks

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type MockPaymentProvider struct {
	CreatePaymentFn func(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error)
	GetPaymentFn    func(ctx context.Context, id string) (*domain.Payment, error)
	ListPaymentsFn  func(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error)
	RefundPaymentFn func(ctx context.Context, req *domain.RefundRequest) (*domain.Refund, error)
	CancelPaymentFn func(ctx context.Context, paymentID string) error
	GetRefundFn     func(ctx context.Context, paymentID, refundID string) (*domain.Refund, error)
	ListRefundsFn   func(ctx context.Context, paymentID string) ([]*domain.Refund, error)
}

func (m *MockPaymentProvider) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	if m.CreatePaymentFn != nil {
		return m.CreatePaymentFn(ctx, req)
	}
	return nil, nil
}

func (m *MockPaymentProvider) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	if m.GetPaymentFn != nil {
		return m.GetPaymentFn(ctx, id)
	}
	return nil, nil
}

func (m *MockPaymentProvider) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	if m.ListPaymentsFn != nil {
		return m.ListPaymentsFn(ctx, filters)
	}
	return nil, nil
}

func (m *MockPaymentProvider) RefundPayment(ctx context.Context, req *domain.RefundRequest) (*domain.Refund, error) {
	if m.RefundPaymentFn != nil {
		return m.RefundPaymentFn(ctx, req)
	}
	return nil, nil
}

func (m *MockPaymentProvider) CancelPayment(ctx context.Context, paymentID string) error {
	if m.CancelPaymentFn != nil {
		return m.CancelPaymentFn(ctx, paymentID)
	}
	return nil
}

func (m *MockPaymentProvider) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	if m.GetRefundFn != nil {
		return m.GetRefundFn(ctx, paymentID, refundID)
	}
	return nil, nil
}

func (m *MockPaymentProvider) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
	if m.ListRefundsFn != nil {
		return m.ListRefundsFn(ctx, paymentID)
	}
	return nil, nil
}
