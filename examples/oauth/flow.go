package oauthexample

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/tmp-it-play/datagsm-sdk-go/oauth"
)

func Begin(client *oauth.Client) (*oauth.Authorization, error) {
	return client.NewAuthorization(oauth.ScopeOpenID, oauth.ScopeAccountRead, oauth.ScopeStudentRead)
}

// Complete receives a saved, single-use attempt from the user's server-side session.
func Complete(ctx context.Context, client *oauth.Client, config oauth.Config, attempt oauth.Authorization, callback url.Values) (*oauth.Token, *oauth.UserInfo, error) {
	code, err := attempt.Code(callback)
	if err != nil {
		return nil, nil, err
	}
	token, err := client.Exchange(ctx, oauth.CodeExchange{Code: code, CodeVerifier: attempt.CodeVerifier})
	if err != nil {
		return nil, nil, err
	}
	if token.IDToken == nil {
		return nil, nil, errors.New("missing ID token")
	}
	keys, err := client.JWKS(ctx)
	if err != nil {
		return nil, nil, err
	}
	issuer := config.AuthorizationBaseURL
	if issuer == "" {
		issuer = oauth.DefaultAuthorizationURL
	}
	claims, err := keys.VerifyIDToken(*token.IDToken, issuer, config.ClientID, attempt.Nonce, time.Now())
	if err != nil {
		return nil, nil, err
	}
	user, err := client.UserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, err
	}
	if claims.Subject != user.Subject {
		return nil, nil, errors.New("UserInfo subject does not match ID token")
	}
	return token, user, nil
}
