package w32

import (
	"math"
)

type D2D1_SIZE_U = D2D_SIZE_U
type D2D_SIZE_U Size[UINT32]

func (su D2D_SIZE_U) ToUint64() (out uint64) {
	return (uint64(su.Height) << 32) | uint64(su.Width)
	//return *(*uintptr)(unsafe.Pointer(&su))
}
func UintptrTo_D2D_SIZE_U(in uintptr) (out D2D_SIZE_U) {
	return D2D1_SIZE_U{
		Width:  uint32(in & 0xFFFFFFFF),
		Height: uint32(in >> 32),
	}
	//return *(*D2D_SIZE_U)(unsafe.Pointer(&in))
}

type D2D1_SIZE_F = D2D_SIZE_F
type D2D_SIZE_F Size[FLOAT]

func (sf D2D_SIZE_F) ToUint64() (out uint64) {
	wBits := uint64(math.Float32bits(sf.Width))
	hBits := uint64(math.Float32bits(sf.Height))
	return ((hBits) << 32) | wBits
	//return *(*uintptr)(unsafe.Pointer(&sf))
}
func UintptrTo_D2D_SIZE_F(in uintptr) (out D2D_SIZE_F) {
	x := uint64(in)
	wBits := uint32(x & 0xFFFFFFFF)
	hBits := uint32(x >> 32)
	return D2D1_SIZE_F{
		Width:  math.Float32frombits(wBits),
		Height: math.Float32frombits(hBits),
	}
	//return *(*D2D_SIZE_F)(unsafe.Pointer(&in))
}

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
type D2D_MATRIX_3X2_F struct {
	// Horizontal scaling / cosine of rotation
	M11 FLOAT

	// Vertical shear / sine of rotation
	M12 FLOAT

	// Horizontal shear / negative sine of rotation
	M21 FLOAT

	// Vertical scaling / cosine of rotation
	M22 FLOAT

	// Horizontal shift (always orthogonal regardless of rotation)
	Dx FLOAT

	// Vertical shift (always orthogonal regardless of rotation)
	Dy FLOAT
}

// Array returns the matrix data as a 3x2 array,
// mimicking the C-style FLOAT m[3][2] union field.
func (m *D2D_MATRIX_3X2_F) Array() [3][2]FLOAT {
	return [3][2]FLOAT{
		{m.M11, m.M12},
		{m.M21, m.M22},
		{m.Dx, m.Dy},
	}
}

// FromArray populates the matrix from a 3x2 array.
func (m *D2D_MATRIX_3X2_F) FromArray(a [3][2]FLOAT) {
	m.M11, m.M12 = a[0][0], a[0][1]
	m.M21, m.M22 = a[1][0], a[1][1]
	m.Dx, m.Dy = a[2][0], a[2][1]
}

func D2D1_Matrix3x2F_Identity() D2D_MATRIX_3X2_F {
	return D2D_MATRIX_3X2_F{
		M11: 1, M12: 0,
		M21: 0, M22: 1,
		Dx: 0, Dy: 0,
	}
}
