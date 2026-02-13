package domain

type Region struct {
	CountryCode  string
	CurrencyCode string
	Locale       string
	TimezoneIANA string
}

type RegionCapabilities struct {
	Region     Region
	Payment    PaymentCapabilities
	Shipment   ShipmentCapabilities
	QR         QRCapabilities
	RateLimits RateLimits
}

type PaymentCapabilities struct {
	SupportedMethods       []PaymentMethodInfo
	MinAmount              Money
	MaxAmount              Money
	MaxAmountWithoutKYC    *Money
	SupportsRefunds        bool
	SupportsPartialRefunds bool
	SupportsInstallments   bool
	MaxInstallments        int
	SupportedCurrencies    []string
	RequiresKYC            bool
	RequiresTaxID          bool
}

func (c PaymentCapabilities) IsMethodSupported(methodType PaymentMethod) bool {
	for _, m := range c.SupportedMethods {
		if m.Type == methodType {
			return true
		}
	}
	return false
}

func (c PaymentCapabilities) GetMethodInfo(methodID string) *PaymentMethodInfo {
	for _, m := range c.SupportedMethods {
		if m.ID == methodID {
			return &m
		}
	}
	return nil
}

type PaymentMethodInfo struct {
	ID             string
	Type           PaymentMethod
	Name           string
	MinAmount      Money
	MaxAmount      Money
	ProcessingTime string
	Metadata       map[string]interface{}
}

type ShipmentCapabilities struct {
	SupportedCarriers       []CarrierInfo
	SupportsTracking        bool
	SupportsLabelPrint      bool
	SupportsScheduledPickup bool
	SupportsCancellation    bool
	MaxWeightKg             float64
	MaxDimensionsCm         Dimensions
}

func (c ShipmentCapabilities) IsCarrierSupported(carrierID string) bool {
	for _, carrier := range c.SupportedCarriers {
		if carrier.ID == carrierID {
			return true
		}
	}
	return false
}

type CarrierInfo struct {
	ID            string
	Name          string
	ServiceTypes  []string
	CoverageZones []string
}

type QRCapabilities struct {
	Supported               bool
	SupportsDynamicQR       bool
	SupportsStaticQR        bool
	MaxExpirationMinutes    int
	MinAmount               Money
	MaxAmount               Money
	RequiresPOSRegistration bool
}

type RateLimits struct {
	RequestsPerSecond     int
	RequestsPerMinute     int
	ConcurrentConnections int
}
