//go:build tufty2040

// Command bringup checks the Tufty 2040 hardware one piece at a time, with no
// TinyGUI, no framebuffer and no PNG decoder:
//
//  1. user LED blinks twice per second for the whole run (firmware is alive);
//
//  2. backlight GPIO2 driven high as a plain output (no PWM) before the panel
//     is initialised: the glass should glow;
//
//  3. panel init, then the whole screen cycles red, green, blue, white, black,
//     one second each, forever;
//
//  4. USB serial heartbeat once per colour with the heap state.
//
//     tinygo flash -target=tufty2040 ./examples/tufty2040/bringup
package main

import (
	"machine"
	"runtime"
	"time"

	"github.com/itohio/tinygui/examples/tufty2040/board"
)

var colours = []struct {
	name string
	c565 uint16
}{
	{"red", 0xF800}, {"green", 0x07E0}, {"blue", 0x001F}, {"white", 0xFFFF}, {"black", 0x0000},
}

func main() {
	// give a host up to 3 s to open the USB serial port so the log isn't lost
	for end := time.Now().Add(3 * time.Second); time.Now().Before(end) && !machine.Serial.DTR(); {
		time.Sleep(10 * time.Millisecond)
	}
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	go func() {
		for {
			led.High()
			time.Sleep(250 * time.Millisecond)
			led.Low()
			time.Sleep(250 * time.Millisecond)
		}
	}()

	bl := machine.LCD_BACKLIGHT
	bl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	bl.High()
	log("backlight on (GPIO2 high), panel not initialised yet")
	time.Sleep(2 * time.Second)

	var lcd board.LCD
	lcd.Configure()
	log("panel initialised, cpu_hz " + itoa(int(machine.CPUFrequency())))

	for frame := 0; ; frame++ {
		c := colours[frame%len(colours)]
		t0 := time.Now()
		lcd.Fill(c.c565)
		log("fill " + c.name + " us " + itoa(int(time.Since(t0).Microseconds())))
		time.Sleep(time.Second)
	}
}

func log(s string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	println("bringup:", s, "heap_alloc", m.HeapAlloc, "heap_sys", m.HeapSys)
}

func itoa(v int) string {
	var b [12]byte
	i := len(b)
	for {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			return string(b[i:])
		}
	}
}
