package main

/*
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"strings"
	"unsafe"

	_ "embed"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed inter.ttf
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
	Orientation  string `json:"orientation"`
}

//export FreeMemory
func FreeMemory(ptr unsafe.Pointer) {
	C.free(ptr)
}

func evalInt(expr string) (string, error) {
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
func ImagePreview(path *C.char, outLength *C.int, showLogo C.int) *C.uchar {
	img := imageEdit(C.GoString(path), showLogo == 1)
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

func imageSave(path string, output string, showLogo bool) {
	result := imageEdit(path, showLogo)
	imaging.Save(result, output)
}

func getEXIF(path string) EXIFInfo {
	f, _ := os.Open(path)
	defer f.Close()

	data, _ := exif.Decode(f)
	return formatExif(data)
}

//export ImageSave
func ImageSave(path *C.char, output *C.char, showLogo C.int) {
	imageSave(C.GoString(path), C.GoString(output), showLogo == 1)
}

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	// info := getEXIF(C.GoString(path))
	data, _ := json.Marshal(getEXIF(C.GoString(path)))
	return C.CString(string(data))
}

func main() {
	imageSave("/Users/zhoucheng/Downloads/测试照片/索尼.JPG", "/Users/zhoucheng/Downloads/测试照片/输出.jpg", true)
}
