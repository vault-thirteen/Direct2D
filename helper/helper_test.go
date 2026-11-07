package h

import (
	"testing"

	img "github.com/vault-thirteen/Direct2D/Image"
	"github.com/vault-thirteen/auxie/tester"
)

func Test_SaveBytesToFile(t *testing.T) {
	aTest := tester.New(t)
	srcFilePath := `..\img\Row_of_Town_Houses.jpg`
	dstFilePath := `..\test\pixels.bin`

	image, err := img.NewImageFromFile(srcFilePath, false)
	aTest.MustBeNoError(err)

	var pixelData []byte
	pixelData, err = image.PixelDataAsBGRA()
	aTest.MustBeNoError(err)

	err = SaveBytesToFile(dstFilePath, pixelData)
	aTest.MustBeNoError(err)
}
