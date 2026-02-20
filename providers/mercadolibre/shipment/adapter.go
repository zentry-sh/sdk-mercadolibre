package shipment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

var formatNewHeader = httputil.WithHeader("x-format-new", "true")

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

func (a *Adapter) CreateShipment(ctx context.Context, req *domain.CreateShipmentRequest) (*domain.Shipment, error) {
	return nil, errors.NewError(errors.ErrCodeInvalidRequest,
		"shipments are created automatically by Mercado Libre when an order is paid; use GetShipment or GetShipmentByOrder instead")
}

func (a *Adapter) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	a.log.Debug("get_shipment", "id", id)

	path := fmt.Sprintf("/shipments/%s", url.PathEscape(id))

	var mlResp MLShipmentResponse
	if err := a.http.GetWithOptions(ctx, path, &mlResp, formatNewHeader); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainShipment(&mlResp), nil
}

func (a *Adapter) GetShipmentByOrder(ctx context.Context, orderID string) (*domain.Shipment, error) {
	a.log.Debug("get_shipment_by_order", "order_id", orderID)

	path := fmt.Sprintf("/orders/%s/shipments", url.PathEscape(orderID))

	var mlResp MLShipmentResponse
	if err := a.http.GetWithOptions(ctx, path, &mlResp, formatNewHeader); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainShipment(&mlResp), nil
}

func (a *Adapter) ListShipments(ctx context.Context, filters domain.ShipmentFilters) ([]*domain.Shipment, error) {
	a.log.Debug("list_shipments")

	query := a.mapper.BuildShipmentSearchQuery(filters)
	path := fmt.Sprintf("/shipments/search%s", query)

	var mlResp MLShipmentSearchResponse
	if err := a.http.GetWithOptions(ctx, path, &mlResp, formatNewHeader); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainShipments(mlResp.Results), nil
}

func (a *Adapter) UpdateShipment(ctx context.Context, id string, req *domain.UpdateShipmentRequest) (*domain.Shipment, error) {
	a.log.Debug("update_shipment", "id", id)

	mlReq := a.mapper.ToMLUpdateRequest(req)
	path := fmt.Sprintf("/shipments/%s", url.PathEscape(id))

	if err := a.http.PutWithOptions(ctx, path, mlReq, nil, formatNewHeader); err != nil {
		return nil, err
	}

	return a.GetShipment(ctx, id)
}

func (a *Adapter) CancelShipment(ctx context.Context, id string) error {
	a.log.Debug("cancel_shipment", "id", id)

	body := map[string]string{"status": "cancelled"}
	path := fmt.Sprintf("/shipments/%s", url.PathEscape(id))

	return a.http.PutWithOptions(ctx, path, body, nil, formatNewHeader)
}

func (a *Adapter) GetTracking(ctx context.Context, shipmentID string) ([]domain.ShipmentEvent, error) {
	a.log.Debug("get_tracking", "shipment_id", shipmentID)

	path := fmt.Sprintf("/shipments/%s/history", url.PathEscape(shipmentID))

	var mlResp []MLShipmentHistoryEntry
	if err := a.http.GetWithOptions(ctx, path, &mlResp, formatNewHeader); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainTrackingEvents(mlResp), nil
}

func (a *Adapter) GetLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	a.log.Debug("get_label", "shipment_id", shipmentID)

	path := fmt.Sprintf("/shipments/%s/labels", url.PathEscape(shipmentID))

	return a.http.DoRaw(ctx, http.MethodGet, path,
		httputil.WithHeader("Accept", "application/pdf"),
		formatNewHeader,
	)
}
