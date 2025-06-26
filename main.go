package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"image"
	"image/color"
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

func savePhoto(path string, output string) {
	// 读取图片
	img, err := imaging.Open(path)
	if err != nil {
		return
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	// 计算新高度：原高度 + 0.15倍高度
	extendHeight := int(float64(h) * 0.15)
	newHeight := h + extendHeight

	// 新建一张白色背景图片
	whiteBg := imaging.New(w, newHeight, color.White)

	// 把原图粘贴到新图的顶部(0,0)
	result := imaging.Paste(whiteBg, img, image.Pt(0, 0))

	// 保存新图
	imaging.Save(result, output)
}

//export SavePhoto
func SavePhoto(path *C.char, output *C.char) {
	savePhoto(C.GoString(path), C.GoString(output))
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

//export FreeCString
func FreeCString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	info := getEXIF(C.GoString(path))
	return C.CString(string(info))
}

func main() {
	savePhoto("/Users/zhoucheng/Downloads/DSC_0092.jpg", "/Users/zhoucheng/Downloads/DSC_0092.jpg")
}
