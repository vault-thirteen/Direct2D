package p

import (
	"sync"

	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/Start-Controller"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"
)

const (
	Msg_PrinterStart = "Printer Start"
	Msg_PrinterStop  = "Printer Stop"
)

const (
	Err_NullPointer     = "null pointer"
	Err_ChannelIsClosed = "channel is closed"
)

type Printer[T any] struct {
	settings  *PrinterSettings
	sc        *stct.StartController
	printJobs *chan T
	printerWg *sync.WaitGroup
}

func NewPrinter[T any](settings *PrinterSettings) (pr *Printer[T], err error) {
	if settings == nil {
		return nil, errors.ErrorFromText(Err_NullPointer)
	}

	err = settings.Validate()
	if err != nil {
		return nil, err
	}

	ch := make(chan T, settings.printJobsChannelSize)

	pr = &Printer[T]{
		settings:  settings,
		sc:        stct.NewStartController(),
		printJobs: &ch,
		printerWg: new(sync.WaitGroup),
	}

	return pr, nil
}
func (p *Printer[T]) Start() (err error) {
	err = p.sc.BeginStart()
	if err != nil {
		return err
	}

	p.printerWg.Add(1)
	go p.runWorker()

	err = p.sc.CompleteStart()
	if err != nil {
		return err
	}

	return nil
}
func (p *Printer[T]) Stop() (err error) {
	err = p.sc.BeginStop()
	if err != nil {
		return err
	}

	if p.printJobs == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	// Unfortunately, the architecture of Go language is so terrible, that you
	// can not safely check whether the channel was already closed or not.
	defer func() {
		x := recover()
		if x != nil {
			// Channel was already closed.
			err = c.RecoveryError(x, Err_ChannelIsClosed)
		}
	}()
	close(*p.printJobs)

	p.printerWg.Wait()

	err = p.sc.CompleteStop()
	if err != nil {
		return err
	}

	return nil
}
func (p *Printer[T]) runWorker() {
	cio.PrintMessage(Msg_PrinterStart)
	defer func() {
		p.printerWg.Done()
		cio.PrintMessage(Msg_PrinterStop)
	}()

	for printJob := range *p.printJobs {
		cio.PrintMessage(printJob)
	}
}
func (p *Printer[T]) ReceiveJob(job T) (err error) {
	if !p.sc.IsRunning() {
		return errors.ErrorFromText(stct.Err_NotRunning)
	}

	if p.printJobs == nil {
		return errors.ErrorFromText(Err_NullPointer)
	}

	// Unfortunately, the architecture of Go language is so terrible, that you
	// can not safely check whether the channel was closed or not.
	defer func() {
		x := recover()
		if x != nil {
			// Channel is closed.
			err = c.RecoveryError(x, Err_ChannelIsClosed)
		}
	}()
	*p.printJobs <- job

	return nil
}
