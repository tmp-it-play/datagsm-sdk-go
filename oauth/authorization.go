package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	datagsm "github.com/tmp-it-play/datagsm-sdk-go"
)

type AuthorizationParams struct {
	State         string
	Nonce         string
	CodeChallenge string
	Scopes        []string
}

// Authorization holds one login attempt; store it in the user's server-side session.
type Authorization struct {
	URL          string
	State        string
	Nonce        string
	CodeVerifier string
}

func NewCodeVerifier() string {
	var value [32]byte
	_, _ = rand.Read(value[:])
	return base64.RawURLEncoding.EncodeToString(value[:])
}

func CodeChallenge(verifier string) (string, error) {
	if len(verifier) < 43 || len(verifier) > 128 {
		return "", errors.New("datagsm: code verifier must have 43-128 characters")
	}
	for _, c := range verifier {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-._~", c)) {
			return "", errors.New("datagsm: code verifier contains an invalid character")
		}
	}
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:]), nil
}

func (c *Client) AuthorizeURL(params AuthorizationParams) (string, error) {
	u, err := url.Parse(c.config.RedirectURI)
	if err != nil || u.Scheme == "" || u.Fragment != "" || u.User != nil {
		return "", errors.New("datagsm: valid redirect URI is required")
	}
	if params.State == "" {
		return "", errors.New("datagsm: state is required")
	}
	q := url.Values{"client_id": {c.config.ClientID}, "redirect_uri": {c.config.RedirectURI}, "response_type": {"code"}, "state": {params.State}}
	if params.Nonce != "" {
		q.Set("nonce", params.Nonce)
	}
	if len(params.Scopes) > 0 {
		q.Set("scope", strings.Join(params.Scopes, " "))
	}
	if params.CodeChallenge != "" {
		challenge, err := base64.RawURLEncoding.DecodeString(params.CodeChallenge)
		if err != nil || len(challenge) != sha256.Size {
			return "", errors.New("datagsm: invalid S256 code challenge")
		}
		q.Set("code_challenge", params.CodeChallenge)
		q.Set("code_challenge_method", "S256")
	}
	return c.authorization.BaseURL + "/v1/oauth/authorize?" + q.Encode(), nil
}

func (c *Client) NewAuthorization(scopes ...string) (*Authorization, error) {
	a := &Authorization{State: NewCodeVerifier(), Nonce: NewCodeVerifier(), CodeVerifier: NewCodeVerifier()}
	challenge, err := CodeChallenge(a.CodeVerifier)
	if err != nil {
		return nil, err
	}
	a.URL, err = c.AuthorizeURL(AuthorizationParams{State: a.State, Nonce: a.Nonce, CodeChallenge: challenge, Scopes: scopes})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// Code validates state before returning a code or a provider error. Consume the saved attempt once.
func (a Authorization) Code(callback url.Values) (string, error) {
	if a.State == "" || len(callback["state"]) != 1 || subtle.ConstantTimeCompare([]byte(a.State), []byte(callback.Get("state"))) != 1 {
		return "", errors.New("datagsm: OAuth state mismatch")
	}
	if len(callback["code"]) > 1 || len(callback["error"]) > 1 || len(callback["error_description"]) > 1 {
		return "", errors.New("datagsm: duplicate OAuth callback parameter")
	}
	if failure := callback.Get("error"); failure != "" {
		if callback.Get("code") != "" {
			return "", errors.New("datagsm: callback contains both code and error")
		}
		return "", &datagsm.APIError{Code: failure, Message: callback.Get("error_description")}
	}
	if callback.Get("code") == "" {
		return "", errors.New("datagsm: callback has no authorization code")
	}
	return callback.Get("code"), nil
}
