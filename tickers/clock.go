package tickers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Time denominations commonly used
const (
	daily     = 24 * time.Hour
	halfdaily = 12 * time.Hour
	hourly    = 1 * time.Hour
	minutely  = 1 * time.Minute
	secondly  = 1 * time.Second
)

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
