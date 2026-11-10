package mc

import (
	c "github.com/vault-thirteen/Direct2D/common"
)

const (
	Err_UnsupportedButton = "unsupported button"
)

// MouseButtonDownPos stores position where each mouse button was pressed.
// If mouse button was released at the same position where it was pressed,
// then this action is considered as a click event.
type MouseButtonDownPos struct {
	Left   *c.Position
	Middle *c.Position
	Right  *c.Position
	X1     *c.Position
	X2     *c.Position
}

func NewMouseButtonDownPos() (mbdp *MouseButtonDownPos) {
	mbdp = &MouseButtonDownPos{
		Left:   nil,
		Middle: nil,
		Right:  nil,
		X1:     nil,
		X2:     nil,
	}

	return mbdp
}

func (mbdp *MouseButtonDownPos) RegisterPressPosition(mouseButtonType MouseButtonType, mx int, my int) {
	switch mouseButtonType {
	case MouseButtonType_Left:
		mbdp.Left = &c.Position{X: mx, Y: my}

	case MouseButtonType_Middle:
		mbdp.Middle = &c.Position{X: mx, Y: my}

	case MouseButtonType_Right:
		mbdp.Right = &c.Position{X: mx, Y: my}

	case MouseButtonType_X1:
		mbdp.X1 = &c.Position{X: mx, Y: my}

	case MouseButtonType_X2:
		mbdp.X2 = &c.Position{X: mx, Y: my}
	}

	return
}
func (mbdp *MouseButtonDownPos) GetPressPosition(mouseButtonType MouseButtonType) (pos *c.Position) {
	switch mouseButtonType {
	case MouseButtonType_Left:
		return mbdp.Left

	case MouseButtonType_Middle:
		return mbdp.Middle

	case MouseButtonType_Right:
		return mbdp.Right

	case MouseButtonType_X1:
		return mbdp.X1

	case MouseButtonType_X2:
		return mbdp.X2
	}

	return nil
}
func (mbdp *MouseButtonDownPos) UnRegisterPressPosition(mouseButtonType MouseButtonType) {
	switch mouseButtonType {
	case MouseButtonType_Left:
		mbdp.Left = nil

	case MouseButtonType_Middle:
		mbdp.Middle = nil

	case MouseButtonType_Right:
		mbdp.Right = nil

	case MouseButtonType_X1:
		mbdp.X1 = nil

	case MouseButtonType_X2:
		mbdp.X2 = nil
	}

	return
}
