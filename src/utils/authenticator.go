package utils

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authenticator is used to authenticate our users.
type Authenticator struct {
	provider *oidc.Provider
	config   oauth2.Config
	ctx      context.Context
}

// New instantiates the *Authenticator.
func NewAuthenticator(ctx context.Context, auth0Domain string, auth0CallbackUrl string, auth0ClientId string, auth0ClientSecret string) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		ctx,
		"https://"+auth0Domain+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     auth0ClientId,
		ClientSecret: auth0ClientSecret,
		RedirectURL:  auth0CallbackUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		provider: provider,
		config:   conf,
		ctx:      ctx,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (authenticator *Authenticator) VerifyIDToken(token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)

	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: authenticator.config.ClientID,
	}

	return authenticator.provider.Verifier(oidcConfig).Verify(authenticator.ctx, rawIDToken)
}

// Create a login URL with a new state
func (authenticator *Authenticator) GetAuthURL(state string) string {
	return authenticator.config.AuthCodeURL(state)
}

// Exchange an authorization code for a token
func (authenticator *Authenticator) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	token, err := authenticator.config.Exchange(authenticator.ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
