package as

type AppSettings struct {
	Debug                                    AppDebugSettings
	LockThread                               bool
	Window                                   AppWindowSettings
	PrintJobsChannelSize                     int
	ImagesToLoad                             []string
	UseHardwareRendering                     bool
	UseStraightAlphaChannelForImportedImages bool
	RenderTargetFormat                       AppRenderTargetFormatSettings
	BitmapFormat                             AppBitmapFormatSettings
}

// Limits for formats are described at the following link:
// https://learn.microsoft.com/en-us/windows/win32/direct2d/supported-pixel-formats-and-alpha-modes#supported-formats-for-id2d1hwndrendertarget

// Software rendering supports only BGRA format.
// Hardware rendering supports BGRA and RGBA formats.

// Premultiplied alpha channel is supported by both software and hardware
// renderers. Straight alpha channel is not supported by all renderers.

// Bitmaps support only BGRA format with premultiplied alpha channel.
