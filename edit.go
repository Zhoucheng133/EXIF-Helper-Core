package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func calMargin(w int) int {
	return int(math.Floor(float64(w) * 0.03))
}

func evalNum(expr string) (string, error) {
	parts := strings.Split(expr, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("不支持的表达式格式")
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || b == 0 {
		return "", fmt.Errorf("格式错误或除数为零")
	}
	return fmt.Sprintf("%d", a/b), nil
}

func evalExposure(expr string) (string, error) {
	parts := strings.Split(expr, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("不支持的表达式格式")
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || b == 0 {
		return "", fmt.Errorf("格式错误或除数为零")
	}
	if a/b < 1 {
		return expr, nil
	}
	return fmt.Sprintf("%d", a/b), nil
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

	focal, _ := evalNum(exif.Focal)

	text := fmt.Sprintf("%s (%smm)", exif.LenModel, focal)

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

	fNum, _ := evalNum(exif.Fnum)
	exp, _ := evalExposure(exif.ExposureTime)

	text := fmt.Sprintf("F%s    %ss    ISO%s", fNum, exp, exif.Iso)
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
