package config

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"gopkg.in/yaml.v3"
)

//go:embed capabilities/*.yaml
var capabilitiesFS embed.FS

type CapabilitiesYAML struct {
	Region     RegionYAML     `yaml:"region"`
	Payment    PaymentYAML    `yaml:"payment"`
	Shipment   ShipmentYAML   `yaml:"shipment"`
	QR         QRYAML         `yaml:"qr"`
	RateLimits RateLimitsYAML `yaml:"rate_limits"`
}

type RegionYAML struct {
	CountryCode  string `yaml:"country_code"`
	CurrencyCode string `yaml:"currency_code"`
	Locale       string `yaml:"locale"`
	TimezoneIANA string `yaml:"timezone_iana"`
}

type PaymentYAML struct {
	SupportedMethods       []PaymentMethodYAML `yaml:"supported_methods"`
	MinAmount              float64             `yaml:"min_amount"`
	MaxAmount              float64             `yaml:"max_amount"`
	MaxAmountWithoutKYC    float64             `yaml:"max_amount_without_kyc"`
	SupportsRefunds        bool                `yaml:"supports_refunds"`
	SupportsPartialRefunds bool                `yaml:"supports_partial_refunds"`
	SupportsInstallments   bool                `yaml:"supports_installments"`
	MaxInstallments        int                 `yaml:"max_installments"`
	SupportedCurrencies    []string            `yaml:"supported_currencies"`
	RequiresKYC            bool                `yaml:"requires_kyc"`
	RequiresTaxID          bool                `yaml:"requires_tax_id"`
}

type PaymentMethodYAML struct {
	ID             string  `yaml:"id"`
	Type           string  `yaml:"type"`
	Name           string  `yaml:"name"`
	MinAmount      float64 `yaml:"min_amount"`
	MaxAmount      float64 `yaml:"max_amount"`
	ProcessingTime string  `yaml:"processing_time"`
}

type ShipmentYAML struct {
	SupportedCarriers       []CarrierYAML  `yaml:"supported_carriers"`
	SupportsTracking        bool           `yaml:"supports_tracking"`
	SupportsLabelPrint      bool           `yaml:"supports_label_print"`
	SupportsScheduledPickup bool           `yaml:"supports_scheduled_pickup"`
	SupportsCancellation    bool           `yaml:"supports_cancellation"`
	MaxWeightKg             float64        `yaml:"max_weight_kg"`
	MaxDimensionsCm         DimensionsYAML `yaml:"max_dimensions_cm"`
}

type CarrierYAML struct {
	ID            string   `yaml:"id"`
	Name          string   `yaml:"name"`
	ServiceTypes  []string `yaml:"service_types"`
	CoverageZones []string `yaml:"coverage_zones"`
}

type DimensionsYAML struct {
	Length float64 `yaml:"length"`
	Width  float64 `yaml:"width"`
	Height float64 `yaml:"height"`
}

type QRYAML struct {
	Supported               bool    `yaml:"supported"`
	SupportsDynamicQR       bool    `yaml:"supports_dynamic_qr"`
	SupportsStaticQR        bool    `yaml:"supports_static_qr"`
	MaxExpirationMinutes    int     `yaml:"max_expiration_minutes"`
	MinAmount               float64 `yaml:"min_amount"`
	MaxAmount               float64 `yaml:"max_amount"`
	RequiresPOSRegistration bool    `yaml:"requires_pos_registration"`
}

type RateLimitsYAML struct {
	RequestsPerSecond     int `yaml:"requests_per_second"`
	RequestsPerMinute     int `yaml:"requests_per_minute"`
	ConcurrentConnections int `yaml:"concurrent_connections"`
}

type Loader struct {
	cache map[string]*domain.RegionCapabilities
	mu    sync.RWMutex
}

func NewLoader() *Loader {
	return &Loader{
		cache: make(map[string]*domain.RegionCapabilities),
	}
}

func (l *Loader) Load(countryCode string) (*domain.RegionCapabilities, error) {
	l.mu.RLock()
	if caps, ok := l.cache[countryCode]; ok {
		l.mu.RUnlock()
		return caps, nil
	}
	l.mu.RUnlock()

	filename := fmt.Sprintf("capabilities/%s.yaml", strings.ToLower(countryCode))
	data, err := capabilitiesFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("capabilities not found for country %s: %w", countryCode, err)
	}

	var yamlCaps CapabilitiesYAML
	if err := yaml.Unmarshal(data, &yamlCaps); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities for %s: %w", countryCode, err)
	}

	caps := l.toDomain(&yamlCaps)

	l.mu.Lock()
	l.cache[countryCode] = caps
	l.mu.Unlock()

	return caps, nil
}

