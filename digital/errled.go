package digital

import (
	"github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

type ErrLED struct {
	*gpio.DirectPinDriver
	state bool // represents the state of the pin
}

func NewErrLED(pin string, adp gobot.Adaptor) *ErrLED {
	return &ErrLED{
		state:           false,
		DirectPinDriver: gpio.NewDirectPinDriver(adp, pin),
	}
}
func (el *ErrLED) Log(err error) {
	logrus.Error(err)
	el.DirectPinDriver.DigitalWrite(1)
}
func (el *ErrLED) Boot() *ErrLED {
	el.DirectPinDriver.DigitalWrite(0) // to start with the pin is off
	return el
}

func (el *ErrLED) IsHigh() bool {
	return el.state
}

func (el *ErrLED) ShutD() {
	el.DirectPinDriver.DigitalWrite(0)
}
