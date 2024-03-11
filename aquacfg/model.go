package aquacfg

type ScheduleType uint8

const (
	TICK_EVERY ScheduleType = iota // for every interval the relay just ticks
	TICK_EVERY_DAYAT
	PULSE_EVERY
	PULSE_EVERY_DAYAT
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
	Rabbit struct {
		Server string `json:"server"`
		Login  string `json:"login"`
	} `json:"rabbit"`
}
