package usecases

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/ports"
)

type CapabilitiesService struct {
	provider ports.CapabilitiesProvider
}

func NewCapabilitiesService(provider ports.CapabilitiesProvider) *CapabilitiesService {
	return &CapabilitiesService{
		provider: provider,
	}
}

func (s *CapabilitiesService) GetCapabilities(ctx context.Context, countryCode string) (*domain.RegionCapabilities, error) {
	return s.provider.GetCapabilities(ctx, countryCode)
}

func (s *CapabilitiesService) ListSupportedRegions(ctx context.Context) ([]domain.Region, error) {
	return s.provider.ListSupportedRegions(ctx)
}

func (s *CapabilitiesService) GetPaymentMethods(ctx context.Context, countryCode string) ([]domain.PaymentMethodInfo, error) {
	caps, err := s.provider.GetCapabilities(ctx, countryCode)
	if err != nil {
		return nil, err
	}
	return caps.Payment.SupportedMethods, nil
}

func (s *CapabilitiesService) GetCarriers(ctx context.Context, countryCode string) ([]domain.CarrierInfo, error) {
	caps, err := s.provider.GetCapabilities(ctx, countryCode)
	if err != nil {
		return nil, err
	}
	return caps.Shipment.SupportedCarriers, nil
}

func (s *CapabilitiesService) IsQRSupported(ctx context.Context, countryCode string) (bool, error) {
	caps, err := s.provider.GetCapabilities(ctx, countryCode)
	if err != nil {
		return false, err
	}
	return caps.QR.Supported, nil
}

func (s *CapabilitiesService) ValidatePaymentRequest(ctx context.Context, countryCode string, req *domain.CreatePaymentRequest) error {
	return s.provider.ValidatePaymentRequest(ctx, countryCode, req)
}

func (s *CapabilitiesService) ValidateShipmentRequest(ctx context.Context, countryCode string, req *domain.CreateShipmentRequest) error {
	return s.provider.ValidateShipmentRequest(ctx, countryCode, req)
}

func (s *CapabilitiesService) ValidateQRRequest(ctx context.Context, countryCode string, req *domain.CreateQRRequest) error {
	return s.provider.ValidateQRRequest(ctx, countryCode, req)
}

func (s *CapabilitiesService) GetCurrency(ctx context.Context, countryCode string) (string, error) {
	caps, err := s.provider.GetCapabilities(ctx, countryCode)
	if err != nil {
		return "", err
	}
	return caps.Region.CurrencyCode, nil
}

func (s *CapabilitiesService) GetMaxInstallments(ctx context.Context, countryCode string) (int, error) {
	caps, err := s.provider.GetCapabilities(ctx, countryCode)
	if err != nil {
		return 0, err
	}
	return caps.Payment.MaxInstallments, nil
}
