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

// Verifies  a token is valid
func (a *Authenticator) VerifyIDToken(token string) (*oidc.IDToken, error) {
	split := strings.Split(token, " ")
	if len(split) != 2 {
		return nil, errors.New("invalid token format")
	}
	token = split[1]

	fmt.Println(token)

	oidcConfig := &oidc.Config{
		ClientID: a.config.ClientID,
	}

	return a.provider.Verifier(oidcConfig).Verify(a.ctx, token)
}

// Create a login URL with a new state
func (a *Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state)
}

// Exchange an authorization code for a token
func (a *Authenticator) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	token, err := a.config.Exchange(a.ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}
