package mc

import "github.com/vault-thirteen/auxie/errors"

const (
	Errf_ErrorInField = "error in field: %s"
)

const (
	Field_MessageChannelSize = "messageChannelSize"
)

type MainControllerSettings struct {
	isDebug            bool
	messageChannelSize int
}

func NewMainControllerSettings(messageChannelSize int, isDebug bool) (mcs *MainControllerSettings, err error) {
	mcs = &MainControllerSettings{
		isDebug:            isDebug,
		messageChannelSize: messageChannelSize,
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

	return nil
}
