package w32

type Size[T any] struct {
	Width  T
	Height T
}

func D2D1_SizeU(width UINT32, height UINT32) D2D1_SIZE_U {
	return D2D1_SIZE_U{Width: width, Height: height}
}

type Rect[T any] struct {
	Left   T
	Top    T
	Right  T
	Bottom T
}

func D2D1_RectF(left, top, right, bottom FLOAT) D2D1_RECT_F {
	return D2D1_RECT_F{Left: left, Top: top, Right: right, Bottom: bottom}
}

const (
	DefaultDpiX FLOAT = 96.0
	DefaultDpiY FLOAT = 96.0
)

type Matrix3x2F = D2D1_MATRIX_3X2_F
