package main

/* ===========
What started out to be a project for simple patio lighting with a set of relays that are driven by the clock likke a cron job has evolved into a prototype project to control the water pump of the aquaponics system. Relays are now thrown to drive the pump for scheduled intervals like pulse

author		: kneerunjun@gmail.com
date		: 21-2-2024
place		: Pune
=============== */
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type ScheduleType uint8

const (
	TICK_EVERY ScheduleType = iota // for every interval the relay just ticks
	TICK_EVERY_DAYAT
	PULSE_EVERY
	PULSE_EVERY_DAYAT
)

var (
	config = AppConfig{}
)

// AppConfig : object model that captures the configuration for the app in a single run
// configuration is loaded in the memory once in init, and then stays for the life of the appliation
// Any change in the configuration has to be enforced my restarting the application
type AppConfig struct {
	AppName  string `json:"appname"`
	Schedule struct {
		Config   ScheduleType `json:"config"`
		TickAt   string       `json:"tickat"`
		PulseGap int          `json:"pulsegap,omitempty"`
	} `json:"schedule"`
	Gpio struct {
		Touch  string `json:"touch"`
		ErrLed string `json:"errled"`
		Relays struct {
			Pump string `json:"pump"`
		} `json:"relays"`
	} `json:"gpio"`
}

func init() {
	// Setting up the loggin framework
	log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: false})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// TODO: read configuration file and hoist all the vars
	f, err := os.Open("/etc/aquapone.config.json")
	if err != nil || f == nil {
		log.Panicf("failed to access config.json %s", err)
		return
	}
	byt, err := io.ReadAll(f)
	if err != nil {
		log.Panicf("failed to read config.json %s", err)
		return
	}

	if err := json.Unmarshal(byt, &config); err != nil {
		log.Panicf("failed to unmarshal config.json %s", err)
		return
	}
	log.WithFields(log.Fields{
		"name":         config.AppName,
		"sched":        config.Schedule.Config,
		"tick":         config.Schedule.TickAt,
		"pulsegap":     config.Schedule.PulseGap,
		"touchpin":     config.Gpio.Touch,
		"errledpin":    config.Gpio.ErrLed,
		"pumprelaypin": config.Gpio.Relays.Pump,
	}).Debug("read in app config")
}

// this main loop would only setup the tickers
func main() {
	log.WithFields(log.Fields{
		"time": time.Now().Format(time.RFC822),
	}).Debugf("Starting %s", config.AppName)
	defer log.Warn("Closing the application now..")

	// Contexts and waitgroups
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// initialized hardware drivers
	r := raspi.NewAdaptor()
	r.Connect()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for intr := range interrupt.TouchOrSysSignal(config.Gpio.Touch, digital.SLOW_WATCH_3_3V, r, ctx, &wg) {
			log.WithFields(log.Fields{
				"time": intr.Format(time.RFC822),
			}).Warn("Interrupted...")
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

		rs := digital.NewRelaySwitch(config.Gpio.Relays.Pump, false, r).Boot()
		ticks, _ := tickers.PulseEveryDayAt(config.Schedule.TickAt, time.Duration(config.Schedule.PulseGap)*time.Second, ctx, &wg)

		for t := range ticks {
			log.Debugf("Flipping the relay state: %s", t.Format(time.RFC822))
			rs.Toggle()
		}

		rs.Low()
		log.Warn("Now shutting down relay..")
	}()
	// Flushing the hardware states
	wg.Wait()

}
