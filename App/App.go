package a

import (
	"runtime"
	"syscall"

	"unsafe"

	as "github.com/vault-thirteen/Direct2D/AppSettings"
	img "github.com/vault-thirteen/Direct2D/Image"
	mc "github.com/vault-thirteen/Direct2D/MainController"
	p "github.com/vault-thirteen/Direct2D/Printer"
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"
	"github.com/vault-thirteen/auxie/os/win32"
	b "github.com/vault-thirteen/auxie/os/win32/base"
	col "github.com/vault-thirteen/auxie/os/win32/class/D2D1/ColorF/enum"
	e "github.com/vault-thirteen/auxie/os/win32/enum"
	i "github.com/vault-thirteen/auxie/os/win32/interface"
	t "github.com/vault-thirteen/auxie/os/win32/type"
	"golang.org/x/sys/windows"
)

type App struct {
	settings                   *as.AppSettings
	instance                   b.HINSTANCE
	printer                    *p.Printer[any]
	mc                         *mc.MainController
	images                     []*img.Image
	hwnd                       b.HWND
	width                      b.UINT32
	height                     b.UINT32
	direct2DFactory            *i.ID2D1Factory
	windowClassName            b.LPCWSTR
	pIWICFactory               *i.IWICImagingFactory
	renderTargetProperties     *t.D2D1_RENDER_TARGET_PROPERTIES
	hwndRenderTargetProperties *t.D2D1_HWND_RENDER_TARGET_PROPERTIES
	renderTarget               *i.ID2D1HwndRenderTarget
	bitmapProperties           *t.D2D1_BITMAP_PROPERTIES
	bitmapSize                 t.D2D1_SIZE_U
	bitmap                     *i.ID2D1Bitmap
}

