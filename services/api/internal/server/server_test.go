package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/accounts"
	"github.com/kurtschwarz/homie/services/api/internal/auth"
	"github.com/kurtschwarz/homie/services/api/internal/graph"
	"github.com/kurtschwarz/homie/services/api/internal/photos"
	"github.com/kurtschwarz/homie/services/api/internal/server"
	"github.com/kurtschwarz/homie/services/api/internal/store"
)

func TestGraphQLMultipartUploadTransport(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "photos.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	library, err := photos.NewLibrary(filepath.Join(t.TempDir(), "photos"))
	if err != nil {
		t.Fatal(err)
	}
	photosService := photos.New(photos.NewSQLiteRepository(db), library)

	accountsDB, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer accountsDB.Close()
	accountsService, err := accounts.New(context.Background(), accounts.NewSQLiteRepository(accountsDB), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	// uploadPhoto is USER-only: Photo Sync's own bearer token does not
	// satisfy it yet, so the test proves the upload with a session instead.
	session, _, err := accountsService.Setup(context.Background(), "test", accountsService.SetupCode(), "kurt", "password123", "test-agent")
	if err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	operations, err := writer.CreateFormField("operations")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(operations).Encode(map[string]any{
		"query": "mutation($input: PhotoUploadInput!, $file: Upload!) { uploadPhoto(input: $input, file: $file) { id url } }",
		"variables": map[string]any{
			"input": map[string]any{"assetID": "asset", "version": "1", "takenAt": 1, "width": 1, "height": 1},
			"file":  nil,
		},
	}); err != nil {
		t.Fatal(err)
	}

	fileMap, err := writer.CreateFormField("map")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(fileMap, `{"0":["variables.file"]}`); err != nil {
		t.Fatal(err)
	}
	filePartHeader := make(textproto.MIMEHeader)
	filePartHeader.Set("Content-Disposition", `form-data; name="0"; filename="photo.jpg"`)
	filePartHeader.Set("Content-Type", "image/jpeg")
	filePart, err := writer.CreatePart(filePartHeader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(filePart, "photo-bytes"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	handler := server.New(server.Config{Resolver: &graph.Resolver{Photos: photosService, Accounts: accountsService}})
	request := httptest.NewRequest(http.MethodPost, "/graphql", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session.Token})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d response=%s", response.Code, response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte("\"errors\"")) {
		t.Fatalf("multipart upload failed: %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("/photos/")) {
		t.Fatalf("multipart upload did not store a photo: %s", response.Body.String())
	}
}

func TestPhotosEndpoint(t *testing.T) {
	library, err := photos.NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := library.Save("photo.jpg", bytes.NewReader([]byte("jpeg-bytes"))); err != nil {
		t.Fatal(err)
	}
	handler := server.New(server.Config{Photos: library.FS()})

	request := httptest.NewRequest(http.MethodGet, "/photos/photo.jpg", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "jpeg-bytes" {
		t.Fatalf("hit: status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("cache-control=%q", response.Header().Get("Cache-Control"))
	}

	for _, path := range []string{"/photos/missing.jpg", "/photos/photo.jpg.tmp", "/photos/sub/photo.jpg"} {
		request = httptest.NewRequest(http.MethodGet, path, nil)
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("path=%q status=%d, want 404", path, response.Code)
		}
	}
}

// The manager is mounted under a prefix, so its asset paths are the ones
// at risk of being swallowed by the SPA index.html fallback.
func TestManagerAssetsAreServedNotFallenBackTo(t *testing.T) {
	manager := fstest.MapFS{
		"index.html":             {Data: []byte("<html>manager</html>")},
		"assets/index-abc123.js": {Data: []byte("console.log('manager')")},
	}
	// Manager is only mounted on the authenticated listener.
	handler := server.NewLAN(server.Config{Resolver: &graph.Resolver{}, Manager: manager})

	request := httptest.NewRequest(http.MethodGet, "/manage/assets/index-abc123.js", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", response.Code)
	}
	if response.Body.String() != "console.log('manager')" {
		t.Fatalf("body=%q, want the asset and not index.html", response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/javascript") {
		t.Fatalf("content-type=%q, want text/javascript", contentType)
	}

	// Client-side routes still need the index.html fallback.
	for _, path := range []string{"/manage/", "/manage/settings"} {
		request = httptest.NewRequest(http.MethodGet, path, nil)
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.String() != "<html>manager</html>" {
			t.Fatalf("path=%q status=%d body=%q", path, response.Code, response.Body.String())
		}
	}
}

func TestPhotosEndpointWithoutLibrary(t *testing.T) {
	handler := server.New(server.Config{})
	request := httptest.NewRequest(http.MethodGet, "/photos/photo.jpg", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404", response.Code)
	}
}

func TestSingleListenerRequiresAuthentication(t *testing.T) {
	const token = "dev-token"
	handler := server.New(server.Config{Resolver: &graph.Resolver{}, Token: token})
	for _, tc := range []struct {
		name          string
		authority     string
		wantAuthError bool
	}{
		{name: "missing token", wantAuthError: true},
		{name: "valid token", authority: "Bearer " + token},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ timer { title } }"}`))
			request.Header.Set("Content-Type", "application/json")
			if tc.authority != "" {
				request.Header.Set("Authorization", tc.authority)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			gotAuthError := strings.Contains(response.Body.String(), "unauthorized")
			if gotAuthError != tc.wantAuthError {
				t.Fatalf("body=%s, wantAuthError=%v", response.Body.String(), tc.wantAuthError)
			}
		})
	}
}

func TestLANListenerProtectsGraphQLAndServesManagerAndSignedPhotos(t *testing.T) {
	const token = "test-token"
	library, err := photos.NewLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := library.Save("photo.jpg", strings.NewReader("jpeg-bytes")); err != nil {
		t.Fatal(err)
	}
	signer := auth.NewSigner(token)
	handler := server.NewLAN(server.Config{
		Resolver: &graph.Resolver{}, Token: token, Signer: signer,
		Photos: library.FS(), Manager: fstest.MapFS{"index.html": {Data: []byte("manager")}},
	})
	// The gate is a GraphQL-level concern, not a transport-level one: a
	// rejected operation comes back as a normal GraphQL error, the same
	// convention as every other resolver failure.
	for _, tc := range []struct {
		name          string
		header        string
		wantAuthError bool
	}{{"missing", "", true}, {"wrong", "Bearer wrong", true}, {"correct", "Bearer " + token, false}} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ timer { title } }"}`))
			req.Header.Set("Content-Type", "application/json")
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != http.StatusOK {
				t.Fatalf("status=%d, want 200", res.Code)
			}
			gotAuthError := strings.Contains(res.Body.String(), "unauthorized")
			if gotAuthError != tc.wantAuthError {
				t.Fatalf("body=%s, wantAuthError=%v", res.Body.String(), tc.wantAuthError)
			}
		})
	}
	manager := httptest.NewRecorder()
	handler.ServeHTTP(manager, httptest.NewRequest(http.MethodGet, "/manage/", nil))
	if manager.Code != http.StatusOK || manager.Body.String() != "manager" {
		t.Fatalf("manager status=%d body=%q", manager.Code, manager.Body.String())
	}
	for _, path := range []string{"/", "/healthz", "/playground"} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != http.StatusNotFound {
			t.Fatalf("path=%q status=%d", path, res.Code)
		}
	}
	expiry := time.Now().Add(time.Hour).Unix()
	signed := fmt.Sprintf("/photos/photo.jpg?exp=%d&sig=%s", expiry, signer.Sign("photo.jpg", expiry))
	photo := httptest.NewRecorder()
	handler.ServeHTTP(photo, httptest.NewRequest(http.MethodGet, signed, nil))
	if photo.Code != http.StatusOK || photo.Body.String() != "jpeg-bytes" {
		t.Fatalf("signed photo status=%d body=%q", photo.Code, photo.Body.String())
	}
	for _, path := range []string{fmt.Sprintf("/photos/photo.jpg?exp=%d&sig=bad", expiry), fmt.Sprintf("/photos/photo.jpg?exp=%d&sig=%s", time.Now().Add(-time.Hour).Unix(), signer.Sign("photo.jpg", time.Now().Add(-time.Hour).Unix()))} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != http.StatusNotFound {
			t.Fatalf("invalid signed photo status=%d", res.Code)
		}
	}
}

