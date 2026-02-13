package domain

import "time"

type QRCode struct {
	ID                string
	POSID             string
	StoreID           string
	CollectorID       string
	Type              QRType
	Status            QRStatus
	ExternalReference string
	Amount            *Money
	Description       string
	QRData            string
	ImageURL          string
	ExpiresAt         *time.Time
	Payment           *Payment
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (q *QRCode) IsExpired() bool {
	if q.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*q.ExpiresAt)
}

func (q *QRCode) IsPaid() bool {
	return q.Status == QRStatusApproved && q.Payment != nil
}

func (q *QRCode) CanCancel() bool {
	return q.Status == QRStatusActive || q.Status == QRStatusPending
}

type CreateQRRequest struct {
	POSID             string
	CollectorID       string
	StoreID           string
	Type              QRType
	ExternalReference string
	Amount            *Money
	Description       string
	ExpirationMinutes int
	Items             []QRItem
	NotificationURL   string
}

type QRItem struct {
	Title       string
	Description string
	Quantity    int
	UnitPrice   Money
	TotalAmount Money
}

type POSInfo struct {
	ID              string
	Name            string
	StoreID         string
	ExternalID      string
	FixedAmount     bool
	QRCode          string
	Category        int
	URL             string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RegisterPOSRequest struct {
	Name        string
	StoreID     string
	ExternalID  string
	FixedAmount bool
	Category    int
}

type StoreInfo struct {
	ID           string
	Name         string
	ExternalID   string
	BusinessHours map[string]string
	Location     Address
	CreatedAt    time.Time
}

type RegisterStoreRequest struct {
	Name          string
	ExternalID    string
	BusinessHours map[string]string
	Location      Address
}
