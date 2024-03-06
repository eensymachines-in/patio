package tickers

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// PulseEvery : after every d duration it would tick twice separated by w duration
// canc channel will kill the loop and close the channel
// d > w always
func PulseEvery(d, w time.Duration, ctx context.Context, wg *sync.WaitGroup) chan time.Time {
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
				case <-ctx.Done():
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return ticks
}

// PulseEveryDayAt : This is the same as TickEveryDay but involves 2 ticks in every call. - hence the name pulse
//
//   - clock	: time at which the pulse is initiated everyday
//
//   - pulse	: gap in the pulse - since 2 ticks make a pulse
//
//   - canc		: interruption channel
func PulseEveryDayAt(clock string, pulse time.Duration, ctx context.Context, wg *sync.WaitGroup) (chan time.Time, error) {
	ticks := make(chan time.Time, 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ticks)
		// time calculations have to be done only inside the go routine since scheduling time of this routine is indeterminate
		// only when the coroutine gets scheduled can you do all the time calculations.
		// offDuration, offset := calc_tickOffset(hr, min)
		tt, offset, err := tickTimeUnix(clock)
		if err != nil {
			log.Error(err)
			return
		}
		offDuration, _ := time.ParseDuration(fmt.Sprintf("%ds", offset))
		if offset >= 0 {
			// case where time until tick duration, so sleeping
			/*
				|--- now < you are here
				------- offset (sleep)
				|--- tick
				--- pulse duration(sleep)
				|--- pulse
			*/
			log.WithFields(log.Fields{"offset": offDuration}).Debug("Time until tick")
			select {
			case <-time.After(offDuration):
				ticks <- time.Now()
			case <-ctx.Done():
				return
			}
			select {
			case <-time.After(pulse):
				ticks <- time.Now()
			case <-ctx.Done():
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
				/*
					------ tick
					|- Sleep (pulse duration)
					------ now < you are here (send the tick that was missed)
					------- pulse
				*/
				ticks <- time.Now()
				log.Debug("we are between the pulses")
				select {
				case <-time.After(time.Until(end)): // this will be less than the pulse since the elapsed time has to be subtracted
					ticks <- time.Now()
				case <-ctx.Done():
					return
				}
			} else {
				/*
					------ tick
					|- Sleep(pulse duration)
					------ pulse
					------- now < you are here (no ticks are sent, since the entire pulse is missed)
				*/
				log.WithFields(log.Fields{
					"elapsed": time.Until(end),
				}).Info("We are beyond ticking time & pulse duration")
			}
			// Since we are already pass the pulsing time, the next one pulse shall start for less than 24 hours
			// offset here is negative, hence the final offset calculated would have to be less than 24 hours / 86400 seconds
			/*
				I --- tick time
				I |- sleep
				I --- pulse end
				I |- now (in either of the cases since now is beyond tick time, the second cycle is offset-shy of 24 hours
			*/
			seconds := 86400 - (time.Now().Unix() - tt) // for the second tick
			dur, _ := time.ParseDuration(fmt.Sprintf("%ds", seconds))
			log.WithFields(log.Fields{
				"duration": dur,
			}).Debug("ofsetted day")
			select {
			case <-time.After(dur):
				ticks <- time.Now()
			case <-ctx.Done():
				return
			}
			select {
			case <-time.After(pulse):
				ticks <- time.Now()
			case <-ctx.Done():
				return
			}
		}
		for t := range PulseEvery(daily, pulse, ctx, wg) {
			ticks <- t
		}
	}()
	return ticks, nil
}
