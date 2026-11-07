package w32

import (
	"testing"
	"unsafe"

	"github.com/vault-thirteen/auxie/tester"
)

func Test_D2D_SIZE_U(t *testing.T) {
	aTest := tester.New(t)

	x := D2D_SIZE_U{Width: 1000000001, Height: 2000000002}
	aTest.MustBeEqual(int(unsafe.Sizeof(x)), 8)
}
func Test_D2D_SIZE_F(t *testing.T) {
	aTest := tester.New(t)

	x := D2D_SIZE_F{Width: 11.11, Height: 22.22}
	aTest.MustBeEqual(int(unsafe.Sizeof(x)), 8)
}
