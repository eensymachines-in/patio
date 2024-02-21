package main

import(
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	log "github.com/sirupsen/logrus"
	_ "gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
  "strings"
  "strconv"
)
const (
  daily = 1*time.Hour
)
func init(){
	log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: false})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
//TickEveryDayAt : for a given clock time like 13:40,it will send ticks separated by 24 hours 
// closing the canc channel will bring down the loop and close all the ticks
func TickEveryDayAt(clock string, canc <-chan bool) (chan time.Time, error){
  ticks  := make(chan time.Time, 1)
  // get time of the day, and figure out if the now time is ahead or past the time of tick 
  hrmin := strings.Split(clock, ":") // clock format expected - 13:09, 24 hour clock with minute resolution
  if len(hrmin) != 2 {
    // this is when the clock isnt as expected. 
    return nil, fmt.Errorf("invalid clock format, expected format is 13:04")
  }
  hr, _ := strconv.ParseInt(hrmin[0], 10, 32)
  min, _ := strconv.ParseInt(hrmin[1], 10, 32)
  go func(){
    defer close(ticks) 
    // time calculations have to be done only inside the go routine since scheduling time of this routine is indeterminate
    // only when the coroutine gets scheduled can you do all the time calculations.
    now := time.Now()// getting the current time 
    h,m,s := now.Clock()
    midnight := now.Unix() - int64(s +(m*60)+(h*3600)) // tracing the midnight time 
    tickTime := midnight + (hr*3600) + (min*60)
    log.WithFields(log.Fields{
      "ticktime": tickTime,
      "nowtime" : now.Unix(),
    }).Debug("clock times")
    offset := tickTime - now.Unix() // getting the offset of the now to the time of tick
    offDuration,_ := time.ParseDuration(fmt.Sprintf("%ds", offset)) // converting those seconds of ffset to duration offset
    log.WithFields(log.Fields{"offset": offDuration}).Debug("time until next tick")

    if offset >=0 {
      // case where time until tick duration, so sleeping
      <-time.After(offDuration)
      ticks <- now
    } else {
      //this is a tricky situation when the ticking time for the day has already elapsed 
      // 24 hour cycle does not apply, for the next tick but you have to send an extra tick for the tick that has elapsed 
      ticks <- now
      then := now.Add(daily).Add(offDuration)  // offDuration is negative in this case
      <-time.After(then.Sub(now))//this is less than daily duration  
      ticks <- time.Now()
    }
    for t := range TickEvery(daily, canc){
      ticks <- t
    }
  }()
  return ticks, nil
}

// PulseEvery : after every d duration it would tick twice separated by w duration
// canc channel will kill the loop and close the channel
// d > w always
func PulseEvery(d, w time.Duration, canc <-chan bool) chan time.Time{
	ticks := make(chan time.Time, 1)
	go func (){
		defer close(ticks)
		for {
			select {
			case <-time.After(d):
				ticks <- time.Now()
				<-time.After(w)
				ticks <-time.Now()
			case <- canc:
				return
			}
		}
	}()
	return ticks
}
// For the given duration this can send channel messages for the current time over an over till canceled 
// use the canc channel to kill the loop 
func TickEvery(d time.Duration, canc <- chan bool) chan time.Time {
	ticks := make(chan time.Time, 1)
	//sets up the loop for ticking, can be closed only if the cancel channel closed.
	go func (){
		defer close(ticks)
		for {
			select {
			case <-time.After(d):
				ticks <- time.Now()
			case <- canc:
				return
			}
		}
	}()
	return ticks
}
// this main loop would only setup the tickers
func main(){
	fmt.Println("We are inside the patio program ..")
	
	cancel := make(chan bool)
	signals := make(chan os.Signal,1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer close(signals)
  go func(){
	// cannot flush hardware here since after cancel is closed, the program will exit
	// upon getting the signal we just call off all the loops.
		<- signals
		close(cancel)
	}() // flushing the hardware states 

	// initialized hardware drivers
	r := raspi.NewAdaptor()
  relay := gpio.NewDirectPinDriver(r, "40")
	relay.On() // the chinese relay has closure on GPIO low, open on GPIO high
  //Setup work for the bot
	log.Debug("Initialized Pi connection..")
  ticks, _ := TickEveryDayAt("05:20", cancel)
  for t:= range ticks{
		log.Debug(t.String())
    on,_ := relay.DigitalRead()
		if on ==1 {
			relay.Off()
		}else {
			relay.On()
		}
	}
	// Flushing the hardware states 
	relay.On()
}
