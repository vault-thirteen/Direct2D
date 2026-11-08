package mc

type MainControllerMessageType byte

const (
	MainControllerMessageType_Unknown       = 0
	MainControllerMessageType_MouseEvent    = 1
	MainControllerMessageType_KeyboardEvent = 2
	MainControllerMessageType_Quit          = 0xFF
)
