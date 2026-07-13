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
	"image/jpeg"
	"unsafe"

	_ "embed"
)

//export FreeMemory
func FreeMemory(ptr unsafe.Pointer) {
	C.free(ptr)
}

//export ImagePreview
func ImagePreview(path *C.char, outLength *C.int, showLogo C.int, showF C.int, showExposureTime C.int, showISO C.int) *C.uchar {
	img := utils.ImageEdit(C.GoString(path), showLogo == 1, showF == 1, showExposureTime == 1, showISO == 1, 1000)
	if img == nil {
		*outLength = 0
		return nil
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		*outLength = 0
		return nil
	}
	data := buf.Bytes()
	*outLength = C.int(len(data))
	ptr := C.malloc(C.size_t(len(data)))
	C.memcpy(ptr, unsafe.Pointer(&data[0]), C.size_t(len(data)))

	return (*C.uchar)(ptr)
}

//export ImageSave
func ImageSave(path *C.char, output *C.char, showLogo C.int, showF C.int, showExposureTime C.int, showISO C.int) {
	utils.ImageSave(C.GoString(path), C.GoString(output), showLogo == 1, showF == 1, showExposureTime == 1, showISO == 1)
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

func main() {}
