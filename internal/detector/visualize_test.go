package detector

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"gocv.io/x/gocv"
)

// TestVisualizeCandidates draws the detected candidates over each sample image
// and saves the result to VISUALIZE_DIR. It is a tuning aid, skipped by default.
func TestVisualizeCandidates(t *testing.T) {
	dir := os.Getenv("VISUALIZE_DIR")
	if dir == "" {
		t.Skip("set VISUALIZE_DIR to write annotated images")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob("../../testdata/sample-*.png")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		visualize(t, path, dir)
	}
}

func visualize(t *testing.T, path, dir string) {
	t.Helper()

	img := gocv.IMRead(path, gocv.IMReadColor)
	if img.Empty() {
		t.Fatalf("could not read %s", path)
	}
	defer img.Close()

	bin := preprocess(img)
	defer bin.Close()

	candidates := findCandidates(bin)
	green := color.RGBA{G: 255, A: 255}
	for _, r := range candidates {
		gocv.Rectangle(&img, r, green, 3)
	}

	name := filepath.Base(path)
	if !gocv.IMWrite(filepath.Join(dir, name), img) {
		t.Fatalf("could not write %s", name)
	}
	t.Logf("%s: %d candidates", name, len(candidates))
}
