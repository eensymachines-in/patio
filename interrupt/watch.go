package interrupt

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eensymachines-in/patio/digital"
	"github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
)

// SysSignalWatch : watches incoming system interrupts and notifies all other threads via unbuffered channel
func SysSignalWatch() chan bool {
	interrupt := make(chan bool)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		// thread that watches sys signals
		defer close(signals)
		// no need to close interrupt, since either of the cases it will be closed
		for {
			select {
			case <-signals: // when system interrupts, notify all processes
				close(interrupt)
				return
			case <-interrupt: // when other processes interrupt, stop this watch on signals
				return
			}
		}
	}()
	return interrupt
}

// TouchSensorWatch : watches grove touch sensor signal and interprets the same as interrupt signal
// pin 				: pin on the SoC where the touch sensor is connected
// adp				: connection adaptor for the SoC
func TouchSensorWatch(pin string, adp gobot.Adaptor) chan bool {
	interrupt := make(chan bool)
	touch := digital.NewTouchSensor(pin, adp).Boot()
	go func() {
		for t := range touch.Watch(interrupt) {
			logrus.WithFields(logrus.Fields{
				"time": t.Format(time.RFC822),
			}).Warn("touch interrupt..")
			close(interrupt)
			return
		}
	}()
	return interrupt
}

// TouchOrSysSignal : Or combination for system interrupts & touch sensor button whichever occurs first
func TouchOrSysSignal(pin string, adp gobot.Adaptor) chan bool {
	interrupt := make(chan bool)
	touch := digital.NewTouchSensor(pin, adp).Boot()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		defer close(signals)
		for {
			select {
			case <-touch.Watch(interrupt):
				logrus.WithFields(logrus.Fields{
					"time": time.Now().Format(time.RFC822),
				}).Warn("button interrupt..")
				close(interrupt)
				return
			case <-signals:
				logrus.WithFields(logrus.Fields{
					"time": time.Now().Format(time.RFC822),
				}).Warn("system interrupt..")
				close(interrupt)
				return
			case <-interrupt:
				return
			}
		}
	}()
	return interrupt
}
