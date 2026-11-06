package as

import w32 "github.com/vault-thirteen/Direct2D/os/win32"

type AppRenderTargetFormatSettings struct {
	PixelFormat w32.DXGI_FORMAT
	AlphaMode   w32.D2D1_ALPHA_MODE
}
