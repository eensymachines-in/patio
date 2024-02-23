package digital

/* ====================
23-FEB-2024 | kneerunjun@gmail.com | eensymachines.in |
When working with relays its st fwd that we choose the gobot drivers available.
- one they are untested, and thus has some sweet spots
- two we need clock assisted relays that work on a cron basis
- plus some available hardware here in india is chinese made. Such relays are thrown when pin goes digitally low. - inverted relays
Here we develop a thick wrapper around gpio.DirecPinDriver which can substitute RelayDriver.
Testing platform with Raspberry Pi Zero W rev 1.1, BCM2835

==================== */
import (
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

// RelaySwitch : for purposes of simple relay operations, this encapsulates DirectPinDriver
// gobot package does provide a similar datatype but found that to be unreliable
type RelaySwitch struct {
	*gpio.DirectPinDriver
	Inverted bool
	state    bool // state of the pin

}

// NewRelaySwitch : ctor for relay wrapper.
// pin 		: hardware pin index as string. This is the pin index and not the gpio number
// invrtd 	: flag true when digital low throws the relay
// conn 	: instance of one of the adptors from the various provided under the gobot package
//
/*
	// for pin 35 (GPIO19) on RPi 0w with inverted relays connected
	rs := digital.NewRelaySwitch("35", true, raspi.NewAdaptor())
*/
func NewRelaySwitch(pin string, invrtd bool, conn gobot.Adaptor) *RelaySwitch {
	return &RelaySwitch{
		DirectPinDriver: gpio.NewDirectPinDriver(conn, pin),
		Inverted:        invrtd,
	}

}

// Boot : call this immediately after constructor.
// will set the pin to low - for inverted relays will set the pin to high
// copy the pin state back onto the field
func (rs *RelaySwitch) Boot() *RelaySwitch {
	if rs.Inverted {
		rs.DirectPinDriver.DigitalWrite(1)
	} else {
		rs.DirectPinDriver.DigitalWrite(0)
	}
	time.Sleep(1 * time.Second)
	val, _ := rs.DirectPinDriver.DigitalRead()
	rs.state = val == 1 // state is independent of the inversion
	// state just reflects if high or low
	// high / low is equivalent to on / off
	// digital pin state is inverted incase of inverted flag
	return rs
}

// IsHigh : returns the internal state of RelaySwitch
// This is always in sync with actual pin state high/low
func (rs *RelaySwitch) IsHigh() bool {
	return rs.state
}

// Toggle: sets the relay the opposite of the current state
// this helps when operating the relay on cron ticks
func (rs *RelaySwitch) Toggle() {
	if rs.IsHigh() {
		rs.Low()
	} else {
		rs.High()
	}
}

// Low : Relay switch opens
// for inverted relays, pin is set to high
func (rs *RelaySwitch) Low() error {
	if !rs.Inverted {
		rs.DirectPinDriver.DigitalWrite(0)
	} else {
		rs.DirectPinDriver.DigitalWrite(1)
	}
	rs.state = false
	return nil
}

// High: relay switch closes.
// for inverted relays the pin set to low
func (rs *RelaySwitch) High() error {
	if !rs.Inverted {
		rs.DirectPinDriver.DigitalWrite(1)
	} else {
		rs.DirectPinDriver.DigitalWrite(0)
	}
	rs.state = true
	return nil

}
