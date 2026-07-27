package utils

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	xdraw "golang.org/x/image/draw"
)

func logoNameHandler(camMake string) (string, float64) {
	lower := strings.ToLower(camMake)
	switch {
	case strings.Contains(lower, "nikon"):
		return "nikon", 2.5
	case strings.Contains(lower, "sony"):
		return "sony", 5.0
	case strings.Contains(lower, "apple"):
		return "apple", 2.5
	case strings.Contains(lower, "canon"):
		return "canon", 5.0
	case strings.Contains(lower, "panasonic"):
		return "panasonic", 5.0
	case strings.Contains(lower, "leica"):
		return "leica", 2.5
	case strings.Contains(lower, "fujifilm"):
		return "fujifilm", 5.0
	case strings.Contains(lower, "xiaomi"):
		return "xiaomi", 2.5
	case strings.Contains(lower, "huawei"):
		return "huawei", 2.5
	case strings.Contains(lower, "oppo"):
		return "oppo", 5.0
	case strings.Contains(lower, "vivo"):
		return "vivo", 5.0
	case strings.Contains(lower, "oneplus"):
		return "oneplus", 2.5
	case strings.Contains(lower, "honor"):
		return "honor", 5.0
	case strings.Contains(lower, "google"):
		return "google", 2.5
	case strings.Contains(lower, "samsung"):
		return "samsung", 5.0
	case strings.Contains(lower, "om digital") || strings.Contains(lower, "olympus"):
		return "olympus", 5.0
	default:
		return "", 2.5
	}
}

func calMargin(w int) int {
	return int(math.Floor(float64(w) * 0.03))
}

var logoCache sync.Map

func getDecodedLogo(name string) image.Image {
	if cached, ok := logoCache.Load(name); ok {
		return cached.(image.Image)
	}
	file, err := modelImages.Open("assets/" + name + ".png")
	if err != nil {
		return nil
	}
	defer file.Close()
	img, _ := png.Decode(file)
	if img != nil {
		logoCache.Store(name, img)
	}
	return img
}

func drawLen(dst draw.Image, h int, w int, extendHeight int, exif EXIFInfo) {
	fontSize := float64(extendHeight) * 0.15
	face := loadFontFace(fontSize)
	if face == nil {
		return
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.RGBA{180, 180, 200, 255}),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	text := fmt.Sprintf("%s (%smm)", exif.LenModel, exif.Focal)

	textWidth := font.MeasureString(face, text).Round()

	x := w - calMargin(w) - textWidth
	y := h + (extendHeight+ascent-descent)/2 - int(math.Floor(float64(extendHeight)*0.17))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(text)
}

func drawInfos(dst draw.Image, h int, w int, extendHeight int, exif EXIFInfo, showF bool, showExposureTime bool, showISO bool) {
	fontSize := float64(extendHeight) * 0.2
	face := loadFontFace(fontSize)
	if face == nil {
		return
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	text := ""

	var parts []string
	if showExposureTime {
		parts = append(parts, fmt.Sprintf("%ss", exif.ExposureTime))
	}
	if showF {
		parts = append(parts, fmt.Sprintf("f/%s", exif.Fnum))
	}
	if showISO {
		parts = append(parts, fmt.Sprintf("ISO%s", exif.Iso))
	}
	text = strings.Join(parts, ", ")

	textWidth := font.MeasureString(face, text).Round()

	x := w - calMargin(w) - textWidth
	y := h + (extendHeight+ascent-descent)/2 + int(math.Floor(float64(extendHeight)*0.13))

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(text)
}

func drawDatetime(dst draw.Image, h int, w int, extendHeight int, dateTime string) {
	fontSize := float64(extendHeight) * 0.15
	face := loadFontFace(fontSize)
	if face == nil {
		return
	}
	drawer := &font.Drawer{
		Dst:  dst,
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
}

//go:embed assets/*
var modelImages embed.FS

func drawLogo(dst draw.Image, h, w, extendHeight int, camMake string) {
	logoName, ratio := logoNameHandler(camMake)
	if logoName == "" {
		return
	}
	logoImg := getDecodedLogo(logoName)
	if logoImg == nil {
		return
	}

	targetHeight := int(float64(extendHeight) / ratio)
	scaleW := logoImg.Bounds().Dx() * targetHeight / logoImg.Bounds().Dy()
	scaledRect := image.Rect(0, 0, scaleW, targetHeight)
	scaled := image.NewRGBA(scaledRect)
	xdraw.CatmullRom.Scale(scaled, scaledRect, logoImg, logoImg.Bounds(), draw.Over, nil)

	x := (w - scaleW) / 2
	y := h + (extendHeight-targetHeight)/2

	offset := image.Pt(x, y)
	rect := image.Rectangle{Min: offset, Max: offset.Add(scaled.Bounds().Size())}
	draw.Draw(dst, rect, scaled, image.Point{}, draw.Over)
}

func drawModel(dst draw.Image, h int, w int, extendHeight int, camModel string, camMake string, showLogo bool) {
	fontSize := float64(extendHeight) * 0.28
	face := loadFontFace(fontSize)
	if face == nil {
		return
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	metrics := face.Metrics()
	ascent := metrics.Ascent.Round()
	descent := metrics.Descent.Round()

	x := calMargin(w)
	y := h + (extendHeight+ascent-descent)/2 - int(math.Floor(float64(extendHeight)*0.13))

	if showLogo {
		drawLogo(dst, h, w, extendHeight, camMake)
	}

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	drawer.DrawString(camModel)
}

func ImageEdit(path string, showLogo bool, showF bool, showExposureTime bool, showISO bool, maxDim int) *image.NRGBA {
	img, err := imaging.Open(path)
	if err != nil {
		return nil
	}

	exif, err := GetEXIF(path)

	if err != nil {
		return nil
	}

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

	if maxDim > 0 && (w > maxDim || h > maxDim) {
		var newW, newH int
		if w >= h {
			newW = maxDim
			newH = h * maxDim / w
		} else {
			newH = maxDim
			newW = w * maxDim / h
		}
		img = imaging.Resize(img, newW, newH, imaging.Lanczos)
		w = newW
		h = newH
	}

	extendHeight := int(float64(h) * 0.12)
	newHeight := h + extendHeight
	whiteBg := imaging.New(w, newHeight, color.White)
	result := imaging.Paste(whiteBg, img, image.Pt(0, 0))

	drawModel(result, h, w, extendHeight, exif.CamModel, exif.CamMake, showLogo)
	drawDatetime(result, h, w, extendHeight, exif.CaptureTime)
	drawLen(result, h, w, extendHeight, exif)
	drawInfos(result, h, w, extendHeight, exif, showF, showExposureTime, showISO)

	return result
}
