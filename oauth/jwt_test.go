package oauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func TestJWTValidation(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	key := JWK{KeyType: "RSA", KeyID: "test-key", N: base64.RawURLEncoding.EncodeToString(private.N.Bytes()), E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(private.E)).Bytes())}
	keys := JWKS{Keys: []JWK{key}}
	now := time.Unix(1800000000, 0)
	sign := func(header any, claims any) string {
		t.Helper()
		h, err := json.Marshal(header)
		if err != nil {
			t.Fatal(err)
		}
		p, err := json.Marshal(claims)
		if err != nil {
			t.Fatal(err)
		}
		message := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(p)
		digest := sha256.Sum256([]byte(message))
		sig, err := rsa.SignPKCS1v15(rand.Reader, private, crypto.SHA256, digest[:])
		if err != nil {
			t.Fatal(err)
		}
		return message + "." + base64.RawURLEncoding.EncodeToString(sig)
	}
	header := map[string]any{"alg": "RS256", "kid": "test-key"}
	id := IDTokenClaims{Issuer: DefaultAuthorizationURL, Subject: "user@gsm.hs.kr", Audience: []string{"client"}, IssuedAt: now.Unix() - 10, ExpiresAt: now.Unix() + 300, Nonce: "nonce"}
	valid := sign(header, id)
	if _, err := keys.VerifyIDToken(valid, DefaultAuthorizationURL, "client", "nonce", now); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		change func(*IDTokenClaims)
	}{
		{"issuer", func(c *IDTokenClaims) { c.Issuer = "https://other.example" }},
		{"audience", func(c *IDTokenClaims) { c.Audience = []string{"other"} }},
		{"authorized party", func(c *IDTokenClaims) { c.AuthorizedParty = "other" }},
		{"multiple audiences", func(c *IDTokenClaims) { c.Audience = []string{"client", "other"} }},
		{"nonce", func(c *IDTokenClaims) { c.Nonce = "other" }},
		{"expired", func(c *IDTokenClaims) { c.ExpiresAt = now.Unix() }},
		{"future", func(c *IDTokenClaims) { c.IssuedAt = now.Unix() + 1 }},
		{"not before", func(c *IDTokenClaims) { c.NotBefore = now.Unix() + 1 }},
		{"missing subject", func(c *IDTokenClaims) { c.Subject = "" }},
		{"missing expiry", func(c *IDTokenClaims) { c.ExpiresAt = 0 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			claims := id
			tt.change(&claims)
			if _, err := keys.VerifyIDToken(sign(header, claims), DefaultAuthorizationURL, "client", "nonce", now); err == nil {
				t.Fatal("accepted invalid claims")
			}
		})
	}
	for _, h := range []map[string]any{{"alg": "none", "kid": "test-key"}, {"alg": "HS256", "kid": "test-key"}, {"alg": "RS256", "kid": "unknown"}, {"alg": "RS256", "kid": "test-key", "crit": []string{"unknown"}}} {
		if _, err := keys.VerifyIDToken(sign(h, id), DefaultAuthorizationURL, "client", "nonce", now); err == nil {
			t.Fatal("accepted invalid header")
		}
	}
	for _, token := range []string{"", "a.b.c", valid[:len(valid)-4] + "AAAA"} {
		if _, err := keys.VerifyIDToken(token, DefaultAuthorizationURL, "client", "nonce", now); err == nil {
			t.Fatal("accepted invalid signature")
		}
	}
	for _, mutate := range []func(*JWK){func(k *JWK) { k.N = "!" }, func(k *JWK) { k.E = "Ag" }, func(k *JWK) { k.Use = "enc" }, func(k *JWK) { k.Algorithm = "HS256" }} {
		bad := key
		mutate(&bad)
		if _, err := (JWKS{Keys: []JWK{bad}}).VerifyIDToken(valid, DefaultAuthorizationURL, "client", "nonce", now); err == nil {
			t.Fatal("accepted invalid key")
		}
	}
	if _, err := (JWKS{Keys: []JWK{key, key}}).VerifyIDToken(valid, DefaultAuthorizationURL, "client", "nonce", now); err == nil {
		t.Fatal("accepted duplicate kid")
	}
	access := AccessTokenClaims{Subject: "client", ClientID: "client", Scopes: []string{"app:read"}, GrantType: "client_credentials", IssuedAt: now.Unix() - 10, ExpiresAt: now.Unix() + 300}
	accessToken := sign(header, access)
	if _, err := keys.VerifyAccessToken(accessToken, "client", []string{"app:read"}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := keys.VerifyAccessToken(accessToken, "other", nil, now); err == nil {
		t.Fatal("accepted wrong client")
	}
	if _, err := keys.VerifyAccessToken(accessToken, "client", []string{"app:write"}, now); err == nil {
		t.Fatal("accepted missing scope")
	}
	if _, err := keys.VerifyIDToken(accessToken, DefaultAuthorizationURL, "client", "nonce", now); err == nil {
		t.Fatal("accepted access token as ID token")
	}
	if _, err := keys.VerifyAccessToken(valid, "client", nil, now); err == nil {
		t.Fatal("accepted ID token as access token")
	}
}
