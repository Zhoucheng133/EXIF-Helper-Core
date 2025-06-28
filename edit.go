package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func calMargin(w int) int {
	return int(math.Floor(float64(w) * 0.03))
}

func drawLen(rgba *image.RGBA, h int, w int, extendHeight int, exif EXIFInfo) *image.RGBA {
	fontSize := float64(extendHeight) * 0.15
	face, _ := loadFontFace(fontBytes, fontSize)
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.RGBA{180, 180, 200, 255}),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	// focal, _ := evalNum(exif.Focal)

	text := fmt.Sprintf("%s (%smm)", exif.LenModel, exif.Focal)

	textWidth := font.MeasureString(face, text).Round()

	x := w - calMargin(w) - textWidth
	y := h + (extendHeight+ascent-descent)/2 - int(math.Floor(float64(extendHeight)*0.17))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(text)

	return rgba
}

func drawInfos(rgba *image.RGBA, h int, w int, extendHeight int, exif EXIFInfo) *image.RGBA {
	fontSize := float64(extendHeight) * 0.25
	face, _ := loadFontFace(fontBytes, fontSize)
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	// fNum, _ := evalNum(exif.Fnum)
	// exp, _ := evalExposure(exif.ExposureTime)

	text := fmt.Sprintf("F%s    %ss    ISO%s", exif.Fnum, exif.ExposureTime, exif.Iso)
	textWidth := font.MeasureString(face, text).Round()

	x := w - calMargin(w) - textWidth
	y := h + (extendHeight+ascent-descent)/2 + int(math.Floor(float64(extendHeight)*0.13))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(text)

	return rgba
}

func drawDatetime(rgba *image.RGBA, h int, w int, extendHeight int, dateTime string) *image.RGBA {
	fontSize := float64(extendHeight) * 0.15
	face, _ := loadFontFace(fontBytes, fontSize)
	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(color.RGBA{180, 180, 200, 255}),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	x := calMargin(w)
	y := h + (extendHeight+ascent-descent)/2 + int(math.Floor(float64(extendHeight)*0.17))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	time := strings.Replace(dateTime, ":", "-", 2)

	drawer.DrawString(time)

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

	x := calMargin(w)
	y := h + (extendHeight+ascent-descent)/2 - int(math.Floor(float64(extendHeight)*0.13))

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
	drawDatetime(rgba, h, w, extendHeight, exif.CaptureTime)
	drawLen(rgba, h, w, extendHeight, exif)
	drawInfos(rgba, h, w, extendHeight, exif)

	nrgba := imaging.Clone(rgba)

	return nrgba
}
