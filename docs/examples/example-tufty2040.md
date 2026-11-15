---
aliases:
  - "Tufty 2040 showcase"
  - "tinytufty2040"
tags:
  - tinygui
  - examples
  - tufty2040
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# tinytufty2040 — showcase on the Pimoroni Tufty 2040

A demo of TinyGUI on the owner's Tufty 2040 badge (owner request 2026-11-10). It began as `examples/tufty2040` here and moved to its own repo, `github.com/itohio/tinytufty2040` (owner 2026-11-12), which imports tinygui (and `examples/framebuffer`, `examples/choice/icons`); paths below are in that repo, and its `docs/` is the current reference. Part of [[examples]]. Scope rules: [[0001-library-scope-and-layers]].

## Layout of the code

| Path | Build | Role |
|---|---|---|
| `tinytufty2040/app/` | any | the whole UI, TinyGUI only: pages, theme, example widgets, input mapping, serial commands; no `machine` import |
| `tinytufty2040/main_tufty2040.go` | `tufty2040` | firmware: frame loop, buttons, USER via `ui.PeekButton`, USB-serial console via `ui.CommandStreamMux` + `ui.SerialReader`, framebuffer dump |
| `tinytufty2040/board/` | `tufty2040` | `board.LCD`: ST7789 320×240 on the 8-bit parallel bus, bit-banged via RP2040 SIO; init and orientation from Pimoroni's driver; write timing sized for the 200 MHz CPU clock (83 ns per byte, ST7789 needs ≥ 66 ns) |
| `tinytufty2040/bringup/` | `tufty2040` | bring-up firmware without TinyGUI: LED heartbeat, backlight, panel init, full-screen colour cycle, serial log with heap state |
| `tinytufty2040/badge_tufty2040.go` | `tufty2040` | light sensor, battery (1.24 V reference), VBUS, user LED and backlight PWM, button debouncing/auto-repeat |
| `tinytufty2040/sim/` | host (`!tinygo`) | simulator: same `app`, fake hardware, scripted button presses, PNG screenshots |
| `examples/framebuffer/` | any | RGB565 RAM framebuffer implementing the full `ui.Displayer`; boards push it, the simulator encodes it |

The firmware draws each frame in three 80-row bands into a 51 200 B RAM buffer (`framebuffer.NewBand`) and pushes each band in one burst, so the panel never shows the clear-then-draw sequence ([[findings]] F-06 avoided in the example, not fixed in the library). A full 153 600 B frame buffer ran the heap out of memory on the badge; two 120-row bands still did once the heap fragmented ([[findings]] F-30, F-31). Commands from USB serial and from the buttons share one channel: [[tufty2040-commands]].

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
| Flash | 8 MB on the board, but TinyGo's `tufty2040` target links with `--defsym=__flash_size=1020K` (`targets/tufty2040.json`), so firmware plus any assets kept in flash must fit in 1 020 KB unless the linker setting is changed | TinyGo 0.43 `targets/tufty2040.json`; Pimoroni (8 MB) |

Not verified on the device yet: orientation, colour order and inversion, button polarity, ADC scaling. The firmware has a `madctl <n>` serial command to probe orientation and `pins` to print button levels.

## Pages and the features they exercise

| Page | TinyGUI features |
|---|---|
| Demos | the home page: every demo runs live in a tile (`CellGrid` 1 px, the focused tile 2 px and larger), placed by the `layout.Wrap` arranger (rows re-wrap when the focused tile grows); A opens the demo full screen, B there returns; a speed `Dropdown` (½, ×1, ×2) among the tiles; B on the page opens a choice `Dialog` (Pause/Resume all, Reset all… → modal confirm, About → popup), all drawn over the running tiles by `container.Overlay`; `tiles` console command for each tile's budget |
| Dash | `Label` with zero-allocation text providers, `VolumeGauge`, `SolidGauge`, custom sparkline (ring buffer) and bar chart widgets, `HList`/`VList` nesting, `Selectable` to skip static rows |
| Gauge | `Gauge` vertical and horizontal, `MultiGauge`, `VolumeGauge`, four `animation` curves (linear, ease-in, ease-out, ease-in-out) ping-ponging, `InteractiveGauge`, `InteractiveMultiGauge`, `InteractiveLabel[int]` with commit callback restarting the animations |
| Ctrl | `Toggle` (drives the user LED), `InteractiveLabel[int]`/`[float32]` with formatters (LED level, backlight, gain), `InteractiveLabelChoice` (accent colour recolours the UI), example `durationLabel` using `ui.DurationAdd/Sub`, `MultilineLabel` |
| Image | `InteractiveIcon` and `InteractiveIconChoice` (PNG strings decoded by `ui.DrawPng`), `InteractiveWidgetChoice` over `Bitmap16` icons, animated `Bitmap16` plasma, `Bitmap8` (RGB332), `Separator` |
| Flow | `Grid`, `HFlow`, `VFlow`, `HList` side by side |
| List | `ScrollChoice` with 20 rows, `ScrollObserver` scrollbar, navigator descent (A enters, B leaves) |
| Log | `InteractiveLog` fed by `NavigatorObserver`, button and serial events; navigator depth and `Walk` node count |
| Life | Game of Life on a 64×48 `CellGrid` (round LEDs, ages fade; RGB565 scanline fast path), 3 px in the page; Run/Pause (`Toggle`), pattern (`InteractiveLabelChoice`: Random, glider gun, Acorn, R-pentomino); **Full** at 5 px, where UP/DOWN change the pattern, A runs/pauses, B returns; `ui.Clipper` band skipping |
| Water | falling-liquid automaton on a 64×64 grid, `CellGrid` at 3 px (192×192), foam/depth shading; gravity in 8 directions (`InteractiveLabelChoice`), Pour and Fill (`Toggle`); 3 steps per frame; **Full** shows the grid alone on the whole screen (LEDs 5 px: 64×48 at 320×240, or 4 px: 64×60), rows following gravity; there UP/DOWN rotate gravity, A pours, B returns, C changes page |
| Fire | a burning liquid (owner 2026-11-12): fuel on the shared liquid engine (`liquid.go`, as water) pools in a corner or spreads and follows **Gravity** (8 directions); its exposed surface burns and is used up (about a minute per pool, a new splash after it runs out); flames rise against gravity (Doom-fire heat spread turned with gravity) and light fuel they touch; **Fuel** pours a stream, **Splash** drops burning fuel, Flames low/mid/tall; `CellGrid` 3 px in the page, **Full** 5 px: UP/DOWN turn gravity, A splash, B back |
| Pong | 64×48 LED court on `CellGrid`: fixed-point physics, 3×5 LED score digits, ball trail; CPU vs CPU in the page (level choice), **You** (PLAY) goes full screen at 5 px with your paddle on UP/DOWN, A pause, B back |
| Boids | Reynolds flocking of 32/64/128 boids in a **world frame**; the screen is a 64×48 box moving through it (owner 2026-11-12): turning (UP/DOWN in full screen, Spin on the page) or shaking (A) the box sweeps its walls through the flock, boids keep their world heading and only react when a wall reaches them; fixed point, heads coloured by on-screen heading, fading trails; an IMU's orientation will drive the box once one is attached |
| (any) | tab bar with an eased underline (`animation`), scrolling to keep the active tab in view (6 visible), `InteractiveWidgetChoice` as the page pager, `container` idle timeout, `RandomContext` screensaver (hold USER > 1 s) |

