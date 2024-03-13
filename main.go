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
	"strconv"
	"sync"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/eensymachines-in/patio/digital"
	"github.com/eensymachines-in/patio/interrupt"
	"github.com/eensymachines-in/patio/tickers"
	oled "github.com/eensymachines-in/ssd1306"
	log "github.com/sirupsen/logrus"
	_ "gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	config = aquacfg.AppConfig{}
)

func init() {
	// required environment variables
	/*
		PATH_APPCONFIG
		NAME_SYSCTLSERVICE
		MODE_DEBUGLVL
		AMQP_LOGIN
		AMQP_SERVER
		AMQP_CFGCHNNL
		GPIO_TOUCH
		GPIO_ERRLED
		GPIO_PUMP_MAIN
	*/
	for _, v := range []string{
		"PATH_APPCONFIG",
		"NAME_SYSCTLSERVICE",
		"MODE_DEBUGLVL",
		"AMQP_LOGIN",
		"AMQP_SERVER",
		"AMQP_CFGCHNNL",
		"GPIO_TOUCH",
		"GPIO_ERRLED",
		"GPIO_PUMP_MAIN",
	} {
		if val := os.Getenv(v); val == "" {
			log.Panicf("Required environment variable missing in ~/.bashrc: %s", v)
		}
	}
	// Setting up the loggin framework
	log.SetFormatter(&log.TextFormatter{DisableColors: false, FullTimestamp: false})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)

	lvl, err := strconv.Atoi(os.Getenv("MODE_DEBUGLVL"))
	if err != nil {
		log.Warnf("invalid env var value for logging level, only integers %s", os.Getenv("MODE_DEBUGLVL"))
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.Level(lvl)) // sets from the environment
	}

	// TODO: read configuration file and hoist all the vars
	f, err := os.Open(os.Getenv("PATH_APPCONFIG"))
	if err != nil || f == nil {
		log.Panicf("failed to access application configuration file %s", err)
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
		"name":     config.AppName,
		"sched":    config.Schedule.Config,
		"tick":     config.Schedule.TickAt,
		"pulsegap": config.Schedule.PulseGap,
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
		for intr := range interrupt.TouchOrSysSignal(os.Getenv("GPIO_TOUCH"), digital.SLOW_WATCH_3_3V, r, ctx, &wg) {
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

		rs := digital.NewRelaySwitch(os.Getenv("GPIO_PUMP_MAIN"), false, r).Boot()
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
