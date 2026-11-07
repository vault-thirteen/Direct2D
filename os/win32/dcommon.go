package w32

type D2D_SIZE_U Size[UINT32]
type D2D_SIZE_F Size[FLOAT]

type D2D1_PIXEL_FORMAT struct {
	Format    DXGI_FORMAT
	AlphaMode D2D1_ALPHA_MODE
}

type D2D1_RECT_U = D2D_RECT_U
type D2D_RECT_U struct {
	Left   UINT32
	Top    UINT32
	Right  UINT32
	Bottom UINT32
}

type D2D1_RECT_F = D2D_RECT_F
type D2D_RECT_F struct {
	Left   FLOAT
	Top    FLOAT
	Right  FLOAT
	Bottom FLOAT
}

type D2D1_MATRIX_3X2_F = D2D_MATRIX_3X2_F
