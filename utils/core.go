package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	goexif3 "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	rwexif "github.com/rwcarlsen/goexif/exif"
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

	data, err := rwexif.Decode(f)

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

func RemoveExif(inputPath string, outputPath string) error {
	img, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	return imaging.Save(img, outputPath)
}

func EditEXIF(inputPath string, outputDir string, exifJSON string) error {
	var info EXIFInfo
	if err := json.Unmarshal([]byte(exifJSON), &info); err != nil {
		return fmt.Errorf("解析EXIF JSON失败: %w", err)
	}

	jmp := jis.NewJpegMediaParser()
	intfc, err := jmp.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("解析JPEG文件失败: %w", err)
	}
	sl := intfc.(*jis.SegmentList)

	// 尝试拿到已有EXIF builder；如果图片本身没有EXIF段则新建一个
	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		im, ierr := exifcommon.NewIfdMappingWithStandard()
		if ierr != nil {
			return fmt.Errorf("创建IFD映射失败: %w", ierr)
		}
		ti := goexif3.NewTagIndex()
		rootIb = goexif3.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)
	}

	exifIb, err := goexif3.GetOrCreateIbFromRootIb(rootIb, "IFD/Exif")
	if err != nil {
		return fmt.Errorf("获取Exif SubIFD失败: %w", err)
	}

	// ---------- IFD0 ----------
	if info.CamMake != "" {
		if err := rootIb.SetStandardWithName("Make", info.CamMake); err != nil {
			return fmt.Errorf("设置相机厂商失败: %w", err)
		}
	}
	if info.CamModel != "" {
		if err := rootIb.SetStandardWithName("Model", info.CamModel); err != nil {
			return fmt.Errorf("设置相机型号失败: %w", err)
		}
	}
	if info.Orientation != "" {
		o, err := strconv.ParseUint(info.Orientation, 10, 16)
		if err != nil {
			return fmt.Errorf("解析Orientation失败: %w", err)
		}
		if err := rootIb.SetStandardWithName("Orientation", []uint16{uint16(o)}); err != nil {
			return fmt.Errorf("设置Orientation失败: %w", err)
		}
	}

	// ---------- Exif SubIFD ----------
	if info.LenMake != "" {
		if err := exifIb.SetStandardWithName("LensMake", info.LenMake); err != nil {
			return fmt.Errorf("设置镜头厂商失败: %w", err)
		}
	}
	if info.LenModel != "" {
		if err := exifIb.SetStandardWithName("LensModel", info.LenModel); err != nil {
			return fmt.Errorf("设置镜头型号失败: %w", err)
		}
	}
	if info.CaptureTime != "" {
		t, err := normalizeExifTime(info.CaptureTime)
		if err != nil {
			return fmt.Errorf("解析拍摄时间失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("DateTimeOriginal", t); err != nil {
			return fmt.Errorf("设置DateTimeOriginal失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("DateTimeDigitized", t); err != nil {
			return fmt.Errorf("设置DateTimeDigitized失败: %w", err)
		}
		if err := rootIb.SetStandardWithName("DateTime", t); err != nil {
			return fmt.Errorf("设置DateTime失败: %w", err)
		}
	}
	if info.ExposureTime != "" {
		r, err := parseRational(info.ExposureTime)
		if err != nil {
			return fmt.Errorf("解析曝光时间失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("ExposureTime", []exifcommon.Rational{r}); err != nil {
			return fmt.Errorf("设置ExposureTime失败: %w", err)
		}
	}
	if info.Fnum != "" {
		r, err := parseRational(info.Fnum)
		if err != nil {
			return fmt.Errorf("解析光圈失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("FNumber", []exifcommon.Rational{r}); err != nil {
			return fmt.Errorf("设置FNumber失败: %w", err)
		}
	}
	if info.Iso != "" {
		iso, err := strconv.ParseUint(info.Iso, 10, 16)
		if err != nil {
			return fmt.Errorf("解析ISO失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("ISOSpeedRatings", []uint16{uint16(iso)}); err != nil {
			return fmt.Errorf("设置ISO失败: %w", err)
		}
	}
	if info.Focal != "" {
		r, err := parseRational(info.Focal)
		if err != nil {
			return fmt.Errorf("解析焦距失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("FocalLength", []exifcommon.Rational{r}); err != nil {
			return fmt.Errorf("设置FocalLength失败: %w", err)
		}
	}
	if info.Focal35 != "" {
		f35, err := strconv.ParseUint(info.Focal35, 10, 16)
		if err != nil {
			return fmt.Errorf("解析35mm等效焦距失败: %w", err)
		}
		if err := exifIb.SetStandardWithName("FocalLengthIn35mmFilm", []uint16{uint16(f35)}); err != nil {
			return fmt.Errorf("设置FocalLengthIn35mmFilm失败: %w", err)
		}
	}

	// 把修改后的EXIF写回segment list
	if err := sl.SetExif(rootIb); err != nil {
		return fmt.Errorf("写入EXIF到segment失败: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	outPath := filepath.Join(outputDir, filepath.Base(inputPath))
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	if err := sl.Write(outFile); err != nil {
		return fmt.Errorf("写出文件失败: %w", err)
	}

	return nil
}

func parseRational(s string) (exifcommon.Rational, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return exifcommon.Rational{}, fmt.Errorf("空值")
	}

	if strings.Contains(s, "/") {
		parts := strings.SplitN(s, "/", 2)
		num, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
		if err != nil {
			return exifcommon.Rational{}, fmt.Errorf("分子解析失败: %w", err)
		}
		den, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
		if err != nil {
			return exifcommon.Rational{}, fmt.Errorf("分母解析失败: %w", err)
		}
		if den == 0 {
			return exifcommon.Rational{}, fmt.Errorf("分母不能为0")
		}
		return exifcommon.Rational{Numerator: uint32(num), Denominator: uint32(den)}, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return exifcommon.Rational{}, fmt.Errorf("数值解析失败: %w", err)
	}

	const precision = 10000
	num := uint32(f * precision)
	den := uint32(precision)

	if g := gcd(num, den); g > 0 {
		num /= g
		den /= g
	}

	return exifcommon.Rational{Numerator: num, Denominator: den}, nil
}

func gcd(a, b uint32) uint32 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func normalizeExifTime(s string) (string, error) {
	s = strings.TrimSpace(s)

	layouts := []string{
		"2006:01:02 15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006:01:02 15:04:05"), nil
		}
	}

	return "", fmt.Errorf("无法识别的时间格式: %s", s)
}