func NewApp(settings *as.AppSettings, instance b.HINSTANCE) (app *App, err error) {
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

	// Main controller.
	var mainControllerSettings *mc.MainControllerSettings
	mainControllerSettings, err = mc.NewMainControllerSettings(app.settings.MainControllerMessageChannelSize, app.settings.Debug.IsEnabled)
	if err != nil {
		return nil, err
	}

	app.mc, err = mc.NewMainController(mainControllerSettings)
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

	// 1. Lock the O.S. thread.
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

	// 2. Start the printer.
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

	// 3. Start the main controller.
	{
		err = app.mc.Start()
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				derr := app.mc.Stop()
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 4. Initialise the COM library.
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

	// 5. Create the Direct2D factory.
	{
		var factoryOptions t.D2D1_FACTORY_OPTIONS
		if app.settings.Debug.IsEnabled {
			if app.settings.Debug.UseDirect2dDebug {
				factoryOptions = t.D2D1_FACTORY_OPTIONS{DebugLevel: e.D2D1_DEBUG_LEVEL_INFORMATION}
			} else {
				factoryOptions = t.D2D1_FACTORY_OPTIONS{DebugLevel: e.D2D1_DEBUG_LEVEL_NONE}
			}
		}

		var hr b.HRESULT
		hr = win32.D2D1CreateFactory(e.D2D1_FACTORY_TYPE_MULTI_THREADED, &b.IID_ID2D1Factory, &factoryOptions, &app.direct2DFactory)
		if hr != b.S_OK {
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

	// 6. Register the window class.
	var wc t.WNDCLASSEXW
	{
		var cursor b.HCURSOR
		cursor, err = win32.LoadCursorW(0, b.IDC_ARROW)
		if err != nil {
			return err
		}

		app.windowClassName, err = syscall.UTF16PtrFromString("Direct2DBitmapApp")
		if err != nil {
			return err
		}

		wc = t.WNDCLASSEXW{
			Size:       uint32(unsafe.Sizeof(wc)),
			Style:      b.CS_HREDRAW | b.CS_VREDRAW,
			WndProc:    app.wndProc,
			ClsExtra:   0,
			WndExtra:   int32(unsafe.Sizeof(uintptr(0))),
			Instance:   app.instance,
			Icon:       0,
			Cursor:     cursor,
			Background: windows.Handle(b.COLOR_WINDOWFRAME),
			ClassName:  app.windowClassName,
		}
		wcg := t.WNDCLASSEXW_ForGolang(wc)

		_, err = win32.RegisterClassExW(&wcg)
		if err != nil {
			return err
		}

		defer func() {
			if needRollBack {
				_, derr := win32.UnregisterClassW(app.windowClassName, app.instance)
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 7. Create a window.
	{
		var windowName b.LPCWSTR
		windowName, err = syscall.UTF16PtrFromString(app.settings.Window.Name)
		if err != nil {
			return err
		}

		app.hwnd, err = win32.CreateWindowExW(
			0,
			wc.ClassName,
			windowName,
			b.WS_OVERLAPPEDWINDOW|b.WS_VISIBLE,
			b.CW_USEDEFAULT,
			b.CW_USEDEFAULT,
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
				_, derr := win32.DestroyWindow(app.hwnd)
				err = errors.Combine(err, derr)
			}
		}()
	}

	// 8. Show the window.
	{
		win32.ShowWindow(app.hwnd, windows.SW_SHOWNORMAL)

		var ret b.BOOL
		ret = win32.UpdateWindow(app.hwnd)
		if ret == 0 {
			return c.FunctionFailure("UpdateWindow")
		}
	}

	// 9. Initialise the Windows Imaging Component (WIC).
	{
		app.pIWICFactory = nil
		var hr b.HRESULT
		hr = win32.CoCreateInstance(
			&i.CLSID_WICImagingFactory1, // TODO: 0,1,2 ???
			nil,
			b.DWORD(e.CLSCTX_INPROC_SERVER),
			&(i.IID_IWICImagingFactory),
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

	var msg t.MSG
	var ret b.BOOL

	for {
		ret, err = win32.GetMessageW(&msg, 0, 0, 0)
		if err != nil {
			return err
		}

		// Windows O.S. documentation suggests that we should stop reading O.S.
		// messages when 'GetMessage' function returns zero.
		if ret == 0 {
			break
		}

		_ = win32.TranslateMessage(&msg)
		_ = win32.DispatchMessageW(&msg)
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
	var msg t.MSG
	var ret = win32.PeekMessageW(&msg, 0, 0, 0, b.PM_NOREMOVE)
	return ret != 0
}
func (app *App) wndProc(hwnd b.HWND, message e.WM, wParam b.WPARAM, lParam b.LPARAM) b.LRESULT {
	if app.settings.Debug.IsEnabled {
		if app.settings.Debug.ShowOsMessages {
			txt := cio.FormatMessage("OS Message [%v] wParam:%+v lParam:%+v", message, wParam, lParam)
			app.printMessage(txt)
		}
	}

	// WM_CREATE message can return '-1' error.
	if message == e.WM_CREATE {
		if app.settings.Debug.IsEnabled {
			app.printMessage(c.Msg_WM_CREATE)
		}

		pcs := (*t.CREATESTRUCT)(unsafe.Pointer(lParam))
		pApp := (*App)(pcs.LpCreateParams)

		win32.SetLastError(0)
		pv := win32.SetWindowLongPtrW(hwnd, b.GWLP_USERDATA, uintptr(unsafe.Pointer(pApp)))
		if pv == 0 {
			err := windows.GetLastError()
			if err != nil {
				app.printMessage(err)
				return ^uintptr(0) // -1.
			}
		}

		return 0
	}

	x, err := win32.GetWindowLongPtrW(hwnd, b.GWLP_USERDATA)
	if err != nil {
		app.printMessage(err)
		return win32.DefWindowProcW(hwnd, message, wParam, lParam)
	}

	pApp := (*App)(unsafe.Pointer(x))
	if pApp == nil {
		err = errors.ErrorFromText(c.Err_NullPointer)
		app.printMessage(err)
		return win32.DefWindowProcW(hwnd, message, wParam, lParam)
	}

	switch message {

	case e.WM_PAINT:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_PAINT)
			}

			err = app.onRender()
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var ret b.BOOL
			ret = win32.ValidateRect(app.hwnd, nil)
			if ret == 0 {
				err = c.FunctionFailure("ValidateRect")
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_MOUSEMOVE:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_None, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_LBUTTONDOWN:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Left, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_LBUTTONUP:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Left, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_RBUTTONDOWN:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Right, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_RBUTTONUP:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Right, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_MBUTTONDOWN:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Middle, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_MBUTTONUP:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, mc.MouseButtonType_Middle, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_KEYDOWN:
		{
			keyCode := cio.KeyCode(wParam)
			isRepeated := (lParam & b.RepeatedKeyMarker) != 0
			if isRepeated {
				return 0
			}

			var keyboardEvent *mc.MainControllerMessageKeyboardEvent
			keyboardEvent, err = mc.NewMainControllerMessageKeyboardEvent(keyCode, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			if app.settings.Debug.IsEnabled {
				keyName, ok := cio.KeyCodeNames[keyCode]
				if ok {
					keyboardEvent.ButtonName = &keyName
				}
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_KeyboardEvent, keyboardEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_SYSKEYDOWN:
		{
			keyCode := cio.KeyCode(wParam)
			isRepeated := (lParam & b.RepeatedKeyMarker) != 0
			if isRepeated {
				return 0
			}

			var keyboardEvent *mc.MainControllerMessageKeyboardEvent
			keyboardEvent, err = mc.NewMainControllerMessageKeyboardEvent(keyCode, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			if app.settings.Debug.IsEnabled {
				keyName, ok := cio.KeyCodeNames[keyCode]
				if ok {
					keyboardEvent.ButtonName = &keyName
				}
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_KeyboardEvent, keyboardEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_KEYUP:
		{
			keyCode := cio.KeyCode(wParam)

			var keyboardEvent *mc.MainControllerMessageKeyboardEvent
			keyboardEvent, err = mc.NewMainControllerMessageKeyboardEvent(keyCode, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			if app.settings.Debug.IsEnabled {
				keyName, ok := cio.KeyCodeNames[keyCode]
				if ok {
					keyboardEvent.ButtonName = &keyName
				}
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_KeyboardEvent, keyboardEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_SYSKEYUP:
		{
			keyCode := cio.KeyCode(wParam)

			var keyboardEvent *mc.MainControllerMessageKeyboardEvent
			keyboardEvent, err = mc.NewMainControllerMessageKeyboardEvent(keyCode, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			if app.settings.Debug.IsEnabled {
				keyName, ok := cio.KeyCodeNames[keyCode]
				if ok {
					keyboardEvent.ButtonName = &keyName
				}
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_KeyboardEvent, keyboardEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_MOUSEWHEEL:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			mwDelta := int16(wParam >> 16)
			var btnType mc.MouseButtonType
			if mwDelta > 0 {
				btnType = mc.MouseButtonType_WheelUp
			} else {
				btnType = mc.MouseButtonType_WheelDown
			}

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, btnType, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_XBUTTONDOWN:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			xButton := uint16(wParam >> 16)
			var btnType mc.MouseButtonType
			switch xButton {
			case b.XBUTTON1:
				btnType = mc.MouseButtonType_X1
			case b.XBUTTON2:
				btnType = mc.MouseButtonType_X2
			default:
				err = errors.ErrorFromText(c.Err_UnknownType)
				app.printMessage(err)
				return 0
			}

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, btnType, mc.ButtonState_Pressed)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_XBUTTONUP:
		{
			mx := int(int16(lParam & 0xffff))
			my := int(int16((lParam >> 16) & 0xffff))

			xButton := uint16(wParam >> 16)
			var btnType mc.MouseButtonType
			switch xButton {
			case b.XBUTTON1:
				btnType = mc.MouseButtonType_X1
			case b.XBUTTON2:
				btnType = mc.MouseButtonType_X2
			default:
				err = errors.ErrorFromText(c.Err_UnknownType)
				app.printMessage(err)
				return 0
			}

			var mouseEvent *mc.MainControllerMessageMouseEvent
			mouseEvent, err = mc.NewMainControllerMessageMouseEvent(mx, my, btnType, mc.ButtonState_Released)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_MouseEvent, mouseEvent)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			return 0
		}

	// WM_SIZE message can not return errors.
	case e.WM_SIZE:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_SIZE)
			}

			app.width = b.UINT(b.LOWORD(uintptr(lParam)))
			app.height = b.UINT(b.HIWORD(uintptr(lParam)))

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

			var ret b.BOOL
			ret = win32.InvalidateRect(app.hwnd, nil, 1)
			if ret == 0 {
				err = c.FunctionFailure("InvalidateRect")
				app.printMessage(err)
				return 0
			}

			return 0
		}

	case e.WM_CLOSE:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_CLOSE)
			}

			var mcm *mc.MainControllerMessage
			mcm, err = mc.NewMainControllerMessage(mc.MainControllerMessageType_Quit, nil)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			err = app.mc.ReceiveMessage(mcm)
			if err != nil {
				app.printMessage(err)
				return 0
			}

			//ShowWindow(hwnd, windows.SW_HIDE)
			win32.PostQuitMessage(0) // Send WM_QUIT message to itself.
			return 0
		}

	case e.WM_DESTROY:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_DESTROY)
			}

			return 0
		}

	case e.WM_NCDESTROY:
		{
			if app.settings.Debug.IsEnabled {
				app.printMessage(c.Msg_WM_NCDESTROY)
			}

			return 0
		}
	}

	return win32.DefWindowProcW(hwnd, message, wParam, lParam)
}
func (app *App) onResize(wParam b.WPARAM) (err error) {
	if app.settings.Debug.IsEnabled {
		switch wParam {
		case b.SIZE_RESTORED:
			app.printMessage("SIZE_RESTORED")
		case b.SIZE_MINIMIZED:
			app.printMessage("SIZE_MINIMIZED")
		case b.SIZE_MAXIMIZED:
			app.printMessage("SIZE_MAXIMIZED")
		case b.SIZE_MAXSHOW:
			app.printMessage("SIZE_MAXSHOW")
		case b.SIZE_MAXHIDE:
			app.printMessage("SIZE_MAXHIDE")
		}
	}

	if app.renderTarget == nil {
		return nil
	}

	size := t.D2D1_SizeU(app.width, app.height)

	var mbs b.UINT32 = app.renderTarget.GetMaximumBitmapSize()
	if size.Width > mbs {
		return errors.ErrorFromFormattedText(c.Errf_SizeExceedsMaximumSize, size.Width, mbs)
	}
	if size.Height > mbs {
		return errors.ErrorFromFormattedText(c.Errf_SizeExceedsMaximumSize, size.Height, mbs)
	}

	var hr b.HRESULT
	hr, err = app.renderTarget.Resize(&size)
	if err != nil {
		if hr == b.D2DERR_RECREATE_TARGET {
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

	var hr b.HRESULT
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
	if windowsState != e.D2D1_WINDOW_STATE_NONE {
		if windowsState == e.D2D1_WINDOW_STATE_OCCLUDED {
			// Do not render an occluded window.
			return nil
		}
		hr = b.HRESULT(windowsState)
		return hr
	}

	app.renderTarget.BeginDraw()
	var identityMatrix = t.D2D_Matrix3x2F_Identity()
	app.renderTarget.SetTransform(&identityMatrix)
	var color = t.NewColorF_2B(col.Gray)
	app.renderTarget.Clear(&color)

	if app.bitmap != nil {
		var size t.D2D1_SIZE_F = app.bitmap.GetSize()
		var destinationRect = t.D2D1_RectF(0.0, 0.0, size.Width, size.Height)
		app.renderTarget.DrawBitmap(app.bitmap, &destinationRect, b.FLOAT(1.0), e.D2D1_BITMAP_INTERPOLATION_MODE_LINEAR, nil)
	}

	hr, err = app.renderTarget.EndDraw()
	if err != nil {
		if hr == b.D2DERR_RECREATE_TARGET {
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

	var rect t.RECT
	_, err = win32.GetClientRect(app.hwnd, &rect)
	if err != nil {
		return err
	}

	var size = t.D2D1_SizeU(uint32(rect.Right-rect.Left), uint32(rect.Bottom-rect.Top))

	var renderTargetProperties = t.RenderTargetProperties()
	if app.settings.UseHardwareRendering {
		renderTargetProperties.Type = e.D2D1_RENDER_TARGET_TYPE_HARDWARE
	} else {
		renderTargetProperties.Type = e.D2D1_RENDER_TARGET_TYPE_SOFTWARE
	}
	renderTargetProperties.PixelFormat = t.D2D1_PIXEL_FORMAT{
		Format:    app.settings.RenderTargetFormat.PixelFormat,
		AlphaMode: app.settings.RenderTargetFormat.AlphaMode,
	}
	app.renderTargetProperties = &renderTargetProperties

	var hwndRenderTargetProperties = t.HwndRenderTargetProperties(app.hwnd, size, e.D2D1_PRESENT_OPTIONS_NONE)
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
	if expectedPixelFormat != e.DXGI_FORMAT_B8G8R8A8_UNORM {
		return errors.ErrorFromText(c.Err_UnsupportedPixelFormat)
	}

	// Convert image from RGBA into BGRA model.
	var pixelData []byte
	pixelData, err = image.PixelDataAsBGRA()
	if err != nil {
		return err
	}

	app.bitmapSize = t.D2D1_SIZE_U{
		Width:  b.UINT32(image.Size.Width),
		Height: b.UINT32(image.Size.Height),
	}

	// While we do not support YCbCr format, we process only RGBA images.
	var lineStride int
	lineStride, _, _ = image.LineStride()
	var pitch = b.UINT32(lineStride)

	var bitmapProperties = t.D2D1_BITMAP_PROPERTIES{
		PixelFormat: t.D2D1_PIXEL_FORMAT{
			Format:    expectedPixelFormat,
			AlphaMode: image.AlphaChannelMode,
		},
		DpiX: b.DefaultDpiX,
		DpiY: b.DefaultDpiY,
	}
	app.bitmapProperties = &bitmapProperties

	// List of formats supported by ID2D1Bitmap in Direct2D:
	// https: //learn.microsoft.com/en-us/windows/win32/direct2d/supported-pixel-formats-and-alpha-modes#specifying-a-pixel-format-for-an-id2d1bitmap
	_, err = app.renderTarget.CreateBitmap(app.bitmapSize, pixelData, pitch, app.bitmapProperties, &app.bitmap)
	if err != nil {
		return err
	}

	sizeInPx := app.bitmap.GetPixelSize()
	if (sizeInPx.Width != app.bitmapSize.Width) || (sizeInPx.Height != app.bitmapSize.Height) {
		return errors.ErrorFromText(c.Err_SizeMismatch)
	}

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
	var x byte = 0xFF
	if unsafe.Sizeof(x) != 1 {
		return errors.ErrorFromText(c.Err_UnsupportedArchitecture)
	}

	var upA uintptr = 0xFFFFFFFFFFFFFFFF
	if unsafe.Sizeof(upA) != 8 {
		return errors.ErrorFromText(c.Err_UnsupportedPointerSize)
	}

	return nil
}
