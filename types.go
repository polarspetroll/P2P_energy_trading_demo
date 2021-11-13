package main

import (
	"time"

	"github.com/polarspetroll/gopio"
)

type Config struct {
	Pins     []int  `json:"relays"`
	Interval string `json:"interval"`
	InaAddrs []int  `json:"ina_addresses"`
}

type HTML struct {
	Username string
	Message  string
}

type Trial struct {
	TimeLeft  time.Duration
	UnitsLeft float64
	Price     int
}

type User struct {
	Username  string
	Password  string
	Condition Trial
	RelayPin  gopio.WiringPiPin
	InaAddr   int
}

type SID struct {
	Sid      string
	Username string
}
