// Image preprocessing: turns a color document image into a binary mask.
package detector

import "gocv.io/x/gocv"

// preprocess converts a color image to grayscale and applies an inverted
// adaptive threshold, so ink becomes white (255) and paper black (0) even
// with uneven scan lighting. The caller must Close the returned Mat.
func preprocess(img gocv.Mat) gocv.Mat {
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	bin := gocv.NewMat()
	gocv.AdaptiveThreshold(gray, &bin, 255,
		gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinaryInv, 25, 10)
	return bin
}
