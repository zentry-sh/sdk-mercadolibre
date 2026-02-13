package core

import (
	"context"
	"testing"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/providers/mercadolibre"
)

func TestCapabilitiesAdapter_GetCapabilities_PE(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()

	caps, err := adapter.GetCapabilities(context.Background(), "PE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if caps.Region.CountryCode != "PE" {
		t.Errorf("expected country code 'PE', got '%s'", caps.Region.CountryCode)
	}

	if caps.Region.CurrencyCode != "PEN" {
		t.Errorf("expected currency 'PEN', got '%s'", caps.Region.CurrencyCode)
	}

	if len(caps.Payment.SupportedMethods) == 0 {
		t.Error("expected at least one payment method")
	}

	if !caps.QR.Supported {
		t.Error("expected QR to be supported in PE")
	}
}

func TestCapabilitiesAdapter_GetCapabilities_MX(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()

	caps, err := adapter.GetCapabilities(context.Background(), "MX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if caps.Region.CountryCode != "MX" {
		t.Errorf("expected country code 'MX', got '%s'", caps.Region.CountryCode)
	}

	if caps.Region.CurrencyCode != "MXN" {
		t.Errorf("expected currency 'MXN', got '%s'", caps.Region.CurrencyCode)
	}

	hasOxxo := false
	for _, method := range caps.Payment.SupportedMethods {
		if method.ID == "oxxo" {
			hasOxxo = true
			break
		}
	}
	if !hasOxxo {
		t.Error("expected OXXO payment method in MX")
	}
}

func TestCapabilitiesAdapter_GetCapabilities_AllCountries(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()
	countries := []string{"PE", "MX", "AR", "BR", "CL", "CO"}

	for _, country := range countries {
		t.Run(country, func(t *testing.T) {
			caps, err := adapter.GetCapabilities(context.Background(), country)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", country, err)
			}

			if caps.Region.CountryCode != country {
				t.Errorf("expected country code '%s', got '%s'", country, caps.Region.CountryCode)
			}

			if caps.Region.CurrencyCode == "" {
				t.Errorf("expected currency code for %s", country)
			}

			if len(caps.Payment.SupportedMethods) == 0 {
				t.Errorf("expected at least one payment method for %s", country)
			}

			if caps.Payment.MaxAmount.Amount <= 0 {
				t.Errorf("expected positive max amount for %s", country)
			}
		})
	}
}

func TestCapabilitiesAdapter_ValidatePaymentRequest(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()
	ctx := context.Background()

	tests := []struct {
		name    string
		country string
		req     *domain.CreatePaymentRequest
		wantErr bool
	}{
		{
			name:    "valid PE payment",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount: domain.Money{Amount: 100, Currency: "PEN"},
				Payer:  domain.Payer{Email: "test@example.com"},
			},
			wantErr: false,
		},
		{
			name:    "amount below minimum",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount: domain.Money{Amount: 0.001, Currency: "PEN"},
				Payer:  domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name:    "amount above maximum",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount: domain.Money{Amount: 999999999, Currency: "PEN"},
				Payer:  domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name:    "invalid currency",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount: domain.Money{Amount: 100, Currency: "USD"},
				Payer:  domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name:    "valid MX payment with OXXO",
			country: "MX",
			req: &domain.CreatePaymentRequest{
				Amount:   domain.Money{Amount: 500, Currency: "MXN"},
				MethodID: "oxxo",
				Payer:    domain.Payer{Email: "test@example.com"},
			},
			wantErr: false,
		},
		{
			name:    "unsupported payment method",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount:   domain.Money{Amount: 100, Currency: "PEN"},
				MethodID: "unknown_method",
				Payer:    domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
		{
			name:    "installments exceeds max",
			country: "PE",
			req: &domain.CreatePaymentRequest{
				Amount:       domain.Money{Amount: 100, Currency: "PEN"},
				Installments: 99,
				Payer:        domain.Payer{Email: "test@example.com"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidatePaymentRequest(ctx, tt.country, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePaymentRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCapabilitiesAdapter_ValidateQRRequest(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()
	ctx := context.Background()

	tests := []struct {
		name    string
		country string
		req     *domain.CreateQRRequest
		wantErr bool
	}{
		{
			name:    "valid dynamic QR",
			country: "PE",
			req: &domain.CreateQRRequest{
				ExternalReference: "test-123",
				Type:              domain.QRTypeDynamic,
				Amount:            &domain.Money{Amount: 100, Currency: "PEN"},
			},
			wantErr: false,
		},
		{
			name:    "QR amount exceeds max",
			country: "PE",
			req: &domain.CreateQRRequest{
				ExternalReference: "test-123",
				Type:              domain.QRTypeDynamic,
				Amount:            &domain.Money{Amount: 999999, Currency: "PEN"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateQRRequest(ctx, tt.country, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQRRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCapabilitiesAdapter_ListSupportedRegions(t *testing.T) {
	adapter := mercadolibre.NewCapabilitiesAdapter()

	regions, err := adapter.ListSupportedRegions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(regions) < 6 {
		t.Errorf("expected at least 6 regions, got %d", len(regions))
	}

	countryMap := make(map[string]bool)
	for _, r := range regions {
		countryMap[r.CountryCode] = true
	}

	expectedCountries := []string{"PE", "MX", "AR", "BR", "CL", "CO"}
	for _, country := range expectedCountries {
		if !countryMap[country] {
			t.Errorf("expected country %s in supported regions", country)
		}
	}
}
