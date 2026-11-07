package p

import (
	"github.com/vault-thirteen/auxie/errors"
)

const (
	Errf_ErrorInField = "error in field: %s"
)

const (
	Field_PrintJobsChannelSize = "printJobsChannelSize"
)

type PrinterSettings struct {
	printJobsChannelSize int
}

func NewPrinterSettings(printJobsChannelSize int) (ps *PrinterSettings, err error) {
	ps = &PrinterSettings{
		printJobsChannelSize: printJobsChannelSize,
	}

	err = ps.Validate()
	if err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *PrinterSettings) Validate() (err error) {
	if ps == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	if ps.printJobsChannelSize <= 0 {
		return errors.ErrorFromFormattedText(Errf_ErrorInField, Field_PrintJobsChannelSize)
	}

	return nil
}
