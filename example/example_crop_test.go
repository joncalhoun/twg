package example_test

import (
	"fmt"
	"io"

	// Needed for initialize side effect
	_ "image/png"

	"github.com/joncalhoun/twg/example"
)

var file string = "this is not used"

func Example_crop() {
	var r io.Reader
	img, err := example.Decode(r)
	if err != nil {
		panic(err)
	}
	err = example.Crop(img, 0, 0, 20, 20)
	if err != nil {
		panic(err)
	}
	var w io.Writer
	if err != nil {
		panic(err)
	}
	err = example.Encode(img, w)
	if err != nil {
		panic(err)
	}
	fmt.Println("See out.jpg for the cropped image.")

	// Output:
	// See out.jpg for the cropped image.
}
