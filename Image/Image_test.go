package img

import (
	"github.com/vault-thirteen/auxie/tester"

	"testing"
)

func Test_NewImageFromFile(t *testing.T) {
	aTest := tester.New(t)
	filePath := `..\img\Row_of_Town_Houses.jpg`
	image, err := NewImageFromFile(filePath)
	aTest.MustBeNoError(err)
	aTest.MustBeEqual(image.FilePath, filePath)
	aTest.MustBeEqual(image.Name, "Row_of_Town_Houses")
	aTest.MustBeEqual(image.Size, ImageSize{Width: 900, Height: 506})
	aTest.MustBeEqual(image.FormatName, "jpeg")
	aTest.MustBeEqual(image.ColourModelType, ColourModelType_RGBA)
	aTest.MustBeEqual(image.PixelDataSize(), 900*506*4)
	aTest.MustBeEqual(image.LineStride(), 900*4)
}
