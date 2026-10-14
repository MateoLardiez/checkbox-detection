package detector

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
)

// newBox returns a mask with a 40x40 outline of 3px borders at (100,100).
func newBox() (gocv.Mat, image.Rectangle) {
	bin := gocv.NewMatWithSize(300, 300, gocv.MatTypeCV8UC1)
	box := image.Rect(100, 100, 140, 140)
	fillRect(&bin, image.Rect(100, 100, 140, 103))
	fillRect(&bin, image.Rect(100, 137, 140, 140))
	fillRect(&bin, image.Rect(100, 100, 103, 140))
	fillRect(&bin, image.Rect(137, 100, 140, 140))
	return bin, box
}

func TestIsCheckedEmptyBox(t *testing.T) {
	bin, box := newBox()
	defer bin.Close()

	if isChecked(bin, box) {
		t.Fatal("empty box reported as checked")
	}
}

func TestIsCheckedBoxWithSpeck(t *testing.T) {
	bin, box := newBox()
	defer bin.Close()
	fillRect(&bin, image.Rect(118, 118, 120, 120))

	if isChecked(bin, box) {
		t.Fatal("box with a tiny speck reported as checked")
	}
}

func TestIsCheckedBoxWithX(t *testing.T) {
	bin, box := newBox()
	defer bin.Close()
	for i := 4; i < 36; i++ {
		fillRect(&bin, image.Rect(100+i, 100+i, 103+i, 103+i))
		fillRect(&bin, image.Rect(100+i, 137-i, 103+i, 140-i))
	}

	if !isChecked(bin, box) {
		t.Fatal("box with an X reported as empty")
	}
}

func TestIsCheckedTinyBox(t *testing.T) {
	bin := gocv.NewMatWithSize(50, 50, gocv.MatTypeCV8UC1)
	defer bin.Close()

	if isChecked(bin, image.Rect(10, 10, 12, 12)) {
		t.Fatal("tiny box reported as checked")
	}
}
