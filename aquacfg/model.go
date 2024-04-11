package aquacfg

import "regexp"

type ScheduleType uint8

const (
	INTERVAL_MIN = 10 // short intervals can be detrimental to the life of the relays
)

const (
	TICK_EVERY ScheduleType = iota // for every interval the relay just ticks
	TICK_EVERY_DAYAT
	PULSE_EVERY
	PULSE_EVERY_DAYAT
)

type Schedule struct {
	Config   ScheduleType `json:"config"`             // ticking algorithm
	TickAt   string       `json:"tickat"`             // time of the day ticking /pulsing starts at
	PulseGap int          `json:"pulsegap,omitempty"` // pulse width incase its pulsing
	Interval int          `json:"interval,omitempty"` // ticking interval incase its ticking
}

// IsValid : for the given schedule it checks to see if configuration is not conflicting
// Depending on what is the schedule type certain values may lead to
func (sched *Schedule) IsValid() bool {
	if (sched.Config == TICK_EVERY || sched.Config == PULSE_EVERY) && sched.Interval <= INTERVAL_MIN {
		// interval cannot be so short - short intervals can lead to shortened life of the relays
		return false
	}
	if (sched.Config == PULSE_EVERY || sched.Config == PULSE_EVERY_DAYAT) && sched.PulseGap <= INTERVAL_MIN {
		// pulse gap cannot be less than a threshold since it would be then detrimental to the relay life
		return false
	}
	if sched.Config == PULSE_EVERY && (sched.Interval <= sched.PulseGap) {
		// At no time the pulsegap can be greater than interval
		return false
	}
	if sched.Config == TICK_EVERY_DAYAT || sched.Config == PULSE_EVERY_DAYAT {
		// time has to specifed for 2 particular configuration that are clock driven
		expr := regexp.MustCompile(`^[0-9]{2}:[0-9]{2}$`)
		if !expr.MatchString(sched.TickAt) {
			return false
		}
	}
	return true
}

// AppConfig : object model that captures the configuration for the app in a single run
// configuration is loaded in the memory once in init, and then stays for the life of the appliation
// Any change in the configuration has to be enforced my restarting the application
type AppConfig struct {
	AppName  string   `json:"appname"`
	Schedule Schedule `json:"schedule"`
}
