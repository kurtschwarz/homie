// Package secrets encrypts application credentials at rest.
package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	DefaultKeyPath = "/var/lib/homie/secret-key"
	version        = byte(1)
)

// Box seals values using NaCl secretbox and one application-level key.
type Box struct{ key [32]byte }

// Load reads HOMIE_SECRET_KEY as base64, or loads/creates a key at path.
func Load(path string) (*Box, error) {
	if encoded := strings.TrimSpace(os.Getenv("HOMIE_SECRET_KEY")); encoded != "" {
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(key) != 32 {
			return nil, errors.New("HOMIE_SECRET_KEY must be base64 for exactly 32 bytes")
		}
		return New(key)
	}
	if path == "" {
		path = DefaultKeyPath
	}
	key, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate secret key: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("create secret key directory: %w", err)
		}
		if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(key)+"\n"), 0o600); err != nil {
			return nil, fmt.Errorf("write secret key: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("read secret key: %w", err)
	} else {
		key, err = base64.StdEncoding.DecodeString(strings.TrimSpace(string(key)))
		if err != nil {
			return nil, fmt.Errorf("decode secret key: %w", err)
		}
	}
	return New(key)
}

// New constructs a Box from exactly 32 bytes.
func New(key []byte) (*Box, error) {
	if len(key) != 32 {
		return nil, errors.New("secret key must be exactly 32 bytes")
	}
	b := &Box{}
	copy(b.key[:], key)
	return b, nil
}

// Seal encrypts plaintext with a fresh nonce. The result includes a format version and nonce.
func (b *Box) Seal(plaintext []byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("generate secret nonce: %w", err)
	}
	out := make([]byte, 1, 1+len(nonce)+len(plaintext)+secretbox.Overhead)
	out[0] = version
	out = append(out, nonce[:]...)
	out = secretbox.Seal(out, plaintext, &nonce, &b.key)
	return out, nil
}

// Open authenticates and decrypts a value produced by Seal.
func (b *Box) Open(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 1+24+secretbox.Overhead || ciphertext[0] != version {
		return nil, errors.New("invalid encrypted value")
	}
	var nonce [24]byte
	copy(nonce[:], ciphertext[1:25])
	plaintext, ok := secretbox.Open(nil, ciphertext[25:], &nonce, &b.key)
	if !ok {
		return nil, errors.New("decrypt encrypted value")
	}
	return plaintext, nil
}
