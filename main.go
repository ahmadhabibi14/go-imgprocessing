package main

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	// Load an image
	imgPath := "input.jpg"
	srcImg, err := imaging.Open(imgPath)
	if err != nil {
		panic(err)
	}

	// Create a new image with the same dimensions
	dstImg := imaging.Clone(srcImg)

	// Define the text to write
	text := "Hello, Golang!"

	// Define the font and color
	drawColor := color.RGBA{255, 255, 255, 255} // white color
	face := &basicfont.Face{
		Advance: 7,
		Width:   70,
		Height:  90,
		Ascent:  11,
		Descent: 2,
		Mask:    basicfont.Face7x13.Mask,
		Ranges:  basicfont.Face7x13.Ranges,
	}
	drawOpts := &drawTextOptions{
		Face:  face,
		Color: drawColor,
	}

	// Draw the text onto the image
	drawText(dstImg, 100, 100, text, drawOpts)

	// Save the result
	outputImagePath := "output.jpg"
	err = imaging.Save(dstImg, outputImagePath)
	if err != nil {
		panic(err)
	}
	println("Text added to the image and saved as", outputImagePath)
}

// drawTextOptions holds options for drawing text on an image.
type drawTextOptions struct {
	Face  font.Face
	Color color.Color
}

// drawText writes text on the given image at the specified position.
func drawText(img *image.NRGBA, x, y int, text string, opts *drawTextOptions) {
	drawer := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(opts.Color),
		Face: opts.Face,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(x * 64),
			Y: fixed.Int26_6(y * 64),
		},
	}
	drawer.DrawString(text)
}
