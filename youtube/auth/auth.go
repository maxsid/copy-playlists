package auth

import (
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"os"
)

var scopes = []string{youtube.YoutubeForceSslScope}

// LoadCredentialFromFile reads a JSON credential file by the path and returns oauth2.Config.
func LoadCredentialFromFile(path string) (*oauth2.Config, error) {
	return loadCredentialFromFile(path, &jsonConfigLoader{})
}

// loadCredentialFromFile reads a credential client secret from a file by the path via LoaderConfigFromJSON.
func loadCredentialFromFile(path string, loader LoaderConfigFromJSON) (*oauth2.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	secretBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return readCredential(secretBytes, loader)
}

func readCredential(credential []byte, loader LoaderConfigFromJSON) (*oauth2.Config, error) {
	return loader.ConfigFromJSON(credential, scopes...)
}
