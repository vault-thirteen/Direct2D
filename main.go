package main

import (
	"unsafe"

	a "github.com/vault-thirteen/Direct2D/App"
	as "github.com/vault-thirteen/Direct2D/AppSettings"
	c "github.com/vault-thirteen/Direct2D/common"
	w32 "github.com/vault-thirteen/Direct2D/os/win32"
	"golang.org/x/sys/windows"
)

// This variable is a module marker, used for Windows API.
var mm byte = 255

func main() {
	appSettings := &as.AppSettings{
		Debug: as.AppDebugSettings{
			IsEnabled:              true,
			UseDirect2dDebug:       false,
			ShowOsMessages:         false,
			ShowWindowSizeOnResize: true,
			ShowMouseCoordinates:   true,
		},
		LockThread: true,
		Window: as.AppWindowSettings{
			Name: "Window for tests with Direct2D API",
			Size: c.Size{
				Width:  640,
				Height: 480,
			},
		},
		PrintJobsChannelSize:             16,
		MainControllerMessageChannelSize: 16,
		ImagesToLoad: []string{
			`img\Row_of_Town_Houses.jpg`,
		},
		UseHardwareRendering:                     false,
		UseStraightAlphaChannelForImportedImages: false,
		RenderTargetFormat: as.AppRenderTargetFormatSettings{
			PixelFormat: w32.DXGI_FORMAT_B8G8R8A8_UNORM,
			AlphaMode:   w32.D2D1_ALPHA_MODE_PREMULTIPLIED,
		},
		BitmapFormat: as.AppBitmapFormatSettings{
			PixelFormat: w32.DXGI_FORMAT_B8G8R8A8_UNORM,
			AlphaMode:   w32.D2D1_ALPHA_MODE_PREMULTIPLIED,
		},
	}

	instance, err := getInstance()
	mustBeNoError(err)

	var app *a.App
	app, err = a.NewApp(appSettings, instance)
	mustBeNoError(err)

	err = app.Run()
	mustBeNoError(err)
}
func getInstance() (instance windows.Handle, err error) {
	// Emulate hInstance for Windows API.
	flags := uint32(windows.GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS | windows.GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT)
	err = windows.GetModuleHandleEx(flags, (*uint16)(unsafe.Pointer(&mm)), &instance)
	if err != nil {
		return 0, err
	}

	return instance, nil
}
func mustBeNoError(err error) {
	if err != nil {
		panic(err)
	}
}
