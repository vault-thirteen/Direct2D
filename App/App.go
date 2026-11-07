package a

import (
	"runtime"
	"syscall"

	"unsafe"

	as "github.com/vault-thirteen/Direct2D/AppSettings"
	img "github.com/vault-thirteen/Direct2D/Image"
	p "github.com/vault-thirteen/Direct2D/Printer"
	c "github.com/vault-thirteen/Direct2D/common"
	w32 "github.com/vault-thirteen/Direct2D/os/win32"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"
	"golang.org/x/sys/windows"
)

type App struct {
	settings                   *as.AppSettings
	instance                   windows.Handle
	printer                    *p.Printer[any]
	images                     []*img.Image
	hwnd                       w32.HWND
	width                      w32.UINT32
	height                     w32.UINT32
	direct2DFactory            *w32.ID2D1Factory
	windowClassName            w32.LPCWSTR
	pIWICFactory               *w32.IWICImagingFactory
	renderTargetProperties     *w32.D2D1_RENDER_TARGET_PROPERTIES
	hwndRenderTargetProperties *w32.D2D1_HWND_RENDER_TARGET_PROPERTIES
	renderTarget               *w32.ID2D1HwndRenderTarget
	bitmapProperties           *w32.D2D1_BITMAP_PROPERTIES
	bitmapSize                 *w32.D2D1_SIZE_U
	bitmapSizePacked           uint64
	bitmap                     *w32.ID2D1Bitmap
}

