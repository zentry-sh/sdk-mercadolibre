package domain

import "time"

type Shipment struct {
	ID                string
	OrderID           string
	ExternalReference string
	Status            ShipmentStatus
	SubStatus         string
	Origin            Address
	Destination       Address
	Package           Package
	Carrier           Carrier
	TrackingNumber    string
	ServiceType       string
	Label             *LabelInfo
	EstimatedDelivery *time.Time
	Events            []ShipmentEvent
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *Shipment) IsDelivered() bool {
	return s.Status == ShipmentStatusDelivered
}

func (s *Shipment) CanCancel() bool {
	return s.Status.CanCancel()
}

func (s *Shipment) LastEvent() *ShipmentEvent {
	if len(s.Events) == 0 {
		return nil
	}
	return &s.Events[len(s.Events)-1]
}

type ShipmentEvent struct {
	Status      ShipmentStatus
	SubStatus   string
	Description string
	Location    string
	Date        time.Time
}

type CreateShipmentRequest struct {
	OrderID           string
	ExternalReference string
	Origin            Address
	Destination       Address
	Package           Package
	ServiceType       string
	CarrierID         string
}

type UpdateShipmentRequest struct {
	Destination *Address
	Package     *Package
}

type ShipmentFilters struct {
	OrderID           string
	ExternalReference string
	Status            *ShipmentStatus
	FromDate          *time.Time
	ToDate            *time.Time
	Limit             int
	Offset            int
}
