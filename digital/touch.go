package digital

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

const (
	FAST_WATCH_5V   = 250 * time.Millisecond // when connected to 5V Vcc
	SLOW_WATCH_3_3V = 600 * time.Millisecond // when connected to 3.3 Vcc
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
func (ts *TouchSensor) Watch(speed time.Duration, ctx context.Context, wg *sync.WaitGroup) chan time.Time {
	// https://stackoverflow.com/questions/25657207/how-to-know-a-buffered-channel-is-full
	// making an unbufferred channel with overflow configuration
	// When the listener isnt ready channel would overflow and hencee only one tick is sent
	touches := make(chan time.Time, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(touches)
		defer logrus.Warn("Now closing touch sensor..")
		for {
			select {
			case <-time.After(speed):
				val, _ := ts.DirectPinDriver.DigitalRead()
				if val == 1 && len(touches) < 1 {
					touches <- time.Now()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return touches
}