func (l *Loader) ListSupportedCountries() []string {
	entries, err := capabilitiesFS.ReadDir("capabilities")
	if err != nil {
		return nil
	}

	countries := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if len(name) > 5 && name[len(name)-5:] == ".yaml" {
				countries = append(countries, strings.ToUpper(name[:len(name)-5]))
			}
		}
	}
	return countries
}

func (l *Loader) toDomain(y *CapabilitiesYAML) *domain.RegionCapabilities {
	currency := y.Region.CurrencyCode

	methods := make([]domain.PaymentMethodInfo, len(y.Payment.SupportedMethods))
	for i, m := range y.Payment.SupportedMethods {
		methods[i] = domain.PaymentMethodInfo{
			ID:             m.ID,
			Type:           mapPaymentMethod(m.Type),
			Name:           m.Name,
			MinAmount:      domain.Money{Amount: m.MinAmount, Currency: currency},
			MaxAmount:      domain.Money{Amount: m.MaxAmount, Currency: currency},
			ProcessingTime: m.ProcessingTime,
		}
	}

	carriers := make([]domain.CarrierInfo, len(y.Shipment.SupportedCarriers))
	for i, c := range y.Shipment.SupportedCarriers {
		carriers[i] = domain.CarrierInfo{
			ID:            c.ID,
			Name:          c.Name,
			ServiceTypes:  c.ServiceTypes,
			CoverageZones: c.CoverageZones,
		}
	}

	var maxAmountWithoutKYC *domain.Money
	if y.Payment.MaxAmountWithoutKYC > 0 {
		maxAmountWithoutKYC = &domain.Money{Amount: y.Payment.MaxAmountWithoutKYC, Currency: currency}
	}

	return &domain.RegionCapabilities{
		Region: domain.Region{
			CountryCode:  y.Region.CountryCode,
			CurrencyCode: y.Region.CurrencyCode,
			Locale:       y.Region.Locale,
			TimezoneIANA: y.Region.TimezoneIANA,
		},
		Payment: domain.PaymentCapabilities{
			SupportedMethods:       methods,
			MinAmount:              domain.Money{Amount: y.Payment.MinAmount, Currency: currency},
			MaxAmount:              domain.Money{Amount: y.Payment.MaxAmount, Currency: currency},
			MaxAmountWithoutKYC:    maxAmountWithoutKYC,
			SupportsRefunds:        y.Payment.SupportsRefunds,
			SupportsPartialRefunds: y.Payment.SupportsPartialRefunds,
			SupportsInstallments:   y.Payment.SupportsInstallments,
			MaxInstallments:        y.Payment.MaxInstallments,
			SupportedCurrencies:    y.Payment.SupportedCurrencies,
			RequiresKYC:            y.Payment.RequiresKYC,
			RequiresTaxID:          y.Payment.RequiresTaxID,
		},
		Shipment: domain.ShipmentCapabilities{
			SupportedCarriers:       carriers,
			SupportsTracking:        y.Shipment.SupportsTracking,
			SupportsLabelPrint:      y.Shipment.SupportsLabelPrint,
			SupportsScheduledPickup: y.Shipment.SupportsScheduledPickup,
			SupportsCancellation:    y.Shipment.SupportsCancellation,
			MaxWeightKg:             y.Shipment.MaxWeightKg,
			MaxDimensionsCm: domain.Dimensions{
				Length: y.Shipment.MaxDimensionsCm.Length,
				Width:  y.Shipment.MaxDimensionsCm.Width,
				Height: y.Shipment.MaxDimensionsCm.Height,
			},
		},
		QR: domain.QRCapabilities{
			Supported:               y.QR.Supported,
			SupportsDynamicQR:       y.QR.SupportsDynamicQR,
			SupportsStaticQR:        y.QR.SupportsStaticQR,
			MaxExpirationMinutes:    y.QR.MaxExpirationMinutes,
			MinAmount:               domain.Money{Amount: y.QR.MinAmount, Currency: currency},
			MaxAmount:               domain.Money{Amount: y.QR.MaxAmount, Currency: currency},
			RequiresPOSRegistration: y.QR.RequiresPOSRegistration,
		},
		RateLimits: domain.RateLimits{
			RequestsPerSecond:     y.RateLimits.RequestsPerSecond,
			RequestsPerMinute:     y.RateLimits.RequestsPerMinute,
			ConcurrentConnections: y.RateLimits.ConcurrentConnections,
		},
	}
}

func mapPaymentMethod(methodType string) domain.PaymentMethod {
	switch methodType {
	case "card":
		return domain.PaymentMethodCard
	case "transfer":
		return domain.PaymentMethodTransfer
	case "cash":
		return domain.PaymentMethodCash
	case "qr":
		return domain.PaymentMethodQR
	case "wallet":
		return domain.PaymentMethodWallet
	default:
		return domain.PaymentMethodCard
	}
}
