package w32

import (
	"math"
	"unsafe"
)

type D2D1_SIZE_F = D2D_SIZE_F

func (sf D2D_SIZE_F) ToUint64() (out uint64) {
	wBits := uint64(math.Float32bits(sf.Width))
	hBits := uint64(math.Float32bits(sf.Height))
	return ((hBits) << 32) | wBits
}
func Uint64To_D2D_SIZE_F(in uint64) (out D2D_SIZE_F) {
	wBits := uint32(in & 0xFFFFFFFF)
	hBits := uint32(in >> 32)
	return D2D1_SIZE_F{
		Width:  math.Float32frombits(wBits),
		Height: math.Float32frombits(hBits),
	}
}
func (sf D2D_SIZE_F) ToUintptr() (out uintptr) {
	return *(*uintptr)(unsafe.Pointer(&sf))
}
func UintptrTo_D2D_SIZE_F(in uintptr) (out D2D_SIZE_F) {
	return *(*D2D_SIZE_F)(unsafe.Pointer(&in))
}
