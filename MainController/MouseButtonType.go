package mc

type MouseButtonType byte

const (
	MouseButtonType_None      MouseButtonType = 0
	MouseButtonType_Left      MouseButtonType = 1
	MouseButtonType_Middle    MouseButtonType = 2
	MouseButtonType_Right     MouseButtonType = 3
	MouseButtonType_WheelUp   MouseButtonType = 4
	MouseButtonType_WheelDown MouseButtonType = 5
	MouseButtonType_X1        MouseButtonType = 6
	MouseButtonType_X2        MouseButtonType = 7
	MouseButtonType_Unknown   MouseButtonType = 8

	MouseButtonTypeMax = MouseButtonType_Unknown
)
