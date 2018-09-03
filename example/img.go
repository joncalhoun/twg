package example

import "io"

// Supported format types
const (
	FormatPNG = "png"
	FormatJPG = "jpg"
	FormatGIF = "gif"
)

// Image is a fake image type
type Image struct{}

// Decode is a fake function to decode images provided via an io.Reader
func Decode(r io.Reader) (*Image, error) {
	return &Image{}, nil
}

// Crop is a fake function to crop images
func Crop(img *Image, x1, y1, x2, y2 int) error {
	return nil
}

// Encode is a fake function to encode images
func Encode(img *Image, format string, w io.Writer) error {
	return nil
}
