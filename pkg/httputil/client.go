package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

const maxResponseBytes = 10 << 20 // 10 MiB

type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
	log         logger.Logger
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
		log = logger.Nop()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:     config.BaseURL,
		accessToken: config.AccessToken,
		log:         log,
		retryConfig: retryConfig,
	}
}

func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *Client) Do(ctx context.Context, method, path string, body any, result any) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "failed to marshal request body", err)
		}
	}
	return c.doWithRetry(ctx, method, path, bodyBytes, result)
}

func (c *Client) doWithRetry(ctx context.Context, method, path string, body []byte, result any) error {
	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			c.log.Debug("retrying request", "attempt", attempt, "backoff_ms", backoff.Milliseconds())
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

		if err := c.doRequest(ctx, method, path, body, result); err != nil {
			lastErr = err
			if !c.shouldRetry(err) {
				return err
			}
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte, result any) error {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "invalid request path", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeInternal, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}

	c.log.Debug("http request", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "request failed", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	respBody, err := io.ReadAll(limited)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "failed to read response body", err)
	}

	c.log.Debug("http response", "status", resp.StatusCode, "bytes", len(respBody))

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

	_ = json.Unmarshal(body, &apiErr)

	message := apiErr.Message
	if message == "" {
		message = apiErr.Error
	}
	if message == "" {
		message = fmt.Sprintf("HTTP %d", statusCode)
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
	sdkErr, ok := err.(*errors.SDKError)
	if !ok {
		return false
	}
	switch sdkErr.Code {
	case errors.ErrCodeRateLimited, errors.ErrCodeTimeout, errors.ErrCodeNetworkError:
		return true
	}
	return false
}

type RequestOption func(*http.Request)

func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) { r.Header.Set(key, value) }
}

func (c *Client) DoWithOptions(ctx context.Context, method, path string, body any, result any, opts ...RequestOption) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "failed to marshal request body", err)
		}
	}
	return c.doWithRetryOpts(ctx, method, path, bodyBytes, result, opts)
}

func (c *Client) DoRaw(ctx context.Context, method, path string, opts ...RequestOption) ([]byte, error) {
	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			c.log.Debug("retrying request", "attempt", attempt, "backoff_ms", backoff.Milliseconds())
			select {
			case <-ctx.Done():
				return nil, errors.NewErrorWithCause(errors.ErrCodeTimeout, "context cancelled", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > c.retryConfig.MaxBackoff {
				backoff = c.retryConfig.MaxBackoff
			}
		}

		data, err := c.doRawRequest(ctx, method, path, opts)
		if err != nil {
			lastErr = err
			if !c.shouldRetry(err) {
				return nil, err
			}
			continue
		}
		return data, nil
	}
	return nil, lastErr
}

func (c *Client) doRawRequest(ctx context.Context, method, path string, opts []RequestOption) ([]byte, error) {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "invalid request path", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, errors.NewErrorWithCause(errors.ErrCodeInternal, "failed to create request", err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
	for _, opt := range opts {
		opt(req)
	}

	c.log.Debug("http request raw", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewErrorWithCause(errors.ErrCodeNetworkError, "request failed", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	respBody, err := io.ReadAll(limited)
	if err != nil {
		return nil, errors.NewErrorWithCause(errors.ErrCodeNetworkError, "failed to read response body", err)
	}

	c.log.Debug("http response raw", "status", resp.StatusCode, "bytes", len(respBody))

	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (c *Client) doWithRetryOpts(ctx context.Context, method, path string, body []byte, result any, opts []RequestOption) error {
	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			c.log.Debug("retrying request", "attempt", attempt, "backoff_ms", backoff.Milliseconds())
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

		if err := c.doRequestOpts(ctx, method, path, body, result, opts); err != nil {
			lastErr = err
			if !c.shouldRetry(err) {
				return err
			}
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Client) doRequestOpts(ctx context.Context, method, path string, body []byte, result any, opts []RequestOption) error {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeInvalidRequest, "invalid request path", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeInternal, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
	for _, opt := range opts {
		opt(req)
	}

	c.log.Debug("http request", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "request failed", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxResponseBytes)
	respBody, err := io.ReadAll(limited)
	if err != nil {
		return errors.NewErrorWithCause(errors.ErrCodeNetworkError, "failed to read response body", err)
	}

	c.log.Debug("http response", "status", resp.StatusCode, "bytes", len(respBody))

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

func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.Do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.Do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.Do(ctx, http.MethodPut, path, body, result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) GetWithOptions(ctx context.Context, path string, result any, opts ...RequestOption) error {
	return c.DoWithOptions(ctx, http.MethodGet, path, nil, result, opts...)
}

func (c *Client) PostWithOptions(ctx context.Context, path string, body any, result any, opts ...RequestOption) error {
	return c.DoWithOptions(ctx, http.MethodPost, path, body, result, opts...)
}

func (c *Client) PutWithOptions(ctx context.Context, path string, body any, result any, opts ...RequestOption) error {
	return c.DoWithOptions(ctx, http.MethodPut, path, body, result, opts...)
}

func (c *Client) DeleteWithOptions(ctx context.Context, path string, opts ...RequestOption) error {
	return c.DoWithOptions(ctx, http.MethodDelete, path, nil, nil, opts...)
}
