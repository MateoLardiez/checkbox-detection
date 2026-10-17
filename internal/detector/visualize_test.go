package detector

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"gocv.io/x/gocv"
)

// TestVisualizeDetections draws what Detect returns over each sample image
// (red = checked, green = unchecked) and saves it to VISUALIZE_DIR.
// It is a tuning aid, skipped by default.
func TestVisualizeDetections(t *testing.T) {
	dir := os.Getenv("VISUALIZE_DIR")
	if dir == "" {
		t.Skip("set VISUALIZE_DIR to write annotated images")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range testImages(t) {
		visualize(t, path, dir)
	}
}

func visualize(t *testing.T, path, dir string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	boxes, err := Detect(data)
	if err != nil {
		t.Fatal(err)
	}

	img := gocv.IMRead(path, gocv.IMReadColor)
	if img.Empty() {
		t.Fatalf("could not read %s", path)
	}
	defer img.Close()

	green := color.RGBA{G: 255, A: 255}
	red := color.RGBA{R: 255, A: 255}
	checked := 0
	for _, b := range boxes {
		c := green
		if b.IsChecked {
			c = red
			checked++
		}
		gocv.Rectangle(&img, image.Rect(b.BBox[0], b.BBox[1], b.BBox[2], b.BBox[3]), c, 3)
	}

	name := filepath.Base(path)
	if !gocv.IMWrite(filepath.Join(dir, name), img) {
		t.Fatalf("could not write %s", name)
	}
	t.Logf("%s: %d boxes, %d checked", name, len(boxes), checked)
}
