package qr

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http *httputil.Client
	log  logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.Nop()
	}
	return &Adapter{http: http, log: log}
}

func (a *Adapter) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	a.log.Debug("create_qr", "external_ref", req.ExternalReference)
	return nil, nil
}

func (a *Adapter) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	return nil, nil
}

func (a *Adapter) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	return nil, nil
}

func (a *Adapter) DeleteQR(ctx context.Context, qrID string) error {
	return nil
}

func (a *Adapter) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	return nil, nil
}

func (a *Adapter) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	return nil, nil
}

func (a *Adapter) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	return nil, nil
}

func (a *Adapter) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	return nil, nil
}

func (a *Adapter) DeletePOS(ctx context.Context, posID string) error {
	return nil
}

func (a *Adapter) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	return nil, nil
}

func (a *Adapter) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	return nil, nil
}

func (a *Adapter) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	return nil, nil
}
