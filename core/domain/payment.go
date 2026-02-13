package domain

import "time"

type Payment struct {
	ID                string
	ExternalReference string
	Amount            Money
	NetAmount         Money
	Description       string
	Method            PaymentMethod
	MethodID          string
	Status            PaymentStatus
	StatusDetail      string
	Payer             Payer
	Installments      int
	Metadata          map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ApprovedAt        *time.Time
}

func (p *Payment) IsApproved() bool {
	return p.Status == PaymentStatusApproved
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusInProcess
}

func (p *Payment) CanRefund() bool {
	return p.Status == PaymentStatusApproved
}

type CreatePaymentRequest struct {
	ExternalReference string
	Amount            Money
	Description       string
	Method            PaymentMethod
	MethodID          string
	Payer             Payer
	Token             string
	Installments      int
	CallbackURL       string
	NotificationURL   string
	Metadata          map[string]interface{}
}

type PaymentFilters struct {
	ExternalReference string
	Status            *PaymentStatus
	MethodID          string
	FromDate          *time.Time
	ToDate            *time.Time
	Limit             int
	Offset            int
}

type RefundRequest struct {
	PaymentID string
	Amount    *Money
	Reason    string
}

type Refund struct {
	ID                string
	PaymentID         string
	Amount            Money
	Status            string
	Reason            string
	ExternalReference string
	CreatedAt         time.Time
}
