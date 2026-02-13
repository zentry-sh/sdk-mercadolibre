package ports

import (
	"context"

	"github.com/zentry/sdk-mercadolibre/core/domain"
)

type CapabilitiesProvider interface {
	GetCapabilities(ctx context.Context, countryCode string) (*domain.RegionCapabilities, error)
	ListSupportedRegions(ctx context.Context) ([]domain.Region, error)
	ValidatePaymentRequest(ctx context.Context, countryCode string, req *domain.CreatePaymentRequest) error
	ValidateShipmentRequest(ctx context.Context, countryCode string, req *domain.CreateShipmentRequest) error
	ValidateQRRequest(ctx context.Context, countryCode string, req *domain.CreateQRRequest) error
}
