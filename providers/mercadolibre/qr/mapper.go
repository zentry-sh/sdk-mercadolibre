package qr

import (
	"fmt"
	"net/url"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type Mapper struct{}

func NewMapper() *Mapper { return &Mapper{} }

func (m *Mapper) ToMLCreateOrderRequest(req *domain.CreateQRRequest) *MLCreateOrderRequest {
	mlReq := &MLCreateOrderRequest{
		Type:              "qr",
		ExternalReference: req.ExternalReference,
		Title:             req.Description,
		Description:       req.Description,
		NotificationURL:   req.NotificationURL,
	}

	if req.Amount != nil {
		mlReq.TotalAmount = req.Amount.Amount
	}

	if req.ExpirationMinutes > 0 {
		exp := time.Now().Add(time.Duration(req.ExpirationMinutes) * time.Minute)
		mlReq.ExpirationDate = &exp
	}

	if len(req.Items) > 0 {
		mlReq.Items = make([]MLOrderItem, len(req.Items))
		for i, item := range req.Items {
			mlReq.Items[i] = MLOrderItem{
				Title:       item.Title,
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice.Amount,
				TotalAmount: item.TotalAmount.Amount,
			}
		}
	}

	return mlReq
}

func (m *Mapper) ToDomainQR(ml *MLOrderResponse) *domain.QRCode {
	if ml == nil {
		return nil
	}

	qr := &domain.QRCode{
		ID:                ml.ID,
		POSID:             ml.POSID,
		StoreID:           ml.StoreID,
		CollectorID:       fmt.Sprintf("%d", ml.CollectorID),
		ExternalReference: ml.ExternalReference,
		Status:            m.MapQRStatus(ml.Status),
		QRData:            ml.QRData,
		Description:       ml.Type,
	}

	if ml.TotalAmount > 0 {
		qr.Amount = &domain.Money{Amount: ml.TotalAmount}
	}
	if ml.CreatedDate != nil {
		qr.CreatedAt = *ml.CreatedDate
	}
	if ml.LastUpdatedDate != nil {
		qr.UpdatedAt = *ml.LastUpdatedDate
	}
	if ml.ExpirationDate != nil {
		qr.ExpiresAt = ml.ExpirationDate
	}

	qr.Type = domain.QRTypeDynamic
	if ml.Type == "static" {
		qr.Type = domain.QRTypeStatic
	}

	if len(ml.Payments) > 0 {
		p := ml.Payments[0]
		qr.Payment = &domain.Payment{
			ID:     fmt.Sprintf("%d", p.ID),
			Amount: domain.Money{Amount: p.TransactionAmount},
			Status: m.mapPaymentStatus(p.Status),
		}
	}

	return qr
}

func (m *Mapper) ToDomainPOS(ml *MLPOSResponse) *domain.POSInfo {
	if ml == nil {
		return nil
	}

	pos := &domain.POSInfo{
		ID:          fmt.Sprintf("%d", ml.ID),
		Name:        ml.Name,
		ExternalID:  ml.ExternalID,
		StoreID:     ml.StoreID,
		FixedAmount: ml.FixedAmount,
		Category:    ml.Category,
		URL:         ml.URL,
	}

	if ml.QR != nil {
		pos.QRCode = ml.QR.Image
	}
	if ml.DateCreated != nil {
		pos.CreatedAt = *ml.DateCreated
	}
	if ml.DateLastUpdated != nil {
		pos.UpdatedAt = *ml.DateLastUpdated
	}

	return pos
}

func (m *Mapper) ToDomainPOSList(items []MLPOSResponse) []*domain.POSInfo {
	result := make([]*domain.POSInfo, len(items))
	for i := range items {
		result[i] = m.ToDomainPOS(&items[i])
	}
	return result
}

func (m *Mapper) ToDomainStore(ml *MLStoreResponse) *domain.StoreInfo {
	if ml == nil {
		return nil
	}

	store := &domain.StoreInfo{
		ID:            ml.ID,
		Name:          ml.Name,
		ExternalID:    ml.ExternalID,
		BusinessHours: ml.BusinessHours,
	}

	if ml.Location != nil {
		store.Location = domain.Address{
			Street:  ml.Location.StreetName,
			Number:  ml.Location.StreetNumber,
			City:    ml.Location.CityName,
			State:   ml.Location.StateName,
			Lat:     ml.Location.Latitude,
			Lon:     ml.Location.Longitude,
		}
	}
	if ml.DateCreated != nil {
		store.CreatedAt = *ml.DateCreated
	}

	return store
}

func (m *Mapper) ToDomainStoreList(items []MLStoreResponse) []*domain.StoreInfo {
	result := make([]*domain.StoreInfo, len(items))
	for i := range items {
		result[i] = m.ToDomainStore(&items[i])
	}
	return result
}

func (m *Mapper) ToMLPOSRequest(req *domain.RegisterPOSRequest) *MLPOSRequest {
	return &MLPOSRequest{
		Name:        req.Name,
		ExternalID:  req.ExternalID,
		StoreID:     req.StoreID,
		FixedAmount: req.FixedAmount,
		Category:    req.Category,
	}
}

func (m *Mapper) ToMLStoreRequest(req *domain.RegisterStoreRequest) *MLStoreRequest {
	mlReq := &MLStoreRequest{
		Name:          req.Name,
		ExternalID:    req.ExternalID,
		BusinessHours: req.BusinessHours,
	}

	if req.Location.Street != "" || req.Location.City != "" {
		mlReq.Location = &MLStoreLocation{
			StreetName:   req.Location.Street,
			StreetNumber: req.Location.Number,
			CityName:     req.Location.City,
			StateName:    req.Location.State,
			Latitude:     req.Location.Lat,
			Longitude:    req.Location.Lon,
		}
	}

	return mlReq
}

func (m *Mapper) BuildExternalRefQuery(ref string) string {
	params := url.Values{}
	params.Set("external_reference", ref)
	return fmt.Sprintf("?%s", params.Encode())
}

func (m *Mapper) BuildStoreSearchQuery(storeID string) string {
	params := url.Values{}
	if storeID != "" {
		params.Set("store_id", storeID)
	}
	encoded := params.Encode()
	if encoded == "" {
		return ""
	}
	return fmt.Sprintf("?%s", encoded)
}

func (m *Mapper) MapQRStatus(status string) domain.QRStatus {
	switch status {
	case "active", "opened":
		return domain.QRStatusActive
	case "pending":
		return domain.QRStatusPending
	case "approved", "closed":
		return domain.QRStatusApproved
	case "rejected":
		return domain.QRStatusRejected
	case "expired":
		return domain.QRStatusExpired
	case "cancelled":
		return domain.QRStatusCancelled
	default:
		return domain.QRStatusUnknown
	}
}

func (m *Mapper) mapPaymentStatus(status string) domain.PaymentStatus {
	switch status {
	case "pending":
		return domain.PaymentStatusPending
	case "approved":
		return domain.PaymentStatusApproved
	case "rejected":
		return domain.PaymentStatusRejected
	case "cancelled":
		return domain.PaymentStatusCancelled
	case "in_process":
		return domain.PaymentStatusInProcess
	case "refunded":
		return domain.PaymentStatusRefunded
	default:
		return domain.PaymentStatusUnknown
	}
}
