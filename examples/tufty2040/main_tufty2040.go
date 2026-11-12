//go:build tufty2040

// Command tufty2040 is the TinyGUI showcase firmware for the Pimoroni Tufty 2040.
//
//	tinygo flash -target=tufty2040 ./examples/tufty2040
//
// Board code lives here; the UI is in ./app and runs unchanged in ./sim.
package main

import (
	"device/arm"
	"io"
	"machine"
	"machine/usb"
	"runtime"
	"time"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/examples/tufty2040/app"
	"github.com/itohio/tinygui/examples/tufty2040/board"
)

// The frame is drawn in three 80-row bands into a 51.2 KB buffer: widgets draw
// at memory speed, each band is pushed in one burst (no clear-then-draw
// flicker), and the rest of the frame's 153.6 KB stays free for the heap. A
// full-frame buffer left ≈ 45 KB of heap and ran out of memory; two 120-row
// bands still ran out when the PNG decoder (≈ 13 KB per decode) met a
// fragmented heap on the Images page after a tour of the other pages.
const bandRows = app.ScreenH / 3

var pixels [app.ScreenW * bandRows]uint16

const frameTime = 33 * time.Millisecond

// setUSBSerial makes the USB serial number the flash chip's unique ID in hex,
// the same value MicroPython reported (df6750955f2a3536 on the owner's badge),
// so host tools find the badge by serial number instead of by port name. It
// runs first in main, before the host asks for the string descriptor; in an
// init() the flash command stops TinyGo from evaluating package initialisers
// at compile time (22 KB of tables moved from .data to runtime init).
func setUSBSerial() {
	const hex = "0123456789abcdef"
	id := machine.DeviceID()
	s := make([]byte, 0, 2*len(id))
	for _, b := range id {
		s = append(s, hex[b>>4], hex[b&15])
	}
	usb.Serial = string(s)
}

func main() {
	setUSBSerial()
	waitForConsole(3 * time.Second)
	boot("start")
	var panel board.LCD
	panel.Configure()
	boot("lcd")
	hw := newBadge()
	hw.SetBacklight(0)
	boot("badge")

	dumping := false
	fb := framebuffer.NewBand(app.ScreenW, app.ScreenH, bandRows, pixels[:], func(f *framebuffer.RGB565) error {
		rows := f.Pix[:int(min(f.H, f.ScreenH-f.Y0))*int(f.W)]
		panel.Push(f.Y0, rows)
		if dumping {
			dumpRows(rows)
		}
		return nil
	})
	a := app.New(fb, hw, app.Config{
		DurationAdd: ui.DurationAdd,
		DurationSub: ui.DurationSub,
		Log:         func(s string) { println(s) },
	}, time.Now())
	boot("app")

	// USB serial console: line commands, see app.Handlers ("help").
	handlers := a.Handlers()
	handlers["madctl"] = func(arg []byte) { // orientation probe: madctl <decimal>
		v := 0
		for _, c := range arg {
			v = v*10 + int(c-'0')
		}
		panel.SetMADCTL(uint8(v))
	}
	handlers["reboot"] = func([]byte) { arm.SystemReset() } // fresh state for scripted tours
	handlers["pins"] = func([]byte) {
		println("A", machine.BUTTON_A.Get(), "B", machine.BUTTON_B.Get(), "C", machine.BUTTON_C.Get(),
			"UP", machine.BUTTON_UP.Get(), "DOWN", machine.BUTTON_DOWN.Get(), "USER", machine.BUTTON_USER.Get(),
			"VBUS", machine.VBUS_DETECT.Get())
	}
	mux := ui.NewCommandStreamMux(idleReader{ui.NewSerialReader(machine.Serial)}, handlers)
	go mux.Run()

	buttons := []*button{
		newButton(machine.BUTTON_UP, app.BtnUp, true),
		newButton(machine.BUTTON_DOWN, app.BtnDown, true),
		newButton(machine.BUTTON_A, app.BtnA, false),
		newButton(machine.BUTTON_B, app.BtnB, false),
		newButton(machine.BUTTON_C, app.BtnC, false),
	}
	machine.BUTTON_USER.Configure(machine.PinConfig{Mode: machine.PinInputPullup}) // active low

	first := true
	var last time.Time
	for {
		now := time.Now()
		for _, b := range buttons {
			b.poll(a, now)
		}
		if !machine.BUTTON_USER.Get() {
			a.UserHold(ui.PeekButton(machine.BUTTON_USER)) // blocks until release
		}
		if now.Sub(last) >= frameTime {
			last = now
			a.Update(now)
			if dumping = a.TakeDump(); dumping {
				println("FB", app.ScreenW, app.ScreenH)
			}
			var draw, push time.Duration
			for i := range fb.Bands() {
				fb.SetY0(int16(i * bandRows))
				t0 := time.Now()
				a.Draw()
				t1 := time.Now()
				_ = fb.Display()
				draw += t1.Sub(t0)
				push += time.Since(t1)
			}
			a.ReportTiming(draw, push)
			if dumping {
				println()
				println("END")
			}
			if first { // light the panel only after the first full frame
				first = false
				hw.SetBacklight(80)
				boot("first frame")
			}
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// idleReader sleeps when the serial port has nothing buffered, so the mux
// goroutine (which yields with runtime.Gosched on every read) stops spinning.
type idleReader struct{ r io.Reader }

func (i idleReader) Read(b []byte) (int, error) {
	n, err := i.r.Read(b)
	if n == 0 && err == nil {
		time.Sleep(5 * time.Millisecond)
	}
	return n, err
}

// dumpRows streams pixels over USB serial as big-endian RGB565. The frame
// loop frames a full dump as: line "FB 320 240", 153 600 bytes (both bands),
// newline, line "END" — a raw screenshot without a camera.
func dumpRows(pix []uint16) {
	var buf [512]byte
	n := 0
	for _, p := range pix {
		buf[n], buf[n+1] = byte(p>>8), byte(p)
		n += 2
		if n == len(buf) {
			_, _ = machine.Serial.Write(buf[:n])
			n = 0
		}
	}
	if n > 0 {
		_, _ = machine.Serial.Write(buf[:n])
	}
}

// waitForConsole gives a host up to d to open the USB serial port, so the
// boot log below is not lost.
func waitForConsole(d time.Duration) {
	for end := time.Now().Add(d); time.Now().Before(end) && !machine.Serial.DTR(); {
		time.Sleep(10 * time.Millisecond)
	}
}

// boot prints a boot stage with the heap state.
func boot(stage string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	println("boot:", stage, "heap_alloc", m.HeapAlloc, "heap_sys", m.HeapSys)
}
