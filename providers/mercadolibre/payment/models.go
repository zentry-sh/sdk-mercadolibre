package payment

import "time"

type MLCreatePaymentRequest struct {
	TransactionAmount float64                `json:"transaction_amount"`
	Description       string                 `json:"description,omitempty"`
	PaymentMethodID   string                 `json:"payment_method_id,omitempty"`
	ExternalReference string                 `json:"external_reference,omitempty"`
	Payer             *MLPayer               `json:"payer,omitempty"`
	Token             string                 `json:"token,omitempty"`
	Installments      int                    `json:"installments,omitempty"`
	NotificationURL   string                 `json:"notification_url,omitempty"`
	CallbackURL       string                 `json:"callback_url,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type MLPayer struct {
	ID             string            `json:"id,omitempty"`
	Email          string            `json:"email,omitempty"`
	FirstName      string            `json:"first_name,omitempty"`
	LastName       string            `json:"last_name,omitempty"`
	Phone          *MLPhone          `json:"phone,omitempty"`
	Identification *MLIdentification `json:"identification,omitempty"`
	Address        *MLAddress        `json:"address,omitempty"`
}

type MLPhone struct {
	AreaCode string `json:"area_code,omitempty"`
	Number   string `json:"number,omitempty"`
}

type MLIdentification struct {
	Type   string `json:"type,omitempty"`
	Number string `json:"number,omitempty"`
}

type MLAddress struct {
	StreetName   string `json:"street_name,omitempty"`
	StreetNumber string `json:"street_number,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
}

type MLPaymentResponse struct {
	ID                  int64                  `json:"id"`
	Status              string                 `json:"status"`
	StatusDetail        string                 `json:"status_detail"`
	ExternalReference   string                 `json:"external_reference"`
	TransactionAmount   float64                `json:"transaction_amount"`
	NetReceivedAmount   float64                `json:"net_received_amount"`
	CurrencyID          string                 `json:"currency_id"`
	Description         string                 `json:"description"`
	PaymentMethodID     string                 `json:"payment_method_id"`
	PaymentTypeID       string                 `json:"payment_type_id"`
	Installments        int                    `json:"installments"`
	Payer               *MLPayer               `json:"payer"`
	Metadata            map[string]interface{} `json:"metadata"`
	DateCreated         time.Time              `json:"date_created"`
	DateApproved        *time.Time             `json:"date_approved"`
	DateLastUpdated     time.Time              `json:"date_last_updated"`
}

type MLPaymentSearchResponse struct {
	Paging  MLPaging            `json:"paging"`
	Results []MLPaymentResponse `json:"results"`
}

type MLPaging struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type MLRefundRequest struct {
	Amount float64 `json:"amount,omitempty"`
}

type MLRefundResponse struct {
	ID                int64     `json:"id"`
	PaymentID         int64     `json:"payment_id"`
	Amount            float64   `json:"amount"`
	Status            string    `json:"status"`
	Source            string    `json:"source"`
	DateCreated       time.Time `json:"date_created"`
	Reason            string    `json:"reason"`
	UniqueSequenceNumber string `json:"unique_sequence_number"`
}

type MLErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Cause   []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"cause"`
}
