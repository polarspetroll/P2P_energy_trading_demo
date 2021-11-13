package main

import (
	"periph.io/x/periph/conn/i2c/i2creg"
	//"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func init() {
	host.Init()
}

func (u *User) AddConsumption() {
	bus, _ := i2creg.Open("")
	defer bus.Close()

	opts := ina219.DefaultOpts
	opts.Address = u.InaAddr

	sensor, _ := ina219.New(bus, &opts)
	measurement, _ := sensor.Sense()
	u.Condition.UnitsLeft -= float64(measurement.Power)
	mutex2.Lock()
	total_consumption += float64(measurement.Power)
	mutex2.Unlock()
}
