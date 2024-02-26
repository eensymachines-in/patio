package digital

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

type TouchSensor struct {
	*gpio.DirectPinDriver
	state bool // represents the state of the pin
}

func NewTouchSensor(pin string, adp gobot.Adaptor) *TouchSensor {
	return &TouchSensor{
		state:           false,
		DirectPinDriver: gpio.NewDirectPinDriver(adp, pin),
	}
}

func (ts *TouchSensor) Boot() *TouchSensor {
	ts.DirectPinDriver.DigitalWrite(0) // to start with the pin is off
	return ts
}

func (ts *TouchSensor) ShutD() {
	ts.DirectPinDriver.DigitalWrite(0)
}
func (ts *TouchSensor) Watch(canc chan bool) chan time.Time {
	// https://stackoverflow.com/questions/25657207/how-to-know-a-buffered-channel-is-full
	// making an unbufferred channel with overflow configuration
	// When the listener isnt ready channel would overflow and hencee only one tick is sent
	touches := make(chan time.Time)
	go func() {
		defer close(touches)
		for {
			select {
			case <-time.After(200 * time.Millisecond):
				val, _ := ts.DirectPinDriver.DigitalRead()
				if val == 1 {
					touches <- time.Now()
				}
			case <-canc:
				return
			default:
				continue
			}
		}
	}()
	return touches
}
