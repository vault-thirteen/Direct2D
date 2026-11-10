package mc

import (
	"sync"

	c "github.com/vault-thirteen/Direct2D/common"
	stct "github.com/vault-thirteen/auxie/Start-Controller"
	"github.com/vault-thirteen/auxie/cio"
	"github.com/vault-thirteen/auxie/errors"
)

const (
	Msg_MainControllerStart = "Main Controller Start"
	Msg_MainControllerStop  = "Main Controller Stop"
)

const (
	Err_NullPointer     = "null pointer"
	Err_ChannelIsClosed = "channel is closed"
)

type MainController struct {
	settings       *MainControllerSettings
	sc             *stct.StartController
	messageChannel *chan *MainControllerMessage
	workWg         *sync.WaitGroup
	mbdp           *MouseButtonDownPos
}

func NewMainController(settings *MainControllerSettings) (mc *MainController, err error) {
	if settings == nil {
		return nil, errors.ErrorFromText(Err_NullPointer)
	}

	err = settings.Validate()
	if err != nil {
		return nil, err
	}

	ch := make(chan *MainControllerMessage, settings.messageChannelSize)

	mc = &MainController{
		settings:       settings,
		sc:             stct.NewStartController(),
		messageChannel: &ch,
		workWg:         new(sync.WaitGroup),
		mbdp:           NewMouseButtonDownPos(),
	}

	return mc, nil
}
func (mc *MainController) Start() (err error) {
	err = mc.sc.BeginStart()
	if err != nil {
		return err
	}

	mc.workWg.Add(1)
	go mc.runWorker()

	err = mc.sc.CompleteStart()
	if err != nil {
		return err
	}

	return nil
}
func (mc *MainController) Stop() (err error) {
	err = mc.sc.BeginStop()
	if err != nil {
		return err
	}

	if mc.messageChannel == nil {
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
	close(*mc.messageChannel)

	mc.workWg.Wait()

	err = mc.sc.CompleteStop()
	if err != nil {
		return err
	}

	return nil
}
func (mc *MainController) runWorker() {
	cio.PrintMessage(Msg_MainControllerStart)
	defer func() {
		mc.workWg.Done()
		cio.PrintMessage(Msg_MainControllerStop)
	}()

	for msg := range *mc.messageChannel {
		if mc.settings.isDebug {
			switch msg.Type {
			case MainControllerMessageType_MouseEvent:
				mc.processMouseEvent(msg)

			case MainControllerMessageType_KeyboardEvent:
				if msg.KeyboardEvent.ButtonName != nil {
					cio.PrintMessage(msg.KeyboardEvent, *msg.KeyboardEvent.ButtonName)
				} else {
					cio.PrintMessage(msg.KeyboardEvent)
				}

			case MainControllerMessageType_Quit:
				cio.PrintMessage(msg)
			}
		}

		//TODO
	}
}
func (mc *MainController) processMouseEvent(msg *MainControllerMessage) {
	if msg.MouseEvent.ButtonState == ButtonState_Pressed {
		mc.mbdp.RegisterPressPosition(msg.MouseEvent.ButtonType, msg.MouseEvent.CursorX, msg.MouseEvent.CursorY)
	}
	if msg.MouseEvent.ButtonState == ButtonState_Released {
		pressPos := mc.mbdp.GetPressPosition(msg.MouseEvent.ButtonType)
		if (pressPos != nil) && (pressPos.X == msg.MouseEvent.CursorX) && (pressPos.Y == msg.MouseEvent.CursorY) {
			//TODO
			cio.PrintMessage("Mouse Click, buttonType:", msg.MouseEvent.ButtonType)
		}
		mc.mbdp.UnRegisterPressPosition(msg.MouseEvent.ButtonType)
	}

	cio.PrintMessage(msg.MouseEvent)
}
func (mc *MainController) ReceiveMessage(msg *MainControllerMessage) (err error) {
	if !mc.sc.IsRunning() {
		return errors.ErrorFromText(stct.Err_NotRunning)
	}

	if mc.messageChannel == nil {
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
	*mc.messageChannel <- msg

	return nil
}
