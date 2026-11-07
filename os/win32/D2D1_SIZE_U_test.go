package w32

import (
	"testing"
	"unsafe"

	"github.com/vault-thirteen/auxie/tester"
)

func Test_D2D_SIZE_U_ToUint64(t *testing.T) {
	aTest := tester.New(t)

	var su = D2D_SIZE_U{Width: 123, Height: 456}
	var su64 = su.ToUint64()
	var su64expected uint64 = uint64(456<<32) | uint64(123)
	aTest.MustBeEqual(su64, su64expected)
}
func Test_Uint64To_D2D_SIZE_U(t *testing.T) {
	aTest := tester.New(t)

	var x uint64 = uint64(456<<32) | uint64(123)
	var su = Uint64To_D2D_SIZE_U(x)
	var suExpected = D2D_SIZE_U{Width: 123, Height: 456}
	aTest.MustBeEqual(su, suExpected)
}
func Test_D2D_SIZE_U_ToUintptr(t *testing.T) {
	aTest := tester.New(t)

	var su = D2D_SIZE_U{Width: 123, Height: 456}
	var x = su.ToUintptr()
	var suExpected = *(*D2D_SIZE_U)(unsafe.Pointer(&x))
	aTest.MustBeEqual(su, suExpected)
	aTest.MustBeEqual(x, uintptr(su.ToUint64()))
}
func Test_UintptrTo_D2D_SIZE_U(t *testing.T) {
	aTest := tester.New(t)

	var su = D2D_SIZE_U{Width: 123, Height: 456}
	var x = uintptr(su.ToUintptr())
	aTest.MustBeEqual(UintptrTo_D2D_SIZE_U(x), su)
	aTest.MustBeEqual(Uint64To_D2D_SIZE_U(uint64(x)), su)
}
