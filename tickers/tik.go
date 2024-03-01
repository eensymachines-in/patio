package tickers

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)
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
