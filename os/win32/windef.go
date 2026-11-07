package w32

import "golang.org/x/sys/windows"

type HBITMAP = windows.Handle
type HBRUSH = windows.Handle
type HCOLORSPACE = windows.Handle
type HDC = windows.Handle
type HGLRC = windows.Handle
type HDESK = windows.Handle
type HENHMETAFILE = windows.Handle
type HFONT = windows.Handle
type HICON = windows.Handle
type HMENU = windows.Handle
type HPALETTE = windows.Handle
type HPEN = windows.Handle
type HWINEVENTHOOK = windows.Handle
type HMONITOR = windows.Handle
type HUMPD = windows.Handle
type LPRECT = *RECT

type HWND = windows.Handle

type HCURSOR = HICON

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type POINT struct {
	X LONG
	Y LONG
}
