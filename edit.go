package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	xdraw "golang.org/x/image/draw"
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

	// fNum, _ := evalNum(exif.Fnum)
	// exp, _ := evalExposure(exif.ExposureTime)

	// text := fmt.Sprintf("F%s  %ss  ISO%s", exif.Fnum, exif.ExposureTime, exif.Iso)
	text := fmt.Sprintf("%ss, f/%s", exif.ExposureTime, exif.Fnum)
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

//go:embed assets/*
var modelImages embed.FS

func drawModel(rgba *image.RGBA, h int, w int, extendHeight int, camModel string, camMake string, showLogo bool) *image.RGBA {
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

	if showLogo {
		logoName := ""
		if strings.Contains(strings.ToLower(camMake), "nikon") {
			logoName = "nikon"
		} else if strings.Contains(strings.ToLower(camMake), "sony") {
			logoName = "sony"
		} else if strings.Contains(strings.ToLower(camMake), "apple") {
			logoName = "apple"
		} else if strings.Contains(strings.ToLower(camMake), "canon") {
			logoName = "canon"
		} else if strings.Contains(strings.ToLower(camMake), "panasonic") {
			logoName = "panasonic"
		} else if strings.Contains(strings.ToLower(camMake), "leica") {
			logoName = "leica"
		}

		logoFile, err := modelImages.Open("assets/" + logoName + ".png")
		if err == nil {
			defer logoFile.Close()
			logoImg, _ := png.Decode(logoFile)

			targetHeight := int(float64(extendHeight) * 0.3)
			scaleW := logoImg.Bounds().Dx() * targetHeight / logoImg.Bounds().Dy()
			scaledRect := image.Rect(0, 0, scaleW, targetHeight)
			scaled := image.NewRGBA(scaledRect)
			xdraw.CatmullRom.Scale(scaled, scaledRect, logoImg, logoImg.Bounds(), draw.Over, nil)

			offset := image.Pt(
				x,
				y-targetHeight/2-int(float64(targetHeight)*0.37),
			)
			rect := image.Rectangle{Min: offset, Max: offset.Add(scaled.Bounds().Size())}
			draw.Draw(rgba, rect, scaled, image.Point{}, draw.Over)

			x += scaleW + int(float64(targetHeight)*0.3)
		}
	}

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(camModel)

	return rgba
}

func imageEdit(path string, showLogo bool) *image.NRGBA {
	img, err := imaging.Open(path)
	if err != nil {
		return nil
	}

	exif := getEXIF(path)

	switch exif.Orientation {
	case "3":
		img = imaging.Rotate180(img)
	case "6":
		img = imaging.Rotate270(img)
	case "8":
		img = imaging.Rotate90(img)
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	extendHeight := int(float64(h) * 0.12)
	newHeight := h + extendHeight
	whiteBg := imaging.New(w, newHeight, color.White)
	result := imaging.Paste(whiteBg, img, image.Pt(0, 0))

	rgba := image.NewRGBA(result.Bounds())
	draw.Draw(rgba, rgba.Bounds(), result, image.Point{}, draw.Src)

	drawModel(rgba, h, w, extendHeight, exif.CamModel, exif.CamMake, showLogo)
	drawDatetime(rgba, h, w, extendHeight, exif.CaptureTime)
	drawLen(rgba, h, w, extendHeight, exif)
	drawInfos(rgba, h, w, extendHeight, exif)

	nrgba := imaging.Clone(rgba)

	return nrgba
}
