package mercadolibre

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/errors"
	"github.com/zentry/sdk-mercadolibre/pkg/httputil"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

type AuthClient struct {
	http         *httputil.Client
	clientID     string
	clientSecret string
	log          logger.Logger
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	UserID       int64  `json:"user_id"`
	RefreshToken string `json:"refresh_token"`
}

type Credentials struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       int64
}

func NewAuthClient(country, clientID, clientSecret string, log logger.Logger) *AuthClient {
	endpoints := GetEndpoints(country)

	if log == nil {
		log = logger.Nop()
	}

	return &AuthClient{
		http: httputil.NewClient(httputil.ClientConfig{
			BaseURL: endpoints.OAuth2URL,
			Timeout: 30 * time.Second,
			Logger:  log,
		}),
		clientID:     clientID,
		clientSecret: clientSecret,
		log:          log,
	}
}

func (c *AuthClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*Credentials, error) {
	req := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}

	var resp TokenResponse
	if err := c.http.Post(ctx, "", req, &resp); err != nil {
		return nil, err
	}

	return c.tokenToCredentials(&resp), nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*Credentials, error) {
	req := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"refresh_token": refreshToken,
	}

	var resp TokenResponse
	if err := c.http.Post(ctx, "", req, &resp); err != nil {
		return nil, err
	}

	return c.tokenToCredentials(&resp), nil
}

func (c *AuthClient) tokenToCredentials(resp *TokenResponse) *Credentials {
	return &Credentials{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
		UserID:       resp.UserID,
	}
}

func (c *Credentials) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Credentials) ShouldRefresh() bool {
	return time.Now().Add(5 * time.Minute).After(c.ExpiresAt)
}

func BuildAuthorizationURL(country, clientID, redirectURI string, scopes []string) string {
	endpoints := GetEndpoints(country)

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	if len(scopes) > 0 {
		params.Set("scope", strings.Join(scopes, " "))
	}

	u, err := url.JoinPath(endpoints.BaseURL, "/authorization")
	if err != nil {
		return ""
	}
	return u + "?" + params.Encode()
}

type TokenManager struct {
	authClient  *AuthClient
	credentials *Credentials
	onRefresh   func(*Credentials)
}

func NewTokenManager(authClient *AuthClient, credentials *Credentials) *TokenManager {
	return &TokenManager{
		authClient:  authClient,
		credentials: credentials,
	}
}

func (m *TokenManager) SetOnRefresh(callback func(*Credentials)) {
	m.onRefresh = callback
}

func (m *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	if m.credentials == nil {
		return "", errors.NewError(errors.ErrCodeUnauthorized, "no credentials available")
	}

	if !m.credentials.ShouldRefresh() {
		return m.credentials.AccessToken, nil
	}

	newCreds, err := m.authClient.RefreshToken(ctx, m.credentials.RefreshToken)
	if err != nil {
		return "", err
	}

	m.credentials = newCreds

	if m.onRefresh != nil {
		m.onRefresh(newCreds)
	}

	return m.credentials.AccessToken, nil
}
