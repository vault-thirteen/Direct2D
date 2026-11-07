package w32

import (
	"syscall"
	"unsafe"

	c "github.com/vault-thirteen/Direct2D/os/win32/common"
)

var IID_IUnknown = IID{
	Data1: 0x00000000,
	Data2: 0x0000,
	Data3: 0x0000,
	Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
}

type IUnknown struct {
	*c.VTable
}

func (x *IUnknown) QueryInterface(riid *IID, ppvObject *unsafe.Pointer) (hr HRESULT, err error) {
	r0, _, _ := syscall.SyscallN(x.VTable[0], uintptr(unsafe.Pointer(x)), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(ppvObject)))
	hr = HRESULT(r0)
	if hr.FAILED() {
		err = syscall.Errno(hr)
	}
	return
}

func (x *IUnknown) AddRef() ULONG {
	r0, _, _ := syscall.SyscallN(x.VTable[1], uintptr(unsafe.Pointer(x)))
	return ULONG(r0)
}

func (x *IUnknown) Release() ULONG {
	r0, _, _ := syscall.SyscallN(x.VTable[2], uintptr(unsafe.Pointer(x)))
	return ULONG(r0)
}
