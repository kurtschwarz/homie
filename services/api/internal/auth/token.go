package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const DefaultTokenPath = "/var/lib/homie/token"

// Load returns the configured token, creating a persistent random token when
// none has been configured or saved yet.
func Load(path string) (string, error) {
	if token := strings.TrimSpace(os.Getenv("HOMIE_AUTH_TOKEN")); token != "" {
		return token, nil
	}
	if path == "" {
		path = DefaultTokenPath
	}
	data, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", errors.New("stored auth token is empty")
		}
		return token, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(random[:])
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", err
	}
	return token, nil
}
