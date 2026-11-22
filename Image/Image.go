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

	c "github.com/vault-thirteen/auxie/colour"
	"github.com/vault-thirteen/auxie/errors"
	e "github.com/vault-thirteen/auxie/os/win32/enum"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
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
	ColourModelType  c.ColourModelType
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
		ColourModelType:  c.GetColourModelType(x.ColorModel()),
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
	case c.ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Pix, nil, nil, nil

	case c.ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Pix, nil, nil, nil

	case c.ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Pix, nil, nil, nil

	case c.ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Pix, nil, nil, nil

	case c.ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Pix, nil, nil, nil

	case c.ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Pix, nil, nil, nil

	case c.ColourModelType_Gray:
		return (*i.src).(*image.Gray).Pix, nil, nil, nil

	case c.ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Pix, nil, nil, nil

	case c.ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return x.Y, x.Cb, x.Cr, nil
		}

	case c.ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return x.Y, x.Cb, x.Cr, x.A
		}

	case c.ColourModelType_CMYK:
		return (*i.src).(*image.CMYK).Pix, nil, nil, nil

	default:
		return nil, nil, nil, nil
	}
}

// PixelDataAsBGRA returns raw pixel data in BGRA order if supported.
func (i *Image) PixelDataAsBGRA() (pixelData []byte, err error) {
	switch i.ColourModelType {
	case c.ColourModelType_RGBA:
		return ConvertPixelDataFromRgbaToBgra((*i.src).(*image.RGBA).Pix)

	case c.ColourModelType_NRGBA:
		return ConvertPixelDataFromRgbaToBgra((*i.src).(*image.NRGBA).Pix)

	default:
		return nil, errors.ErrorFromText(Err_NotSupported)
	}
}

// PixelDataSize returns the size in bytes of raw pixel data in RGBA order if
// supported.
func (i *Image) PixelDataSize() int {
	switch i.ColourModelType {
	case c.ColourModelType_RGBA:
		return len((*i.src).(*image.RGBA).Pix)

	case c.ColourModelType_NRGBA:
		return len((*i.src).(*image.NRGBA).Pix)

	case c.ColourModelType_RGBA64:
		return len((*i.src).(*image.RGBA64).Pix)

	case c.ColourModelType_NRGBA64:
		return len((*i.src).(*image.NRGBA64).Pix)

	case c.ColourModelType_Alpha:
		return len((*i.src).(*image.Alpha).Pix)

	case c.ColourModelType_Alpha16:
		return len((*i.src).(*image.Alpha16).Pix)

	case c.ColourModelType_Gray:
		return len((*i.src).(*image.Gray).Pix)

	case c.ColourModelType_Gray16:
		return len((*i.src).(*image.Gray16).Pix)

	case c.ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return len(x.Y) + len(x.Cb) + len(x.Cr)
		}

	case c.ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return len(x.Y) + len(x.Cb) + len(x.Cr) + len(x.A)
		}

	case c.ColourModelType_CMYK:
		return len((*i.src).(*image.CMYK).Pix)

	default:
		return 0
	}
}

// LineStride returns the line stride, i.e. the size in bytes of each row of
// pixels if supported.
func (i *Image) LineStride() (int, int, int) {
	switch i.ColourModelType {
	case c.ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Stride, 0, 0

	case c.ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Stride, 0, 0

	case c.ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Stride, 0, 0

	case c.ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Stride, 0, 0

	case c.ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Stride, 0, 0

	case c.ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Stride, 0, 0

	case c.ColourModelType_Gray:
		return (*i.src).(*image.Gray).Stride, 0, 0

	case c.ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Stride, 0, 0

	case c.ColourModelType_YCbCr:
		{
			x := (*i.src).(*image.YCbCr)
			return x.YStride, x.CStride, 0
		}

	case c.ColourModelType_NYCbCrA:
		{
			x := (*i.src).(*image.NYCbCrA)
			return x.YStride, x.CStride, x.AStride
		}

	case c.ColourModelType_CMYK:
		return (*i.src).(*image.CMYK).Stride, 0, 0

	default:
		return 0, 0, 0
	}
}

// Rect returns image's bounds.
func (i *Image) Rect() image.Rectangle {
	switch i.ColourModelType {
	case c.ColourModelType_RGBA:
		return (*i.src).(*image.RGBA).Rect

	case c.ColourModelType_NRGBA:
		return (*i.src).(*image.NRGBA).Rect

	case c.ColourModelType_RGBA64:
		return (*i.src).(*image.RGBA64).Rect

	case c.ColourModelType_NRGBA64:
		return (*i.src).(*image.NRGBA64).Rect

	case c.ColourModelType_Alpha:
		return (*i.src).(*image.Alpha).Rect

	case c.ColourModelType_Alpha16:
		return (*i.src).(*image.Alpha16).Rect

	case c.ColourModelType_Gray:
		return (*i.src).(*image.Gray).Rect

	case c.ColourModelType_Gray16:
		return (*i.src).(*image.Gray16).Rect

	case c.ColourModelType_YCbCr:
		return (*i.src).(*image.YCbCr).Rect

	case c.ColourModelType_NYCbCrA:
		return (*i.src).(*image.NYCbCrA).Rect

	case c.ColourModelType_CMYK:
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
