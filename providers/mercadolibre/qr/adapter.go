package qr

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Adapter struct {
	http   *httputil.Client
	logger logger.Logger
}

func NewAdapter(http *httputil.Client, log logger.Logger) *Adapter {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &Adapter{
		http:   http,
		logger: log,
	}
}

func (a *Adapter) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	a.logger.Debug("creating QR", "external_ref", req.ExternalReference)
	return nil, nil
}

func (a *Adapter) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	a.logger.Debug("getting QR", "id", qrID)
	return nil, nil
}

func (a *Adapter) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	a.logger.Debug("getting QR by reference", "ref", ref)
	return nil, nil
}

func (a *Adapter) DeleteQR(ctx context.Context, qrID string) error {
	a.logger.Debug("deleting QR", "id", qrID)
	return nil
}

func (a *Adapter) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	a.logger.Debug("getting QR payment", "id", qrID)
	return nil, nil
}

func (a *Adapter) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	a.logger.Debug("registering POS", "name", req.Name)
	return nil, nil
}

func (a *Adapter) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	a.logger.Debug("getting POS", "id", posID)
	return nil, nil
}

func (a *Adapter) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	a.logger.Debug("listing POS", "store_id", storeID)
	return nil, nil
}

func (a *Adapter) DeletePOS(ctx context.Context, posID string) error {
	a.logger.Debug("deleting POS", "id", posID)
	return nil
}

func (a *Adapter) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	a.logger.Debug("registering store", "name", req.Name)
	return nil, nil
}

func (a *Adapter) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	a.logger.Debug("getting store", "id", storeID)
	return nil, nil
}

func (a *Adapter) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	a.logger.Debug("listing stores")
	return nil, nil
}
