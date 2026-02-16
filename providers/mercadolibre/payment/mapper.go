package payment

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type Mapper struct{}

func NewMapper() *Mapper {
	return &Mapper{}
}

func (m *Mapper) ToMLCreatePaymentRequest(req *domain.CreatePaymentRequest) *MLCreatePaymentRequest {
	mlReq := &MLCreatePaymentRequest{
		TransactionAmount: req.Amount.Amount,
		Description:       req.Description,
		ExternalReference: req.ExternalReference,
		Token:             req.Token,
		Installments:      req.Installments,
		NotificationURL:   req.NotificationURL,
		CallbackURL:       req.CallbackURL,
		Metadata:          req.Metadata,
	}

	if req.MethodID != "" {
		mlReq.PaymentMethodID = req.MethodID
	}

	if req.Payer.Email != "" || req.Payer.FirstName != "" {
		mlReq.Payer = m.toMLPayer(&req.Payer)
	}

	return mlReq
}

func (m *Mapper) toMLPayer(payer *domain.Payer) *MLPayer {
	mlPayer := &MLPayer{
		ID:        payer.ID,
		Email:     payer.Email,
		FirstName: payer.FirstName,
		LastName:  payer.LastName,
	}

	if payer.Phone != "" {
		mlPayer.Phone = &MLPhone{
			Number: payer.Phone,
		}
	}

	if !payer.Identification.IsEmpty() {
		mlPayer.Identification = &MLIdentification{
			Type:   payer.Identification.Type,
			Number: payer.Identification.Number,
		}
	}

	if payer.Address != nil && !payer.Address.IsEmpty() {
		mlPayer.Address = &MLAddress{
			StreetName:   payer.Address.Street,
			StreetNumber: payer.Address.Number,
			ZipCode:      payer.Address.ZipCode,
			City:         payer.Address.City,
			State:        payer.Address.State,
		}
	}

	return mlPayer
}

func (m *Mapper) ToDomainPayment(ml *MLPaymentResponse) *domain.Payment {
	payment := &domain.Payment{
		ID:                fmt.Sprintf("%d", ml.ID),
		ExternalReference: ml.ExternalReference,
		Amount: domain.Money{
			Amount:   ml.TransactionAmount,
			Currency: ml.CurrencyID,
		},
		NetAmount: domain.Money{
			Amount:   ml.NetReceivedAmount,
			Currency: ml.CurrencyID,
		},
		Description:  ml.Description,
		Method:       m.mapPaymentTypeToMethod(ml.PaymentTypeID),
		MethodID:     ml.PaymentMethodID,
		Status:       m.mapStatus(ml.Status),
		StatusDetail: ml.StatusDetail,
		Installments: ml.Installments,
		Metadata:     ml.Metadata,
		CreatedAt:    ml.DateCreated,
		UpdatedAt:    ml.DateLastUpdated,
		ApprovedAt:   ml.DateApproved,
	}

	if ml.Payer != nil {
		payment.Payer = m.toDomainPayer(ml.Payer)
	}

	return payment
}

func (m *Mapper) toDomainPayer(ml *MLPayer) domain.Payer {
	payer := domain.Payer{
		ID:        ml.ID,
		Email:     ml.Email,
		FirstName: ml.FirstName,
		LastName:  ml.LastName,
	}

	if ml.Phone != nil {
		payer.Phone = ml.Phone.Number
	}

	if ml.Identification != nil {
		payer.Identification = domain.Identification{
			Type:   ml.Identification.Type,
			Number: ml.Identification.Number,
		}
	}

	if ml.Address != nil {
		payer.Address = &domain.Address{
			Street:  ml.Address.StreetName,
			Number:  ml.Address.StreetNumber,
			ZipCode: ml.Address.ZipCode,
			City:    ml.Address.City,
			State:   ml.Address.State,
		}
	}

	return payer
}

func (m *Mapper) mapStatus(status string) domain.PaymentStatus {
	switch status {
	case "pending":
		return domain.PaymentStatusPending
	case "approved":
		return domain.PaymentStatusApproved
	case "authorized":
		return domain.PaymentStatusApproved
	case "rejected":
		return domain.PaymentStatusRejected
	case "cancelled":
		return domain.PaymentStatusCancelled
	case "in_process":
		return domain.PaymentStatusInProcess
	case "refunded":
		return domain.PaymentStatusRefunded
	case "charged_back":
		return domain.PaymentStatusChargedBack
	case "in_mediation":
		return domain.PaymentStatusInMediation
	default:
		return domain.PaymentStatusUnknown
	}
}

func (m *Mapper) mapPaymentTypeToMethod(paymentType string) domain.PaymentMethod {
	switch paymentType {
	case "credit_card", "debit_card", "prepaid_card":
		return domain.PaymentMethodCard
	case "bank_transfer":
		return domain.PaymentMethodTransfer
	case "ticket", "atm":
		return domain.PaymentMethodCash
	case "digital_wallet", "account_money":
		return domain.PaymentMethodWallet
	default:
		return domain.PaymentMethodCard
	}
}

func (m *Mapper) ToMLRefundRequest(req *domain.RefundRequest) *MLRefundRequest {
	mlReq := &MLRefundRequest{}
	if req.Amount != nil {
		mlReq.Amount = req.Amount.Amount
	}
	return mlReq
}

func (m *Mapper) ToDomainRefund(ml *MLRefundResponse, currency string) *domain.Refund {
	return &domain.Refund{
		ID:        fmt.Sprintf("%d", ml.ID),
		PaymentID: fmt.Sprintf("%d", ml.PaymentID),
		Amount: domain.Money{
			Amount:   ml.Amount,
			Currency: currency,
		},
		Status:            ml.Status,
		Reason:            ml.Reason,
		ExternalReference: ml.UniqueSequenceNumber,
		CreatedAt:         ml.DateCreated,
	}
}

func (m *Mapper) ToDomainPayments(mlPayments []MLPaymentResponse) []*domain.Payment {
	payments := make([]*domain.Payment, len(mlPayments))
	for i, ml := range mlPayments {
		payments[i] = m.ToDomainPayment(&ml)
	}
	return payments
}

func (m *Mapper) BuildSearchQuery(filters domain.PaymentFilters) string {
	params := url.Values{}

	if filters.ExternalReference != "" {
		params.Set("external_reference", filters.ExternalReference)
	}
	if filters.Status != nil {
		params.Set("status", filters.Status.String())
	}
	if filters.FromDate != nil {
		params.Set("begin_date", filters.FromDate.Format(time.RFC3339))
	}
	if filters.ToDate != nil {
		params.Set("end_date", filters.ToDate.Format(time.RFC3339))
	}
	if filters.Limit > 0 {
		params.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Offset > 0 {
		params.Set("offset", strconv.Itoa(filters.Offset))
	}

	if len(params) == 0 {
		return ""
	}
	return "?" + params.Encode()
}
