package mc

import (
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/errors"
)

type MainControllerMessageMouseEvent struct {
	CursorX     int
	CursorY     int
	ButtonType  MouseButtonType
	ButtonState ButtonState
}

func NewMainControllerMessageMouseEvent(cx int, cy int, btnType MouseButtonType, btnState ButtonState) (me *MainControllerMessageMouseEvent, err error) {
	if btnType > MouseButtonTypeMax {
		return nil, errors.ErrorFromText(c.Err_UnknownType)
	}
	if (btnState != ButtonState_Pressed) && (btnState != ButtonState_Released) {
		return nil, errors.ErrorFromText(c.Err_UnknownState)
	}

	me = &MainControllerMessageMouseEvent{
		CursorX:     cx,
		CursorY:     cy,
		ButtonType:  btnType,
		ButtonState: btnState,
	}

	return me, nil
}
