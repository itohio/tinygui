---
aliases:
  - "Tufty 2040 showcase"
  - "examples/tufty2040"
tags:
  - tinygui
  - examples
  - tufty2040
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# examples/tufty2040 — showcase on the Pimoroni Tufty 2040

A seven-page demo of everything TinyGUI has at `b140156`, written for the owner's Tufty 2040 badge (owner request 2026-11-10). Part of [[examples]]. Scope rules: [[0001-library-scope-and-layers]].

## Layout of the code

| Path | Build | Role |
|---|---|---|
| `examples/tufty2040/app/` | any | the whole UI, TinyGUI only: pages, theme, example widgets, input mapping, serial commands; no `machine` import |
| `examples/tufty2040/main_tufty2040.go` | `tufty2040` | firmware: frame loop, buttons, USER via `ui.PeekButton`, USB-serial console via `ui.CommandStreamMux` + `ui.SerialReader`, framebuffer dump |
| `examples/tufty2040/board/` | `tufty2040` | `board.LCD`: ST7789 320×240 on the 8-bit parallel bus, bit-banged via RP2040 SIO; init and orientation from Pimoroni's driver; write timing sized for the 200 MHz CPU clock (83 ns per byte, ST7789 needs ≥ 66 ns) |
| `examples/tufty2040/bringup/` | `tufty2040` | bring-up firmware without TinyGUI: LED heartbeat, backlight, panel init, full-screen colour cycle, serial log with heap state |
| `examples/tufty2040/badge_tufty2040.go` | `tufty2040` | light sensor, battery (1.24 V reference), VBUS, user LED and backlight PWM, button debouncing/auto-repeat |
| `examples/tufty2040/sim/` | host (`!tinygo`) | simulator: same `app`, fake hardware, scripted button presses, PNG screenshots |
| `examples/framebuffer/` | any | RGB565 RAM framebuffer implementing the full `ui.Displayer`; boards push it, the simulator encodes it |

The firmware draws each frame in two 120-row bands into a 76 800 B RAM buffer (`framebuffer.NewBand`) and pushes each band in one burst, so the panel never shows the clear-then-draw sequence ([[findings]] F-06 avoided in the example, not fixed in the library). A full 153 600 B frame buffer ran the heap out of memory on the badge ([[findings]] F-30).

## Board facts used

From TinyGo's `src/machine/board_tufty2040.go` (0.43.0), Pimoroni's `pimoroni-pico` sources and the TinyGo board page (all read 2026-11-10):

| Item | Pins / value | Source |
|---|---|---|
| Panel | ST7789, 320×240, 8-bit parallel: CS 10, DC 11, WR 12, RD 13, DB0–7 = GPIO14–21, backlight GPIO2 (PWM1 A) | board file; `picographics.cpp` `ParallelPins {10, 11, 12, 13, 14, 2}` |
| Orientation | mounted upside down → ROTATE_180: MADCTL `0xB0`, CASET 0–319, RASET 0–239 | `picographics.cpp` ("Tufty display is upside-down"), `st7789.cpp` `configure_display` |
| Buttons | A 7, B 8, C 9, UP 22, DOWN 6 active-high (pull-down); USER/BOOT 23 active-low | board file; `button_test.py` (`invert=False` / `invert=True`) |
| User LED | GPIO25 (PWM4 B) | board file, `led_pwm.py` |
| Light sensor | ADC0 / GPIO26, powered by SENSOR_POWER GPIO27 | board file, `autobright.py` |
| Battery | ADC3 / GPIO29 ×3 divider, 1.24 V reference on ADC2 / GPIO28, VBUS detect GPIO24 | `battery.py` |

Not verified on the device yet: orientation, colour order and inversion, button polarity, ADC scaling. The firmware has a `madctl <n>` serial command to probe orientation and `pins` to print button levels.

## Pages and the features they exercise

