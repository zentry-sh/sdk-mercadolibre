package domain

type PaymentMethod string

const (
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodQR       PaymentMethod = "qr"
	PaymentMethodWallet   PaymentMethod = "wallet"
)

func (m PaymentMethod) String() string {
	return string(m)
}

func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCard, PaymentMethodTransfer, PaymentMethodCash, PaymentMethodQR, PaymentMethodWallet:
		return true
	}
	return false
}

type PaymentStatus int

const (
	PaymentStatusUnknown PaymentStatus = iota
	PaymentStatusPending
	PaymentStatusApproved
	PaymentStatusRejected
	PaymentStatusCancelled
	PaymentStatusInProcess
	PaymentStatusRefunded
	PaymentStatusChargedBack
	PaymentStatusInMediation
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusApproved:
		return "approved"
	case PaymentStatusRejected:
		return "rejected"
	case PaymentStatusCancelled:
		return "cancelled"
	case PaymentStatusInProcess:
		return "in_process"
	case PaymentStatusRefunded:
		return "refunded"
	case PaymentStatusChargedBack:
		return "charged_back"
	case PaymentStatusInMediation:
		return "in_mediation"
	default:
		return "unknown"
	}
}

func (s PaymentStatus) IsFinal() bool {
	switch s {
	case PaymentStatusApproved, PaymentStatusRejected, PaymentStatusCancelled, PaymentStatusRefunded:
		return true
	}
	return false
}

type ShipmentStatus int

const (
	ShipmentStatusUnknown ShipmentStatus = iota
	ShipmentStatusPending
	ShipmentStatusReadyToShip
	ShipmentStatusShipped
	ShipmentStatusInTransit
	ShipmentStatusOutForDelivery
	ShipmentStatusDelivered
	ShipmentStatusCancelled
	ShipmentStatusReturned
	ShipmentStatusNotDelivered
)

func (s ShipmentStatus) String() string {
	switch s {
	case ShipmentStatusPending:
		return "pending"
	case ShipmentStatusReadyToShip:
		return "ready_to_ship"
	case ShipmentStatusShipped:
		return "shipped"
	case ShipmentStatusInTransit:
		return "in_transit"
	case ShipmentStatusOutForDelivery:
		return "out_for_delivery"
	case ShipmentStatusDelivered:
		return "delivered"
	case ShipmentStatusCancelled:
		return "cancelled"
	case ShipmentStatusReturned:
		return "returned"
	case ShipmentStatusNotDelivered:
		return "not_delivered"
	default:
		return "unknown"
	}
}

func (s ShipmentStatus) IsFinal() bool {
	switch s {
	case ShipmentStatusDelivered, ShipmentStatusCancelled, ShipmentStatusReturned:
		return true
	}
	return false
}

func (s ShipmentStatus) CanCancel() bool {
	switch s {
	case ShipmentStatusPending, ShipmentStatusReadyToShip:
		return true
	}
	return false
}

type QRType string

const (
	QRTypeDynamic QRType = "dynamic"
	QRTypeStatic  QRType = "static"
)

func (t QRType) String() string {
	return string(t)
}

func (t QRType) IsValid() bool {
	return t == QRTypeDynamic || t == QRTypeStatic
}

type QRStatus int

const (
	QRStatusUnknown QRStatus = iota
	QRStatusActive
	QRStatusPending
	QRStatusApproved
	QRStatusRejected
	QRStatusExpired
	QRStatusCancelled
)

func (s QRStatus) String() string {
	switch s {
	case QRStatusActive:
		return "active"
	case QRStatusPending:
		return "pending"
	case QRStatusApproved:
		return "approved"
	case QRStatusRejected:
		return "rejected"
	case QRStatusExpired:
		return "expired"
	case QRStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func (s QRStatus) IsFinal() bool {
	switch s {
	case QRStatusApproved, QRStatusRejected, QRStatusExpired, QRStatusCancelled:
		return true
	}
	return false
}
