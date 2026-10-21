package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"strings"
)

// crockford is Crockford's base32 alphabet: no I, O, 0 or 1, so a code
// survives being read off a terminal and typed on a phone without
// ambiguity.
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// GenerateSetupCode returns an 8-character code formatted as "XXXX-XXXX".
func GenerateSetupCode() (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	var b strings.Builder
	for i, v := range raw {
		if i == 4 {
			b.WriteByte('-')
		}
		b.WriteByte(crockford[int(v)%len(crockford)])
	}
	return b.String(), nil
}

// EqualSetupCode compares a submitted code against the generated one in
// constant time, ignoring surrounding whitespace and case.
func EqualSetupCode(code, submitted string) bool {
	a := strings.ToUpper(strings.TrimSpace(code))
	b := strings.ToUpper(strings.TrimSpace(submitted))
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
