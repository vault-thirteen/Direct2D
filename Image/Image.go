package img

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"  // For the GIF format decoder.
	_ "image/jpeg" // For the JPEG format decoder.
	_ "image/png"  // For the PNG format decoder.
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/vault-thirteen/auxie/errors"
)

const (
	Err_Bounds = "bounds error"
)

type Image struct {
	FilePath        string
	Name            string
	Size            ImageSize
	FormatName      string
	ColourModel     color.Model
	ColourModelType ColourModelType
	src             *image.RGBA
}

func NewImageFromFile(filePath string) (i *Image, err error) {
	var f *os.File
	f, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		derr := f.Close()
		err = errors.Combine(err, derr)
	}()
	fileName := filepath.Base(filePath)

	var img image.Image
	var formatName string
	img, formatName, err = image.Decode(f)
	if err != nil {
		return nil, err
	}

	rgbaImg := convertImageToRgba(img)
	if rgbaImg.Bounds() != img.Bounds() {
		return nil, errors.ErrorFromText(Err_Bounds)
	}

	i = &Image{
		FilePath: filePath,
		Name:     fileNameWithoutExtension(fileName),
		Size: ImageSize{
			Width:  rgbaImg.Bounds().Dx(),
			Height: rgbaImg.Bounds().Dy(),
		},
		FormatName:      formatName,
		ColourModel:     rgbaImg.ColorModel(),
		ColourModelType: getColourModelType(rgbaImg.ColorModel()),
		src:             rgbaImg,
	}

	return i, nil
}

// PixelData returns raw pixel data in RGBA order.
func (i *Image) PixelData() []byte {
	return i.src.Pix
}

// PixelDataSize returns the size in bytes of raw pixel data in RGBA order.
func (i *Image) PixelDataSize() int {
	return len(i.src.Pix)
}

// LineStride returns the line stride, i.e. the size in bytes of each row of
// pixels.
func (i *Image) LineStride() int {
	return i.src.Stride
}

// Rect returns image's bounds.
func (i *Image) Rect() image.Rectangle {
	return i.src.Rect
}

func convertImageToRgba(src image.Image) (dst *image.RGBA) {
	var ok bool
	dst, ok = src.(*image.RGBA)
	if ok {
		return dst
	}

	bounds := src.Bounds()
	dst = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)

	return dst
}
func fileNameWithoutExtension(fileName string) (fnwo string) {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
func getColourModelType(cm color.Model) (cmt ColourModelType) {
	switch cm {
	case color.RGBAModel:
		return ColourModelType_RGBA
	case color.RGBA64Model:
		return ColourModelType_RGBA64
	case color.NRGBAModel:
		return ColourModelType_NRGBA
	case color.NRGBA64Model:
		return ColourModelType_NRGBA64
	case color.AlphaModel:
		return ColourModelType_Alpha
	case color.Alpha16Model:
		return ColourModelType_Alpha16
	case color.GrayModel:
		return ColourModelType_Gray
	case color.Gray16Model:
		return ColourModelType_Gray16
	case color.YCbCrModel:
		return ColourModelType_YCbCr
	case color.NYCbCrAModel:
		return ColourModelType_NYCbCrA
	case color.CMYKModel:
		return ColourModelType_CMYK
	default:
		return ColourModelType_Unknown
	}
}
