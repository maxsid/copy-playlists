package youtube

import (
	"context"
	"golang.org/x/oauth2"
)

type Config interface {
	ConfigCodeURLGenerator
	ConfigCodeExchanger
	ConfigTokenSourceGetter
}

type ConfigCodeExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type ConfigCodeURLGenerator interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
}

type ConfigTokenSourceGetter interface {
	TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
}
