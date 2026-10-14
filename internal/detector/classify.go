// Classification: decides whether a detected checkbox is checked or empty.
package detector

import (
	"image"

	"gocv.io/x/gocv"
)

const (
	innerInset = 0.25 // fraction of each side ignored when measuring ink, to skip the border
	checkedInk = 0.10 // minimum fraction of ink inside the box to call it checked
)

// isChecked reports whether the inside of box, without its border, has enough
// ink in the binary mask to count as a mark.
func isChecked(bin gocv.Mat, box image.Rectangle) bool {
	inner := box.Inset(int(float64(box.Dx()) * innerInset))
	if inner.Empty() {
		return false
	}

	region := bin.Region(inner)
	defer region.Close()

	ink := float64(gocv.CountNonZero(region)) / float64(inner.Dx()*inner.Dy())
	return ink >= checkedInk
}
