package auth

import (
	"errors"
	"golang.org/x/oauth2"
	"os"
	"path"
	"testing"
)

func initFile() (string, func()) {
	filepath := path.Join(os.TempDir(), "load-credential-config-test.json")
	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString(`{"value":"credentials content}"`); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	return filepath, func() {
		if err = os.Remove(filepath); err != nil {
			panic(err)
		}
	}
}

type mockLoader struct {
	err  error
	conf *oauth2.Config
}

func (l *mockLoader) ConfigFromJSON(_ []byte, _ ...string) (*oauth2.Config, error) {
	return l.conf, l.err
}

func newMockLoader(err error, conf *oauth2.Config) *mockLoader {
	if conf == nil {
		conf = new(oauth2.Config)
	}
	return &mockLoader{err: err, conf: conf}
}

func TestLoadCredentialConfig(t *testing.T) {
	filename, remover := initFile()
	defer remover()

	type args struct {
		path   string
		loader LoaderConfigFromJSON
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "OK",
			args:    args{path: filename, loader: newMockLoader(nil, nil)},
			wantErr: false,
		},
		{
			name:    "Reading credentials error",
			args:    args{path: "abcsasd", loader: newMockLoader(nil, nil)},
			wantErr: true,
		},
		{
			name:    "ConfigFromJSON error",
			args:    args{path: "abcsasd", loader: newMockLoader(errors.New("fake error"), nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadCredentialFromFile(tt.args.path, tt.args.loader)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadCredentialConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
