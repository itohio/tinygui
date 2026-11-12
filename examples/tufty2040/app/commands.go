package app

import (
	"runtime"
	"time"
)

// Serial console. The board feeds ui.CommandStreamMux with these handlers;
// handlers only enqueue, and Update executes them on the UI goroutine, so the
// widget tree is touched by one goroutine only.

type command struct {
	name   string
	arg    int
	hasArg bool
	long   bool
}

const helpText = "commands: up|down|a|b|c [long], user <ms>, page <n>, led <0-100>, bl <10-100>, saver [0|1], solid [0-255], stats, tiles, dump, help"

// Handlers returns the command table for ui.NewCommandStreamMux.
func (a *App) Handlers() map[string]func([]byte) {
	m := map[string]func([]byte){}
	for _, name := range []string{"up", "down", "a", "b", "c", "user", "page", "led", "bl", "saver", "solid", "stats", "tiles", "dump", "help"} {
		m[name] = func(arg []byte) { a.Command(name, string(arg)) }
	}
	return m
}

// Command queues one console command (safe to call from another goroutine).
func (a *App) Command(name, arg string) {
	c := command{name: name, long: arg == "long"}
	c.arg, c.hasArg = atoi(arg)
	select {
	case a.cmds <- c:
	default: // queue full: drop
	}
}

func (a *App) exec(c command) {
	switch c.name {
	case "up":
		a.Press(BtnUp, c.long)
	case "down":
		a.Press(BtnDown, c.long)
	case "a":
		a.Press(BtnA, c.long)
	case "b":
		a.Press(BtnB, c.long)
	case "c":
		a.Press(BtnC, c.long)
	case "user":
		a.UserHold(time.Duration(c.arg) * time.Millisecond)
	case "page":
		if c.arg >= 0 && c.arg < len(a.pages) {
			a.showPage(c.arg)
		}
	case "led":
		a.ledOn, a.ledPct = c.arg > 0, clampi(c.arg, 0, 100)
		a.applyLED()
	case "bl":
		a.backlight = clampi(c.arg, 10, 100)
		a.hw.SetBacklight(a.backlight)
	case "saver": // no argument toggles; 1 on, 0 off
		if c.hasArg {
			a.saver = c.arg != 0
		} else {
			a.saver = !a.saver
		}
	case "solid": // a uniform grey screen for camera tests; no argument: back to the UI
		a.solid = -1
		if c.hasArg {
			a.solid = clampi(c.arg, 0, 255)
		}
	case "dump":
		a.dumpReq = true
	case "stats":
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		a.log("fps " + itoa(a.fps) + " draw_us " + itoa(int(a.drawUS)) + " push_us " + itoa(int(a.pushUS)) +
			" page " + itoa(a.cur) + " edit " + itoa(btoi(a.editing)))
		a.log("heap_alloc " + itoa(int(ms.HeapAlloc)) + " heap_sys " + itoa(int(ms.HeapSys)) +
			" mallocs " + itoa(int(ms.Mallocs)) + " gc " + itoa(int(ms.NumGC)))
	case "tiles": // frame budget of each demo tile on the Demos page
		a.menu.report(a.log)
	case "help":
		a.log(helpText)
	}
}

func atoi(s string) (int, bool) {
	n, neg, ok := 0, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '-' && i == 0:
			neg = true
		case c >= '0' && c <= '9':
			n = n*10 + int(c-'0')
			ok = true
		default:
			return 0, false
		}
	}
	if neg {
		n = -n
	}
	return n, ok
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
