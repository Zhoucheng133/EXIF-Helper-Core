package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"os"

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

//export GetEXIF
func GetEXIF(path *C.char) *C.char {
	f, err := os.Open(C.GoString(path))
	if err != nil {
		return C.CString(err.Error())
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return C.CString(err.Error())
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
	return C.CString(string(data))
}

func main() {

}
