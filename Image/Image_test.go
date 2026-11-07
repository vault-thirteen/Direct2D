package img

import (
	w32 "github.com/vault-thirteen/Direct2D/os/win32"
	"github.com/vault-thirteen/auxie/tester"

	"testing"
)

func Test_NewImageFromFile(t *testing.T) {
	aTest := tester.New(t)
	filePath := `..\img\Row_of_Town_Houses.jpg`
	image, err := NewImageFromFile(filePath, true)
	aTest.MustBeNoError(err)
	aTest.MustBeEqual(image.FilePath, filePath)
	aTest.MustBeEqual(image.Name, "Row_of_Town_Houses")
	aTest.MustBeEqual(image.Size, ImageSize{Width: 900, Height: 506})
	aTest.MustBeEqual(image.FormatName, "jpeg")
	aTest.MustBeEqual(image.ColourModelType, ColourModelType_NRGBA)
	aTest.MustBeEqual(image.AlphaChannelMode, w32.D2D1_ALPHA_MODE_STRAIGHT)
	aTest.MustBeEqual(image.PixelDataSize(), 900*506*4)

	var lineStride int
	lineStride, _, _ = image.LineStride()
	aTest.MustBeEqual(lineStride, 900*4)
}
