package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func postFile(t *testing.T, field string, content []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile(field, "upload")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/detect", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	DetectHandler(rec, req)
	return rec
}

func TestDetectHandlerBadRequests(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		content []byte
	}{
		{"missing file field", "other", []byte("data")},
		{"empty file", "file", nil},
		{"not an image", "file", []byte("not an image")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postFile(t, tt.field, tt.content)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}

			var resp errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
			if resp.Error == "" {
				t.Fatal("expected an error message")
			}
		})
	}
}

func TestDetectHandlerNotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/detect", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	DetectHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDetectHandlerValidImage(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample-000.png")
	if err != nil {
		t.Fatal(err)
	}

	rec := postFile(t, "file", data)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"boxes":[]}` {
		t.Fatalf("body = %s, want {\"boxes\":[]}", got)
	}
}
