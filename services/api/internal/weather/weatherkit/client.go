package weatherkit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiBaseURL = "https://weatherkit.apple.com/api/v1/weather"

// Config contains the credentials required to request WeatherKit data.
type Config struct {
	TeamID     string
	ServiceID  string
	KeyID      string
	PrivateKey []byte
	HTTPClient *http.Client
}

// Client requests current conditions from Apple's WeatherKit REST API.
type Client struct {
	teamID    string
	serviceID string
	keyID     string
	key       *ecdsa.PrivateKey
	http      *http.Client

	tokenMu     sync.Mutex
	token       string
	tokenExpiry time.Time
}

// CurrentConditions contains the fields used to power the home-screen glance.
type CurrentConditions struct {
	ConditionCode string    `json:"conditionCode"`
	Temperature   float64   `json:"temperature"`
	Daylight      *bool     `json:"daylight"`
	Expiration    time.Time `json:"-"`
}

type weatherResponse struct {
	CurrentWeather struct {
		ConditionCode string  `json:"conditionCode"`
		Temperature   float64 `json:"temperature"`
		Daylight      bool    `json:"daylight"`
	} `json:"currentWeather"`
	Metadata struct {
		ExpireTime time.Time `json:"expireTime"`
	} `json:"metadata"`
}

// New parses the WeatherKit key once and prepares a bounded-time HTTP client.
func New(config Config) (*Client, error) {
	if config.TeamID == "" || config.ServiceID == "" || config.KeyID == "" {
		return nil, errors.New("weatherkit credentials are incomplete")
	}
	block, _ := pem.Decode(config.PrivateKey)
	if block == nil {
		return nil, errors.New("decode WeatherKit private key: no PEM block found")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse WeatherKit private key: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, errors.New("WeatherKit private key must be an ECDSA P-256 key")
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{teamID: config.TeamID, serviceID: config.ServiceID, keyID: config.KeyID, key: key, http: client}, nil
}

// CurrentWeather fetches conditions for the given coordinates and language.
func (c *Client) CurrentWeather(ctx context.Context, latitude, longitude float64) (CurrentConditions, error) {
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 || math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
		return CurrentConditions{}, errors.New("WeatherKit coordinates are outside valid ranges")
	}
	token, err := c.developerToken(time.Now())
	if err != nil {
		return CurrentConditions{}, err
	}
	endpoint := fmt.Sprintf("%s/en/%s/%s", apiBaseURL, coordinate(latitude), coordinate(longitude))
	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return CurrentConditions{}, fmt.Errorf("build WeatherKit URL: %w", err)
	}
	query := requestURL.Query()
	query.Set("dataSets", "currentWeather")
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return CurrentConditions{}, fmt.Errorf("create WeatherKit request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.http.Do(request)
	if err != nil {
		return CurrentConditions{}, fmt.Errorf("request WeatherKit current weather: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return CurrentConditions{}, fmt.Errorf("WeatherKit returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload weatherResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return CurrentConditions{}, fmt.Errorf("decode WeatherKit response: %w", err)
	}
	if payload.CurrentWeather.ConditionCode == "" {
		return CurrentConditions{}, errors.New("WeatherKit response is missing current weather")
	}
	daylight := payload.CurrentWeather.Daylight
	return CurrentConditions{
		ConditionCode: payload.CurrentWeather.ConditionCode,
		Temperature:   payload.CurrentWeather.Temperature,
		Daylight:      &daylight,
		Expiration:    payload.Metadata.ExpireTime,
	}, nil
}

func (c *Client) developerToken(now time.Time) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && now.Before(c.tokenExpiry.Add(-time.Minute)) {
		return c.token, nil
	}
	header := map[string]string{
		"alg": "ES256",
		"kid": c.keyID,
		"id":  c.teamID + "." + c.serviceID,
	}
	claims := map[string]any{
		"iss": c.teamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
		"sub": c.serviceID,
	}
	encodedHeader, err := encodeSegment(header)
	if err != nil {
		return "", fmt.Errorf("encode WeatherKit JWT header: %w", err)
	}
	encodedClaims, err := encodeSegment(claims)
	if err != nil {
		return "", fmt.Errorf("encode WeatherKit JWT claims: %w", err)
	}
	message := encodedHeader + "." + encodedClaims
	hash := sha256.Sum256([]byte(message))
	r, s, err := ecdsa.Sign(rand.Reader, c.key, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign WeatherKit JWT: %w", err)
	}
	// JWS ES256 signatures use fixed-width R||S, not ASN.1 DER.
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	c.token = message + "." + base64.RawURLEncoding.EncodeToString(signature)
	c.tokenExpiry = now.Add(time.Hour)
	return c.token, nil
}

func encodeSegment(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func coordinate(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
