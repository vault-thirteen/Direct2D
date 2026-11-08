package mc

import (
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"
)

type MainControllerMessageKeyboardEvent struct {
	KeyCode     cio.KeyCode
	ButtonState ButtonState
	ButtonName  *string
}

func NewMainControllerMessageKeyboardEvent(btnCode cio.KeyCode, btnState ButtonState) (ke *MainControllerMessageKeyboardEvent, err error) {
	if (btnState != ButtonState_Pressed) && (btnState != ButtonState_Released) {
		return nil, errors.ErrorFromText(c.Err_UnknownState)
	}

	ke = &MainControllerMessageKeyboardEvent{
		KeyCode:     btnCode,
		ButtonState: btnState,
		ButtonName:  nil,
	}

	return ke, nil
}
