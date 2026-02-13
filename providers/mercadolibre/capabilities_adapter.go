package mercadolibre

import (
	"context"
	"fmt"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre/config"
)

type CapabilitiesAdapter struct {
	loader *config.Loader
}

func NewCapabilitiesAdapter() *CapabilitiesAdapter {
	return &CapabilitiesAdapter{
		loader: config.NewLoader(),
	}
}

func (a *CapabilitiesAdapter) GetCapabilities(ctx context.Context, countryCode string) (*domain.RegionCapabilities, error) {
	caps, err := a.loader.Load(countryCode)
	if err != nil {
		return nil, errors.NewError(errors.ErrCodeNotFound, fmt.Sprintf("capabilities not found for country: %s", countryCode))
	}
	return caps, nil
}

func (a *CapabilitiesAdapter) ListSupportedRegions(ctx context.Context) ([]domain.Region, error) {
	countries := a.loader.ListSupportedCountries()
	regions := make([]domain.Region, 0, len(countries))

	for _, country := range countries {
		caps, err := a.loader.Load(country)
		if err != nil {
			continue
		}
		regions = append(regions, caps.Region)
	}

	return regions, nil
}

func (a *CapabilitiesAdapter) ValidatePaymentRequest(ctx context.Context, countryCode string, req *domain.CreatePaymentRequest) error {
	caps, err := a.GetCapabilities(ctx, countryCode)
	if err != nil {
		return err
	}

	if req.Amount.Amount < caps.Payment.MinAmount.Amount {
		return errors.InvalidRequest(fmt.Sprintf("amount %.2f is below minimum %.2f for %s",
			req.Amount.Amount, caps.Payment.MinAmount.Amount, countryCode))
	}

	if req.Amount.Amount > caps.Payment.MaxAmount.Amount {
		return errors.InvalidRequest(fmt.Sprintf("amount %.2f exceeds maximum %.2f for %s",
			req.Amount.Amount, caps.Payment.MaxAmount.Amount, countryCode))
	}

	currencyValid := false
	for _, c := range caps.Payment.SupportedCurrencies {
		if c == req.Amount.Currency {
			currencyValid = true
			break
		}
	}
	if !currencyValid {
		return errors.InvalidRequest(fmt.Sprintf("currency %s not supported for %s", req.Amount.Currency, countryCode))
	}

	if req.MethodID != "" {
		methodInfo := caps.Payment.GetMethodInfo(req.MethodID)
		if methodInfo == nil {
			return errors.NewError(errors.ErrCodeUnsupportedMethod, fmt.Sprintf("payment method %s not supported for %s", req.MethodID, countryCode))
		}

		if req.Amount.Amount < methodInfo.MinAmount.Amount {
			return errors.InvalidRequest(fmt.Sprintf("amount %.2f is below minimum %.2f for payment method %s",
				req.Amount.Amount, methodInfo.MinAmount.Amount, req.MethodID))
		}

		if req.Amount.Amount > methodInfo.MaxAmount.Amount {
			return errors.InvalidRequest(fmt.Sprintf("amount %.2f exceeds maximum %.2f for payment method %s",
				req.Amount.Amount, methodInfo.MaxAmount.Amount, req.MethodID))
		}
	}

	if req.Installments > 1 {
		if !caps.Payment.SupportsInstallments {
			return errors.InvalidRequest(fmt.Sprintf("installments not supported for %s", countryCode))
		}
		if req.Installments > caps.Payment.MaxInstallments {
			return errors.InvalidRequest(fmt.Sprintf("installments %d exceeds maximum %d for %s",
				req.Installments, caps.Payment.MaxInstallments, countryCode))
		}
	}

	return nil
}

func (a *CapabilitiesAdapter) ValidateShipmentRequest(ctx context.Context, countryCode string, req *domain.CreateShipmentRequest) error {
	caps, err := a.GetCapabilities(ctx, countryCode)
	if err != nil {
		return err
	}

	if req.CarrierID != "" && !caps.Shipment.IsCarrierSupported(req.CarrierID) {
		return errors.InvalidRequest(fmt.Sprintf("carrier %s not supported for %s", req.CarrierID, countryCode))
	}

	if req.Package.Weight > caps.Shipment.MaxWeightKg {
		return errors.InvalidRequest(fmt.Sprintf("package weight %.2f kg exceeds maximum %.2f kg for %s",
			req.Package.Weight, caps.Shipment.MaxWeightKg, countryCode))
	}

	if req.Package.Length > caps.Shipment.MaxDimensionsCm.Length ||
		req.Package.Width > caps.Shipment.MaxDimensionsCm.Width ||
		req.Package.Height > caps.Shipment.MaxDimensionsCm.Height {
		return errors.InvalidRequest(fmt.Sprintf("package dimensions exceed maximum for %s", countryCode))
	}

	return nil
}

func (a *CapabilitiesAdapter) ValidateQRRequest(ctx context.Context, countryCode string, req *domain.CreateQRRequest) error {
	caps, err := a.GetCapabilities(ctx, countryCode)
	if err != nil {
		return err
	}

	if !caps.QR.Supported {
		return errors.InvalidRequest(fmt.Sprintf("QR payments not supported for %s", countryCode))
	}

	if req.Type == domain.QRTypeDynamic && !caps.QR.SupportsDynamicQR {
		return errors.InvalidRequest(fmt.Sprintf("dynamic QR not supported for %s", countryCode))
	}

	if req.Type == domain.QRTypeStatic && !caps.QR.SupportsStaticQR {
		return errors.InvalidRequest(fmt.Sprintf("static QR not supported for %s", countryCode))
	}

	if req.Amount != nil {
		if req.Amount.Amount < caps.QR.MinAmount.Amount {
			return errors.InvalidRequest(fmt.Sprintf("QR amount %.2f is below minimum %.2f for %s",
				req.Amount.Amount, caps.QR.MinAmount.Amount, countryCode))
		}

		if req.Amount.Amount > caps.QR.MaxAmount.Amount {
			return errors.InvalidRequest(fmt.Sprintf("QR amount %.2f exceeds maximum %.2f for %s",
				req.Amount.Amount, caps.QR.MaxAmount.Amount, countryCode))
		}
	}

	if req.ExpirationMinutes > caps.QR.MaxExpirationMinutes {
		return errors.InvalidRequest(fmt.Sprintf("QR expiration %d minutes exceeds maximum %d for %s",
			req.ExpirationMinutes, caps.QR.MaxExpirationMinutes, countryCode))
	}

	return nil
}
