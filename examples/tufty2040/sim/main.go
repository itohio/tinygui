//go:build !tinygo

// Command sim runs the Tufty 2040 showcase on the host and writes PNG frames.
//
//	go run ./examples/tufty2040/sim -out /tmp/tufty            # scripted tour
//	go run ./examples/tufty2040/sim -out dir -script "c,c,a,up,shot:x"
//
// Script steps: up, down, a, b, c (suffix "+" = long press), user:<ms>,
// wait:<frames>, shot:<name>. Each step advances 3 frames (33 ms each).
package main

import (
	"flag"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/examples/tufty2040/app"
)

const tour = "wait:300,shot:01-dashboard," +
	"c,wait:20,shot:02-gauges,a,up,up,up+,shot:03-gauges-edit-level,a," +
	"down,a,up,up,a,up,shot:04-gauges-edit-mix,b," +
	"c,wait:10,a,shot:05-controls-led-on,down,a,up,up,shot:06-controls-edit-led,a," +
	"down,down,down,a,up,shot:07-controls-accent,a," +
	"c,wait:10,a,up,shot:08-images-edit,a,down,down,a,up,shot:09-images-bitmap,a," +
	"c,wait:10,shot:10-layouts," +
	"c,wait:10,a,down,down,down,down,down,down,down,down,down,down,down,down,shot:11-list-scrolled," +
	"c,wait:10,shot:12-log," +
	"user:1500,wait:5,shot:13-saver,a," +
	"c,wait:60,shot:14-life,down,a,up,shot:15-life-gun,a,wait:90,shot:16-life-gun-later," +
	"c,wait:40,shot:17-water,a,up,up,a,wait:60,shot:18-water-left,down,a,wait:60,shot:19-water-pour," +
	"down,down,a,wait:20,shot:20-water-full5,up,up,wait:60,shot:21-water-full5-up,b,down,a,up,a,up,a,wait:30,shot:22-water-full4,b,shot:23-water-back"

type fakeHW struct{ t *time.Time }

func (h fakeHW) Light() uint16 {
	s := float64(h.t.UnixMilli()) / 1000
	return uint16(32000 + 20000*math.Sin(s*1.3) + 6000*math.Sin(s*7.1))
}
func (h fakeHW) Battery() (float32, bool) { return 3.92, true }
func (h fakeHW) SetLED(p int)             { fmt.Println("hw: led", p) }
func (h fakeHW) SetBacklight(p int)       { fmt.Println("hw: backlight", p) }

// durationAdd mirrors ui.DurationAdd, which is only built for TinyGo targets.
func durationAdd(d, delta time.Duration) time.Duration { return d + delta*magnitude(d) }
func durationSub(d, delta time.Duration) time.Duration { return d - delta*magnitude(d) }
func magnitude(d time.Duration) time.Duration {
	m := time.Duration(1)
	for d >= 100 {
		d /= 10
		m *= 10
	}
	return m
}

func main() {
	out := flag.String("out", "sim-out", "output directory for PNG frames")
	script := flag.String("script", tour, "comma-separated steps")
	flag.Parse()
	if err := os.MkdirAll(*out, 0o755); err != nil {
		panic(err)
	}
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	fb := framebuffer.New(app.ScreenW, app.ScreenH, make([]uint16, app.ScreenW*app.ScreenH), nil)
	a := app.New(fb, fakeHW{&now}, app.Config{
		DurationAdd: durationAdd, DurationSub: durationSub,
		Log: func(s string) { fmt.Println("log:", s) },
	}, now)
	frame := func(n int) {
		for range n {
			now = now.Add(33 * time.Millisecond)
			a.Update(now)
			a.Draw()
		}
	}
	buttons := map[string]app.Button{"up": app.BtnUp, "down": app.BtnDown, "a": app.BtnA, "b": app.BtnB, "c": app.BtnC}
	for _, step := range strings.Split(*script, ",") {
		name, arg, _ := strings.Cut(step, ":")
		long := strings.HasSuffix(name, "+")
		name = strings.TrimSuffix(name, "+")
		switch name {
		case "wait":
			n, _ := strconv.Atoi(arg)
			frame(n)
			continue
		case "shot":
			a.Draw()
			path := filepath.Join(*out, arg+".png")
			f, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			if err := png.Encode(f, fb.Image()); err != nil {
				panic(err)
			}
			f.Close()
			fmt.Println("wrote", path)
			continue
		case "user":
			ms, _ := strconv.Atoi(arg)
			a.UserHold(time.Duration(ms) * time.Millisecond)
		default:
			b, ok := buttons[name]
			if !ok {
				panic("unknown step " + step)
			}
			a.Press(b, long)
		}
		frame(3)
	}
}
