package mercadolibre

type Endpoints struct {
	BaseURL      string
	PaymentsAPI  string
	ShipmentsAPI string
	QRAPI        string
	OAuth2URL    string
}

var countryEndpoints = map[string]Endpoints{
	"PE": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
	"MX": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
	"AR": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
	"BR": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
	"CL": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
	"CO": {
		BaseURL:      "https://api.mercadolibre.com",
		PaymentsAPI:  "https://api.mercadopago.com",
		ShipmentsAPI: "https://api.mercadolibre.com",
		QRAPI:        "https://api.mercadopago.com",
		OAuth2URL:    "https://api.mercadolibre.com/oauth/token",
	},
}

func GetEndpoints(country string) Endpoints {
	if endpoints, ok := countryEndpoints[country]; ok {
		return endpoints
	}
	return countryEndpoints["PE"]
}

func ListSupportedCountries() []string {
	countries := make([]string, 0, len(countryEndpoints))
	for country := range countryEndpoints {
		countries = append(countries, country)
	}
	return countries
}

func IsCountrySupported(country string) bool {
	_, ok := countryEndpoints[country]
	return ok
}
