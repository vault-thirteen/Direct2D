package w32

import "unsafe"

//go:generate mkwinsyscall -output zsyscall_windows.go win32.go

// Functions that set last error:
/*
//sys LoadCursorW(hInstance HINSTANCE, lpCursorName *uint16) (hc HCURSOR, err error)[failretval==0] = user32.LoadCursorW
//sys RegisterClassExW(unnamedParam1 *WNDCLASSEXW_Golang) (atom ATOM, err error)[failretval==0] = user32.RegisterClassExW
//sys UnregisterClassW(lpClassName *uint16, hInstance HINSTANCE)(b BOOL, err error)[failretval==0] = user32.UnregisterClassW
//sys CreateWindowExW(dwExStyle DWORD, lpClassName *uint16, lpWindowName *uint16, dwStyle DWORD, x int32, y int32, nWidth int32, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) (hWnd HWND, err error)[failretval==0] = user32.CreateWindowExW
//sys DestroyWindow(hWnd HWND) (b BOOL, err error)[failretval==0] = user32.DestroyWindow
//sys GetMessageW(lpMsg *MSG, hWnd HWND, wMsgFilterMin UINT, wMsgFilterMax UINT)(b BOOL, err error)[failretval==-1] = user32.GetMessageW
//sys GetWindowLongPtrW(hWnd HWND, nIndex int32) (lp LONG_PTR, err error)[failretval==0] = user32.GetWindowLongPtrW
//sys GetClientRect(hWnd HWND, lpRect *RECT)(b BOOL, err error)[failretval==0] = user32.GetClientRect
*/

// Functions that do not set last error:
/*
//sys D2D1CreateFactory(factoryType D2D1_FACTORY_TYPE, riid *IID, pFactoryOptions *D2D1_FACTORY_OPTIONS, ppIFactory **ID2D1Factory) (hr HRESULT) = d2d1.D2D1CreateFactory
//sys ShowWindow(hWnd HWND, nCmdShow int32) (b BOOL) = user32.ShowWindow
//sys UpdateWindow(hWnd HWND)(b BOOL) = user32.UpdateWindow
//sys TranslateMessage(lpMsg *MSG) (b BOOL) = user32.TranslateMessage
//sys DispatchMessageW(lpMsg *MSG) (lr LRESULT) = user32.DispatchMessageW
//sys PeekMessageW(lpMsg *MSG, hWnd HWND, wMsgFilterMin UINT, wMsgFilterMax UINT, wRemoveMsg UINT)(b BOOL) = user32.PeekMessageW
//sys ValidateRect(hWnd HWND, lpRect *RECT)(b BOOL) = user32.ValidateRect
//sys InvalidateRect(hWnd HWND, lpRect *RECT, bErase BOOL)(b BOOL)[failretval==0] = user32.InvalidateRect
//sys DefWindowProcW(hWnd HWND, Msg WM, wParam WPARAM, lParam LPARAM)(lr LRESULT) = user32.DefWindowProcW
//sys PostQuitMessage(nExitCode int32) = user32.PostQuitMessage
//sys CoCreateInstance(rclsid *windows.GUID, pUnkOuter *IUnknown, dwClsContext DWORD, riid *windows.GUID, ppv *unsafe.Pointer) (hr HRESULT) = ole32.CoCreateInstance
//sys GetKeyState(nVirtKey int32)(ks SHORT) = user32.GetKeyState
*/

// Non-standard logics. Be very careful using these functions.
/*
//sys SetLastError(dwErrCode DWORD) = kernel32.SetLastError
//sys SetWindowLongPtrW(hWnd HWND, nIndex int32, dwNewLong LONG_PTR)(lp LONG_PTR) = user32.SetWindowLongPtrW
*/

func StructArg[T any](v T) (x uintptr) {
	if unsafe.Sizeof(v) > unsafe.Sizeof(x) {
		panic("win32: by-value struct is too large to pass in a register")
	}
	*(*T)(unsafe.Pointer(&x)) = v
	return x
}
