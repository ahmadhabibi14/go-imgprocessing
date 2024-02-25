package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/ftrvxmtrx/tga"
	"github.com/nfnt/resize"
)

type MosaicBuilder struct {
	pathToParts string
	partSize    uint
	logger      *log.Logger
}

func NewMosaicBuilder(pathToParts string, partSize uint) *MosaicBuilder {
	return &MosaicBuilder{
		pathToParts: pathToParts,
		partSize:    partSize,
		logger:      log.New(os.Stderr, "[mosaic] ", log.LstdFlags),
	}
}

func (b *MosaicBuilder) Build() error {
	b.logger.Println("Load parts paths ...")
	partsPaths, err := b.getPartsPaths()
	if err != nil {
		return err
	}

	b.logger.Println("Load parts map ...")
	partsMap, err := b.getPartsMap(partsPaths)
	if err != nil {
		return err
	}

	b.logger.Println("Open source image ...")
	img, err := os.Open("img/img-1.jpg")
	defer func(img *os.File) { _ = img.Close() }(img)

	if err != nil {
		return err
	}

	src, err := jpeg.Decode(img)
	if err != nil {
		return err
	}

	b.logger.Println("calculate ...")
	imgSize := src.Bounds().Size()

	if imgSize.X > 300 {
		src = resize.Resize(300, 0, src, resize.Lanczos3)
	}
	if imgSize.Y > 300 {
		src = resize.Resize(0, 300, src, resize.Lanczos3)
	}

	imgSize = src.Bounds().Size()

	partSize := int(b.partSize)
	resImg := image.NewRGBA(
		image.Rectangle{
			Min: image.Point{},
			Max: image.Point{X: imgSize.X * partSize, Y: imgSize.Y * partSize},
		},
	)

	for x := 0; x < imgSize.X; x++ {
		for y := 0; y < imgSize.Y; y++ {
			bcolor := src.At(x, y)
			part := getClosestPart(&partsMap, bcolor)

			for px := 0; px < partSize; px++ {
				for py := 0; py < partSize; py++ {

					partPixel := color.RGBAModel.Convert(part.At(px, py)).(color.RGBA)
					resImg.Set(partSize*x+px, partSize*y+py, partPixel)
				}
			}
		}
	}

	b.logger.Println("save res ...")
	f, err := os.Create("res.png")
	if err != nil {
		return err
	}

	if err := png.Encode(f, resImg); err != nil {
		return err
	}

	b.logger.Println("done!")

	return nil
}

// getPartsPaths - parsing folder tree with paths to mosaic parts
func (b *MosaicBuilder) getPartsPaths() ([]string, error) {
	var res []string

	return res, filepath.Walk(b.pathToParts, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			res = append(res, path)
		}

		return nil
	})
}

type PixelColor [3]uint8

func (b *MosaicBuilder) getPartsMap(parts []string) (map[PixelColor]image.Image, error) {
	partsMap := make(map[PixelColor]image.Image, len(parts))

	for _, path := range parts {
		src, err := b.loadImage(path)
		if err == nil {
			partsMap[calculateModalAverageColour(src)] = src
		}
	}

	if len(partsMap) == 0 {
		return nil, fmt.Errorf("empty map")
	}

	return partsMap, nil
}

func (b *MosaicBuilder) loadImage(path string) (image.Image, error) {
	infile, err := os.Open(path)
	defer func(infile *os.File) {
		_ = infile.Close()
	}(infile)

	if err != nil {
		return nil, err
	}

	src, err := tga.Decode(infile)
	if err != nil {
		return nil, err
	}

	return resize.Resize(b.partSize, b.partSize, src, resize.Lanczos3), nil
}

func calculateModalAverageColour(img image.Image) PixelColor {
	imgSize := img.Bounds().Size()

	var redTotal, greenTotal, blueTotal, pixelsCount int64

	for x := 0; x < imgSize.X; x++ {
		for y := 0; y < imgSize.Y; y++ {
			cc := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)

			redTotal += int64(cc.R)
			greenTotal += int64(cc.G)
			blueTotal += int64(cc.B)

			pixelsCount++
		}
	}

	r := uint8(redTotal / pixelsCount)
	g := uint8(greenTotal / pixelsCount)
	b := uint8(blueTotal / pixelsCount)

	return PixelColor{r, g, b}
}
func getClosestPart(mp *map[PixelColor]image.Image, pix color.Color) image.Image {
	c := color.RGBAModel.Convert(pix).(color.RGBA)
	key := [3]uint8{c.R, c.G, c.B}

	if part, ok := (*mp)[key]; ok {
		return part
	}

	var minD *float64
	var prt *image.Image

	for m, i := range *mp {

		o := int64(m[0])
		a := int64(c.R)
		rr := float64(o - a)

		rd := math.Pow(rr, 2)
		gd := math.Pow(float64(int64(m[1])-int64(c.G)), 2)
		bd := math.Pow(float64(int64(m[2])-int64(c.B)), 2)

		d := math.Sqrt(rd + gd + bd)
		if minD == nil || *minD > d {
			minD = &d
			prt = &i
		}
	}

	(*mp)[key] = *prt

	return *prt
}

func main() {
	builder := NewMosaicBuilder(
		"./img",
		5,
	)

	if err := builder.Build(); err != nil {
		builder.logger.Fatal(err)
	}
}
