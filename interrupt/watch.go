package interrupt

/* ==========================
Inerrupts are a common requirement when it comes to SoCs and IoT devices.
System interrupts when the OS decides to call it off
Manual intervention - tactile buttons & touch sensors
Each of the watches runs an infinite loop to capture all t he interrupts as thety come in,. unless ofcourse the client calls the cancel function to stop the watch
==========================*/
import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eensymachines-in/patio/digital"
	"github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
)

// SysSignalWatch : watches system interruptions and sends the signal over interrupt channel
// loop will continue watching till context is cancelled from client side
//
/*
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for t := range SysSignalWatch(ctx, &wg){
		log.Debug("received system interruption, system closing now")
		cancel()
	}
*/
func SysSignalWatch(ctx context.Context, wg *sync.WaitGroup) chan time.Time {
	interrupt := make(chan time.Time, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(signals)
		defer close(interrupt)
		defer logrus.Warn("Now closing loop for SysSignalWatch")

		for {
			select {
			case <-signals: // when system interrupts, notify all processes
				logrus.WithFields(logrus.Fields{
					"time": time.Now().Format(time.RFC822),
				}).Warn("system interrupt..")
				interrupt <- time.Now()
			case <-ctx.Done(): // when other processes interrupt, stop this watch on signals
				return
			}
		}
	}()
	return interrupt
}

// TouchSensorWatch : watches grove touch sensor signal and interprets the same as interrupt signal
// pin 				: pin on the SoC where the touch sensor is connected
// adp				: connection adaptor for the SoC
//
/*
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	r := raspi.NewAdaptor()
	err := r.Connect()
	if err != nil {
		log.Panicf("failed to connect to raspberry device %s", err)
	}
	for t := range TouchSensorWatch("PHY_PIN_NUM",r,ctx, &wg){
		log.Debug("received system interruption, system closing now")
		cancel()
	}
*/
func TouchSensorWatch(pin string, speed time.Duration, adp gobot.Adaptor, ctx context.Context, wg *sync.WaitGroup) chan time.Time {
	interrupt := make(chan time.Time, 1)
	touch := digital.NewTouchSensor(pin, adp).Boot()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(interrupt)
		defer logrus.Warn("Now closing loop for TouchSensorWatch")
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-touch.Watch(speed, ctx, wg):
				logrus.WithFields(logrus.Fields{
					"time": t.Format(time.RFC822),
				}).Warn("touch interrupt..")
				interrupt <- t
			}
		}
	}()
	return interrupt
}

// TouchOrSysSignal : Or combination for system interrupts & touch sensor button whichever occurs first
// pin		: physical pin at which the button is connected to
// adp		: connection to the device
//
/*
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	r := raspi.NewAdaptor()
	err := r.Connect()
	if err != nil {
		log.Panicf("failed to connect to raspberry device %s", err)
	}
	for t := range TouchOrSysSignal("PHY_PIN_NUM",r,ctx, &wg){
		log.Debug("received interruption")
		cancel()
	}

*/
func TouchOrSysSignal(pin string, speed time.Duration, adp gobot.Adaptor, ctx context.Context, wg *sync.WaitGroup) chan time.Time {
	interrupt := make(chan time.Time, 1)
	touch := digital.NewTouchSensor(pin, adp).Boot()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(signals)
		defer close(interrupt)
		defer logrus.Warn("Now closing loop for TouchOrSysSignal")
		for {
			select {
			case <-touch.Watch(speed, ctx, wg):
				logrus.WithFields(logrus.Fields{
					"time": time.Now().Format(time.RFC822),
				}).Warn("button interrupt..")
				interrupt <- time.Now()
			case <-signals:
				logrus.WithFields(logrus.Fields{
					"time": time.Now().Format(time.RFC822),
				}).Warn("system interrupt..")
				interrupt <- time.Now()
			case <-ctx.Done():
				return
			}
		}
	}()
	return interrupt
}
