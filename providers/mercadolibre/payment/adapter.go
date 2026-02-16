package payment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http   *httputil.Client
	mapper *Mapper
	log    logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.Nop()
	}
	return &Adapter{
		http:   http,
		mapper: NewMapper(),
		log:    log,
	}
}

func (a *Adapter) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.Payment, error) {
	a.log.Debug("create_payment", "external_ref", req.ExternalReference)

	mlReq := a.mapper.ToMLCreatePaymentRequest(req)

	var mlResp MLPaymentResponse
	if err := a.http.Post(ctx, "/v1/payments", mlReq, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	return a.mapper.ToDomainPayment(&mlResp), nil
}

func (a *Adapter) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	a.log.Debug("get_payment", "id", id)

	var mlResp MLPaymentResponse
	path := fmt.Sprintf("/v1/payments/%s", url.PathEscape(id))
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	return a.mapper.ToDomainPayment(&mlResp), nil
}

func (a *Adapter) ListPayments(ctx context.Context, filters domain.PaymentFilters) ([]*domain.Payment, error) {
	a.log.Debug("list_payments")

	query := a.mapper.BuildSearchQuery(filters)

	var mlResp MLPaymentSearchResponse
	path := fmt.Sprintf("/v1/payments/search%s", query)
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	return a.mapper.ToDomainPayments(mlResp.Results), nil
}

func (a *Adapter) RefundPayment(ctx context.Context, req *domain.RefundRequest) (*domain.Refund, error) {
	a.log.Debug("refund_payment", "payment_id", req.PaymentID)

	mlReq := a.mapper.ToMLRefundRequest(req)

	var mlResp MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds", url.PathEscape(req.PaymentID))
	if err := a.http.Post(ctx, path, mlReq, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	payment, err := a.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return a.mapper.ToDomainRefund(&mlResp, ""), nil
	}

	return a.mapper.ToDomainRefund(&mlResp, payment.Amount.Currency), nil
}

func (a *Adapter) CancelPayment(ctx context.Context, paymentID string) error {
	a.log.Debug("cancel_payment", "payment_id", paymentID)

	body := map[string]string{"status": "cancelled"}
	path := fmt.Sprintf("/v1/payments/%s", url.PathEscape(paymentID))

	return a.mapError(a.http.Put(ctx, path, body, nil))
}

func (a *Adapter) GetRefund(ctx context.Context, paymentID, refundID string) (*domain.Refund, error) {
	a.log.Debug("get_refund", "payment_id", paymentID, "refund_id", refundID)

	var mlResp MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds/%s",
		url.PathEscape(paymentID), url.PathEscape(refundID))
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	currency := ""
	if payment, err := a.GetPayment(ctx, paymentID); err == nil {
		currency = payment.Amount.Currency
	}

	return a.mapper.ToDomainRefund(&mlResp, currency), nil
}

func (a *Adapter) ListRefunds(ctx context.Context, paymentID string) ([]*domain.Refund, error) {
	a.log.Debug("list_refunds", "payment_id", paymentID)

	var mlResp []MLRefundResponse
	path := fmt.Sprintf("/v1/payments/%s/refunds", url.PathEscape(paymentID))
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, a.mapError(err)
	}

	currency := ""
	if payment, err := a.GetPayment(ctx, paymentID); err == nil {
		currency = payment.Amount.Currency
	}

	refunds := make([]*domain.Refund, len(mlResp))
	for i := range mlResp {
		refunds[i] = a.mapper.ToDomainRefund(&mlResp[i], currency)
	}

	return refunds, nil
}

func (a *Adapter) mapError(err error) error {
	if err == nil {
		return nil
	}
	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		return err
	}
	switch sdkErr.ProviderCode {
	case "2001", "cc_rejected_insufficient_amount":
		return errors.InsufficientFunds()
	case "2002", "2003", "2004",
		"cc_rejected_bad_filled_security_code",
		"cc_rejected_bad_filled_card_number",
		"cc_rejected_bad_filled_date":
		return errors.InvalidCard(sdkErr.ProviderMessage)
	case "cc_rejected_high_risk":
		return errors.NewError(errors.ErrCodeFraudRejection, "payment rejected due to fraud risk")
	}
	return err
}
