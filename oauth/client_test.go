package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTokenGrantsAndRawResponses(t *testing.T) {
	var form url.Values
	body := `{"access_token":"access","token_type":"Bearer","expires_in":3600,"refresh_token":"next","scope":"openid datagsm:account_read","id_token":"signed"}`
	c, err := NewClient(Config{ClientID: "client", ClientSecret: "test-secret", RedirectURI: "https://example.com/callback", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "oauth.authorization.datagsm.kr" || req.URL.Path != "/v1/oauth/token" {
			t.Fatal("wrong token endpoint")
		}
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		form = req.PostForm
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	token, err := c.Exchange(ctx, CodeExchange{Code: "code", CodeVerifier: NewCodeVerifier()})
	if err != nil || token.AccessToken != "access" || *token.RefreshToken != "next" || *token.IDToken != "signed" {
		t.Fatalf("%v", err)
	}
	if form.Get("grant_type") != "authorization_code" || form.Get("redirect_uri") != c.config.RedirectURI || form.Get("code") == "" || form.Get("code_verifier") == "" {
		t.Fatal("incorrect code exchange form")
	}
	if _, err := c.Refresh(ctx, "previous", ScopeAccountRead); err != nil {
		t.Fatal(err)
	}
	if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "previous" || form.Get("scope") != ScopeAccountRead {
		t.Fatal("incorrect refresh form")
	}
	body = `{"access_token":"access","token_type":"Bearer","expires_in":3600,"refresh_token":null,"id_token":null,"scope":"app:read"}`
	token, err = c.ClientCredentials(ctx, "app:read")
	if err != nil || token.RefreshToken != nil || token.IDToken != nil || form.Get("grant_type") != "client_credentials" {
		t.Fatalf("%v", err)
	}
	for _, invalid := range []string{`{}`, `{"access_token":"access","token_type":"Basic","expires_in":3600}`, `{"access_token":"access","token_type":"Bearer","expires_in":0}`} {
		body = invalid
		if _, err := c.ClientCredentials(ctx); err == nil {
			t.Fatal("accepted malformed token response")
		}
	}
}

func TestTokenRequestJSONFormat(t *testing.T) {
	var contentType string
	var fields map[string]string
	c, err := NewClient(Config{ClientID: "client", ClientSecret: "test-secret", RequestFormat: RequestFormatJSON, HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		contentType = req.Header.Get("Content-Type")
		if err := json.NewDecoder(req.Body).Decode(&fields); err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"access","token_type":"Bearer","expires_in":3600}`))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ClientCredentials(context.Background(), "app:read", "app:write"); err != nil {
		t.Fatal(err)
	}
	if contentType != "application/json" || fields["grant_type"] != "client_credentials" || fields["client_id"] != "client" || fields["client_secret"] != "test-secret" || fields["scope"] != "app:read app:write" {
		t.Fatalf("incorrect JSON token request: %s %v", contentType, fields)
	}
	if _, err := NewClient(Config{ClientID: "client", RequestFormat: "xml"}); err == nil {
		t.Fatal("accepted unsupported request format")
	}
}

func TestJWKSUsesConfiguredClockSkew(t *testing.T) {
	c, err := NewClient(Config{ClientID: "client", ClockSkew: 30 * time.Second, HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"keys":[]}`))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	keys, err := c.JWKS(context.Background())
	if err != nil || keys.ClockSkew != 30*time.Second {
		t.Fatalf("%v %v", keys, err)
	}
}

func TestOAuthResourceAndPublicEndpoints(t *testing.T) {
	c, err := NewClient(Config{ClientID: "client", ClientSecret: "test-secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-API-KEY") != "" || req.Header.Get("Authorization") == "Basic" {
			t.Fatal("unexpected credentials")
		}
		body := ""
		status := 200
		switch req.URL.Path {
		case "/userinfo":
			if req.URL.Host != "oauth.resource.datagsm.kr" || req.Header.Get("Authorization") != "Bearer access" {
				t.Fatal("incorrect UserInfo authentication or host")
			}
			body = `{"sub":"teacher@gsm.hs.kr","objectType":"TEACHER","student":null,"teacher":{"id":1,"department":"GRADE","description":null},"clubs":[],"projects":[]}`
		case "/.well-known/openid-configuration":
			body = `{"issuer":"https://oauth.authorization.datagsm.kr","token_endpoint_auth_methods_supported":["none","client_secret_post"]}`
		case "/v1/oauth/jwks":
			body = `{"keys":[{"kty":"RSA","kid":"key","use":null,"alg":null,"n":"AA","e":"AQAB"}]}`
		case "/v1/oauth/logout":
			if req.Method != "POST" {
				t.Fatal("incorrect logout method")
			}
			status = 204
		default:
			t.Fatal("unexpected path")
		}
		if req.URL.Path != "/userinfo" && req.Header.Get("Authorization") != "" {
			t.Fatal("credentials leaked to public endpoint")
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	user, err := c.UserInfo(ctx, "access")
	if err != nil || user.Student != nil || user.Teacher.Department != "GRADE" || *user.ObjectType != "TEACHER" {
		t.Fatalf("%v", err)
	}
	discovery, err := c.Discovery(ctx)
	if err != nil || discovery.Issuer != DefaultAuthorizationURL || len(discovery.TokenEndpointAuthMethodsSupported) != 2 {
		t.Fatalf("%v", err)
	}
	keys, err := c.JWKS(ctx)
	if err != nil || len(keys.Keys) != 1 || keys.Keys[0].Use != "" || keys.Keys[0].Algorithm != "" {
		t.Fatalf("%v", err)
	}
	if err := c.Logout(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestOAuthInputAndProviderErrors(t *testing.T) {
	for _, config := range []Config{{}, {ClientID: "client", AuthMethod: AuthSecretBasic}, {ClientID: "client", ClientSecret: "test", AuthMethod: AuthNone}, {ClientID: "client", AuthMethod: "unknown"}} {
		if _, err := NewClient(config); err == nil {
			t.Fatal("accepted invalid config")
		}
	}
	c, err := NewClient(Config{ClientID: "client", HTTPClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":"invalid_grant","error_description":"expired"}`))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := c.Exchange(ctx, CodeExchange{}); err == nil {
		t.Fatal("accepted empty code")
	}
	if _, err := c.Refresh(ctx, ""); err == nil {
		t.Fatal("accepted empty refresh token")
	}
	if _, err := c.ClientCredentials(ctx); err == nil {
		t.Fatal("accepted public client credentials")
	}
	if _, err := c.UserInfo(ctx, "access\r\nInjected: value"); err == nil {
		t.Fatal("accepted invalid header")
	}
	_, err = c.Refresh(ctx, "expired")
	var failure *datagsm.APIError
	if !errors.As(err, &failure) || failure.StatusCode != 400 || failure.Code != "invalid_grant" {
		t.Fatalf("%v", err)
	}
}
