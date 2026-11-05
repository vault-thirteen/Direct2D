package w32

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type BYTE = byte
type BOOL = int32
type WORD = uint16
type DWORD = uint32
type FLOAT = float32
type INT = int32
type UINT = uint32
type UCHAR = uint8
type USHORT = uint16
type ULONG = uint32

type WPARAM = UINT_PTR
type LPARAM = LONG_PTR
type LRESULT = LONG_PTR
type HINSTANCE = windows.Handle
type ATOM = WORD

// LPVOID is a Long Pointer to Void.
// typedef void far *LPVOID;
type LPVOID = unsafe.Pointer

func LOWORD(l uintptr) WORD {
	return WORD(l & 0xffff)
}
func HIWORD(l uintptr) WORD {
	return WORD((l >> 16) & 0xffff)
}
