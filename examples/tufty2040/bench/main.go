//go:build tufty2040

// Command bench measures what the Tufty 2040 can do per frame, to plan demos:
// panel pushes at three colour depths and partial windows, simulation steps
// (Game of Life, fire) at several grid sizes, and the cost of rendering a
// grid as round LED-matrix cells or as ×2 upscaled pixels. No TinyGUI: these
// are the raw costs a TinyGUI widget would sit on. Results go to the USB
// console as "bench: <name> <us per op>"; the screen shows each stage.
//
//	tinygo flash -target=tufty2040 ./examples/tufty2040/bench
package main

import (
	"machine"
	"time"
	"unsafe"

	"github.com/itohio/tinygui/examples/tufty2040/board"
)

const W, H = board.Width, board.Height

// One 153.6 KB buffer, viewed as RGB565, 8-bit or 4-bit pixels per test.
var raw [W * H * 2]byte

func fb565() []uint16 { return unsafe.Slice((*uint16)(unsafe.Pointer(&raw[0])), W*H) }

// Simulation grids (bytes per cell), reused across sizes.
var gridA, gridB [(160 + 2) * (120 + 2)]uint8 // room for a one-cell border

var lut [256]uint16
var lut4 [16]uint16

var lcd board.LCD

func main() {
	board.SetUSBSerial()
	for end := time.Now().Add(3 * time.Second); time.Now().Before(end) && !machine.Serial.DTR(); {
		time.Sleep(10 * time.Millisecond)
	}
	bl := machine.LCD_BACKLIGHT
	bl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	bl.High()
	lcd.Configure()
	println("bench: cpu_hz", machine.CPUFrequency())
	firePalette()

	for {
		pushes()
		life(64, 64)
		life(160, 120)
		lifeFast(64, 64)
		lifeFast(160, 120)
		fire(160, 120, 2)
		fire(320, 240, 1)
		leds(64, 48, 5)
		leds(64, 64, 3)
		println("bench: done")
		time.Sleep(2 * time.Second)
	}
}

func report(name string, w, h int, us int64) {
	println("bench:", name, w, "x", h, "us", us)
}

// pushes: full-screen and partial pushes at three colour depths.
func pushes() {
	p := fb565()
	for i := range p {
		x, y := i%W, i/W
		p[i] = uint16(x*31/W)<<11 | uint16(y*63/H)<<5 | 16
	}
	t := time.Now()
	lcd.PushRect(0, 0, W, H, p)
	report("push565", W, H, time.Since(t).Microseconds())

	for i := range lut {
		lut[i] = uint16(i>>3)<<11 | uint16(i>>2)<<5 | uint16(31-i>>3)
	}
	idx := raw[:W*H]
	for i := range idx {
		idx[i] = uint8(i % W * 255 / W)
	}
	t = time.Now()
	lcd.PushIndexed(0, 0, W, H, idx, &lut)
	report("push8", W, H, time.Since(t).Microseconds())

	for i := range lut4 {
		lut4[i] = uint16(i*2)<<11 | uint16(i*4)<<5 | uint16(i*2)
	}
	nib := raw[:W*H/2]
	for i := range nib {
		nib[i] = uint8(i%16)<<4 | uint8((i+1)%16)
	}
	t = time.Now()
	lcd.PushIndexed4(0, 0, W, H, nib, &lut4)
	report("push4", W, H, time.Since(t).Microseconds())

	t = time.Now()
	lcd.PushRect(0, 0, 64, 64, p)
	report("push565-window", 64, 64, time.Since(t).Microseconds())
	time.Sleep(500 * time.Millisecond)
}

var rng uint32 = 2463534242

func rnd() uint32 {
	rng ^= rng << 13
	rng ^= rng >> 17
	rng ^= rng << 5
	return rng
}

// life runs Conway's Game of Life on a w×h torus (one byte per cell) and
// shows it ×k upscaled through the 8-bit LUT.
func life(w, h int) {
	a, b := gridA[:w*h], gridB[:w*h]
	for i := range a {
		a[i] = uint8(rnd() & 1)
	}
	const gens = 20
	var simUS int64
	for range gens {
		t := time.Now()
		for y := range h {
			ym, yp := (y+h-1)%h*w, (y+1)%h*w
			yc := y * w
			for x := range w {
				xm, xp := (x+w-1)%w, (x+1)%w
				n := a[ym+xm] + a[ym+x] + a[ym+xp] + a[yc+xm] + a[yc+xp] + a[yp+xm] + a[yp+x] + a[yp+xp]
				c := a[yc+x]
				if n == 3 || (n == 2 && c == 1) {
					b[yc+x] = 1
				} else {
					b[yc+x] = 0
				}
			}
		}
		simUS += time.Since(t).Microseconds()
		a, b = b, a
	}
	report("life-step", w, h, simUS/gens)
	// show: scale to the screen through the indexed push (k = largest fit)
	k := min(W/w, H/h)
	idx := raw[:w*k*h*k]
	t := time.Now()
	for y := range h * k {
		for x := range w * k {
			idx[y*w*k+x] = a[(y/k)*w+x/k] * 200
		}
	}
	report("life-upscale-x"+itoa(k), w*k, h*k, time.Since(t).Microseconds())
	lcd.Fill(0)
	t = time.Now()
	lcd.PushIndexed(0, 0, int16(w*k), int16(h*k), idx, &lut)
	report("life-push", w*k, h*k, time.Since(t).Microseconds())
	time.Sleep(500 * time.Millisecond)
}