Controls: UP/DOWN move focus (or change the value while editing; held = auto-repeat with the large step), A selects / commits, B cancels / goes back, C next page (held: previous page), USER hold > 1 s screensaver.

Serial console and buttons share one command channel; the full command set, the matching buttons and the host tool are in [[tufty2040-commands]]. The badge reports USB serial number `df6750955f2a3536`.

## Build and run

```sh
cd .
GOTOOLCHAIN=local GOPROXY=off go run ./tinytufty2040/sim -out /tmp/tufty           # 13 PNGs of a scripted tour
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=tufty2040 -size short -o tufty.uf2 ./tinytufty2040
GOTOOLCHAIN=local GOPROXY=off tinygo flash -target=tufty2040 ./tinytufty2040        # owner only
```

Flashing replaces the badge's MicroPython firmware: hold BOOT/USER, press RESET, copy `tufty.uf2` to the `RPI-RP2` drive (or `tinygo flash` with the board in bootloader mode). Restoring MicroPython needs Pimoroni's Tufty UF2.

## Status (2026-11-12)

| Check | Result |
|---|---|
| `go vet` app, sim, framebuffer | clean |
| simulator tour | 13 screenshots, every page and edit mode inspected visually |
| `tinygo build -target=tufty2040` | builds: 225 008 B flash / 95 160 B static RAM (band buffer 51 200, PNG inflate window 34 856); fonts and icon tables now land in flash (`.rodata`); heap 142 KB |
| app allocations (`-print-allocs`) | construction only; per-frame allocations left are the library's `context.go:79` and `container/scroll.go:167` ([[findings]] F-07) |
| on the device (flashed 2026-11-12, owner's approval) | runs; 15–24 fps; draw 24–42 ms (three band passes) and push 19.4 ms per frame; `heap_alloc` 66–116 KB of 142 KB; the full 13-step hardware tour (`badgectl --tour`) completes without errors; USB serial number `df6750955f2a3536`; on the glass: colour cycle (bring-up) and the UI seen through the badge cam |
| not yet confirmed | physical picture orientation and colours (the serial dump shows the buffer, not the glass); light sensor read 0–1 % (raw 368) and battery 0.05 V on USB power without a battery — plausible, needs a lit room and a battery to confirm |

## Size

`tinygo build -target=tufty2040 -size full`: TinyGUI packages 25 582 B code (`widget` 15 632, `ui` 4 574, `container` 3 676, `layout` 870, `animation` 830); app 21 887; PNG decoder 13 596 flash + 34 880 RAM; `fmt` + `strconv` ≈ 34.5 KB flash, linked by the library's default `fmt.Sprintf` formatter in `InteractiveLabel` even though the demo supplies its own formatters ([[findings]] F-26).

## Library defects found while building it

Recorded in [[findings]] as F-22…F-27: the `Scroll` layout bug (fixed with a regression test), `ScrollChoice` change callbacks that miss navigator moves, the idle timer that ignores navigator moves, no active-segment indicator in `InteractiveMultiGauge`, the forced `fmt` dependency, and the 34 KB inflate window the PNG decoder keeps in RAM. Missing clipping (F-13) forced the app to repaint the tab bar and footer after the body (`app/app.go`, `Draw`).
