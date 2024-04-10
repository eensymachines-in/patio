package aquacfg

type ScheduleType uint8

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

// AppConfig : object model that captures the configuration for the app in a single run
// configuration is loaded in the memory once in init, and then stays for the life of the appliation
// Any change in the configuration has to be enforced my restarting the application
type AppConfig struct {
	AppName  string   `json:"appname"`
	Schedule Schedule `json:"schedule"`
}
