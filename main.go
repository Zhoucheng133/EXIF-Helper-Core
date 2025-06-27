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
	"image/color"
	"image/jpeg"
	"os"
	"unsafe"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
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
	return result
}

func imageSave(path string, output string) {
	result := imageEdit(path)
	imaging.Save(result, output)
}

func getEXIF(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return err.Error()
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return err.Error()
	}

	getTagString := func(name exif.FieldName) string {
		tag, err := x.Get(name)
		if err != nil || tag == nil {
			return ""
		}
		return tag.String()
	}

	res := EXIFInfo{
		CamMake:      getTagString(exif.Make),
		CamModel:     getTagString(exif.Model),
		LenMake:      getTagString(exif.LensMake),
		LenModel:     getTagString(exif.LensModel),
		CaptureTime:  getTagString(exif.DateTime),
		ExposureTime: getTagString(exif.ExposureTime),
		Fnum:         getTagString(exif.FNumber),
		Iso:          getTagString(exif.ISOSpeedRatings),
		Focal:        getTagString(exif.FocalLength),
		Focal35:      getTagString(exif.FocalLengthIn35mmFilm),
	}

	data, _ := json.Marshal(res)
	return string(data)
}

//export ImageSave
func ImageSave(path *C.char, output *C.char) {
	imageSave(C.GoString(path), C.GoString(output))
}

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	info := getEXIF(C.GoString(path))
	return C.CString(string(info))
}

func main() {
	// 测试代码
	// imageSave("/Users/zhoucheng/Downloads/DSC08221.jpg", "/Users/zhoucheng/Downloads/DSC08221_output.jpg")
}
