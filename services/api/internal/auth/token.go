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

// DefaultSigningSecretPath holds the HMAC key for signed photo URLs,
// independent of the bearer token so rotating one doesn't rotate the other.
const DefaultSigningSecretPath = "/var/lib/homie/signing-secret"

// Load returns the configured token, creating a persistent random token when
// none has been configured or saved yet.
func Load(path string) (string, error) {
	return loadOrCreateSecret("HOMIE_AUTH_TOKEN", path, DefaultTokenPath)
}

// LoadSigningSecret returns the configured photo-URL signing secret,
// creating a persistent random one when none has been configured or saved
// yet. It is independent of the bearer token: rotating the token must not
// silently invalidate every signed photo URL already handed out.
func LoadSigningSecret(path string) (string, error) {
	return loadOrCreateSecret("HOMIE_SIGNING_SECRET", path, DefaultSigningSecretPath)
}

func loadOrCreateSecret(envVar, path, defaultPath string) (string, error) {
	if token := strings.TrimSpace(os.Getenv(envVar)); token != "" {
		return token, nil
	}
	if path == "" {
		path = defaultPath
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
