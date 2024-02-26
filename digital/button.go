package digital

import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

/*
============================
We need a simple button driver on the digital pin to read the interrupts
Most of the IoT applications would run an infinite loop (main) to do all the tasks
buttons bring in the functionality of interrupting the loop, and doing a task that prempts the main tasks
typically used for emergency / manual intervention.
23-FEB-2024 | kneerunjun@gmail.com | eensymachines.in |
============================
*/
type BTN_PULL uint8 // pull resistor, and the initial state of the button
// pull up resistors will hold one terminal of the button to a high state and when pressed will drain the current into the GPIO
// pull down resistors will hold one terminal of the button to ground and from a high GPIO will drain the current to GND

const (
	BTN_PULLDOWN uint8 = iota
	BTN_PULLUP
)

// When the gpio goes high, this shall interrupt
type InterruptButton struct {
	*gpio.DirectPinDriver
	state bool // state is in synch with the h/w pin state
	pull  uint8
}

// NewInterruptButton : ctor for interrupt buttons
//
/*
	cancel := make(chan bool, 1)
	defer close(cancel)
	r := raspi.NewAdaptor()
	r.Connect()
	btn := digital.NewInterruptButton("33", digital.BTN_PULLUP, r)
	for t := range btn.Start(cancel, 500*time.Millisecond) {
		fmt.Println(t)
		return
	}
*/
func NewInterruptButton(pin string, pullupdown uint8, adp gobot.Adaptor) *InterruptButton {
	return &InterruptButton{
		DirectPinDriver: gpio.NewDirectPinDriver(adp, pin),
		state:           false,
		pull:            pullupdown,
	}
}

func (ib *InterruptButton) Start(canc chan bool, interval time.Duration) chan time.Time {
	chanIntrpt := make(chan time.Time, 200)
	if ib.pull == BTN_PULLUP {
		ib.DirectPinDriver.DigitalWrite(0) // to start with the pin will be low
	} else if ib.pull == BTN_PULLDOWN {
		ib.DirectPinDriver.DigitalWrite(1)
	}
	go func() {
		for {
			select {
			// NOTE: about 500 msecs should a good starting point to test
			case <-time.After(interval):
				val, _ := ib.DirectPinDriver.DigitalRead()
				if (val == 1 && ib.pull == BTN_PULLUP) || (val == 0 && ib.pull == BTN_PULLDOWN) {
					// the button was pressed
					chanIntrpt <- time.Now()
				}
			case <-canc:
				close(chanIntrpt)
				return
			}
		}
	}()
	return chanIntrpt
}
