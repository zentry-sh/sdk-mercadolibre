package mercadolibre

import (
	"time"

	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Config struct {
	AccessToken   string
	ClientID      string
	ClientSecret  string
	Country       string
	Timeout       time.Duration
	Logger        logger.Logger
	WebhookSecret string
}

type Client struct {
	http         *httputil.Client
	config       Config
	log          logger.Logger
	paymentsURL  string
	shipmentsURL string
	qrURL        string
}

func NewClient(config Config) *Client {
	endpoints := GetEndpoints(config.Country)
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	log := config.Logger
	if log == nil {
		log = logger.Nop()
	}

	httpClient := httputil.NewClient(httputil.ClientConfig{
		BaseURL:     endpoints.PaymentsAPI,
		AccessToken: config.AccessToken,
		Timeout:     timeout,
		Logger:      log,
	})

	return &Client{
		http:         httpClient,
		config:       config,
		log:          log,
		paymentsURL:  endpoints.PaymentsAPI,
		shipmentsURL: endpoints.ShipmentsAPI,
		qrURL:        endpoints.QRAPI,
	}
}

func (c *Client) SetAccessToken(token string) {
	c.config.AccessToken = token
	c.http.SetAccessToken(token)
}

func (c *Client) HTTP() *httputil.Client {
	return c.http
}

func (c *Client) PaymentsHTTP() *httputil.Client {
	return httputil.NewClient(httputil.ClientConfig{
		BaseURL:     c.paymentsURL,
		AccessToken: c.config.AccessToken,
		Timeout:     c.config.Timeout,
		Logger:      c.log,
	})
}

func (c *Client) ShipmentsHTTP() *httputil.Client {
	return httputil.NewClient(httputil.ClientConfig{
		BaseURL:     c.shipmentsURL,
		AccessToken: c.config.AccessToken,
		Timeout:     c.config.Timeout,
		Logger:      c.log,
	})
}

func (c *Client) QRHTTP() *httputil.Client {
	return httputil.NewClient(httputil.ClientConfig{
		BaseURL:     c.qrURL,
		AccessToken: c.config.AccessToken,
		Timeout:     c.config.Timeout,
		Logger:      c.log,
	})
}
