package detector

import (
	"errors"

	"gocv.io/x/gocv"
)

type Box struct {
	BBox      [4]int `json:"bbox"`
	IsChecked bool   `json:"is_checked"`
}

var ErrInvalidImage = errors.New("invalid image")

func Detect(data []byte) ([]Box, error) {
	img, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	if img.Empty() {
		return nil, ErrInvalidImage
	}

	return []Box{}, nil
}
