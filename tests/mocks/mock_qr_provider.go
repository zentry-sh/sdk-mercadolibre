package mocks

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type MockQRProvider struct {
	CreateQRFn                 func(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error)
	GetQRFn                    func(ctx context.Context, qrID string) (*domain.QRCode, error)
	GetQRByExternalReferenceFn func(ctx context.Context, ref string) (*domain.QRCode, error)
	DeleteQRFn                 func(ctx context.Context, qrID string) error
	GetQRPaymentFn             func(ctx context.Context, qrID string) (*domain.Payment, error)
	RegisterPOSFn              func(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error)
	GetPOSFn                   func(ctx context.Context, posID string) (*domain.POSInfo, error)
	ListPOSFn                  func(ctx context.Context, storeID string) ([]*domain.POSInfo, error)
	DeletePOSFn                func(ctx context.Context, posID string) error
	RegisterStoreFn            func(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error)
	GetStoreFn                 func(ctx context.Context, storeID string) (*domain.StoreInfo, error)
	ListStoresFn               func(ctx context.Context) ([]*domain.StoreInfo, error)
}

func (m *MockQRProvider) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	if m.CreateQRFn != nil {
		return m.CreateQRFn(ctx, req)
	}
	return nil, nil
}

func (m *MockQRProvider) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	if m.GetQRFn != nil {
		return m.GetQRFn(ctx, qrID)
	}
	return nil, nil
}

func (m *MockQRProvider) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	if m.GetQRByExternalReferenceFn != nil {
		return m.GetQRByExternalReferenceFn(ctx, ref)
	}
	return nil, nil
}

func (m *MockQRProvider) DeleteQR(ctx context.Context, qrID string) error {
	if m.DeleteQRFn != nil {
		return m.DeleteQRFn(ctx, qrID)
	}
	return nil
}

func (m *MockQRProvider) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	if m.GetQRPaymentFn != nil {
		return m.GetQRPaymentFn(ctx, qrID)
	}
	return nil, nil
}

func (m *MockQRProvider) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	if m.RegisterPOSFn != nil {
		return m.RegisterPOSFn(ctx, req)
	}
	return nil, nil
}

func (m *MockQRProvider) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	if m.GetPOSFn != nil {
		return m.GetPOSFn(ctx, posID)
	}
	return nil, nil
}

func (m *MockQRProvider) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	if m.ListPOSFn != nil {
		return m.ListPOSFn(ctx, storeID)
	}
	return nil, nil
}

func (m *MockQRProvider) DeletePOS(ctx context.Context, posID string) error {
	if m.DeletePOSFn != nil {
		return m.DeletePOSFn(ctx, posID)
	}
	return nil
}

func (m *MockQRProvider) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	if m.RegisterStoreFn != nil {
		return m.RegisterStoreFn(ctx, req)
	}
	return nil, nil
}

func (m *MockQRProvider) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	if m.GetStoreFn != nil {
		return m.GetStoreFn(ctx, storeID)
	}
	return nil, nil
}

func (m *MockQRProvider) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	if m.ListStoresFn != nil {
		return m.ListStoresFn(ctx)
	}
	return nil, nil
}
