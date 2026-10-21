package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

type Signer struct{ key []byte }

func NewSigner(token string) *Signer {
	key := sha256.Sum256([]byte(token + "homie-photo-url-v1"))
	return &Signer{key: key[:]}
}

// Expiry returns the end of the current 24-hour URL validity window. The
// boundary keeps URLs stable for a given photo throughout the day.
func Expiry(now time.Time) int64 {
	return (now.Unix()/86400 + 1) * 86400
}

func (s *Signer) Sign(file string, expiry int64) string {
	mac := hmac.New(sha256.New, s.key)
	_, _ = fmt.Fprintf(mac, "%s\n%d", file, expiry)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Signer) Verify(file string, expiry int64, signature string, now time.Time) bool {
	if s == nil || expiry < now.Unix() {
		return false
	}
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	expected, _ := hex.DecodeString(s.Sign(file, expiry))
	return subtle.ConstantTimeCompare(provided, expected) == 1
}

func ParseExpiry(value string) (int64, error) { return strconv.ParseInt(value, 10, 64) }
