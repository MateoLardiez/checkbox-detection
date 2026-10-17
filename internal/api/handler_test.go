package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
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

	var resp detectResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Boxes) == 0 {
		t.Fatal("expected at least one box")
	}
}

func TestDetectHandlerFormats(t *testing.T) {
	pngData, err := os.ReadFile("../../testdata/sample-000.png")
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatal(err)
	}
	var jpgBuf bytes.Buffer
	if err := jpeg.Encode(&jpgBuf, img, nil); err != nil {
		t.Fatal(err)
	}
	jpgData := jpgBuf.Bytes()
	pdfData := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF\n")

	tests := []struct {
		name    string
		content []byte
		want    int
	}{
		{"png", pngData, http.StatusOK},
		{"jpeg", jpgData, http.StatusOK},
		{"pdf", pdfData, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := postFile(t, "file", tt.content); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestDetectHandlerImageWithoutCheckboxes(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 200, 200))
	for i := range gray.Pix {
		gray.Pix[i] = 255
	}
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, gray); err != nil {
		t.Fatal(err)
	}

	rec := postFile(t, "file", pngData.Bytes())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"boxes":[]}` {
		t.Fatalf("body = %s, want {\"boxes\":[]}", got)
	}
}
