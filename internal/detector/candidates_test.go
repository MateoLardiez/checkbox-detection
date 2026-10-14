package detector

import (
	"image"
	"image/color"
	"testing"

	"gocv.io/x/gocv"
)

func TestFindCandidates(t *testing.T) {
	img := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(255, 255, 255, 0), 1000, 1000, gocv.MatTypeCV8UC3)
	defer img.Close()

	black := color.RGBA{A: 255}
	checkbox := image.Rect(100, 100, 120, 120)
	gocv.Rectangle(&img, checkbox, black, 2)
	gocv.Rectangle(&img, image.Rect(300, 300, 500, 302), black, -1) // line, not square
	gocv.Rectangle(&img, image.Rect(600, 600, 700, 700), black, 2)  // square, too big
	gocv.Rectangle(&img, image.Rect(800, 800, 802, 802), black, -1) // dot, too small

	bin := preprocess(img)
	defer bin.Close()

	got := findCandidates(bin)
	if len(got) != 1 {
		t.Fatalf("got %d candidates %v, want exactly 1", len(got), got)
	}
	if area := checkbox.Inset(-5); !got[0].In(area) {
		t.Errorf("candidate %v, want inside %v", got[0], area)
	}
}

func TestFindCandidatesBlankImage(t *testing.T) {
	img := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(255, 255, 255, 0), 500, 500, gocv.MatTypeCV8UC3)
	defer img.Close()

	bin := preprocess(img)
	defer bin.Close()

	if got := findCandidates(bin); got == nil || len(got) != 0 {
		t.Fatalf("got %v, want empty non-nil slice", got)
	}
}
