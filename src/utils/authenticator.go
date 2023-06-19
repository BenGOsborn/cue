package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		fmt.Sprint("https://", auth0Domain, "/"),
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

// Create a login URL with a new state
func (a *Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state)
}

// Verify an id token is valid
func (a *Authenticator) VerifyToken(token string) (*oidc.IDToken, error) {
	token = strings.TrimPrefix(token, "Bearer ")

	oidcConfig := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	return a.provider.Verifier(oidcConfig).Verify(a.ctx, token)
}

// Exchange an authorization code for a token
func (a *Authenticator) ExchangeCodeForToken(code string) (string, *oidc.IDToken, error) {
	token, err := a.config.Exchange(a.ctx, code)
	if err != nil {
		return "", nil, err
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	idToken, err := a.provider.Verifier(oidcConfig).Verify(a.ctx, rawIdToken)
	if err != nil {
		return "", nil, err
	}

	return rawIdToken, idToken, nil
}
