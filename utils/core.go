package utils

import (
	_ "embed"
	"image"
	"os"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type EXIFInfo struct {
	CamMake      string `json:"camMake"`
	CamModel     string `json:"camModel"`
	LenMake      string `json:"lenMake"`
	LenModel     string `json:"lenModel"`
	CaptureTime  string `json:"captureTime"`
	ExposureTime string `json:"exposureTime"`
	Fnum         string `json:"fNum"`
	Iso          string `json:"iso"`
	Focal        string `json:"focal"`
	Focal35      string `json:"focal35"`
	Orientation  string `json:"orientation"`
}

func GetEXIF(path string) (EXIFInfo, error) {
	f, err := os.Open(path)

	if err != nil {
		return EXIFInfo{}, err
	}

	defer f.Close()

	data, err := exif.Decode(f)

	if err != nil {
		return EXIFInfo{}, err
	}

	return formatExif(data), nil
}

//go:embed assets/inter.ttf
var fontBytes []byte

var (
	fontOnce   sync.Once
	parsedFont *opentype.Font
)

func getParsedFont() *opentype.Font {
	fontOnce.Do(func() {
		fnt, err := opentype.Parse(fontBytes)
		if err != nil {
			return
		}
		parsedFont = fnt
	})
	return parsedFont
}

func loadFontFace(fontSize float64) font.Face {
	fnt := getParsedFont()
	if fnt == nil {
		return nil
	}
	face, _ := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	return face
}

func ImageSave(path string, output string, showLogo bool, showF bool, showExposureTime bool, showISO bool) {
	result := ImageEdit(path, showLogo, showF, showExposureTime, showISO, 0)
	imaging.Save(result, output)
}

func PreviewSize(img image.Image, maxDim int) image.Image {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	// 如果都不超过，直接返回原图
	if w <= maxDim && h <= maxDim {
		return img
	}

	var newWidth, newHeight int
	if w >= h {
		newWidth = maxDim
		newHeight = h * maxDim / w
	} else {
		newHeight = maxDim
		newWidth = w * maxDim / h
	}

	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}
