package oauth

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

func TestPKCE(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge, err := CodeChallenge(verifier)
	if err != nil || challenge != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("%s, %v", challenge, err)
	}
	for _, value := range []string{"", strings.Repeat("a", 42), strings.Repeat("a", 129), strings.Repeat("!", 43)} {
		if _, err := CodeChallenge(value); err == nil {
			t.Fatalf("accepted invalid verifier")
		}
	}
	first, second := NewCodeVerifier(), NewCodeVerifier()
	if len(first) != 43 || first == second {
		t.Fatal("invalid random verifier")
	}
	if _, err := CodeChallenge(first); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorizationAndCallback(t *testing.T) {
	c, err := NewClient(Config{ClientID: "client & id", RedirectURI: "https://example.com/callback?from=login"})
	if err != nil {
		t.Fatal(err)
	}
	a, err := c.NewAuthorization(ScopeOpenID, ScopeAccountRead)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(a.URL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	challenge, _ := CodeChallenge(a.CodeVerifier)
	if q.Get("client_id") != "client & id" || q.Get("redirect_uri") != c.config.RedirectURI || q.Get("state") != a.State || q.Get("nonce") != a.Nonce || q.Get("code_challenge") != challenge || q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid datagsm:account_read" {
		t.Fatal("incorrect authorization parameters")
	}
	if code, err := a.Code(url.Values{"state": {a.State}, "code": {"code"}}); err != nil || code != "code" {
		t.Fatalf("%s, %v", code, err)
	}
	for _, callback := range []url.Values{
		{"code": {"code"}}, {"state": {"wrong"}, "code": {"code"}}, {"state": {a.State}},
		{"state": {a.State, a.State}, "code": {"code"}}, {"state": {a.State}, "code": {"one", "two"}},
		{"state": {a.State}, "code": {"code"}, "error": {"access_denied"}},
	} {
		if _, err := a.Code(callback); err == nil {
			t.Fatal("accepted invalid callback")
		}
	}
	_, err = a.Code(url.Values{"state": {a.State}, "error": {"access_denied"}, "error_description": {"denied"}})
	var failure *datagsm.APIError
	if !errors.As(err, &failure) || failure.Code != "access_denied" {
		t.Fatalf("%v", err)
	}
}

func TestTokenAuthentication(t *testing.T) {
	for _, method := range []AuthMethod{AuthNone, AuthSecretPost, AuthSecretBasic} {
		t.Run(string(method), func(t *testing.T) {
			config := Config{ClientID: "client: id", AuthMethod: method}
			if method != AuthNone {
				config.ClientSecret = "secret&value"
			}
			c, err := NewClient(config)
			if err != nil {
				t.Fatal(err)
			}
			req, err := c.tokenRequest(context.Background(), url.Values{"grant_type": {"authorization_code"}, "code": {"a+b&c"}}, []string{ScopeOpenID, ScopeAccountRead})
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatal(err)
			}
			if req.Method != "POST" || req.URL.Path != "/v1/oauth/token" || req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || form.Get("code") != "a+b&c" || form.Get("client_id") != config.ClientID || form.Get("scope") != "openid datagsm:account_read" {
				t.Fatal("incorrect token request")
			}
			user, pass, basic := req.BasicAuth()
			if method == AuthSecretBasic {
				if !basic || user != url.QueryEscape(config.ClientID) || pass != url.QueryEscape(config.ClientSecret) || form.Has("client_secret") {
					t.Fatal("invalid Basic authentication")
				}
			} else if basic || (method == AuthSecretPost && form.Get("client_secret") != config.ClientSecret) || (method == AuthNone && form.Has("client_secret")) {
				t.Fatal("invalid body authentication")
			}
		})
	}
}
