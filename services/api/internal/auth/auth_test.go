package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadCreatesAndReusesToken(t *testing.T) {
	t.Setenv("HOMIE_AUTH_TOKEN", "")
	path := filepath.Join(t.TempDir(), "token")
	one, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 64 {
		t.Fatalf("token length=%d", len(one))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
	two, err := Load(path)
	if err != nil || one != two {
		t.Fatalf("second token=%q err=%v", two, err)
	}
}

func TestLoadHonorsEnvironment(t *testing.T) {
	t.Setenv("HOMIE_AUTH_TOKEN", "configured-token")
	token, err := Load(filepath.Join(t.TempDir(), "token"))
	if err != nil || token != "configured-token" {
		t.Fatalf("token=%q err=%v", token, err)
	}
}

func TestMiddlewareAndPhotoSignature(t *testing.T) {
	const token = "test-token"
	called := false
	handler := RequireBearer(token, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, tc := range []struct {
		header string
		status int
	}{{"", http.StatusUnauthorized}, {"Bearer bad", http.StatusUnauthorized}, {"Bearer " + token, http.StatusNoContent}} {
		called = false
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", tc.header)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != tc.status || called != (tc.status == http.StatusNoContent) {
			t.Fatalf("header=%q status=%d called=%v", tc.header, res.Code, called)
		}
	}
	signer := NewSigner(token)
	exp := time.Now().Add(time.Hour).Unix()
	sig := signer.Sign("family.jpg", exp)
	if !signer.Verify("family.jpg", exp, sig, time.Now()) || signer.Verify("other.jpg", exp, sig, time.Now()) || signer.Verify("family.jpg", exp, strings.Repeat("0", len(sig)), time.Now()) || signer.Verify("family.jpg", time.Now().Add(-time.Hour).Unix(), signer.Sign("family.jpg", time.Now().Add(-time.Hour).Unix()), time.Now()) {
		t.Fatal("signature verification did not enforce file, MAC, and expiry")
	}
}
