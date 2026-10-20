package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

	handler := server.New(server.Config{Resolver: &graph.Resolver{Photos: photosService}})
	request := httptest.NewRequest(http.MethodPost, "/graphql", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
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
	handler := server.New(server.Config{Manager: manager})

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
	for _, tc := range []struct {
		name   string
		header string
		want   int
	}{{"missing", "", http.StatusUnauthorized}, {"wrong", "Bearer wrong", http.StatusUnauthorized}, {"correct", "Bearer " + token, http.StatusOK}} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ health { status } }"}`))
			req.Header.Set("Content-Type", "application/json")
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("status=%d, want %d body=%s", res.Code, tc.want, res.Body.String())
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

func TestLocalListenerServesWithoutCredentials(t *testing.T) {
	handler := server.NewLocal(server.Config{Display: fstest.MapFS{"index.html": {Data: []byte("display")}}})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK || res.Body.String() != "display" {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
}
