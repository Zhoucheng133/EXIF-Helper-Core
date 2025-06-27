package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func drawLen(rgba *image.RGBA, h int, w int, extendHeight int, lenModel string) *image.RGBA {
	fontSize := float64(extendHeight) * 0.2
	face, _ := loadFontFace(fontBytes, fontSize)
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	x := int(math.Floor(float64(w) * 0.01))
	y := h + (extendHeight+ascent-descent)/2 + int(math.Floor(float64(extendHeight)*0.15))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(lenModel)

	return rgba
}

func drawModel(rgba *image.RGBA, h int, w int, extendHeight int, camModel string) *image.RGBA {
	fontSize := float64(extendHeight) * 0.28
	face, _ := loadFontFace(fontBytes, fontSize)
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	x := int(math.Floor(float64(w) * 0.01))
	y := h + (extendHeight+ascent-descent)/2 - int(math.Floor(float64(extendHeight)*0.15))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(camModel)

	return rgba
}

func imageEdit(path string) *image.NRGBA {
	img, err := imaging.Open(path)
	if err != nil {
		return nil
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	extendHeight := int(float64(h) * 0.15)
	newHeight := h + extendHeight
	whiteBg := imaging.New(w, newHeight, color.White)
	result := imaging.Paste(whiteBg, img, image.Pt(0, 0))

	rgba := image.NewRGBA(result.Bounds())
	draw.Draw(rgba, rgba.Bounds(), result, image.Point{}, draw.Src)
	exif := getEXIF(path)

	drawModel(rgba, h, w, extendHeight, exif.CamModel)
	drawLen(rgba, h, w, extendHeight, exif.LenModel)

	nrgba := imaging.Clone(rgba)

	return nrgba
}