| Page | TinyGUI features |
|---|---|
| Dash | `Label` with zero-allocation text providers, `VolumeGauge`, `SolidGauge`, custom sparkline (ring buffer) and bar chart widgets, `HList`/`VList` nesting, `Selectable` to skip static rows |
| Gauges | `Gauge` vertical and horizontal, `MultiGauge`, `VolumeGauge`, four `animation` curves (linear, ease-in, ease-out, ease-in-out) ping-ponging, `InteractiveGauge`, `InteractiveMultiGauge`, `InteractiveLabel[int]` with commit callback restarting the animations |
| Control | `Toggle` (drives the user LED), `InteractiveLabel[int]`/`[float32]` with formatters (LED level, backlight, gain), `InteractiveLabelChoice` (accent colour recolours the UI), example `durationLabel` using `ui.DurationAdd/Sub`, `MultilineLabel` |
| Images | `InteractiveIcon` and `InteractiveIconChoice` (PNG strings decoded by `ui.DrawPng`), `InteractiveWidgetChoice` over `Bitmap16` icons, animated `Bitmap16` plasma, `Bitmap8` (RGB332), `Separator` |
| Layout | `Grid`, `HFlow`, `VFlow`, `HList` side by side |
| List | `ScrollChoice` with 20 rows, `ScrollObserver` scrollbar, navigator descent (A enters, B leaves) |
| Log | `InteractiveLog` fed by `NavigatorObserver`, button and serial events; navigator depth and `Walk` node count |
| (any) | tab bar with an eased underline (`animation`), `InteractiveWidgetChoice` as the page pager, `container` idle timeout, `RandomContext` screensaver (hold USER > 1 s) |

Controls: UP/DOWN move focus (or change the value while editing; held = auto-repeat with the large step), A selects / commits, B cancels / goes back, C next page (held: previous page), USER hold > 1 s screensaver.

Serial console (USB CDC; any baud rate): `up|down|a|b|c [long]`, `user <ms>`, `page <n>`, `led <0-100>`, `bl <10-100>`, `saver`, `stats` (fps, draw/push time, heap, GC count), `dump` (framebuffer: line `FB 320 240`, 153 600 bytes of big-endian RGB565, line `END`), `help`, plus firmware-only `madctl <n>`, `pins`. Handlers only enqueue; `App.Update` executes them on the UI goroutine.

## Build and run

```sh
cd .
GOTOOLCHAIN=local GOPROXY=off go run ./examples/tufty2040/sim -out /tmp/tufty           # 13 PNGs of a scripted tour
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=tufty2040 -size short -o tufty.uf2 ./examples/tufty2040
GOTOOLCHAIN=local GOPROXY=off tinygo flash -target=tufty2040 ./examples/tufty2040        # owner only
```

Flashing replaces the badge's MicroPython firmware: hold BOOT/USER, press RESET, copy `tufty.uf2` to the `RPI-RP2` drive (or `tinygo flash` with the board in bootloader mode). Restoring MicroPython needs Pimoroni's Tufty UF2.

## Status (2026-11-12)

| Check | Result |
|---|---|
| `go vet` app, sim, framebuffer | clean |
| simulator tour | 13 screenshots, every page and edit mode inspected visually |
| `tinygo build -target=tufty2040` | builds: 224 408 B flash / 142 632 B static RAM (band buffer 76 800, PNG inflate window 34 856, choice icons 10 112, fonts 8 558); heap 117 KB |
| app allocations (`-print-allocs`) | construction only; per-frame allocations left are the library's `context.go:79` and `container/scroll.go:167` ([[findings]] F-07) |
| on the device (flashed 2026-11-12, owner's approval relayed by Leader) | runs; 15–21 fps; draw 32–55 ms and push 12.8 ms per frame (both bands); `heap_alloc` 83–107 KB of 117 KB; ≈ 12 GC runs/s; stable over 28+ page changes and edits; buttons idle-low, USER idle-high, VBUS high (`pins`); every page dumped over serial and inspected (`docs/evidence/tufty-device/`) |
| not yet confirmed | physical picture orientation and colours (the serial dump shows the buffer, not the glass); light sensor read 0–1 % (raw 368) and battery 0.05 V on USB power without a battery — plausible, needs a lit room and a battery to confirm |

## Size

`tinygo build -target=tufty2040 -size full`: TinyGUI packages 25 582 B code (`widget` 15 632, `ui` 4 574, `container` 3 676, `layout` 870, `animation` 830); app 21 887; PNG decoder 13 596 flash + 34 880 RAM; `fmt` + `strconv` ≈ 34.5 KB flash, linked by the library's default `fmt.Sprintf` formatter in `InteractiveLabel` even though the demo supplies its own formatters ([[findings]] F-26).

## Library defects found while building it

Recorded in [[findings]] as F-22…F-27: the `Scroll` layout bug (fixed with a regression test), `ScrollChoice` change callbacks that miss navigator moves, the idle timer that ignores navigator moves, no active-segment indicator in `InteractiveMultiGauge`, the forced `fmt` dependency, and the 34 KB inflate window the PNG decoder keeps in RAM. Missing clipping (F-13) forced the app to repaint the tab bar and footer after the body (`app/app.go`, `Draw`).
