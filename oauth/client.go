// Package oauth implements DataGSM OAuth 2.0 and OpenID Connect.
package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tmp-it-play/datagsm-sdk-go/internal/transport"
)

const (
	DefaultAuthorizationURL = "https://oauth.authorization.datagsm.kr"
	DefaultResourceURL      = "https://oauth.resource.datagsm.kr"
)

type AuthMethod string

const (
	AuthSecretPost  AuthMethod = "client_secret_post"
	AuthSecretBasic AuthMethod = "client_secret_basic"
	AuthNone        AuthMethod = "none"
)

type RequestFormat string

const (
	RequestFormatForm RequestFormat = "form"
	RequestFormatJSON RequestFormat = "json"
)

type Config struct {
	ClientID             string
	ClientSecret         string
	RedirectURI          string
	AuthMethod           AuthMethod
	RequestFormat        RequestFormat
	ClockSkew            time.Duration
	AuthorizationBaseURL string
	ResourceBaseURL      string
	HTTPClient           *http.Client
}

type Client struct {
	config        Config
	authorization *transport.Client
	resource      *transport.Client
}

func NewClient(config Config) (*Client, error) {
	if strings.TrimSpace(config.ClientID) == "" {
		return nil, errors.New("datagsm: OAuth client ID is required")
	}
	if config.AuthMethod == "" {
		config.AuthMethod = AuthNone
		if config.ClientSecret != "" {
			config.AuthMethod = AuthSecretPost
		}
	}
	switch config.AuthMethod {
	case AuthNone:
		if config.ClientSecret != "" {
			return nil, errors.New("datagsm: auth method none cannot use a client secret")
		}
	case AuthSecretPost, AuthSecretBasic:
		if config.ClientSecret == "" {
			return nil, errors.New("datagsm: client authentication requires a secret")
		}
	default:
		return nil, errors.New("datagsm: unsupported client authentication method")
	}
	switch config.RequestFormat {
	case "":
		config.RequestFormat = RequestFormatForm
	case RequestFormatForm, RequestFormatJSON:
	default:
		return nil, errors.New("datagsm: unsupported token request format")
	}
	if config.AuthorizationBaseURL == "" {
		config.AuthorizationBaseURL = DefaultAuthorizationURL
	}
	if config.ResourceBaseURL == "" {
		config.ResourceBaseURL = DefaultResourceURL
	}
	a, err := transport.New(config.AuthorizationBaseURL, config.HTTPClient)
	if err != nil {
		return nil, err
	}
	r, err := transport.New(config.ResourceBaseURL, config.HTTPClient)
	if err != nil {
		return nil, err
	}
	return &Client{config: config, authorization: a, resource: r}, nil
}

func (c *Client) Close() {
	c.authorization.HTTP.CloseIdleConnections()
	c.resource.HTTP.CloseIdleConnections()
}

type CodeExchange struct {
	Code         string
	CodeVerifier string
	Scopes       []string
}

func (c *Client) Exchange(ctx context.Context, input CodeExchange) (*Token, error) {
	if input.Code == "" || c.config.RedirectURI == "" {
		return nil, errors.New("datagsm: code and redirect URI are required")
	}
	if input.CodeVerifier != "" {
		if _, err := CodeChallenge(input.CodeVerifier); err != nil {
			return nil, err
		}
	} else if c.config.AuthMethod == AuthNone {
		return nil, errors.New("datagsm: code verifier is required without a client secret")
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {input.Code}, "redirect_uri": {c.config.RedirectURI}}
	if input.CodeVerifier != "" {
		form.Set("code_verifier", input.CodeVerifier)
	}
	return c.token(ctx, form, input.Scopes)
}

// Refresh must be serialized per session; store the returned refresh token before the next call.
func (c *Client) Refresh(ctx context.Context, refreshToken string, scopes ...string) (*Token, error) {
	if refreshToken == "" {
		return nil, errors.New("datagsm: refresh token is required")
	}
	return c.token(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}}, scopes)
}

func (c *Client) ClientCredentials(ctx context.Context, scopes ...string) (*Token, error) {
	if c.config.AuthMethod == AuthNone {
		return nil, errors.New("datagsm: client credentials requires a client secret")
	}
	return c.token(ctx, url.Values{"grant_type": {"client_credentials"}}, scopes)
}

func (c *Client) tokenRequest(ctx context.Context, form url.Values, scopes []string) (*http.Request, error) {
	form.Set("client_id", c.config.ClientID)
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	if c.config.AuthMethod == AuthSecretPost {
		form.Set("client_secret", c.config.ClientSecret)
	}
	contentType, body := "application/x-www-form-urlencoded", form.Encode()
	if c.config.RequestFormat == RequestFormatJSON {
		fields := make(map[string]string, len(form))
		for key := range form {
			fields[key] = form.Get(key)
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			return nil, fmt.Errorf("datagsm: encode token request: %w", err)
		}
		contentType, body = "application/json", string(encoded)
	}
	req, err := c.authorization.Request(ctx, http.MethodPost, "/v1/oauth/token", nil, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if c.config.AuthMethod == AuthSecretBasic {
		req.SetBasicAuth(url.QueryEscape(c.config.ClientID), url.QueryEscape(c.config.ClientSecret))
	}
	return req, nil
}

func (c *Client) token(ctx context.Context, form url.Values, scopes []string) (*Token, error) {
	req, err := c.tokenRequest(ctx, form, scopes)
	if err != nil {
		return nil, err
	}
	var result Token
	if err := c.authorization.Do(req, &result, false); err != nil {
		return nil, err
	}
	if result.AccessToken == "" || !strings.EqualFold(result.TokenType, "Bearer") || result.ExpiresIn <= 0 {
		return nil, errors.New("datagsm: invalid token response")
	}
	return &result, nil
}

func (c *Client) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if strings.TrimSpace(accessToken) == "" || strings.ContainsAny(accessToken, "\r\n") {
		return nil, errors.New("datagsm: valid access token is required")
	}
	req, err := c.resource.Request(ctx, http.MethodGet, "/userinfo", nil, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	var result UserInfo
	if err := c.resource.Do(req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Discovery(ctx context.Context) (*Discovery, error) {
	req, err := c.authorization.Request(ctx, http.MethodGet, "/.well-known/openid-configuration", nil, nil)
	if err != nil {
		return nil, err
	}
	var result Discovery
	if err := c.authorization.Do(req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) JWKS(ctx context.Context) (*JWKS, error) {
	req, err := c.authorization.Request(ctx, http.MethodGet, "/v1/oauth/jwks", nil, nil)
	if err != nil {
		return nil, err
	}
	var result JWKS
	if err := c.authorization.Do(req, &result, false); err != nil {
		return nil, err
	}
	result.ClockSkew = c.config.ClockSkew
	return &result, nil
}

// Logout uses the HTTP client's cookie jar. It does not revoke issued tokens or clear a browser's cookies.
func (c *Client) Logout(ctx context.Context) error {
	req, err := c.authorization.Request(ctx, http.MethodPost, "/v1/oauth/logout", nil, nil)
	if err != nil {
		return err
	}
	return c.authorization.Do(req, nil, false)
}
