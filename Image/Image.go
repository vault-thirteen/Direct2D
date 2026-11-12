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

	e "github.com/vault-thirteen/auxie/os/win32/enum"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"

	"github.com/vault-thirteen/auxie/errors"
)

const (
	Err_Bounds       = "bounds error"
	Err_NotSupported = "not supported"
)

type Image struct {
	FilePath         string
	Name             string
	Size             ImageSize
	FormatName       string
	ColourModel      color.Model
	ColourModelType  ColourModelType
	AlphaChannelMode e.D2D1_ALPHA_MODE
	src              *image.Image
}

func NewImageFromFile(filePath string, useStraightAlphaChannel bool) (i *Image, err error) {
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

	var srcImg image.Image
	var formatName string
	srcImg, formatName, err = image.Decode(f)
	if err != nil {
		return nil, err
	}

	var x image.Image
	if useStraightAlphaChannel {
		x = convertImageToNRgba(srcImg)
	} else {
		x = convertImageToRgba(srcImg)
	}
	if x.Bounds() != srcImg.Bounds() {
		return nil, errors.ErrorFromText(Err_Bounds)
	}

	i = &Image{
		FilePath: filePath,
		Name:     fileNameWithoutExtension(fileName),
		Size: ImageSize{
			Width:  x.Bounds().Dx(),
			Height: x.Bounds().Dy(),
		},
		FormatName:       formatName,
		ColourModel:      x.ColorModel(),
		ColourModelType:  getColourModelType(x.ColorModel()),
		AlphaChannelMode: e.D2D1_ALPHA_MODE_UNKNOWN,
		src:              &x,
	}

	if useStraightAlphaChannel {
		i.AlphaChannelMode = e.D2D1_ALPHA_MODE_STRAIGHT
	} else {
		i.AlphaChannelMode = e.D2D1_ALPHA_MODE_PREMULTIPLIED
	}

	return i, nil
}

// PixelData returns raw pixel data in RGBA order if supported.
func (i *Image) PixelData() ([]byte, []byte, []byte, []byte) {
	switch i.ColourModelType {
	case ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Pix, nil, nil, nil

	case ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Pix, nil, nil, nil

	case ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Pix, nil, nil, nil

	case ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Pix, nil, nil, nil

	case ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Pix, nil, nil, nil

	case ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Pix, nil, nil, nil

	case ColourModelType_Gray:
		return (*i.src).(*image.Gray).Pix, nil, nil, nil

	case ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Pix, nil, nil, nil

	case ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return x.Y, x.Cb, x.Cr, nil
		}

	case ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return x.Y, x.Cb, x.Cr, x.A
		}

	case ColourModelType_CMYK:
		return (*i.src).(*image.CMYK).Pix, nil, nil, nil

	default:
		return nil, nil, nil, nil
	}
}

// PixelDataAsBGRA returns raw pixel data in BGRA order if supported.
func (i *Image) PixelDataAsBGRA() (pixelData []byte, err error) {

	switch i.ColourModelType {
	case ColourModelType_RGBA:
		return ConvertPixelDataFromRgbaToBgra((*i.src).(*image.RGBA).Pix)

	case ColourModelType_NRGBA:
		return ConvertPixelDataFromRgbaToBgra((*i.src).(*image.NRGBA).Pix)

	default:
		return nil, errors.ErrorFromText(Err_NotSupported)
	}
}

// PixelDataSize returns the size in bytes of raw pixel data in RGBA order if
// supported.
func (i *Image) PixelDataSize() int {
	switch i.ColourModelType {
	case ColourModelType_RGBA:
		return len((*i.src).(*image.RGBA).Pix)

	case ColourModelType_NRGBA:
		return len((*i.src).(*image.NRGBA).Pix)

	case ColourModelType_RGBA64:
		return len((*i.src).(*image.RGBA64).Pix)

	case ColourModelType_NRGBA64:
		return len((*i.src).(*image.NRGBA64).Pix)

	case ColourModelType_Alpha:
		return len((*i.src).(*image.Alpha).Pix)

	case ColourModelType_Alpha16:
		return len((*i.src).(*image.Alpha16).Pix)

	case ColourModelType_Gray:
		return len((*i.src).(*image.Gray).Pix)

	case ColourModelType_Gray16:
		return len((*i.src).(*image.Gray16).Pix)

	case ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return len(x.Y) + len(x.Cb) + len(x.Cr)
		}

	case ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return len(x.Y) + len(x.Cb) + len(x.Cr) + len(x.A)
		}

	case ColourModelType_CMYK:
		return len((*i.src).(*image.CMYK).Pix)

	default:
		return 0
	}
}

// LineStride returns the line stride, i.e. the size in bytes of each row of
// pixels if supported.
func (i *Image) LineStride() (int, int, int) {
	switch i.ColourModelType {
	case ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Stride, 0, 0

	case ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Stride, 0, 0

	case ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Stride, 0, 0

	case ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Stride, 0, 0

	case ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Stride, 0, 0

	case ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Stride, 0, 0

	case ColourModelType_Gray:
		return (*i.src).(*image.Gray).Stride, 0, 0

	case ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Stride, 0, 0

	case ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return x.YStride, x.CStride, 0
		}

	case ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return x.YStride, x.CStride, x.AStride
		}

	case ColourModelType_CMYK:
		return (*i.src).(*image.CMYK).Stride, 0, 0

	default:
		return 0, 0, 0
	}
}

// Rect returns image's bounds.
func (i *Image) Rect() image.Rectangle {
	switch i.ColourModelType {
	case ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Rect

	case ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Rect

	case ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Rect

	case ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Rect

	case ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Rect

	case ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Rect

	case ColourModelType_Gray:
		return (*i.src).(*image.Gray).Rect

	case ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Rect

	case ColourModelType_YCbCr:
		return (*i.src).(*image.YCbCr).Rect

	case ColourModelType_NYCbCrA:
		return (*i.src).(*image.NYCbCrA).Rect

	case ColourModelType_CMYK:
		return (*i.src).(*image.CMYK).Rect

	default:
		return image.Rectangle{}
	}
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
func convertImageToNRgba(src image.Image) (dst *image.NRGBA) {
	var ok bool
	dst, ok = src.(*image.NRGBA)
	if ok {
		return dst
	}

	bounds := src.Bounds()
	dst = image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
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
