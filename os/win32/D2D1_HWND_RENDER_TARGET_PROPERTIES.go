package w32

type D2D1_HWND_RENDER_TARGET_PROPERTIES struct {
	Hwnd           HWND
	PixelSize      D2D1_SIZE_U
	PresentOptions D2D1_PRESENT_OPTIONS
}

func HwndRenderTargetProperties(hwnd HWND, pixelSize D2D1_SIZE_U, presentOptions D2D1_PRESENT_OPTIONS) D2D1_HWND_RENDER_TARGET_PROPERTIES {
	return D2D1_HWND_RENDER_TARGET_PROPERTIES{
		Hwnd:           hwnd,
		PixelSize:      pixelSize,
		PresentOptions: presentOptions,
	}
}
