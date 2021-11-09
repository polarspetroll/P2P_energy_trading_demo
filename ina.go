package main

import (
	"periph.io/x/periph/conn/i2c/i2creg"
	//	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func (u *User) AddConsumption() {
	host.Init()
	bus, _ := i2creg.Open("")
	defer bus.Close()
	sensor, _ := ina219.New(bus, &ina219.DefaultOpts)
	measurement, _ := sensor.Sense()
	u.Condition.UnitsLeft -= float64(measurement.Power)
	total_consumption += float64(measurement.Power)
}
