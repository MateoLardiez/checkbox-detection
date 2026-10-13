package detector

import (
	"testing"

	"gocv.io/x/gocv"
)

func TestPreprocess(t *testing.T) {
	img := gocv.IMRead("../../testdata/sample-000.png", gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read sample image")
	}
	defer img.Close()

	bin := preprocess(img)
	defer bin.Close()

	if bin.Rows() != img.Rows() || bin.Cols() != img.Cols() {
		t.Fatalf("size = %dx%d, want %dx%d", bin.Cols(), bin.Rows(), img.Cols(), img.Rows())
	}
	if bin.Channels() != 1 {
		t.Fatalf("channels = %d, want 1", bin.Channels())
	}
	for i, v := range bin.ToBytes() {
		if v != 0 && v != 255 {
			t.Fatalf("pixel %d = %d, want 0 or 255", i, v)
		}
	}
}
