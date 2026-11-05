package w32

import (
	"syscall"
	"unsafe"
)

type ID2D1Factory struct {
	Vtbl *ID2D1FactoryVtbl
}

// ID2D1FactoryVtbl is a VTable for ID2D1Factory.
// All the fields were tested with C++ VTable tester.
type ID2D1FactoryVtbl struct {
	// IUnknown methods
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	// ID2D1Factory methods
	ReloadSystemMetrics            uintptr
	GetDesktopDpi                  uintptr
	CreateRectangleGeometry        uintptr
	CreateRoundedRectangleGeometry uintptr
	CreateEllipseGeometry          uintptr
	CreateGeometryGroup            uintptr
	CreateTransformedGeometry      uintptr
	CreatePathGeometry             uintptr
	CreateStrokeStyle              uintptr
	CreateDrawingStateBlock        uintptr
	CreateWicBitmapRenderTarget    uintptr
	CreateHwndRenderTarget         uintptr
	CreateDxgiSurfaceRenderTarget  uintptr
	CreateDCRenderTarget           uintptr
}

func (f *ID2D1Factory) Release() ULONG {
	r0, _, _ := syscall.SyscallN(f.Vtbl.Release, uintptr(unsafe.Pointer(f)))
	return ULONG(r0)
}
func (f *ID2D1Factory) CreateHwndRenderTarget(renderTargetProperties *D2D1_RENDER_TARGET_PROPERTIES, hwndRenderTargetProperties *D2D1_HWND_RENDER_TARGET_PROPERTIES, hwndRenderTarget **ID2D1HwndRenderTarget) (hr HRESULT, err error) {
	r0, _, _ := syscall.SyscallN(f.Vtbl.CreateHwndRenderTarget, uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(renderTargetProperties)), uintptr(unsafe.Pointer(hwndRenderTargetProperties)), uintptr(unsafe.Pointer(hwndRenderTarget)))
	hr = HRESULT(r0)
	if hr != S_OK {
		err = syscall.Errno(hr)
	}
	return
}
