package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"os"
	"strings"
	"unsafe"

	_ "embed"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed Roboto.ttf
var fontBytes []byte

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
}

//export FreeMemory
func FreeMemory(ptr unsafe.Pointer) {
	C.free(ptr)
}

func previewSize(img image.Image, maxDim int) image.Image {
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

//export ImagePreview
func ImagePreview(path *C.char, outLength *C.int) *C.uchar {
	img := imageEdit(C.GoString(path))
	if img == nil {
		*outLength = 0
		return nil
	}
	previewImg := previewSize(img, 1000)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, previewImg, nil); err != nil {
		*outLength = 0
		return nil
	}
	data := buf.Bytes()
	*outLength = C.int(len(data))
	ptr := C.malloc(C.size_t(len(data)))
	C.memcpy(ptr, unsafe.Pointer(&data[0]), C.size_t(len(data)))

	return (*C.uchar)(ptr)
}

func loadFontFace(fontBytes []byte, fontSize float64) (font.Face, error) {
	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	return face, err
}

func imageSave(path string, output string) {
	result := imageEdit(path)
	imaging.Save(result, output)
}

func getEXIF(path string) EXIFInfo {
	f, _ := os.Open(path)
	defer f.Close()

	x, _ := exif.Decode(f)
	getTagString := func(name exif.FieldName) string {
		tag, err := x.Get(name)
		if err != nil || tag == nil {
			return ""
		}
		return tag.String()
	}

	res := EXIFInfo{
		CamMake:      strings.ReplaceAll(getTagString(exif.Make), "\"", ""),
		CamModel:     strings.ReplaceAll(getTagString(exif.Model), "\"", ""),
		LenMake:      strings.ReplaceAll(getTagString(exif.LensMake), "\"", ""),
		LenModel:     strings.ReplaceAll(getTagString(exif.LensModel), "\"", ""),
		CaptureTime:  strings.ReplaceAll(getTagString(exif.DateTime), "\"", ""),
		ExposureTime: strings.ReplaceAll(getTagString(exif.ExposureTime), "\"", ""),
		Fnum:         strings.ReplaceAll(getTagString(exif.FNumber), "\"", ""),
		Iso:          strings.ReplaceAll(getTagString(exif.ISOSpeedRatings), "\"", ""),
		Focal:        strings.ReplaceAll(getTagString(exif.FocalLength), "\"", ""),
		Focal35:      strings.ReplaceAll(getTagString(exif.FocalLengthIn35mmFilm), "\"", ""),
	}

	return res
}

//export ImageSave
func ImageSave(path *C.char, output *C.char) {
	imageSave(C.GoString(path), C.GoString(output))
}

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	// info := getEXIF(C.GoString(path))
	data, _ := json.Marshal(getEXIF(C.GoString(path)))
	return C.CString(string(data))
}

func main() {
	// 测试代码
	imageSave("/Users/zhoucheng/Downloads/DSC08041.jpg", "/Users/zhoucheng/Downloads/DSC08041_output.jpg")
	// fmt.Println(getEXIF("/Users/zhoucheng/Downloads/DSC08041.JPG"))
}
