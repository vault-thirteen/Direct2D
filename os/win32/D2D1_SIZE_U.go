package w32

import "unsafe"

type D2D1_SIZE_U = D2D_SIZE_U

func (su D2D_SIZE_U) ToUint64() (out uint64) {
	return (uint64(su.Height) << 32) | uint64(su.Width)
}
func Uint64To_D2D_SIZE_U(in uint64) (out D2D_SIZE_U) {
	return D2D1_SIZE_U{
		Width:  uint32(in & 0xFFFFFFFF),
		Height: uint32(in >> 32),
	}
}
func (su D2D_SIZE_U) ToUintptr() (out uintptr) {
	return *(*uintptr)(unsafe.Pointer(&su))
}
func UintptrTo_D2D_SIZE_U(in uintptr) (out D2D_SIZE_U) {
	return *(*D2D_SIZE_U)(unsafe.Pointer(&in))
}