func NewApp(settings *as.AppSettings, instance windows.Handle) (app *App, err error) {
	app = &App{
		settings: settings,
		instance: instance,
		images:   make([]*img.Image, 0, len(settings.ImagesToLoad)),
	}

	err = app.selfCheck()
	if err != nil {
		return nil, err
	}

	// Printer.
	var printerSettings *p.PrinterSettings
	printerSettings, err = p.NewPrinterSettings(app.settings.PrintJobsChannelSize)
	if err != nil {
		return nil, err
	}

	app.printer, err = p.NewPrinter[any](printerSettings)
	if err != nil {
		return nil, err
	}

	// Images.
	var image *img.Image
	for _, filePath := range settings.ImagesToLoad {
		image, err = img.NewImageFromFile(filePath, app.settings.UseStraightAlphaChannelForImportedImages)
		if err != nil {
			return nil, err
		}

		app.images = append(app.images, image)
	}

	return app, nil
}
func (app *App) Run() (err error) {
	var needRollBack = true

	// 0. Lock the O.S. thread.
	{
		if app.settings.LockThread {
			runtime.LockOSThread()
			defer func() {
				if needRollBack {
					runtime.UnlockOSThread()
				}
			}()
		}
	}

	// 1. Start the printer.
	{
		err = app.printer.Start()
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				derr := app.printer.Stop()
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 2. Initialise the COM library.
	{
		err = windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED)
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				windows.CoUninitialize()
			}
		}()
	}

	// 3. Create the Direct2D factory.
	{
		var factoryOptions w32.D2D1_FACTORY_OPTIONS
		if app.settings.Debug.IsEnabled {
			if app.settings.Debug.UseDirect2dDebug {
				factoryOptions = w32.D2D1_FACTORY_OPTIONS{DebugLevel: w32.D2D1_DEBUG_LEVEL_INFORMATION}
			} else {
				factoryOptions = w32.D2D1_FACTORY_OPTIONS{DebugLevel: w32.D2D1_DEBUG_LEVEL_NONE}
			}
		}

		var hr w32.HRESULT
		hr = w32.D2D1CreateFactory(w32.D2D1_FACTORY_TYPE_MULTI_THREADED, &w32.IID_ID2D1Factory, &factoryOptions, &app.direct2DFactory)
		if hr != w32.S_OK {
			return hr
		}

		defer func() {
			if needRollBack {
				if app.direct2DFactory != nil {
					_ = app.direct2DFactory.Release()
				}
			}
		}()
	}

	// 4. Register the window class.
	var wc w32.WNDCLASSEXW
	{
		var cursor w32.HCURSOR
		cursor, err = w32.LoadCursorW(0, w32.IDC_ARROW)
		if err != nil {
			return err
		}

		app.windowClassName, err = syscall.UTF16PtrFromString("Direct2DBitmapApp")
		if err != nil {
			return err
		}

		wc = w32.WNDCLASSEXW{
			Size:       uint32(unsafe.Sizeof(wc)),
			Style:      w32.CS_HREDRAW | w32.CS_VREDRAW,
			WndProc:    app.wndProc,
			ClsExtra:   0,
			WndExtra:   int32(unsafe.Sizeof(uintptr(0))),
			Instance:   app.instance,
			Icon:       0,
			Cursor:     cursor,
			Background: windows.Handle(w32.COLOR_WINDOWFRAME),
			ClassName:  app.windowClassName,
		}
		wcg := w32.WNDCLASSEXW_ForGolang(wc)

		_, err = w32.RegisterClassExW(&wcg)
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				_, derr := w32.UnregisterClassW(app.windowClassName, app.instance)
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 5. Create a window.
	{
		var windowName w32.LPCWSTR
		windowName, err = syscall.UTF16PtrFromString(app.settings.Window.Name)
		if err != nil {
			return err
		}

		app.hwnd, err = w32.CreateWindowExW(
			0,
			wc.ClassName,
			windowName,
			w32.WS_OVERLAPPEDWINDOW|w32.WS_VISIBLE,
			w32.CW_USEDEFAULT,
			w32.CW_USEDEFAULT,
			int32(app.settings.Window.Size.Width),
			int32(app.settings.Window.Size.Height),
			0,
			0,
			app.instance,
			unsafe.Pointer(app),
		)
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				_, derr := w32.DestroyWindow(app.hwnd)
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 6. Show the window.
	{
		w32.ShowWindow(app.hwnd, windows.SW_SHOWNORMAL)

		var b w32.BOOL
		b = w32.UpdateWindow(app.hwnd)
		if b == 0 {
			return c.FunctionFailure("UpdateWindow")
		}
	}

	// 7. Initialise the Windows Imaging Component (WIC).
	{
		app.pIWICFactory = nil
		var hr w32.HRESULT
		hr = w32.CoCreateInstance(
			&w32.CLSID_WICImagingFactory1, // TODO: 0,1,2 ???
			nil,
			w32.DWORD(w32.CLSCTX_INPROC_SERVER),
			&(w32.IID_IWICImagingFactory),
			(*unsafe.Pointer)(unsafe.Pointer(&app.pIWICFactory)),
		)
		if hr.FAILED() {
			return hr
		}

		defer func() {
			if needRollBack {
				_ = app.pIWICFactory.Release()
			}
		}()
	}

	// 8. Read O.S. messages.
	// Unfortunately, Windows API does not support message readers in separate
	// threads. We must read messages in the same O.S. thread where our window
	// was created.
	{
		err = app.readOsMessages()
		if err != nil {
			return err
		}
	}

	return nil
}
func (app *App) readOsMessages() (err error) {
	app.printMessage(c.Msg_MessageReaderStart)
	defer func() {
		app.printMessage(c.Msg_MessageReaderStop)
	}()

	var msg w32.MSG
	var b w32.BOOL

	for {
		b, err = w32.GetMessageW(&msg, 0, 0, 0)
		if err != nil {
			return err
		}

		// Windows O.S. documentation suggests that we should stop reading O.S.
		// messages when 'GetMessage' function returns zero.
		if b == 0 {
			break
		}

		_ = w32.TranslateMessage(&msg)
		_ = w32.DispatchMessageW(&msg)
	}

	return nil
}
func (app *App) printMessage(message any) {
	err := app.printer.ReceiveJob(message)
	if err != nil {
		panic(err)
	}
}
func (app *App) peekMessage() (isMessageAvailable bool) {
	var msg w32.MSG
	var b = w32.PeekMessageW(&msg, 0, 0, 0, w32.PM_NOREMOVE)
	return b != 0
}
func (app *App) wndProc(hwnd w32.HWND, message w32.WM, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	if app.settings.Debug.IsEnabled {
		if app.settings.Debug.ShowOsMessages {
			txt := cio.FormatMessage("OS Message [%v] wParam:%+v lParam:%+v", message, wParam, lParam)
			app.printMessage(txt)
		}
	}

	// WM_CREATE message can return '-1' error.
	if message == w32.WM_CREATE {
		if app.settings.Debug.IsEnabled {
			app.printMessage(c.Msg_WM_CREATE)
		}

		pcs := (*w32.CREATESTRUCT)(unsafe.Pointer(lParam))
		pApp := (*App)(pcs.LpCreateParams)

		w32.SetLastError(0)
		pv := w32.SetWindowLongPtrW(hwnd, w32.GWLP_USERDATA, uintptr(unsafe.Pointer(pApp)))
		if pv == 0 {
			err := windows.GetLastError()
			if err != nil {
				app.printMessage(err)
				return ^uintptr(0) // -1.
			}
		}

		return 0
	}

	x, err := w32.GetWindowLongPtrW(hwnd, w32.GWLP_USERDATA)
	if err != nil {
		app.printMessage(err)
		return w32.DefWindowProcW(hwnd, message, wParam, lParam)
	}

	pApp := (*App)(unsafe.Pointer(x))
	if pApp == nil {
		err = errors.ErrorFromText(c.Err_NullPointer)
		app.printMessage(err)
		return w32.DefWindowProcW(hwnd, message, wParam, lParam)
	}

	switch message {

	case w32.WM_PAINT:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_PAINT)
			}

			err = app.onRender()
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var b w32.BOOL
			b = w32.ValidateRect(app.hwnd, nil)
			if b == 0 {
				err = c.FunctionFailure("ValidateRect")
				app.printMessage(err)
				return 0
			}

			return 0
		}

	// WM_SIZE message can not return errors.
	case w32.WM_SIZE:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_SIZE)
			}

			app.width = w32.UINT(w32.LOWORD(lParam))
			app.height = w32.UINT(w32.HIWORD(lParam))

			if (app.width <= 0) || (app.height <= 0) {
				err = errors.ErrorFromFormattedText(c.Errf_InvalidSizeRequest, app.width, app.height)
				app.printMessage(err)
				return 0
			}

			err = app.onResize(wParam)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var b w32.BOOL
			b = w32.InvalidateRect(app.hwnd, nil, 1)
			if b == 0 {
				err = c.FunctionFailure("InvalidateRect")
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case w32.WM_CLOSE:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_CLOSE)
			}

			//ShowWindow(hwnd, windows.SW_HIDE)
			w32.PostQuitMessage(0) // Send WM_QUIT message to itself.
			return 0
		}

	case w32.WM_DESTROY:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_DESTROY)
			}

			return 0
		}

	case w32.WM_NCDESTROY:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_NCDESTROY)
			}

			return 0
		}
	}

	return w32.DefWindowProcW(hwnd, message, wParam, lParam)
}
func (app *App) onResize(wParam w32.WPARAM) (err error) {
	if app.settings.Debug.IsEnabled {
		switch wParam {
		case w32.SIZE_RESTORED:
			app.printMessage("SIZE_RESTORED")
		case w32.SIZE_MINIMIZED:
			app.printMessage("SIZE_MINIMIZED")
		case w32.SIZE_MAXIMIZED:
			app.printMessage("SIZE_MAXIMIZED")
		case w32.SIZE_MAXSHOW:
			app.printMessage("SIZE_MAXSHOW")
		case w32.SIZE_MAXHIDE:
			app.printMessage("SIZE_MAXHIDE")
		}
	}

	if app.renderTarget == nil {
		return nil
	}

	size := w32.D2D1_SizeU(app.width, app.height)

	var mbs w32.UINT32 = app.renderTarget.GetMaximumBitmapSize()
	if size.Width > mbs {
		return errors.ErrorFromFormattedText(c.Errf_SizeExceedsMaximumSize, size.Width, mbs)
	}
	if size.Height > mbs {
		return errors.ErrorFromFormattedText(c.Errf_SizeExceedsMaximumSize, size.Height, mbs)
	}

	var hr w32.HRESULT
	hr, err = app.renderTarget.Resize(&size)
	if err != nil {
		if hr == w32.D2DERR_RECREATE_TARGET {
			app.discardDeviceResources()
		}
		return err
	}

	if app.settings.Debug.IsEnabled {
		if app.settings.Debug.ShowWindowSizeOnResize {
			txt := cio.FormatMessage("Window is resized. width:%+v height:%+v wParam:%+v", app.width, app.height, wParam)
			app.printMessage(txt)
		}
	}

	return nil
}
func (app *App) onRender() (err error) {
	err = app.createDeviceResources()
	if err != nil {
		return err
	}

	var hr w32.HRESULT
	// ID2D1HwndRenderTarget::CheckWindowState method (d2d1.h)
	// https://learn.microsoft.com/en-gb/windows/win32/api/d2d1/nf-d2d1-id2d1hwndrendertarget-checkwindowstate
	/*
		Return value
		Type: D2D1_WINDOW_STATE
		A value that indicates whether the HWND associated with this render target is occluded.
	*/
	// D2D1_WINDOW_STATE enumeration (d2d1.h)
	// https://learn.microsoft.com/en-us/windows/win32/api/d2d1/ne-d2d1-d2d1_window_state
	/*
		typedef enum D2D1_WINDOW_STATE {
		  D2D1_WINDOW_STATE_NONE = 0x0000000,
		  D2D1_WINDOW_STATE_OCCLUDED = 0x0000001,
		  D2D1_WINDOW_STATE_FORCE_DWORD = 0xffffffff
		} ;
	*/
	// If window state can only be 0 or 1, then why do I receive value 2291728413 ?????
	// Microsoft will not tell you. Microsoft Direct2D API must burn in hell.
	var windowsState = app.renderTarget.CheckWindowState()
	if windowsState != w32.D2D1_WINDOW_STATE_NONE {
		if windowsState == w32.D2D1_WINDOW_STATE_OCCLUDED {
			// Do not render an occluded window.
			return nil
		}
		hr = w32.HRESULT(windowsState)
		return hr
	}

	app.renderTarget.BeginDraw()
	var identityMatrix = w32.D2D_Matrix3x2F_Identity()
	app.renderTarget.SetTransform(&identityMatrix)
	var color = w32.D2D1_ColorF(w32.D2D1_ColorF_Gray)
	app.renderTarget.Clear(&color)

	if app.bitmap != nil {
		var size w32.D2D1_SIZE_F = app.bitmap.GetSize()
		var destinationRect = w32.D2D1_RectF(0.0, 0.0, size.Width, size.Height)
		app.renderTarget.DrawBitmap(app.bitmap, &destinationRect, w32.FLOAT(1.0), w32.D2D1_BITMAP_INTERPOLATION_MODE_LINEAR, nil)
	}

	hr, err = app.renderTarget.EndDraw()
	if err != nil {
		if hr == w32.D2DERR_RECREATE_TARGET {
			app.discardDeviceResources()
		}
		return err
	}

	// https://learn.microsoft.com/en-us/windows/win32/api/d2d1/ne-d2d1-d2d1_window_state
	_ = app.renderTarget.CheckWindowState()

	return nil
}
func (app *App) createDeviceResources() (err error) {
	err = app.createRenderTarget()
	if err != nil {
		return err
	}

	err = app.loadBitmap()
	if err != nil {
		return err
	}

	return nil
}
func (app *App) discardDeviceResources() {
	app.unloadBitmap()
	app.destroyRenderTarget()
}
func (app *App) createRenderTarget() (err error) {
	if app.renderTarget != nil {
		return nil
	}

	var rect w32.RECT
	_, err = w32.GetClientRect(app.hwnd, &rect)
	if err != nil {
		return err
	}

	var size = w32.D2D1_SizeU(uint32(rect.Right-rect.Left), uint32(rect.Bottom-rect.Top))

	var renderTargetProperties = w32.RenderTargetProperties()
	if app.settings.UseHardwareRendering {
		renderTargetProperties.Type = w32.D2D1_RENDER_TARGET_TYPE_HARDWARE
	} else {
		renderTargetProperties.Type = w32.D2D1_RENDER_TARGET_TYPE_SOFTWARE
	}
	renderTargetProperties.PixelFormat = w32.D2D1_PIXEL_FORMAT{
		Format:    app.settings.RenderTargetFormat.PixelFormat,
		AlphaMode: app.settings.RenderTargetFormat.AlphaMode,
	}
	app.renderTargetProperties = &renderTargetProperties

	var hwndRenderTargetProperties = w32.HwndRenderTargetProperties(app.hwnd, size, w32.D2D1_PRESENT_OPTIONS_NONE)
	app.hwndRenderTargetProperties = &hwndRenderTargetProperties

	// List of formats supported by ID2D1HwndRenderTarget in Direct2D:
	// https://learn.microsoft.com/en-us/windows/win32/direct2d/supported-pixel-formats-and-alpha-modes#supported-formats-for-id2d1hwndrendertarget
	_, err = app.direct2DFactory.CreateHwndRenderTarget(app.renderTargetProperties, app.hwndRenderTargetProperties, &app.renderTarget)
	if err != nil {
		return err
	}

	return nil
}
func (app *App) destroyRenderTarget() {
	if app.renderTarget == nil {
		return
	}

	_ = app.renderTarget.Release()
	app.renderTarget = nil
}
func (app *App) loadBitmap() (err error) {
	if app.bitmap != nil {
		return nil
	}

	if len(app.images) == 0 {
		return nil
	}

	var image = app.images[0]

	// Check formats.
	if image.AlphaChannelMode != app.settings.BitmapFormat.AlphaMode {
		return errors.ErrorFromFormattedText(c.Errf_BitmapAlphaModeMismatch, app.settings.BitmapFormat.AlphaMode, image.AlphaChannelMode)
	}

	// Of all the colour models which are supported in Go language, only RGBA
	// model has chances to be used. Unfortunately, Google corporation does not
	// support alternative technologies and standards.
	if image.ColourModelType != img.ColourModelType_RGBA {
		return errors.ErrorFromText(c.Err_UnsupportedPixelFormat)
	}

	// Of all the colour models which are supported by Direct2D API, only BGRA
	// model has chances to be used. Unfortunately, Microsoft corporation does
	// not support alternative technologies and standards.
	var expectedPixelFormat = app.settings.BitmapFormat.PixelFormat
	if expectedPixelFormat != w32.DXGI_FORMAT_B8G8R8A8_UNORM {
		return errors.ErrorFromText(c.Err_UnsupportedPixelFormat)
	}

	// Convert image from RGBA into BGRA model.
	var pixelData []byte
	pixelData, err = image.PixelDataAsBGRA()
	if err != nil {
		return err
	}

	var bitmapSize = w32.D2D1_SIZE_U{
		Width:  w32.UINT32(image.Size.Width),
		Height: w32.UINT32(image.Size.Height),
	}
	app.bitmapSize = &bitmapSize
	app.bitmapSizePacked = app.bitmapSize.ToUint64()

	// While we do not support YCbCr format, we process only RGBA images.
	var lineStride int
	lineStride, _, _ = image.LineStride()
	var pitch = w32.UINT32(lineStride)

	var bitmapProperties = w32.D2D1_BITMAP_PROPERTIES{
		PixelFormat: w32.D2D1_PIXEL_FORMAT{
			Format:    expectedPixelFormat,
			AlphaMode: image.AlphaChannelMode,
		},
		DpiX: w32.DefaultDpiX,
		DpiY: w32.DefaultDpiY,
	}
	app.bitmapProperties = &bitmapProperties

	//TODO
	app.printMessage(cio.FormatMessage("Bitmap size: %v", app.bitmapSize))
	app.printMessage(cio.FormatMessage("Bitmap size packed: %v", app.bitmapSizePacked))

	// List of formats supported by ID2D1Bitmap in Direct2D:
	// https: //learn.microsoft.com/en-us/windows/win32/direct2d/supported-pixel-formats-and-alpha-modes#specifying-a-pixel-format-for-an-id2d1bitmap
	_, err = app.renderTarget.CreateBitmap(app.bitmapSizePacked, pixelData, pitch, app.bitmapProperties, &app.bitmap)
	if err != nil {
		return err
	}

	//TODO
	app.printMessage(cio.FormatMessage("Bitmap size: %v", app.bitmapSize))
	app.printMessage(cio.FormatMessage("Bitmap size packed: %v", app.bitmapSizePacked))
	sizeInDip := app.bitmap.GetSize()
	app.printMessage(cio.FormatMessage("Bitmap size in DIP: %v", sizeInDip))
	sizeInPx := app.bitmap.GetPixelSize()
	app.printMessage(cio.FormatMessage("Bitmap size in pixels: %v", sizeInPx))

	return nil
}
func (app *App) unloadBitmap() {
	if app.bitmap == nil {
		return
	}

	if len(app.images) == 0 {
		return
	}

	_ = app.bitmap.Release()
	app.bitmap = nil
}
func (app *App) selfCheck() (err error) {
	var b byte = 0xFF
	if unsafe.Sizeof(b) != 1 {
		return errors.ErrorFromText(c.Err_UnsupportedArchitecture)
	}

	var upA uintptr = 0xFFFFFFFFFFFFFFFF
	if unsafe.Sizeof(upA) != 8 {
		return errors.ErrorFromText(c.Err_UnsupportedPointerSize)
	}

	return nil
}
