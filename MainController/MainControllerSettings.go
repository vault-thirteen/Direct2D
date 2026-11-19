package mc

import (
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/errors"
	b "github.com/vault-thirteen/auxie/os/win32/base"
)

const (
	Errf_ErrorInField = "error in field: %s"
)

const (
	Field_MessageChannelSize = "messageChannelSize"
)

type MainControllerSettings struct {
	isDebug            bool
	messageChannelSize int
	backChannel        *chan c.OsChanMsg
	hDataEvent         b.HANDLE
}

func NewMainControllerSettings(isDebug bool, messageChannelSize int, backChannel *chan c.OsChanMsg, hDataEvent b.HANDLE) (mcs *MainControllerSettings, err error) {
	mcs = &MainControllerSettings{
		isDebug:            isDebug,
		messageChannelSize: messageChannelSize,
		backChannel:        backChannel,
		hDataEvent:         hDataEvent,
	}

	err = mcs.Validate()
	if err != nil {
		return nil, err
	}

	return mcs, nil
}

func (mcs *MainControllerSettings) Validate() (err error) {
	if mcs == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	if mcs.messageChannelSize <= 0 {
		return errors.ErrorFromFormattedText(Errf_ErrorInField, Field_MessageChannelSize)
	}

	if mcs.backChannel == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	if mcs.hDataEvent == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	return nil
}
