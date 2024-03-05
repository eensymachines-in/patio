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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// Time denominations commonly used
const (
	daily     = 24 * time.Hour
	halfdaily = 12 * time.Hour
	hourly    = 1 * time.Hour
	minutely  = 1 * time.Minute
	secondly  = 1 * time.Second
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

// parse_clock: for typical applications we need to set clocks as tick time preferences.
// format of clock expected - 20:35
//
// Format does not include seconds, since we arent expected to plan relay operations to the level of seconds.
//
/*
	hr, min, err := parse_clock("20:35")
	if err !=nil{
		return fmt.Errorf("invalid clock format, check and send again")
	}
*/
func parse_clock(clock string) (int64, int64, error) {
	hrmin := strings.Split(clock, ":") // clock format expected - 13:09, 24 hour clock with minute resolution
	if len(hrmin) != 2 {
		// this is when the clock isnt as expected.
		return 0, 0, fmt.Errorf("invalid clock format, expected format is 13:04")
	}
	hr, _ := strconv.ParseInt(hrmin[0], 10, 32)
	min, _ := strconv.ParseInt(hrmin[1], 10, 32)
	return hr, min, nil
}

// calc_tickOffset : for the hr,min time this can get the time elapsed,until from the current time
// returns the time in duration and seconds elapsed / until
func calc_tickOffset(hr, min int64) (time.Duration, int64) {
	now := time.Now() // getting the current time
	h, m, s := now.Clock()
	midnight := now.Unix() - int64(s+(m*60)+(h*3600)) // tracing the midnight time
	tickTime := midnight + (hr * 3600) + (min * 60)
	log.WithFields(log.Fields{"ticktime": tickTime, "nowtime": now.Unix()}).Debug("clock times")
	offset := tickTime - now.Unix()                                  // getting the offset of the now to the time of tick
	offDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // converting those seconds of ffset to duration offset
	return offDuration, offset
}

func PulseEveryDayAt(clock string, pulse time.Duration, canc <-chan bool) (chan time.Time, error) {
	ticks := make(chan time.Time, 2)
	hr, min, _ := parse_clock(clock)
	go func() {
		defer close(ticks)
		// time calculations have to be done only inside the go routine since scheduling time of this routine is indeterminate
		// only when the coroutine gets scheduled can you do all the time calculations.
		offDuration, offset := calc_tickOffset(hr, min)
		if offset >= 0 {
			// case where time until tick duration, so sleeping
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time until tick")
			<-time.After(offDuration)
			ticks <- time.Now()
			<-time.After(pulse)
			ticks <- time.Now()
		} else {
			//this is a tricky situation when the ticking time for the day has already elapsed
			// 24 hour cycle does not apply, for the next tick but you have to send an extra tick for the tick that has elapsed
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time since tick")
			ticks <- time.Now()
			start := time.Now().Add(offDuration) // offDuration is negative, hence it would give the pulse start
			end := start.Add(pulse)
			<-time.After(time.Until(end)) // this will be less than the pulse since the elapsed time has to be subtracted
			ticks <- time.Now()

			offset = int64(86400) + offset                                    // offset here is negative, hence the final offset calculated would have to be less than 24 hours / 86400 seconds
			offsettedDay, _ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // a day is about 86400 seconds
			log.WithFields(log.Fields{"offset": offsettedDay}).Debug("time until next tick, offset day")
			<-time.After(offsettedDay)
			ticks <- time.Now()
			<-time.After(pulse)
			ticks <- time.Now()
		}
		for t := range PulseEvery(daily, pulse, canc) {
			ticks <- t
		}
	}()
	return ticks, nil
}

// TickEveryDayAt : for a given clock time like 13:40,it will send ticks separated by 24 hours
// closing the canc channel will bring down the loop and close all the ticks
// Ticks setup before the ticking time : the loop starts with the delay to compensate
// Ticks setup after the ticking time : immediate tick then offsets the 24 hour cycle for the elapsed time, sends another tick after the day after which regular 24 hours cycle starts
// clock	: string of the clock, example 13:35
// canc 	: interrupt channel to kill the loop
func TickEveryDayAt(clock string, canc <-chan bool) (chan time.Time, error) {
	ticks := make(chan time.Time, 1)
	hr, min, _ := parse_clock(clock)
	go func() {
		defer close(ticks)
		// time calculations have to be done only inside the go routine since scheduling time of this routine is indeterminate
		// only when the coroutine gets scheduled can you do all the time calculations.
		offDuration, offset := calc_tickOffset(hr, min)
		if offset >= 0 {
			// case where time until tick duration, so sleeping
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time until tick")
			<-time.After(offDuration)
			ticks <- time.Now()
		} else {
			//this is a tricky situation when the ticking time for the day has already elapsed
			// 24 hour cycle does not apply, for the next tick but you have to send an extra tick for the tick that has elapsed
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time since tick")
			ticks <- time.Now()
			offset = int64(86400) + offset                                    // offset here is negative, hence the final offset calculated would have to be less than 24 hours / 86400 seconds
			offsettedDay, _ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // a day is about 86400 seconds
			log.WithFields(log.Fields{"offset": offsettedDay}).Debug("time until next tick, offset day")
			<-time.After(offsettedDay)
			ticks <- time.Now()
		}
		for t := range TickEvery(daily, canc) {
			ticks <- t
		}
	}()
	return ticks, nil
}

// PulseEvery : after every d duration it would tick twice separated by w duration
// canc channel will kill the loop and close the channel
// d > w always
func PulseEvery(d, w time.Duration, canc <-chan bool) chan time.Time {
	ticks := make(chan time.Time, 1)
	go func() {
		defer close(ticks)
		for {
			select {
			case <-time.After(d):
				ticks <- time.Now()
				<-time.After(w)
				ticks <- time.Now()
			case <-canc:
				return
			}
		}
	}()
	return ticks
}

// For the given duration this can send channel messages for the current time over an over till canceled
// use the canc channel to kill the loop
func TickEvery(d time.Duration, canc <-chan bool) chan time.Time {
	ticks := make(chan time.Time, 1)
	//sets up the loop for ticking, can be closed only if the cancel channel closed.
	go func() {
		defer close(ticks)
		for {
			select {
			case <-time.After(d):
				ticks <- time.Now()
			case <-canc:
				return
			}
		}
	}()
	return ticks
}

// this main loop would only setup the tickers
func main() {
	fmt.Println("We are inside the patio program ..")
	log.WithFields(log.Fields{
		"time": time.Now(),
	}).Debug("time now")

	// System signal or touch interrupt watch
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// initialized hardware drivers
	r := raspi.NewAdaptor()
	r.Connect()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for intr := range interrupt.TouchOrSysSignal(PIN_TOUCH, digital.SLOW_WATCH_3_3V, r, ctx, &wg) {
			log.WithFields(log.Fields{
				"time": intr,
			}).Warn("received interrupt..")
			cancel() // time for all the program to go down
		}
	}()
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
			case <-ctx.Done():
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
		ticks, _ := tickers.PulseEveryDayAt("09:05", 1*time.Hour, ctx, &wg)
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
