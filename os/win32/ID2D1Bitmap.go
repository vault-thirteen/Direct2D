package w32

import (
	"syscall"
	"unsafe"
)

type ID2D1Bitmap struct {
	Vtbl *ID2D1BitmapVtbl
}

// ID2D1BitmapVtbl is a VTable for ID2D1Bitmap.
// All the fields were tested with C++ VTable tester.
type ID2D1BitmapVtbl struct {
	// IUnknown methods
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// ID2D1Resource methods
	GetFactory uintptr

	// ID2D1Image methods (if compiling for newer Direct2D versions, otherwise flat)
	// ID2D1Bitmap methods
	GetSize              uintptr
	GetPixelSize         uintptr
	GetPixelFormat       uintptr
	GetDpi               uintptr
	CopyFromBitmap       uintptr
	CopyFromRenderTarget uintptr
	CopyFromMemory       uintptr
}

func (b *ID2D1Bitmap) Release() ULONG {
	ret, _, _ := syscall.SyscallN(b.Vtbl.Release, uintptr(unsafe.Pointer(b)))
	return ULONG(ret)
}
func (b *ID2D1Bitmap) GetSize() (sf D2D1_SIZE_F) {
	r0, _, _ := syscall.SyscallN(b.Vtbl.GetSize, uintptr(unsafe.Pointer(b)))
	return UintptrTo_D2D_SIZE_F(r0)
}
func (b *ID2D1Bitmap) GetPixelSize() (su D2D1_SIZE_U) {
	r0, _, _ := syscall.SyscallN(b.Vtbl.GetPixelSize, uintptr(unsafe.Pointer(b)))
	return UintptrTo_D2D_SIZE_U(r0)
}
func (b *ID2D1Bitmap) CopyFromMemory(dstRect *D2D1_RECT_U, srcData []byte, pitch UINT32) (hr HRESULT, err error) {
	r0, _, _ := syscall.SyscallN(b.Vtbl.CopyFromMemory, uintptr(unsafe.Pointer(b)), uintptr(unsafe.Pointer(dstRect)), uintptr(unsafe.Pointer(unsafe.SliceData(srcData))), uintptr(pitch))
	hr = HRESULT(r0)
	if hr != S_OK {
		err = syscall.Errno(hr)
	}
	return
}
