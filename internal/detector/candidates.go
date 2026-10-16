// Candidate search: finds contours that look like checkboxes in a binary mask.
package detector

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

const (
	minSideRatio = 0.005 // smallest checkbox side, as a fraction of image width
	maxSideRatio = 0.03  // largest checkbox side, as a fraction of image width
	minAspect    = 0.67  // narrowest width/height ratio accepted as a checkbox
	maxAspect    = 1.5   // widest width/height ratio accepted as a checkbox
	approxEps    = 0.02  // how far a contour may deviate when reduced to corners (2% of its perimeter)
	maxCornerGap = 0.10  // how far a corner may be from the box's ideal corner, as a fraction of its longer side
	holeInflate  = 0.08  // growth per side of a hole's box to recover the border thickness, as a fraction of its side
)

// findCandidates returns the bounding box of every small, roughly square,
// four-cornered contour in the mask. A checkbox fused with a table line has no
// square outer contour, so its inner border (a hole) is used instead.
func findCandidates(bin gocv.Mat) []image.Rectangle {
	width := float64(bin.Cols())
	minSide := width * minSideRatio
	maxSide := width * maxSideRatio

	hierarchy := gocv.NewMat()
	defer hierarchy.Close()
	contours := gocv.FindContoursWithParams(bin, &hierarchy, gocv.RetrievalCComp, gocv.ChainApproxSimple)
	defer contours.Close()

	n := contours.Size()
	rects := make([]image.Rectangle, n)
	passed := make([]bool, n)
	for i := 0; i < n; i++ {
		contour := contours.At(i)
		rects[i] = gocv.BoundingRect(contour)
		passed[i] = looksLikeCheckbox(rects[i], contour, minSide, maxSide)
	}

	bounds := image.Rect(0, 0, bin.Cols(), bin.Rows())
	candidates := []image.Rectangle{}
	for i := 0; i < n; i++ {
		if !passed[i] {
			continue
		}
		parent := hierarchy.GetVeciAt(0, i)[3]
		switch {
		case parent == -1:
			candidates = append(candidates, rects[i])
		case !passed[parent]: // hole of a shape that is not a checkbox itself
			candidates = append(candidates, inflate(rects[i]).Intersect(bounds))
		}
	}
	return candidates
}

// looksLikeCheckbox applies the size, squareness and corner filters.
func looksLikeCheckbox(r image.Rectangle, contour gocv.PointVector, minSide, maxSide float64) bool {
	w, h := float64(r.Dx()), float64(r.Dy())
	if w < minSide || w > maxSide || h < minSide || h > maxSide {
		return false
	}
	if aspect := w / h; aspect < minAspect || aspect > maxAspect {
		return false
	}
	return isRectangle(contour)
}

// inflate grows a hole's box by the typical border thickness.
func inflate(r image.Rectangle) image.Rectangle {
	pad := int(math.Round(float64(r.Dx()) * holeInflate))
	return r.Inset(-pad)
}

// isRectangle reports whether the contour simplifies to four corners that each
// sit close to a corner of its minimum-area rotated rectangle. Tilted squares
// pass, while slanted or irregular shapes (like a hand-drawn scribble) do not.
func isRectangle(contour gocv.PointVector) bool {
	approx := gocv.ApproxPolyDP(contour, approxEps*gocv.ArcLength(contour, true), true)
	defer approx.Close()
	if approx.Size() != 4 {
		return false
	}

	corners := approx.ToPoints()
	ideal := gocv.MinAreaRect(contour)
	maxGap := maxCornerGap * math.Max(float64(ideal.Width), float64(ideal.Height))
	for _, want := range ideal.Points {
		nearest := math.MaxFloat64
		for _, got := range corners {
			nearest = math.Min(nearest, math.Hypot(float64(got.X-want.X), float64(got.Y-want.Y)))
		}
		if nearest > maxGap {
			return false
		}
	}
	return true
}
