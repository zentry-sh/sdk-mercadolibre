package sdk

import (
	"time"

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

func (c *Config) Validate() error {
	if c.Country == "" {
		c.Country = "PE"
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		Country: "PE",
		Timeout: 30 * time.Second,
	}
}
