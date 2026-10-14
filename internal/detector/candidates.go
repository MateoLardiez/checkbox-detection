// Candidate search: finds contours that look like checkboxes in a binary mask.
package detector

import (
	"image"

	"gocv.io/x/gocv"
)

const (
	minSideRatio = 0.005 // smallest checkbox side, as a fraction of image width
	maxSideRatio = 0.03  // largest checkbox side, as a fraction of image width
	minAspect    = 0.8   // narrowest width/height ratio accepted as a square
	maxAspect    = 1.25  // widest width/height ratio accepted as a square
	approxEps    = 0.02  // how far a contour may deviate when reduced to corners (2% of its perimeter)
)

// findCandidates returns the bounding box of every small, roughly square,
// four-cornered outer contour in the mask. Holes (inner borders) are skipped.
func findCandidates(bin gocv.Mat) []image.Rectangle {
	width := float64(bin.Cols())
	minSide := width * minSideRatio
	maxSide := width * maxSideRatio

	hierarchy := gocv.NewMat()
	defer hierarchy.Close()
	contours := gocv.FindContoursWithParams(bin, &hierarchy, gocv.RetrievalCComp, gocv.ChainApproxSimple)
	defer contours.Close()

	candidates := []image.Rectangle{}
	for i := 0; i < contours.Size(); i++ {
		if hierarchy.GetVeciAt(0, i)[3] != -1 { // has a parent: it is a hole
			continue
		}
		contour := contours.At(i)
		r := gocv.BoundingRect(contour)
		w, h := float64(r.Dx()), float64(r.Dy())

		if w < minSide || w > maxSide || h < minSide || h > maxSide {
			continue
		}
		if aspect := w / h; aspect < minAspect || aspect > maxAspect {
			continue
		}
		if !hasFourCorners(contour) {
			continue
		}
		candidates = append(candidates, r)
	}
	return candidates
}

// hasFourCorners reports whether the contour simplifies to a quadrilateral.
func hasFourCorners(contour gocv.PointVector) bool {
	approx := gocv.ApproxPolyDP(contour, approxEps*gocv.ArcLength(contour, true), true)
	defer approx.Close()
	return approx.Size() == 4
}
