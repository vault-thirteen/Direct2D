package c

import "github.com/vault-thirteen/auxie/errors"

const (
	Err_NullPointer             = "null pointer"
	Err_UnsupportedArchitecture = "unsupported architecture"
	Err_UnsupportedPointerSize  = "unsupported pointer size"

	Errf_FunctionFailure        = "function failure: %s"
	Errf_SizeExceedsMaximumSize = "requested size exceeds maximum value: %d vs %d"
	Errf_InvalidSizeRequest     = "invalid size was requested: width=%d, height=%d"
)

func RecoveryError(x any, messagePrefix string) (err error) {
	errX := errors.ErrorFromFormattedText("%+v", x)
	return errors.ErrorWithPrefix(errX, messagePrefix)
}

func FunctionFailure(funcName string) (err error) {
	return errors.ErrorFromFormattedText(Errf_FunctionFailure, funcName)
}
