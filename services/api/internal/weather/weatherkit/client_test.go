package weatherkit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCurrentWeatherDecodesPayloadAndSendsToken(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("dataSets"); got != "currentWeather" {
			t.Errorf("unexpected dataSets: %q", got)
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == r.Header.Get("Authorization") || !validToken(token) {
			t.Errorf("invalid bearer token")
		}
		_, _ = w.Write([]byte(`{"currentWeather":{"conditionCode":"Clear","temperature":12.4,"daylight":true},"metadata":{"expireTime":"2030-01-01T00:00:00Z"}}`))
	}))
	defer server.Close()

	client, err := New(Config{
		TeamID: "TEAMID", ServiceID: "com.example.weather", KeyID: "KEYID", PrivateKey: privateKey,
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Point the fixed WeatherKit endpoint at the local test server via a custom
	// transport which rewrites the host while preserving the request path.
	client.http = &http.Client{Transport: rewriteTransport{base: server.URL}}
	got, err := client.CurrentWeather(context.Background(), 43.4422469, -80.4699426)
	if err != nil {
		t.Fatal(err)
	}
	if got.ConditionCode != "Clear" || got.Temperature != 12.4 || got.Daylight == nil || !*got.Daylight {
		t.Fatalf("unexpected conditions: %+v", got)
	}
	if !got.Expiration.Equal(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected expiration: %s", got.Expiration)
	}
}

func validToken(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || !strings.Contains(string(header), `"alg":"ES256"`) {
		return false
	}
	claims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !strings.Contains(string(claims), `"sub":"com.example.weather"`) {
		return false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	return err == nil && len(signature) == 64
}

type rewriteTransport struct{ base string }

func (r rewriteTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = strings.TrimPrefix(r.base, "http://")
	return http.DefaultTransport.RoundTrip(clone)
}
