package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
	logger      logger.Logger
	retryConfig RetryConfig
}

type ClientConfig struct {
	BaseURL     string
	AccessToken string
	Timeout     time.Duration
	Logger      logger.Logger
	RetryConfig *RetryConfig
}

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}
}

func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	retryConfig := DefaultRetryConfig()
	if config.RetryConfig != nil {
		retryConfig = *config.RetryConfig
	}

	log := config.Logger
	if log == nil {
		log = logger.NewNopLogger()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:     config.BaseURL,
		accessToken: config.AccessToken,
		logger:      log,
		retryConfig: retryConfig,
	}
}

func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *Client) Do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doWithRetry(ctx, method, path, body, result)
}

func (c *Client) doWithRetry(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("retrying request", "attempt", attempt, "backoff", backoff)
			select {
			case <-ctx.Done():
				return errors.NewErrorWithCause(errors.ErrCodeTimeout, "context cancelled", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > c.retryConfig.MaxBackoff {
				backoff = c.retryConfig.MaxBackoff
			}
		}

		err := c.doRequest(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		if !c.shouldRetry(err) {
			return err
		}
	}

	return lastErr
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "failed to marshal request body", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeInternal, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	c.logger.Debug("sending request", "method", method, "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "request failed", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "failed to read response body", err)
	}

	c.logger.Debug("received response", "status", resp.StatusCode)

	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return errors.NewErrorWithCause(errors.ErrCodeInternal, "failed to unmarshal response", err)
		}
	}

	return nil
}

func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var apiErr struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Status  int    `json:"status"`
		Cause   []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"cause"`
	}

	json.Unmarshal(body, &apiErr)

	message := apiErr.Message
	if message == "" {
		message = apiErr.Error
	}
	if message == "" {
		message = fmt.Sprintf("HTTP %d error", statusCode)
	}

	providerCode := ""
	providerMessage := message
	if len(apiErr.Cause) > 0 {
		providerCode = apiErr.Cause[0].Code
		providerMessage = apiErr.Cause[0].Description
	}

	switch statusCode {
	case http.StatusBadRequest:
		return errors.NewProviderError(errors.ErrCodeInvalidRequest, message, providerCode, providerMessage)
	case http.StatusUnauthorized:
		return errors.NewError(errors.ErrCodeUnauthorized, "unauthorized")
	case http.StatusForbidden:
		return errors.NewError(errors.ErrCodeForbidden, "forbidden")
	case http.StatusNotFound:
		return errors.NewError(errors.ErrCodeNotFound, "resource not found")
	case http.StatusConflict:
		return errors.NewProviderError(errors.ErrCodeConflict, message, providerCode, providerMessage)
	case http.StatusTooManyRequests:
		return errors.RateLimited()
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return errors.Timeout()
	default:
		return errors.NewProviderError(errors.ErrCodeProviderError, message, providerCode, providerMessage)
	}
}

func (c *Client) shouldRetry(err error) bool {
	if sdkErr, ok := err.(*errors.SDKError); ok {
		switch sdkErr.Code {
		case errors.ErrCodeRateLimited, errors.ErrCodeTimeout, errors.ErrCodeNetworkError:
			return true
		}
	}
	return false
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.Do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Do(ctx, http.MethodPut, path, body, result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}