// lifeFast is the same rule without modulo or division in the inner loop
// (the M0+ has no divide instruction): a one-cell border holds the wrapped
// edges, rows are walked with running indices, and the ×k upscale copies
// each cell into k pixels and each output row k times.
func lifeFast(w, h int) {
	pw := w + 2
	a, b := gridA[:pw*(h+2)], gridB[:pw*(h+2)]
	for i := range a {
		a[i] = uint8(rnd() & 1)
	}
	const gens = 20
	var simUS int64
	for range gens {
		t := time.Now()
		// wrap: copy edge rows/columns into the border
		copy(a[1:1+w], a[h*pw+1:h*pw+1+w])
		copy(a[(h+1)*pw+1:(h+1)*pw+1+w], a[pw+1:pw+1+w])
		for y := 0; y < h+2; y++ {
			r := y * pw
			a[r] = a[r+w]
			a[r+w+1] = a[r+1]
		}
		for y := 1; y <= h; y++ {
			up, mid, dn := (y-1)*pw, y*pw, (y+1)*pw
			for x := 1; x <= w; x++ {
				n := a[up+x-1] + a[up+x] + a[up+x+1] + a[mid+x-1] + a[mid+x+1] + a[dn+x-1] + a[dn+x] + a[dn+x+1]
				if n == 3 || (n == 2 && a[mid+x] == 1) {
					b[mid+x] = 1
				} else {
					b[mid+x] = 0
				}
			}
		}
		simUS += time.Since(t).Microseconds()
		a, b = b, a
	}
	report("lifefast-step", w, h, simUS/gens)
	k := min(W/w, H/h)
	ow := w * k
	idx := raw[:ow*h*k]
	t := time.Now()
	o := 0
	for y := 1; y <= h; y++ {
		row := idx[o : o+ow]
		i := 0
		for x := 1; x <= w; x++ {
			v := a[y*pw+x] * 200
			for range k {
				row[i] = v
				i++
			}
		}
		o += ow
		for range k - 1 {
			copy(idx[o:o+ow], row)
			o += ow
		}
	}
	report("upscalefast-x"+itoa(k), ow, h*k, time.Since(t).Microseconds())
	lcd.Fill(0)
	lcd.PushIndexed(0, 0, int16(ow), int16(h*k), idx, &lut)
	time.Sleep(500 * time.Millisecond)
}

// fire is the classic "Doom fire": heat rises with random cooling and drift.
// The grid is the 8-bit frame itself when k == 1.
func fire(w, h, k int) {
	g := raw[:w*h]
	if k > 1 {
		g = gridA[:w*h]
	}
	clear(g)
	for x := range w {
		g[(h-1)*w+x] = 36
	}
	const steps = 20
	var simUS int64
	for range steps {
		t := time.Now()
		for y := 1; y < h; y++ {
			for x := range w {
				src := y*w + x
				v := g[src]
				r := rnd() & 3
				dst := src - w - int(r) + 1
				if dst < 0 {
					dst = 0
				}
				if v >= uint8(r&1) {
					g[dst] = v - uint8(r&1)
				} else {
					g[dst] = 0
				}
			}
		}
		simUS += time.Since(t).Microseconds()
	}
	report("fire-step", w, h, simUS/steps)
	if k == 1 {
		t := time.Now()
		lcd.PushIndexed(0, 0, int16(w), int16(h), g, &lut)
		report("fire-push", w, h, time.Since(t).Microseconds())
	} else {
		idx := raw[:w*k*h*k]
		t := time.Now()
		for y := range h * k {
			for x := range w * k {
				idx[y*w*k+x] = g[(y/k)*w+x/k]
			}
		}
		report("fire-upscale-x"+itoa(k), w*k, h*k, time.Since(t).Microseconds())
		t = time.Now()
		lcd.PushIndexed(0, 0, int16(w*k), int16(h*k), idx, &lut)
		report("fire-push", w*k, h*k, time.Since(t).Microseconds())
	}
	time.Sleep(500 * time.Millisecond)
}

func firePalette() {
	for i := range 37 {
		r := min(255, i*7*2)
		g := max(0, min(255, (i-12)*11))
		b := max(0, min(255, (i-28)*30))
		lut[i] = uint16(r>>3)<<11 | uint16(g>>2)<<5 | uint16(b>>3)
	}
}

// leds renders a w×h grid as round LEDs with a dark gap into the RGB565
// frame (pitch p pixels per cell) and pushes it: the LED-matrix look.
func leds(w, h, p int) {
	a := gridA[:w*h]
	for i := range a {
		a[i] = uint8(rnd() % 37)
	}
	fb := fb565()[:w*p*h*p]
	off := lut[0]
	t := time.Now()
	for cy := range h {
		for cx := range w {
			c := lut[a[cy*w+cx]]
			for dy := range p {
				row := (cy*p+dy)*w*p + cx*p
				for dx := range p {
					// round dot: skip the corners and the last row/column (gap)
					edge := dx == p-1 || dy == p-1
					corner := p >= 4 && (dx == 0 || dx == p-2) && (dy == 0 || dy == p-2)
					if edge || corner {
						fb[row+dx] = off
					} else {
						fb[row+dx] = c
					}
				}
			}
		}
	}
	report("leds-render-pitch"+itoa(p), w*p, h*p, time.Since(t).Microseconds())
	lcd.Fill(0)
	t = time.Now()
	lcd.PushRect(int16((W-w*p)/2), int16((H-h*p)/2), int16(w*p), int16(h*p), fb)
	report("leds-push", w*p, h*p, time.Since(t).Microseconds())
	time.Sleep(time.Second)
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
