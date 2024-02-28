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
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eensymachines-in/patio/digital"
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

// calc_tickOffset : for the hr,min time this can get the time elapsed/until from the current time
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

// tickTimeUnix :  given the clock as string , this can give the tick time as unix elaspsed seconds, and the offset seconds from now. Use this instead of calc_tickOffset
//
// offset > 0  would mean there is time until today's tick
//
// offset <0 would mean time has elapsed since tick
//
// tt		: ticktime as unix elapsed seconds
//
// offset 	: seconds since / until tick time
//
// e		: error in parsing ticktime from clock string
//
/*
	tt, offset, err := tickTimeUnix("13:45")
	OffsetAsdur :=time.ParseDuration(fmt.Sprintf("%ds", offset))
	if err == nil {
		if offset >0 {
			// time until tick
		} else if offset <0 {
			// time elapased since tick
		}else {
			// tick time right now
		}
	}
*/
func tickTimeUnix(clock string) (tt int64, offset int64, e error) {
	hr, min, e := parse_clock(clock)
	if e != nil {
		return
	}
	/* For getting the elapsed / until time between now and the tick time (offset) we compare both of them to same day's midnight
	offset thus can be +ve / -ve depending on when it was assessed. - if the tick is ahead or past the now time
	*/
	now := time.Now()
	h, m, s := now.Clock()
	midnight := now.Unix() - int64(s+(m*60)+(h*3600)) // tracing the midnight time
	tt = midnight + (hr * 3600) + (min * 60)
	log.WithFields(log.Fields{"ticktime": tt, "nowtime": now.Unix()}).Debug("clock times")
	offset = tt - now.Unix() // getting the offset of the now to the time of tick
	return
}

// PulseEveryDayAt : This is the same as TickEveryDay but involves 2 ticks in every call. - hence the name pulse
//
//   - clock	: time at which the pulse is initiated everyday
//
//   - pulse	: gap in the pulse - since 2 ticks make a pulse
//
//   - canc		: interruption channel
func PulseEveryDayAt(clock string, pulse time.Duration, canc <-chan bool) (chan time.Time, error) {
	ticks := make(chan time.Time, 2)
	// hr, min, _ := parse_clock(clock)
	go func() {
		defer close(ticks)
		// time calculations have to be done only inside the go routine since scheduling time of this routine is indeterminate
		// only when the coroutine gets scheduled can you do all the time calculations.
		// offDuration, offset := calc_tickOffset(hr, min)
		_, offset, err := tickTimeUnix(clock)
		if err != nil {
			log.Error(err)
			return
		}
		offDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", offset))
		if offset >= 0 {
			// case where time until tick duration, so sleeping
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time until tick")
			select {
			case <-time.After(offDuration):
				ticks <- time.Now()
			case <-canc:
				return
			}
			select {
			case <-time.After(pulse):
				ticks <- time.Now()
			case <-canc:
				return
			}
		} else {
			// start + pulse > start + offset+ pulse
			//this is a tricky situation when the ticking time for the day has already elapsed
			// 24 hour cycle does not apply, for the next tick but you have to send an extra tick for the tick that has elapsed
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time since tick")
			// 2 ticks make a pulse, Below we are arriving at the pulse start and end
			start := time.Now().Add(offDuration) // offDuration is negative, hence it would give the pulse start
			end := start.Add(pulse)
			if time.Until(end) > 0 {
				ticks <- time.Now()
				log.Debug("we are between the pulses")
				select {
				case <-time.After(time.Until(end)): // this will be less than the pulse since the elapsed time has to be subtracted
					ticks <- time.Now()
				case <-canc:
					return
				}
			}
			log.Debug("we are past both the pulses, hence no ticks")
			// Since we are already pass the pulsing time, the next one pulse shall start for less than 24 hours
			// offset here is negative, hence the final offset calculated would have to be less than 24 hours / 86400 seconds
			offset = int64(86400) + offset
			offsettedDay, _ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // a day is about 86400 seconds
			log.WithFields(log.Fields{"offset": offsettedDay}).Debug("time until next tick, offset day")
			select {
			case <-time.After(offsettedDay):
				ticks <- time.Now()
			case <-canc:
				return
			}
			select {
			case <-time.After(pulse):
				ticks <- time.Now()
			case <-canc:
				return
			}
		}
		for t := range PulseEvery(hourly, pulse, canc) {
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
			// NOTE: incase offDuration is long, still responsive to system interruption
			select {
			case <-time.After(offDuration):
				ticks <- time.Now()
			case <-canc:
				break
			}
		} else {
			//this is a tricky situation when the ticking time for the day has already elapsed
			// 24 hour cycle does not apply, for the next tick but you have to send an extra tick for the tick that has elapsed
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time since tick")
			ticks <- time.Now()
			offset = int64(86400) + offset                                    // offset here is negative, hence the final offset calculated would have to be less than 24 hours / 86400 seconds
			offsettedDay, _ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // a day is about 86400 seconds
			log.WithFields(log.Fields{"offset": offsettedDay}).Debug("time until next tick, offset day")
			// NOTE: offsetedDay is indeed a long duration, using select case will help to be responsive to system interruptions
			select {
			case <-time.After(offsettedDay):
				ticks <- time.Now()
			case <-canc:
				break
			}
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
				// NOTE: for long sleep times to make it responsive to sys interrupts
				select {
				case <-time.After(w):
					ticks <- time.Now()
				case <-canc:
					return
				}

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
	log.Info("Starting the clocked relay now..")
	var wg sync.WaitGroup // unless we are allowing all the threads to exit we cannot close main

	cancel := make(chan bool)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer close(signals)
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-signals:
			// cannot flush hardware here since after cancel is closed, the program will exit
			// upon getting the signal we just call off all the loops.
			log.Warn("Received system interrupt, closing down program..")
			close(cancel)
		case <-cancel:
			return
		}
	}() // flushing the hardware states
	log.Info("Listening for system interruptions now..")
	// initialized hardware drivers
	r := raspi.NewAdaptor()
	err := r.Connect()
	if err != nil {
		log.Panicf("failed to connect to raspberry device %s", err)
	}
	log.Info("Connected to raspberry device..")

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

	digital.NewErrLED(PIN_ERRLED, r).Boot() // created a new err led that can indicate and lg errors
	// errled.Log(fmt.Errorf("test error for the led"))

	touch := digital.NewTouchSensor(PIN_TOUCH, r).Boot()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for t := range touch.Watch(cancel) {
			log.Warnf("Touch interrupt, %s", t.Format(time.RFC3339))
			close(cancel)
			return
		}
	}()
	log.Info("Touch interrupt setup..")

	wg.Add(1)
	go func() {
		defer wg.Done()
		rs := digital.NewRelaySwitch(PIN_RELAY, false, r).Boot()
		ticks, _ := PulseEveryDayAt("18:00", 1*time.Minute, cancel)
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
