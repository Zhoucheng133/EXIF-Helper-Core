package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	goexif3 "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
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

func getTagString(tags map[string]goexif3.ExifTag, name string) string {
	tag, ok := tags[name]
	if !ok {
		return ""
	}
	return strings.TrimSpace(tag.FormattedFirst)
}

func buildTagMap(entries []goexif3.ExifTag) map[string]goexif3.ExifTag {
	tags := make(map[string]goexif3.ExifTag, len(entries))
	for _, e := range entries {
		if _, exists := tags[e.TagName]; !exists {
			tags[e.TagName] = e
		}
	}
	return tags
}

func formatExif(entries []goexif3.ExifTag) EXIFInfo {
	tags := buildTagMap(entries)

	res := EXIFInfo{
		CamMake:      getTagString(tags, "Make"),
		CamModel:     getTagString(tags, "Model"),
		LenMake:      getTagString(tags, "LensMake"),
		LenModel:     getTagString(tags, "LensModel"),
		CaptureTime:  getTagString(tags, "DateTimeOriginal"),
		ExposureTime: evalExposure(getTagString(tags, "ExposureTime")),
		Fnum:         evalFloat(getTagString(tags, "FNumber")),
		Iso:          getTagString(tags, "ISOSpeedRatings"),
		Focal:        evalFloat(getTagString(tags, "FocalLength")),
		Focal35:      evalFloat(getTagString(tags, "FocalLengthIn35mmFilm")),
		Orientation:  getTagString(tags, "Orientation"),
	}

	// ---------- GPS 经纬度 ----------
	latTag, hasLat := tags["GPSLatitude"]
	latRefTag, hasLatRef := tags["GPSLatitudeRef"]
	lonTag, hasLon := tags["GPSLongitude"]
	lonRefTag, hasLonRef := tags["GPSLongitudeRef"]

	if hasLat && hasLatRef {
		if latRationals, ok := latTag.Value.([]exifcommon.Rational); ok && len(latRationals) == 3 {
			if deg, err := goexif3.NewGpsDegreesFromRationals(strings.TrimSpace(latRefTag.FormattedFirst), latRationals); err == nil {
				lat := deg.Decimal()
				res.Latitude = &lat
			}
		}
	}

	if hasLon && hasLonRef {
		if lonRationals, ok := lonTag.Value.([]exifcommon.Rational); ok && len(lonRationals) == 3 {
			if deg, err := goexif3.NewGpsDegreesFromRationals(strings.TrimSpace(lonRefTag.FormattedFirst), lonRationals); err == nil {
				lon := deg.Decimal()
				res.Longitude = &lon
			}
		}
	}

	// ---------- GPS 海拔 ----------
	if altTag, ok := tags["GPSAltitude"]; ok {
		if altRationals, ok := altTag.Value.([]exifcommon.Rational); ok && len(altRationals) > 0 && altRationals[0].Denominator != 0 {
			alt := float64(altRationals[0].Numerator) / float64(altRationals[0].Denominator)
			if altRefTag, ok := tags["GPSAltitudeRef"]; ok {
				if refBytes, ok := altRefTag.Value.([]byte); ok && len(refBytes) > 0 && refBytes[0] == 1 {
					alt = -alt
				}
			}
			res.Altitude = &alt
		}
	}

	return res
}
