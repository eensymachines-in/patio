package main

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/assert"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

func TestRelays(t *testing.T) {
	r := raspi.NewAdaptor()
	relay := gpio.NewDirectPinDriver(r, "40")
	relay.DigitalWrite(0)
	state, _ := relay.DigitalRead()
	assert.IsEqual(t, 0, state, "Unexpcted state of the digital pin")
}
