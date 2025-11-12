//go:build stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid

package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// ScanI2CDevices scans the I2C bus for connected devices and returns their addresses.
func ScanI2CDevices(bus drivers.I2C) []uint16 {
	var devices []uint16
	for addr := uint16(1); addr < 127; addr++ {
		err := bus.Tx(addr, nil, nil)
		if err == nil {
			devices = append(devices, addr)
		}
		time.Sleep(1 * time.Millisecond) // avoid bus flooding
	}
	return devices
}

func main() {
	bus := machine.I2C0
	bus.Configure(machine.I2CConfig{})

	time.Sleep(500 * time.Millisecond) // wait for devices to power up

	devices := ScanI2CDevices(bus)
	for _, addr := range devices {
		println("Found device at address: 0x", hex(addr))
	}
}

// Helper: convert uint8 to hex string
func hex(b uint16) string {
	const hexChars = "0123456789ABCDEF"
	return string([]byte{hexChars[b>>4], hexChars[b&0x0F]})
}