func TestSetupLoginSessionCookieAndLogout(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	accountsService, err := accounts.New(context.Background(), accounts.NewSQLiteRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	handler := server.NewLAN(server.Config{Resolver: &graph.Resolver{Accounts: accountsService}})

	post := func(query string, cookies []*http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":`+strconv.Quote(query)+`}`))
		req.Header.Set("Content-Type", "application/json")
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res
	}

	// An unauthenticated people query is rejected...
	if res := post(`{ people { id } }`, nil); !strings.Contains(res.Body.String(), "unauthorized") {
		t.Fatalf("unauthenticated people query: body=%s", res.Body.String())
	}
	// ...while session and health are public before login. Health reports
	// first-run setup status.
	statusRes := post(`{ session { user { username } } health { status } }`, nil)
	if !strings.Contains(statusRes.Body.String(), `"session":null`) || !strings.Contains(statusRes.Body.String(), `"status":"SETUP_REQUIRED"`) {
		t.Fatalf("health before setup: body=%s", statusRes.Body.String())
	}

	if res := post(`mutation { setup(code: "wrong-code", username: "kurt", password: "password123") { id } }`, nil); !strings.Contains(res.Body.String(), "invalid setup code") {
		t.Fatalf("wrong setup code: body=%s", res.Body.String())
	}

	code := accountsService.SetupCode()
	setupRes := post(`mutation { setup(code: "`+code+`", username: "kurt", password: "password123") { username role } }`, nil)
	cookies := setupRes.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || !cookies[0].HttpOnly {
		t.Fatalf("setup should set an HttpOnly session cookie: cookies=%+v", cookies)
	}
	if !strings.Contains(setupRes.Body.String(), `"username":"kurt"`) {
		t.Fatalf("setup response: body=%s", setupRes.Body.String())
	}

	// The cookie now authenticates a protected query.
	if res := post(`{ people { id } }`, cookies); strings.Contains(res.Body.String(), "unauthorized") {
		t.Fatalf("session-authenticated people query: body=%s", res.Body.String())
	}

	// session is itself public, so the gate never resolves a current user
	// into its context - it has to check the cookie on its own.
	if res := post(`{ session { user { username role } } health { status } }`, cookies); !strings.Contains(res.Body.String(), `"username":"kurt"`) || !strings.Contains(res.Body.String(), `"role":"ADMIN"`) || !strings.Contains(res.Body.String(), `"status":"OK"`) {
		t.Fatalf("session query should resolve the logged-in user: body=%s", res.Body.String())
	}

	logoutRes := post(`mutation { logout }`, cookies)
	cleared := logoutRes.Result().Cookies()
	if len(cleared) != 1 || cleared[0].MaxAge >= 0 {
		t.Fatalf("logout should clear the session cookie: cookies=%+v", cleared)
	}
	if res := post(`{ people { id } }`, cookies); !strings.Contains(res.Body.String(), "unauthorized") {
		t.Fatalf("people query after logout: body=%s", res.Body.String())
	}
}

// TestSetupCodeReachesTheKioskAndNobodyElse pins the property the display's
// pairing screen rests on. Whoever reads the one-time code can claim the
// admin account, so it is served on the loopback listener the kiosk uses and
// nowhere else without a credential.
func TestSetupCodeReachesTheKioskAndNobodyElse(t *testing.T) {
	const token = "test-token"
	db, err := store.Open(filepath.Join(t.TempDir(), "accounts.db"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	accountsService, err := accounts.New(context.Background(), accounts.NewSQLiteRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	code := accountsService.SetupCode()
	if code == "" {
		t.Fatal("expected a setup code before any user exists")
	}
	cfg := server.Config{Resolver: &graph.Resolver{Accounts: accountsService}, Token: token}

	post := func(handler http.Handler, authorization string) string {
		req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ setupCode }"}`))
		req.Header.Set("Content-Type", "application/json")
		if authorization != "" {
			req.Header.Set("Authorization", authorization)
		}
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res.Body.String()
	}

	// The LAN listener refuses it whatever the caller presents. The bearer
	// token is the case that matters: it proves the display, and the
	// display itself has no business learning the code over the network -
	// only by being read off the loopback listener's screen.
	lan := server.NewLAN(cfg)
	for _, authorization := range []string{"", "Bearer wrong", "Bearer " + token} {
		if body := post(lan, authorization); strings.Contains(body, code) || !strings.Contains(body, "unauthorized") {
			t.Fatalf("setup code readable from the LAN with authorization %q: body=%s", authorization, body)
		}
	}
	// The kiosk's loopback-only listener has no gate at all: this is how the
	// code reaches the screen someone is standing in front of.
	if body := post(server.NewLocal(cfg), ""); !strings.Contains(body, code) {
		t.Fatalf("the loopback listener should serve the setup code to the kiosk: body=%s", body)
	}
	// A single listener serving both the display and the network still needs
	// the credential, but does serve the field: that is the dev topology,
	// where one listener stands in for both.
	single := server.New(cfg)
	if body := post(single, ""); strings.Contains(body, code) {
		t.Fatalf("single listener served the setup code with no credential: body=%s", body)
	}
	if body := post(single, "Bearer "+token); !strings.Contains(body, code) {
		t.Fatalf("single listener should serve the setup code to a credentialed display: body=%s", body)
	}

	// A signed-in admin is not the display, and must not learn the code
	// either, even though they share the single dev listener with it.
	session, _, err := accountsService.Setup(context.Background(), "test", code, "kurt", "password123", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ setupCode }"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session.Token})
	res := httptest.NewRecorder()
	single.ServeHTTP(res, req)
	if body := res.Body.String(); strings.Contains(body, code) || !strings.Contains(body, "unauthorized") {
		t.Fatalf("setup code readable by a signed-in admin: body=%s", body)
	}
}

func TestLocalListenerServesWithoutCredentials(t *testing.T) {
	handler := server.NewLocal(server.Config{Display: fstest.MapFS{"index.html": {Data: []byte("display")}}})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK || res.Body.String() != "display" {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
}
