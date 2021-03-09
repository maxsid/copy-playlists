package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type LoaderConfigFromJSON interface {
	ConfigFromJSON(jsonKey []byte, scope ...string) (*oauth2.Config, error)
}

type jsonConfigLoader struct{}

func (j jsonConfigLoader) ConfigFromJSON(jsonKey []byte, scope ...string) (*oauth2.Config, error) {
	return google.ConfigFromJSON(jsonKey, scope...)
}
