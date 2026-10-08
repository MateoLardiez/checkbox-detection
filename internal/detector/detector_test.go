package detector

import (
	"errors"
	"os"
	"testing"
)

func TestDetectInvalidImage(t *testing.T) {
	_, err := Detect([]byte("not an image"))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("got %v, want ErrInvalidImage", err)
	}
}

func TestDetectValidImage(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample-000.png")
	if err != nil {
		t.Fatal(err)
	}

	boxes, err := Detect(data)
	if err != nil {
		t.Fatal(err)
	}
	if boxes == nil || len(boxes) != 0 {
		t.Fatalf("got %v, want empty non-nil slice", boxes)
	}
}
