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
	"exif_helper/utils"
	"fmt"
	"image/jpeg"
	"runtime"
	"unsafe"

	_ "embed"
)

func parseImageOptions(options *C.char) utils.ImageOptions {
	var imageOptions utils.ImageOptions
	if options == nil {
		return imageOptions
	}
	_ = json.Unmarshal([]byte(C.GoString(options)), &imageOptions)
	return imageOptions
}

//export FreeMemory
func FreeMemory(ptr unsafe.Pointer) {
	C.free(ptr)
}

//export ImagePreview
func ImagePreview(path *C.char, outLength *C.int, options *C.char) *C.uchar {
	imageOptions := parseImageOptions(options)
	img := utils.ImageDraw(C.GoString(path), imageOptions, 1000)
	if img == nil {
		*outLength = 0
		return nil
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		*outLength = 0
		return nil
	}
	img = nil
	data := buf.Bytes()
	*outLength = C.int(len(data))
	ptr := C.malloc(C.size_t(len(data)))
	C.memcpy(ptr, unsafe.Pointer(&data[0]), C.size_t(len(data)))

	runtime.GC()
	return (*C.uchar)(ptr)
}

//export ImageSave
func ImageSave(path *C.char, output *C.char, options *C.char) {
	imageOptions := parseImageOptions(options)
	utils.ImageSave(C.GoString(path), C.GoString(output), imageOptions)
}

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	info, err := utils.GetEXIF(C.GoString(path))
	if err != nil {
		return C.CString("")
	}
	data, _ := json.Marshal(info)
	return C.CString(string(data))
}

//export RemoveExif
func RemoveExif(inputPath *C.char, output *C.char) {
	utils.RemoveExif(C.GoString(inputPath), C.GoString(output))
}

//export EditEXIF
func EditEXIF(inputPath *C.char, output *C.char, exif *C.char) {
	utils.EditEXIF(C.GoString(inputPath), C.GoString(output), C.GoString(exif))
}

// 测试用例
// func test() {
// 	err := utils.EditEXIF("/Users/zhoucheng/Downloads/DSC_2400.jpg", "/Users/zhoucheng/Downloads/DSC_2400_out.jpg", `{
// 	"camMake": "NIKON CORPORATION",
// 	"camModel": "NIKON Z 30",
// 	"lenMake": "NIKKOR",
// 	"lenModel": "NIKKOR Z DX 16-50mm f/3.5-6.3 VR",
// 	"captureTime": "2026:07:27 22:15:30",
// 	"exposureTime": "3/2",
// 	"fNum": "5.6",
// 	"iso": "100",
// 	"focal": "35",
// 	"focal35": "52",
// 	"orientation": "1"
// 	}`)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
// }

func test2() {
	data, err := utils.GetEXIF("/Users/zhoucheng/Downloads/IMG_6837.JPG")
	if err != nil {
		fmt.Println(err.Error())
	}
	info, _ := json.Marshal(data)
	fmt.Println(string(info))
}

func main() {}
