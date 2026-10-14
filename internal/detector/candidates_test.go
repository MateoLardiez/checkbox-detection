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

func TestFindCandidatesCheckboxTouchingLine(t *testing.T) {
	bin := gocv.NewMatWithSize(1000, 1000, gocv.MatTypeCV8UC1)
	defer bin.Close()

	// 20x20 outline with 3px borders, its bottom border fused with a long line.
	checkbox := image.Rect(100, 100, 120, 120)
	fillRect(&bin, image.Rect(100, 100, 120, 103))
	fillRect(&bin, image.Rect(100, 117, 120, 120))
	fillRect(&bin, image.Rect(100, 100, 103, 120))
	fillRect(&bin, image.Rect(117, 100, 120, 120))
	fillRect(&bin, image.Rect(30, 117, 400, 120))

	got := findCandidates(bin)
	if len(got) != 1 {
		t.Fatalf("got %d candidates %v, want exactly 1", len(got), got)
	}
	if area := checkbox.Inset(-5); !got[0].In(area) {
		t.Errorf("candidate %v, want inside %v", got[0], area)
	}
	if got[0].Dx() <= 14 {
		t.Errorf("candidate %v is not wider than the 14px hole, it was not inflated", got[0])
	}
}

// fillRect sets every pixel of r to white in a single-channel mask.
func fillRect(m *gocv.Mat, r image.Rectangle) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.SetUCharAt(y, x, 255)
		}
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
