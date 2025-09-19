package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

// 注意，针对光圈和焦距
func evalFloat(input string) string {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return input
	}
	a, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	b, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil || b == 0 {
		return input
	}
	result := a / b
	if math.Mod(result, 1.0) == 0 {
		return fmt.Sprintf("%.0f", result)
	}
	return fmt.Sprintf("%.1f", result)
}

// 针对快门速度
func evalExposure(input string) string {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return input
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return input
	}
	actual := num / den
	if actual >= 1.0 {
		if math.Mod(actual, 1.0) == 0 {
			return fmt.Sprintf("%.0f", actual)
		}
		return fmt.Sprintf("%.1f", actual)
	}
	x := math.Round(1.0 / actual)
	return fmt.Sprintf("1/%.0f", x)
}

func getTagString(x *exif.Exif, name exif.FieldName) string {
	tag, err := x.Get(name)
	if err != nil || tag == nil {
		return ""
	}
	str := strings.ReplaceAll(tag.String(), "\"", "")
	re := regexp.MustCompile(`\s+`)
	str = re.ReplaceAllString(str, " ")

	return strings.TrimSpace(str)
}

func formatExif(data *exif.Exif) EXIFInfo {
	res := EXIFInfo{
		CamMake:      getTagString(data, exif.Make),
		CamModel:     getTagString(data, exif.Model),
		LenMake:      getTagString(data, exif.LensMake),
		LenModel:     getTagString(data, exif.LensModel),
		CaptureTime:  getTagString(data, exif.DateTimeOriginal),
		ExposureTime: evalExposure(getTagString(data, exif.ExposureTime)),
		Fnum:         evalFloat(getTagString(data, exif.FNumber)),
		Iso:          getTagString(data, exif.ISOSpeedRatings),
		Focal:        evalFloat(getTagString(data, exif.FocalLength)),
		Focal35:      evalFloat(getTagString(data, exif.FocalLengthIn35mmFilm)),
		Orientation:  getTagString(data, exif.Orientation),
	}

	return res
}
