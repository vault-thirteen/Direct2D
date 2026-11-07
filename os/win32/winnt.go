package w32

import "syscall"

/*
HRESULT is a numeric code of result used in Microsoft Windows API.

Definition in 'winnt.h' file:

	typedef _Return_type_success_(return >= 0) long HRESULT;

Helper macros in 'intsafe.h' file:

	#define SUCCEEDED(hr)   (((HRESULT)(hr)) >= 0)
	#define FAILED(hr)      (((HRESULT)(hr)) < 0)

Notes

	It is critically important that Microsoft uses 'HRESULT' as a signed
	integer. Go language's 'uintptr' type is an unsigned integer which is not
	directly compatible with Microsoft's API !

	It is crucially important first to cast type of values returned by
	Microsoft's API into a signed integer before any further processing !
*/
type HRESULT int32

func (hr HRESULT) Error() string {
	return syscall.Errno(hr).Error()
}

func unchecked_hresult(val uint32) HRESULT {
	return HRESULT(val)
}

type LPCWSTR = *uint16
type CHAR = byte
type SHORT = int16
type LONG = int32
