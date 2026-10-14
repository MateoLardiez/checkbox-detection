package detector

import (
	"cmp"
	"errors"
	"slices"

	"gocv.io/x/gocv"
)

type Box struct {
	BBox      [4]int `json:"bbox"`
	IsChecked bool   `json:"is_checked"`
}

var ErrInvalidImage = errors.New("invalid image")

// Detect finds the checkboxes in an encoded image (PNG, JPEG...) and returns
// them sorted top to bottom, then left to right, with coordinates in pixels
// of the given image.
func Detect(data []byte) ([]Box, error) {
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	if img.Empty() {
		return nil, ErrInvalidImage
	}

	bin := preprocess(img)
	defer bin.Close()

	candidates := findCandidates(bin)
	boxes := make([]Box, 0, len(candidates))
	for _, r := range candidates {
		boxes = append(boxes, Box{
			BBox:      [4]int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y},
			IsChecked: isChecked(bin, r),
		})
	}
	slices.SortFunc(boxes, compareBoxes)
	return boxes, nil
}

// compareBoxes gives a total order (y1, x1, x2, y2) so the output is deterministic.
func compareBoxes(a, b Box) int {
	return cmp.Or(
		cmp.Compare(a.BBox[1], b.BBox[1]),
		cmp.Compare(a.BBox[0], b.BBox[0]),
		cmp.Compare(a.BBox[2], b.BBox[2]),
		cmp.Compare(a.BBox[3], b.BBox[3]),
	)
}
