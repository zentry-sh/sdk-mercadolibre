package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/core/ports"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
	"github.com/zentry/sdk-mercadolibre/pkg/sanitize"
)

type QRService struct {
	provider ports.QRProvider
	log      logger.Logger
}

func NewQRService(provider ports.QRProvider, log logger.Logger) *QRService {
	if log == nil {
		log = logger.Nop()
	}
	return &QRService{
		provider: provider,
		log:      log,
	}
}

func (s *QRService) CreateQR(ctx context.Context, req *domain.CreateQRRequest) (*domain.QRCode, error) {
	req.ExternalReference = sanitize.String(req.ExternalReference)
	req.POSID = sanitize.ID(req.POSID)
	req.CollectorID = sanitize.ID(req.CollectorID)
	req.StoreID = sanitize.ID(req.StoreID)
	req.Description = sanitize.String(req.Description)
	req.NotificationURL = sanitize.String(req.NotificationURL)

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	s.log.Debug("create_qr", "pos_id", req.POSID, "type", req.Type, "external_ref", req.ExternalReference)
	return s.provider.CreateQR(ctx, req)
}

func (s *QRService) GetQR(ctx context.Context, qrID string) (*domain.QRCode, error) {
	qrID = sanitize.ID(qrID)
	if qrID == "" {
		return nil, errors.InvalidRequest("QR id is required")
	}
	return s.provider.GetQR(ctx, qrID)
}

func (s *QRService) GetQRByExternalReference(ctx context.Context, ref string) (*domain.QRCode, error) {
	ref = sanitize.String(ref)
	if ref == "" {
		return nil, errors.InvalidRequest("external reference is required")
	}
	return s.provider.GetQRByExternalReference(ctx, ref)
}

func (s *QRService) DeleteQR(ctx context.Context, qrID string) error {
	qrID = sanitize.ID(qrID)
	if qrID == "" {
		return errors.InvalidRequest("QR id is required")
	}
	return s.provider.DeleteQR(ctx, qrID)
}

func (s *QRService) GetQRPayment(ctx context.Context, qrID string) (*domain.Payment, error) {
	qrID = sanitize.ID(qrID)
	if qrID == "" {
		return nil, errors.InvalidRequest("QR id is required")
	}
	return s.provider.GetQRPayment(ctx, qrID)
}

func (s *QRService) RegisterPOS(ctx context.Context, req *domain.RegisterPOSRequest) (*domain.POSInfo, error) {
	req.Name = sanitize.String(req.Name)
	req.ExternalID = sanitize.ID(req.ExternalID)
	req.StoreID = sanitize.ID(req.StoreID)

	if err := s.validateRegisterPOSRequest(req); err != nil {
		return nil, err
	}
	s.log.Debug("register_pos", "name", req.Name, "store_id", req.StoreID)
	return s.provider.RegisterPOS(ctx, req)
}

func (s *QRService) GetPOS(ctx context.Context, posID string) (*domain.POSInfo, error) {
	posID = sanitize.ID(posID)
	if posID == "" {
		return nil, errors.InvalidRequest("POS id is required")
	}
	return s.provider.GetPOS(ctx, posID)
}

func (s *QRService) ListPOS(ctx context.Context, storeID string) ([]*domain.POSInfo, error) {
	storeID = sanitize.ID(storeID)
	return s.provider.ListPOS(ctx, storeID)
}

func (s *QRService) DeletePOS(ctx context.Context, posID string) error {
	posID = sanitize.ID(posID)
	if posID == "" {
		return errors.InvalidRequest("POS id is required")
	}
	return s.provider.DeletePOS(ctx, posID)
}

func (s *QRService) RegisterStore(ctx context.Context, req *domain.RegisterStoreRequest) (*domain.StoreInfo, error) {
	req.Name = sanitize.String(req.Name)
	req.ExternalID = sanitize.ID(req.ExternalID)
	req.Location = sanitizeAddress(req.Location)

	if req.Name == "" {
		return nil, errors.InvalidRequest("store name is required")
	}
	return s.provider.RegisterStore(ctx, req)
}

func (s *QRService) GetStore(ctx context.Context, storeID string) (*domain.StoreInfo, error) {
	storeID = sanitize.ID(storeID)
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
