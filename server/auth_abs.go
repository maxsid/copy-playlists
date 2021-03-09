package server

import (
	"context"
	"golang.org/x/oauth2"
)

// AuthConfigurator is a oauth2.Config abstraction.
type AuthConfigurator interface {
	authCodeURLGenerator
	authTokenSourceGetter
	authExchanger
}

type authExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type authCodeURLGenerator interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
}

type authTokenSourceGetter interface {
	TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
}
