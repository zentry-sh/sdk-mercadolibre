package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type QRService struct {
	provider ports.QRProvider
	logger   logger.Logger
}

func NewQRService(provider ports.QRProvider, log logger.Logger) *QRService {
	if log == nil {
		log = logger.NewNopLogger()
	}
	return &QRService{
		provider: provider,
		logger:   log,
	}
}

func (s *QRService) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	s.logger.Info("creating QR",
		"pos_id", req.POSID,
		"external_ref", req.ExternalReference,
		"type", req.Type)

	qr, err := s.provider.CreateQR(ctx, req)
	if err != nil {
		s.logger.Error("failed to create QR", "error", err)
		return nil, err
	}

	s.logger.Info("QR created",
		"id", qr.ID,
		"status", qr.Status.String())

	return qr, nil
}

func (s *QRService) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	if qrID == "" {
		return nil, errors.InvalidRequest("QR id is required")
	}

	return s.provider.GetQR(ctx, qrID)
}

func (s *QRService) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	if ref == "" {
		return nil, errors.InvalidRequest("external reference is required")
	}

	return s.provider.GetQRByExternalReference(ctx, ref)
}

func (s *QRService) DeleteQR(ctx context.Context, qrID string) error {
	if qrID == "" {
		return errors.InvalidRequest("QR id is required")
	}

	s.logger.Info("deleting QR", "id", qrID)

	err := s.provider.DeleteQR(ctx, qrID)
	if err != nil {
		s.logger.Error("failed to delete QR", "error", err)
		return err
	}

	s.logger.Info("QR deleted", "id", qrID)
	return nil
}

func (s *QRService) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	if qrID == "" {
		return nil, errors.InvalidRequest("QR id is required")
	}

	return s.provider.GetQRPayment(ctx, qrID)
}

func (s *QRService) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	if err := s.validateRegisterPOSRequest(req); err != nil {
		return nil, err
	}

	s.logger.Info("registering POS",
		"name", req.Name,
		"store_id", req.StoreID)

	pos, err := s.provider.RegisterPOS(ctx, req)
	if err != nil {
		s.logger.Error("failed to register POS", "error", err)
		return nil, err
	}

	s.logger.Info("POS registered", "id", pos.ID)
	return pos, nil
}

func (s *QRService) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	if posID == "" {
		return nil, errors.InvalidRequest("POS id is required")
	}

	return s.provider.GetPOS(ctx, posID)
}

func (s *QRService) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	return s.provider.ListPOS(ctx, storeID)
}

func (s *QRService) DeletePOS(ctx context.Context, posID string) error {
	if posID == "" {
		return errors.InvalidRequest("POS id is required")
	}

	s.logger.Info("deleting POS", "id", posID)

	err := s.provider.DeletePOS(ctx, posID)
	if err != nil {
		s.logger.Error("failed to delete POS", "error", err)
		return err
	}

	s.logger.Info("POS deleted", "id", posID)
	return nil
}

func (s *QRService) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	if req.Name == "" {
		return nil, errors.InvalidRequest("store name is required")
	}

	s.logger.Info("registering store", "name", req.Name)

	store, err := s.provider.RegisterStore(ctx, req)
	if err != nil {
		s.logger.Error("failed to register store", "error", err)
		return nil, err
	}

	s.logger.Info("store registered", "id", store.ID)
	return store, nil
}

func (s *QRService) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	if storeID == "" {
		return nil, errors.InvalidRequest("store id is required")
	}

	return s.provider.GetStore(ctx, storeID)
}

func (s *QRService) ListStores(ctx context.Context) ([]*domain.StoreInfo, error) {
	return s.provider.ListStores(ctx)
}

func (s *QRService) validateCreateRequest(req *domain.CreateQRRequest) error {
	if req.ExternalReference == "" {
		return errors.InvalidRequest("external_reference is required")
	}
	if !req.Type.IsValid() {
		return errors.InvalidRequest("invalid QR type")
	}
	if req.Type == domain.QRTypeDynamic && (req.Amount == nil || req.Amount.Amount <= 0) {
		return errors.InvalidRequest("amount is required for dynamic QR")
	}
	return nil
}

func (s *QRService) validateRegisterPOSRequest(req *domain.RegisterPOSRequest) error {
	if req.Name == "" {
		return errors.InvalidRequest("POS name is required")
	}
	if req.ExternalID == "" {
		return errors.InvalidRequest("external_id is required")
	}
	return nil
}
