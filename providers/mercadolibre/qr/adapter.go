package qr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/idempotency"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http   *httputil.Client
	mapper *Mapper
	log    logger.Logger
	userID int64
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

func (a *Adapter) SetUserID(id int64) {
	a.userID = id
}

func (a *Adapter) ResolveUserID(ctx context.Context) (int64, error) {
	if a.userID != 0 {
		return a.userID, nil
	}

	var user MLUserResponse
	if err := a.http.Get(ctx, "/users/me", &user); err != nil {
		return 0, errors.NewErrorWithCause(errors.ErrCodeUnauthorized, "failed to resolve user_id", err)
	}

	a.userID = user.ID
	a.log.Debug("resolved_user_id", "user_id", a.userID)
	return a.userID, nil
}

func (a *Adapter) idempotentPost(ctx context.Context, path string, body any, result any) error {
	return a.http.PostWithOptions(ctx, path, body, result,
		httputil.WithHeader("X-Idempotency-Key", idempotency.NewKey()),
	)
}

func (a *Adapter) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	a.log.Debug("create_qr", "external_ref", req.ExternalReference, "type", req.Type)

	mlReq := a.mapper.ToMLCreateOrderRequest(req)

	var mlResp MLOrderResponse
	if err := a.idempotentPost(ctx, "/v1/orders", mlReq, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainQR(&mlResp), nil
}

func (a *Adapter) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	a.log.Debug("get_qr", "id", qrID)

	path := fmt.Sprintf("/v1/orders/%s", url.PathEscape(qrID))

	var mlResp MLOrderResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainQR(&mlResp), nil
}

func (a *Adapter) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	a.log.Debug("get_qr_by_external_ref", "ref", ref)

	query := a.mapper.BuildExternalRefQuery(ref)
	path := fmt.Sprintf("/v1/orders%s", query)

	var mlResp MLOrderSearchResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	if len(mlResp.Elements) == 0 {
		return nil, errors.NotFound("QR order")
	}

	return a.mapper.ToDomainQR(&mlResp.Elements[0]), nil
}

func (a *Adapter) DeleteQR(ctx context.Context, qrID string) error {
	a.log.Debug("delete_qr", "id", qrID)

	path := fmt.Sprintf("/v1/orders/%s/cancel", url.PathEscape(qrID))
	return a.idempotentPost(ctx, path, nil, nil)
}

func (a *Adapter) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	a.log.Debug("get_qr_payment", "id", qrID)

	qr, err := a.GetQR(ctx, qrID)
	if err != nil {
		return nil, err
	}
	if qr.Payment == nil {
		return nil, errors.NotFound("payment for QR order")
	}

	return qr.Payment, nil
}

func (a *Adapter) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	a.log.Debug("register_pos", "name", req.Name, "external_id", req.ExternalID)

	mlReq := a.mapper.ToMLPOSRequest(req)

	var mlResp MLPOSResponse
	if err := a.idempotentPost(ctx, "/pos", mlReq, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainPOS(&mlResp), nil
}

func (a *Adapter) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	a.log.Debug("get_pos", "id", posID)

	path := fmt.Sprintf("/pos/%s", url.PathEscape(posID))

	var mlResp MLPOSResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainPOS(&mlResp), nil
}

func (a *Adapter) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	a.log.Debug("list_pos", "store_id", storeID)

	query := a.mapper.BuildStoreSearchQuery(storeID)
	path := fmt.Sprintf("/pos%s", query)

	var mlResp MLPOSSearchResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainPOSList(mlResp.Results), nil
}

func (a *Adapter) DeletePOS(ctx context.Context, posID string) error {
	a.log.Debug("delete_pos", "id", posID)

	path := fmt.Sprintf("/pos/%s", url.PathEscape(posID))
	return a.http.Delete(ctx, path)
}

func (a *Adapter) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	a.log.Debug("register_store", "name", req.Name)

	userID, err := a.ResolveUserID(ctx)
	if err != nil {
		return nil, err
	}

	mlReq := a.mapper.ToMLStoreRequest(req)
	path := fmt.Sprintf("/users/%d/stores", userID)

	var mlResp MLStoreResponse
	if err := a.idempotentPost(ctx, path, mlReq, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainStore(&mlResp), nil
}

func (a *Adapter) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	a.log.Debug("get_store", "id", storeID)

	path := fmt.Sprintf("/stores/%s", url.PathEscape(storeID))

	var mlResp MLStoreResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainStore(&mlResp), nil
}

func (a *Adapter) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	a.log.Debug("list_stores")

	userID, err := a.ResolveUserID(ctx)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/users/%d/stores/search", userID)

	var mlResp MLStoreSearchResponse
	if err := a.http.Get(ctx, path, &mlResp); err != nil {
		return nil, err
	}

	return a.mapper.ToDomainStoreList(mlResp.Results), nil
}
