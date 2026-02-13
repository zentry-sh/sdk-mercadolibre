package payment

import (
	"context"
	"fmt"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http   *httputil.Client
	mapper *Mapper
	logger logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &Adapter{
		http:   http,
		mapper: NewMapper(),
		logger: log,
	}
}

func (a *Adapter) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	a.logger.Debug("creating payment", "external_ref", req.ExternalReference)

	mlReq := a.mapper.ToMLCreatePaymentRequest(req)

	var mlResp MLPaymentResponse
	err := a.http.Post(ctx, "/v1/payments", mlReq, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	payment := a.mapper.ToDomainPayment(&mlResp)
	a.logger.Info("payment created", "id", payment.ID, "status", payment.Status.String())
	
	return payment, nil
}

func (a *Adapter) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	a.logger.Debug("getting payment", "id", id)

	var mlResp MLPaymentResponse
	err := a.http.Get(ctx, "/v1/payments/"+id, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	return a.mapper.ToDomainPayment(&mlResp), nil
}

func (a *Adapter) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	a.logger.Debug("listing payments")

	query := a.mapper.BuildSearchQuery(filters)

	var mlResp MLPaymentSearchResponse
	err := a.http.Get(ctx, "/v1/payments/search"+query, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	return a.mapper.ToDomainPayments(mlResp.Results), nil
}

func (a *Adapter) RefundPayment(ctx context.Context, req *domain.RefundRequest) (*domain.Refund, error) {
	a.logger.Debug("refunding payment", "payment_id", req.PaymentID)

	mlReq := a.mapper.ToMLRefundRequest(req)
	
	var mlResp MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds", req.PaymentID)
	err := a.http.Post(ctx, path, mlReq, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	payment, err := a.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return a.mapper.ToDomainRefund(&mlResp, ""), nil
	}

	refund := a.mapper.ToDomainRefund(&mlResp, payment.Amount.Currency)
	a.logger.Info("payment refunded", "refund_id", refund.ID)
	
	return refund, nil
}

func (a *Adapter) CancelPayment(ctx context.Context, paymentID string) error {
	a.logger.Debug("cancelling payment", "payment_id", paymentID)

	req := map[string]string{
		"status": "cancelled",
	}

	path := fmt.Sprintf("/v1/payments/%s", paymentID)
	err := a.http.Put(ctx, path, req, nil)
	if err != nil {
		return a.mapError(err)
	}

	a.logger.Info("payment cancelled", "payment_id", paymentID)
	return nil
}

func (a *Adapter) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	a.logger.Debug("getting refund", "payment_id", paymentID, "refund_id", refundID)

	var mlResp MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds/%s", paymentID, refundID)
	err := a.http.Get(ctx, path, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	payment, _ := a.GetPayment(ctx, paymentID)
	currency := ""
	if payment != nil {
		currency = payment.Amount.Currency
	}

	return a.mapper.ToDomainRefund(&mlResp, currency), nil
}

func (a *Adapter) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
	a.logger.Debug("listing refunds", "payment_id", paymentID)

	var mlResp []MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds", paymentID)
	err := a.http.Get(ctx, path, &mlResp)
	if err != nil {
		return nil, a.mapError(err)
	}

	payment, _ := a.GetPayment(ctx, paymentID)
	currency := ""
	if payment != nil {
		currency = payment.Amount.Currency
	}

	refunds := make([]*domain.Refund, len(mlResp))
	for i, ml := range mlResp {
		refunds[i] = a.mapper.ToDomainRefund(&ml, currency)
	}

	return refunds, nil
}

func (a *Adapter) mapError(err error) error {
	if sdkErr, ok := err.(*errors.SDKError); ok {
		switch sdkErr.ProviderCode {
		case "2001":
			return errors.InsufficientFunds()
		case "2002", "2003", "2004":
			return errors.InvalidCard(sdkErr.ProviderMessage)
		case "cc_rejected_insufficient_amount":
			return errors.InsufficientFunds()
		case "cc_rejected_bad_filled_security_code", "cc_rejected_bad_filled_card_number", "cc_rejected_bad_filled_date":
			return errors.InvalidCard(sdkErr.ProviderMessage)
		case "cc_rejected_high_risk":
			return errors.NewError(errors.ErrCodeFraudRejection, "payment rejected due to fraud risk")
		}
	}
	return err
}
