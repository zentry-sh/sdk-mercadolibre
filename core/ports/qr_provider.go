package ports

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type QRProvider interface {
	CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error)
	GetQR(ctx context.Context, qrID string) (*domain.QRCode, error)
	GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error)
	DeleteQR(ctx context.Context, qrID string) error
	GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error)
	RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error)
	GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error)
	ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error)
	DeletePOS(ctx context.Context, posID string) error
	RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error)
	GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error)
	ListStores(ctx context.Context) ([]*domain.StoreInfo, error)
}
