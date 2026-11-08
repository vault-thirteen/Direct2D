package mc

import (
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/errors"
)

type MainControllerMessage struct {
	Type          MainControllerMessageType
	MouseEvent    *MainControllerMessageMouseEvent
	KeyboardEvent *MainControllerMessageKeyboardEvent
}

func NewMainControllerMessage(messageType MainControllerMessageType, data any) (mcm *MainControllerMessage, err error) {
	mcm = &MainControllerMessage{}

	switch messageType {
	case MainControllerMessageType_MouseEvent:
		mcm.Type = MainControllerMessageType_MouseEvent
		mcm.MouseEvent = data.(*MainControllerMessageMouseEvent)

	case MainControllerMessageType_KeyboardEvent:
		mcm.Type = MainControllerMessageType_KeyboardEvent
		mcm.KeyboardEvent = data.(*MainControllerMessageKeyboardEvent)

	case MainControllerMessageType_Quit:
		mcm.Type = MainControllerMessageType_Quit

	default:
		return nil, errors.ErrorFromText(c.Err_UnknownType)
	}

	return mcm, nil
}
