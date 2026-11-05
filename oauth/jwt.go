package oauth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"slices"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("datagsm: invalid or expired JWT")

type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	N         string `json:"n"`
	E         string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type IDTokenClaims struct {
	Issuer          string   `json:"iss"`
	Subject         string   `json:"sub"`
	Audience        []string `json:"aud"`
	AuthorizedParty string   `json:"azp"`
	ExpiresAt       int64    `json:"exp"`
	IssuedAt        int64    `json:"iat"`
	NotBefore       int64    `json:"nbf"`
	Nonce           string   `json:"nonce"`
	Email           string   `json:"email"`
}

type AccessTokenClaims struct {
	Subject   string   `json:"sub"`
	ClientID  string   `json:"clientId"`
	Scopes    []string `json:"scopes"`
	Role      string   `json:"role"`
	GrantType string   `json:"grant_type"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	NotBefore int64    `json:"nbf"`
}

// VerifyIDToken requires the issuer, client ID, and nonce from the saved authorization attempt.
func (keys JWKS) VerifyIDToken(token, issuer, clientID, nonce string, now time.Time) (*IDTokenClaims, error) {
	if issuer == "" || clientID == "" || nonce == "" {
		return nil, ErrInvalidToken
	}
	var claims IDTokenClaims
	if err := keys.verify(token, &claims); err != nil {
		return nil, err
	}
	if claims.Issuer != issuer || claims.Subject == "" || !slices.Contains(claims.Audience, clientID) ||
		(claims.AuthorizedParty != "" || len(claims.Audience) > 1) && claims.AuthorizedParty != clientID ||
		subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(nonce)) != 1 ||
		!validTimes(claims.IssuedAt, claims.ExpiresAt, claims.NotBefore, now) {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

// VerifyAccessToken uses clientId because DataGSM access tokens do not contain iss or aud.
func (keys JWKS) VerifyAccessToken(token, clientID string, requiredScopes []string, now time.Time) (*AccessTokenClaims, error) {
	if clientID == "" {
		return nil, ErrInvalidToken
	}
	var claims AccessTokenClaims
	if err := keys.verify(token, &claims); err != nil {
		return nil, err
	}
	if claims.ClientID != clientID || claims.Subject == "" || !validTimes(claims.IssuedAt, claims.ExpiresAt, claims.NotBefore, now) {
		return nil, ErrInvalidToken
	}
	for _, scope := range requiredScopes {
		if !slices.Contains(claims.Scopes, scope) {
			return nil, ErrInvalidToken
		}
	}
	return &claims, nil
}

func validTimes(issuedAt, expiresAt, notBefore int64, now time.Time) bool {
	return !now.IsZero() && issuedAt > 0 && issuedAt <= now.Unix() && expiresAt > now.Unix() && expiresAt > issuedAt && notBefore <= now.Unix()
}

func (keys JWKS) verify(token string, claims any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ErrInvalidToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalidToken
	}
	var header struct {
		Algorithm string   `json:"alg"`
		KeyID     string   `json:"kid"`
		Critical  []string `json:"crit"`
	}
	if json.Unmarshal(headerBytes, &header) != nil || header.Algorithm != "RS256" || header.KeyID == "" || len(header.Critical) > 0 {
		return ErrInvalidToken
	}
	var key *JWK
	for i := range keys.Keys {
		if keys.Keys[i].KeyID == header.KeyID {
			if key != nil {
				return ErrInvalidToken
			}
			key = &keys.Keys[i]
		}
	}
	if key == nil || key.KeyType != "RSA" || key.Use != "" && key.Use != "sig" || key.Algorithm != "" && key.Algorithm != "RS256" {
		return ErrInvalidToken
	}
	n, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil || len(n) > 1024 {
		return ErrInvalidToken
	}
	e, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil || len(e) == 0 || len(e) > 4 {
		return ErrInvalidToken
	}
	modulus := new(big.Int).SetBytes(n)
	exponent := new(big.Int).SetBytes(e).Int64()
	if modulus.BitLen() < 2048 || exponent < 3 || exponent > 1<<31-1 || exponent%2 == 0 {
		return ErrInvalidToken
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrInvalidToken
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(&rsa.PublicKey{N: modulus, E: int(exponent)}, crypto.SHA256, digest[:], signature) != nil {
		return ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(payload, claims) != nil {
		return ErrInvalidToken
	}
	return nil
}
