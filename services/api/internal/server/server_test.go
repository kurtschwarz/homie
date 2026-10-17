package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kurtschwarz/homie/services/api/internal/server"
)

func TestGraphQLMultipartUploadTransport(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	operations, err := writer.CreateFormField("operations")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(operations).Encode(map[string]any{
		"query": "mutation($input: PhotoUploadInput!, $file: Upload!) { uploadPhoto(input: $input, file: $file) { id } }",
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
	filePart, err := writer.CreateFormFile("0", "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(filePart, "photo-bytes"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	handler := server.New(server.Config{})
	request := httptest.NewRequest(http.MethodPost, "/graphql", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d response=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("photo uploads are not implemented")) {
		t.Fatalf("multipart upload did not reach resolver: %s", response.Body.String())
	}
}
