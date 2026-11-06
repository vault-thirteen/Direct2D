package w32

import (
	"math"
	"testing"
	"unsafe"

	"github.com/vault-thirteen/auxie/tester"
)

func Test_D2D_SIZE_U_ToUintptr(t *testing.T) {
	aTest := tester.New(t)

	var su = D2D_SIZE_U{Width: 123, Height: 456}
	var x = uintptr(su.ToUint64())
	var unpackedSU D2D_SIZE_U = *(*D2D_SIZE_U)(unsafe.Pointer(&x))
	aTest.MustBeEqual(su, unpackedSU)

	widthBits := uint32(x & 0xFFFFFFFF)
	heightBits := uint32(x >> 32)
	width := UINT32(widthBits)
	height := UINT32(heightBits)
	aTest.MustBeEqual(width, UINT32(123))
	aTest.MustBeEqual(height, UINT32(456))
}
func Test_UintptrTo_D2D_SIZE_U(t *testing.T) {
	aTest := tester.New(t)

	var su = D2D_SIZE_U{Width: 123, Height: 456}
	var x = uintptr(su.ToUint64())
	aTest.MustBeEqual(UintptrTo_D2D_SIZE_U(x), su)

	widthBits := uint32(x & 0xFFFFFFFF)
	heightBits := uint32(x >> 32)
	width := UINT32(widthBits)
	height := UINT32(heightBits)
	aTest.MustBeEqual(width, UINT32(123))
	aTest.MustBeEqual(height, UINT32(456))
}

func Test_D2D_SIZE_F_ToUintptr(t *testing.T) {
	aTest := tester.New(t)

	var sf = D2D_SIZE_F{Width: 123.01, Height: 456.02}
	var x = uintptr(sf.ToUint64())
	var unpackedSF D2D_SIZE_F = *(*D2D_SIZE_F)(unsafe.Pointer(&x))
	aTest.MustBeEqual(sf, unpackedSF)

	widthBits := uint32(x & 0xFFFFFFFF)
	heightBits := uint32(x >> 32)
	width := math.Float32frombits(widthBits)
	height := math.Float32frombits(heightBits)
	aTest.MustBeEqual(width, FLOAT(123.01))
	aTest.MustBeEqual(height, FLOAT(456.02))
}
func Test_UintptrTo_D2D_SIZE_F(t *testing.T) {
	aTest := tester.New(t)

	var sf = D2D_SIZE_F{Width: 123.01, Height: 456.02}
	var x = uintptr(sf.ToUint64())
	aTest.MustBeEqual(UintptrTo_D2D_SIZE_F(x), sf)

	widthBits := uint32(x & 0xFFFFFFFF)
	heightBits := uint32(x >> 32)
	width := math.Float32frombits(widthBits)
	height := math.Float32frombits(heightBits)
	aTest.MustBeEqual(width, FLOAT(123.01))
	aTest.MustBeEqual(height, FLOAT(456.02))
}

func Test_D2D1_SIZE_U(t *testing.T) {
	aTest := tester.New(t)

	x := D2D1_SIZE_U{Width: 1000000001, Height: 2000000002}
	aTest.MustBeEqual(int(unsafe.Sizeof(x)), 8)
}
