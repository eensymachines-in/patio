package main

/* ===========
Household patio lighting needs simple clock driven logic of turning OF / ON the lights for intervals in every day
Driven by a Raspberry pi processor, the logic can involve simple peripherals of relays, buttons, LEDs.
The logic involves intelligently driving the relays at desired times while covering all the possible exceptions

author		: kneerunjun@gmail.com
date		: 21-2-2024
place		: Pune
=============== */
import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/eensymachines-in/patio/digital"
	"github.com/eensymachines-in/patio/interrupt"
	"github.com/eensymachines-in/patio/tickers"
	oled "github.com/eensymachines-in/ssd1306"
	log "github.com/sirupsen/logrus"
	_ "gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	PIN_INTERRRUPT = ""   // there was a time when we were planning a shutdown tactile button for interrupt
	PIN_TOUCH      = "31" // touch sensor digital in
	PIN_ERRLED     = "33"
	PIN_RELAY      = "35"
)

func init() {
	// Setting up the loggin framework
	log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: false})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

// this main loop would only setup the tickers
func main() {
	log.Info("Starting the clocked relay now..")
	var wg sync.WaitGroup // unless we are allowing all the threads to exit we cannot close main

	// initialized hardware drivers
	r := raspi.NewAdaptor()
	err := r.Connect()
	if err != nil {
		log.Panicf("failed to connect to raspberry device %s", err)
	}
	log.Info("Connected to raspberry device..")
	cancel := interrupt.TouchOrSysSignal(PIN_TOUCH, r)

	wg.Add(1)
	go func() {
		// display thread
		defer wg.Done()
		disp := oled.NewSundingOLED("oled", r)
		flush_display := func() { // helps clear the display for prep and shutdown
			log.Debug("Flushing display..")
			disp.Clean()
			disp.ResetImage().Render()
		}
		flush_display()
		defer flush_display()
		disp_date := func() string { // helps format current date as string
			now := time.Now()
			_, mn, dd := now.Date()
			hr, min, _ := now.Clock()
			return fmt.Sprintf("%s-%02d %02d:%02d", mn.String()[:3], dd, hr, min)
		}
		disp.Message(10, 10, disp_date()).Render()
		for {
			select {
			case <-cancel:
				return
			case <-time.After(1 * time.Minute):
				disp.Clean()
				disp.Message(10, 10, disp_date()).Render()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rs := digital.NewRelaySwitch(PIN_RELAY, false, r).Boot()
		ticks, _ := tickers.PulseEveryDayAt("18:00", 4*time.Hour, cancel)
		for t := range ticks {
			log.Debugf("Flipping the relay state: %s", t.Format(time.RFC3339))
			rs.Toggle()
		}
		rs.Low()
		log.Warn("Now shutting down relay..")
	}()
	log.Info("Clock relay now setup..")
	// Flushing the hardware states
	wg.Wait()
}
