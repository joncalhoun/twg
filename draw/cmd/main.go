package main

import (
	"image"
	imgdraw "image/draw"
	"image/png"
	"os"

	"github.com/joncalhoun/twg/draw"
)

func main() {
	const w, h = 1000, 1000
	var im imgdraw.Image
	im = image.NewRGBA(image.Rectangle{Max: image.Point{X: w, Y: h}})
	im = draw.FibGradient(im)
	f, err := os.Create("image.png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, im)
	if err != nil {
		panic(err)
	}
}
