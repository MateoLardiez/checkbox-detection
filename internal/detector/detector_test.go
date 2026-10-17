package detector

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg" // registers the JPEG decoder for image.DecodeConfig
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDetectInvalidImage(t *testing.T) {
	_, err := Detect([]byte("not an image"))
	if !errors.Is(err, ErrInvalidImage) {
		t.Fatalf("got %v, want ErrInvalidImage", err)
	}
}

func TestDetectBlankImage(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 200, 200))
	for i := range gray.Pix {
		gray.Pix[i] = 255
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, gray); err != nil {
		t.Fatal(err)
	}

	boxes, err := Detect(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if boxes == nil || len(boxes) != 0 {
		t.Fatalf("got %v, want empty non-nil slice", boxes)
	}
}

// testImages returns every PNG and JPEG in testdata.
func testImages(t *testing.T) []string {
	t.Helper()

	var paths []string
	for _, ext := range []string{"png", "jpg", "jpeg"} {
		found, err := filepath.Glob("../../testdata/*." + ext)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, found...)
	}
	if len(paths) == 0 {
		t.Fatal("no images found in testdata")
	}
	return paths
}

func TestDetectSamples(t *testing.T) {
	for _, path := range testImages(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}

			boxes, err := Detect(data)
			if err != nil {
				t.Fatal(err)
			}
			if len(boxes) == 0 {
				t.Fatal("expected at least one checkbox")
			}
			for _, b := range boxes {
				x1, y1, x2, y2 := b.BBox[0], b.BBox[1], b.BBox[2], b.BBox[3]
				if x1 < 0 || y1 < 0 || x2 > cfg.Width || y2 > cfg.Height || x1 >= x2 || y1 >= y2 {
					t.Errorf("bbox %v out of bounds for a %dx%d image", b.BBox, cfg.Width, cfg.Height)
				}
			}
			if !slices.IsSortedFunc(boxes, compareBoxes) {
				t.Error("boxes are not sorted")
			}

			again, err := Detect(data)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(boxes, again) {
				t.Error("two runs on the same image gave different results")
			}
		})
	}
}
