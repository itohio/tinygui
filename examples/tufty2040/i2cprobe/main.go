//go:build tufty2040

// i2cprobe answers "what is on the Tufty 2040's I2C bus?" over USB serial.
// The bus (GPIO4 SDA, GPIO5 SCL, GPIO3 INT) goes to the Qw/ST connector and
// the debug pads. Every 5 s it prints:
//
//  1. the idle levels of SDA/SCL with the RP2040's pulls off and with
//     pull-downs (high in both: external pull-ups on the bus);
//  2. a scan of 0x08..0x77 with the hardware I2C controller, with
//     WHO_AM_I-style registers of every device that answers;
//  3. a positive control: the RP2040's own I2C0 block listens as a target
//     at 0x42 on the same pins, and a software controller (driving the pads
//     through their output overrides, so the target still sees the bus)
//     scans 0x08..0x77. 0x42 must answer: that proves an ACK on these wires
//     would be seen.
package main

import (
	"device/rp"
	"machine"
	"runtime/volatile"
	"time"

	"github.com/itohio/tinygui/examples/tufty2040/board"
)

const controlAddr = 0x42

func main() {
	board.SetUSBSerial()
	for {
		time.Sleep(5 * time.Second)
		println("i2cprobe: --- pass ---")
		pullups()
		hwScan()
		control()
	}
}

func pullups() {
	sda, scl := machine.GPIO4, machine.GPIO5
	sda.Configure(machine.PinConfig{Mode: machine.PinInput})
	scl.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(time.Millisecond)
	a, b := sda.Get(), scl.Get()
	sda.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	scl.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(time.Millisecond)
	c, d := sda.Get(), scl.Get()
	println("i2cprobe: idle SDA", b2i(a), "SCL", b2i(b), "| with pull-downs SDA", b2i(c), "SCL", b2i(d))
}

func hwScan() {
	bus := machine.I2C0
	if err := bus.Configure(machine.I2CConfig{SDA: machine.GPIO4, SCL: machine.GPIO5, Frequency: 100_000}); err != nil {
		println("i2cprobe: configure:", err.Error())
		return
	}
	var buf [1]byte
	acks, noack, other := 0, 0, 0
	for addr := uint16(0x08); addr <= 0x77; addr++ {
		err := bus.Tx(addr, nil, buf[:])
		switch {
		case err == nil:
			acks++
			println("i2cprobe: hw ACK at", hex(addr))
			whoAmI(bus, addr)
		case err.Error() == "i2c: generic error": // TinyGo's "7-bit address no ack"
			noack++
		default:
			other++
			println("i2cprobe: hw", hex(addr), err.Error())
		}
	}
	println("i2cprobe: hw scan 0x08..0x77: acks", acks, "no-ack", noack, "other", other)
}

// whoAmI reads the identity registers of common IMUs and sensors.
func whoAmI(bus *machine.I2C, addr uint16) {
	var v [1]byte
	for _, reg := range []byte{0x0F, 0x75, 0x00, 0x0D, 0xD0} {
		if err := bus.Tx(addr, []byte{reg}, v[:]); err == nil {
			println("i2cprobe:   reg", hex(uint16(reg)), "=", hex(uint16(v[0])))
		}
	}
}

func control() {
	bus := machine.I2C0
	if err := bus.Configure(machine.I2CConfig{SDA: machine.GPIO4, SCL: machine.GPIO5, Mode: machine.I2CModeTarget}); err != nil {
		println("i2cprobe: target configure:", err.Error())
		return
	}
	if err := bus.Listen(controlAddr); err != nil {
		println("i2cprobe: listen:", err.Error())
		return
	}
	acks := 0
	for addr := uint8(0x08); addr <= 0x77; addr++ {
		if bbProbe(addr) {
			acks++
			println("i2cprobe: control ACK at", hex(uint16(addr)))
		}
	}
	println("i2cprobe: control scan (RP2040 target at", hex(controlAddr), "): acks", acks)
}

// ---- software controller through the pads' output overrides ----

var (
	sdaCtrl = &rp.IO_BANK0.GPIO4_CTRL
	sclCtrl = &rp.IO_BANK0.GPIO5_CTRL
)

const (
	ovMask = 3<<12 | 3<<8 // OEOVER, OUTOVER
	ovLow  = 3<<12 | 2<<8 // output enabled, driven low
)

func low(r *volatile.Register32)     { r.Set(r.Get()&^ovMask | ovLow) }
func release(r *volatile.Register32) { r.Set(r.Get() &^ ovMask) } // back to the peripheral (open drain)

func half() { time.Sleep(20 * time.Microsecond) }

func sclHigh() {
	release(sclCtrl)
	for i := 0; i < 1000 && !machine.GPIO5.Get(); i++ { // clock stretching
		time.Sleep(time.Microsecond)
	}
	half()
}

// bbProbe sends START, the address with W, reads the ACK bit, STOP.
func bbProbe(addr uint8) bool {
	release(sdaCtrl)
	release(sclCtrl)
	half()
	low(sdaCtrl) // START
	half()
	low(sclCtrl)
	b := addr << 1
	for i := 0; i < 8; i++ {
		if b&0x80 != 0 {
			release(sdaCtrl)
		} else {
			low(sdaCtrl)
		}
		b <<= 1
		half()
		sclHigh()
		low(sclCtrl)
	}
	release(sdaCtrl) // ACK clock
	half()
	sclHigh()
	ack := !machine.GPIO4.Get()
	low(sclCtrl)
	low(sdaCtrl) // STOP
	half()
	sclHigh()
	release(sdaCtrl)
	half()
	return ack
}

func b2i(v bool) int {
	if v {
		return 1
	}
	return 0
}

func hex(v uint16) string {
	const d = "0123456789abcdef"
	return "0x" + string([]byte{d[v>>4&15], d[v&15]})
}
