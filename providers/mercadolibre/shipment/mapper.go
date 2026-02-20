package shipment

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type Mapper struct{}

func NewMapper() *Mapper { return &Mapper{} }

func (m *Mapper) ToDomainShipment(ml *MLShipmentResponse) *domain.Shipment {
	if ml == nil {
		return nil
	}

	s := &domain.Shipment{
		ID:                fmt.Sprintf("%d", ml.ID),
		OrderID:           fmt.Sprintf("%d", ml.OrderID),
		ExternalReference: ml.ExternalReference,
		Status:            m.MapShipmentStatus(ml.Status),
		SubStatus:         ml.SubStatus,
		TrackingNumber:    ml.TrackingNumber,
		ServiceType:       ml.LogisticType,
		CreatedAt:         ml.DateCreated,
		UpdatedAt:         ml.LastUpdated,
	}

	if ml.SenderAddress != nil {
		s.Origin = m.ToDomainAddress(ml.SenderAddress)
	}
	if ml.ReceiverAddress != nil {
		s.Destination = m.ToDomainAddress(ml.ReceiverAddress)
	}
	if ml.Carrier != nil {
		s.Carrier = domain.Carrier{
			ID:   fmt.Sprintf("%d", ml.Carrier.ID),
			Name: ml.Carrier.Name,
		}
	}
	if len(ml.ShippingItems) > 0 {
		item := ml.ShippingItems[0]
		s.Package = domain.Package{
			Weight:      item.Weight,
			Length:      item.Length,
			Width:       item.Width,
			Height:      item.Height,
			Description: item.Description,
		}
	}
	if ml.ShippingOption != nil && ml.ShippingOption.EstimatedDeliveryTime != nil {
		if ml.ShippingOption.EstimatedDeliveryTime.Date != nil {
			s.EstimatedDelivery = ml.ShippingOption.EstimatedDeliveryTime.Date
		}
	}
	if ml.DateFirstPrinted != nil {
		s.Label = &domain.LabelInfo{
			Format:    "PDF",
			CreatedAt: *ml.DateFirstPrinted,
		}
	}

	return s
}

func (m *Mapper) ToDomainShipments(items []MLShipmentResponse) []*domain.Shipment {
	result := make([]*domain.Shipment, len(items))
	for i := range items {
		result[i] = m.ToDomainShipment(&items[i])
	}
	return result
}

func (m *Mapper) ToDomainAddress(ml *MLShippingAddress) domain.Address {
	addr := domain.Address{
		Street:  ml.StreetName,
		Number:  ml.StreetNumber,
		ZipCode: ml.ZipCode,
		Lat:     ml.Latitude,
		Lon:     ml.Longitude,
	}
	if ml.City != nil {
		addr.City = ml.City.Name
	}
	if ml.State != nil {
		addr.State = ml.State.Name
	}
	if ml.Country != nil {
		addr.Country = ml.Country.Name
	}
	return addr
}

func (m *Mapper) ToDomainTrackingEvents(entries []MLShipmentHistoryEntry) []domain.ShipmentEvent {
	events := make([]domain.ShipmentEvent, len(entries))
	for i, e := range entries {
		events[i] = domain.ShipmentEvent{
			Status:      m.MapShipmentStatus(e.Status),
			SubStatus:   e.SubStatus,
			Description: e.Description,
			Date:        e.Date,
		}
	}
	return events
}

func (m *Mapper) MapShipmentStatus(status string) domain.ShipmentStatus {
	switch status {
	case "pending":
		return domain.ShipmentStatusPending
	case "handling", "ready_to_ship":
		return domain.ShipmentStatusReadyToShip
	case "shipped":
		return domain.ShipmentStatusShipped
	case "in_transit":
		return domain.ShipmentStatusInTransit
	case "out_for_delivery":
		return domain.ShipmentStatusOutForDelivery
	case "delivered":
		return domain.ShipmentStatusDelivered
	case "cancelled":
		return domain.ShipmentStatusCancelled
	case "returned", "returning_to_sender":
		return domain.ShipmentStatusReturned
	case "not_delivered":
		return domain.ShipmentStatusNotDelivered
	default:
		return domain.ShipmentStatusUnknown
	}
}

func (m *Mapper) BuildShipmentSearchQuery(filters domain.ShipmentFilters) string {
	params := url.Values{}

	if filters.OrderID != "" {
		params.Set("order_id", filters.OrderID)
	}
	if filters.ExternalReference != "" {
		params.Set("external_reference", filters.ExternalReference)
	}
	if filters.Status != nil {
		params.Set("status", filters.Status.String())
	}
	if filters.FromDate != nil {
		params.Set("date_from", filters.FromDate.Format("2006-01-02T15:04:05.000-07:00"))
	}
	if filters.ToDate != nil {
		params.Set("date_to", filters.ToDate.Format("2006-01-02T15:04:05.000-07:00"))
	}
	if filters.Limit > 0 {
		params.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Offset > 0 {
		params.Set("offset", strconv.Itoa(filters.Offset))
	}

	encoded := params.Encode()
	if encoded == "" {
		return ""
	}
	return fmt.Sprintf("?%s", encoded)
}

func (m *Mapper) ToMLUpdateRequest(req *domain.UpdateShipmentRequest) *MLUpdateShipmentRequest {
	if req == nil {
		return nil
	}

	mlReq := &MLUpdateShipmentRequest{}

	if req.Destination != nil {
		mlReq.ReceiverAddress = &MLUpdateAddress{
			StreetName:   req.Destination.Street,
			StreetNumber: req.Destination.Number,
			ZipCode:      req.Destination.ZipCode,
		}
	}

	return mlReq
}
