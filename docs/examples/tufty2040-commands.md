---
aliases:
  - "Tufty 2040 command set"
  - "Badge console"
tags:
  - tinygui
  - examples
  - tufty2040
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# Tufty 2040 showcase: command set

One command channel, two sources (owner, 2026-11-12: "inject commands to the command channel from either source - uart or actual button"). Physical buttons and USB-serial lines end in the same calls, `App.Press(button, long)` and `App.UserHold(duration)` (`tinytufty2040/app/app.go`), so a serial `a` and a press of A are the same event; both log `press a` to the Log page and the console. Part of [[example-tufty2040]].

## Button convention

Same meaning on every page (owner 2026-11-12: "up and down … A, B, and C … so you decide what these do"); the footer always shows the hints for the current mode.

| Button | Browsing | Editing a value | Footer hint |
|---|---|---|---|
| **UP / DOWN** | move the focus to the previous / next row (held: repeats) | increase / decrease the value (held: large steps) | `up/down focus` · `up/down change` |
| **A** | select: start editing the focused row, flip a toggle, enter a list | confirm (commit); on a multi-gauge: next segment | `A select` · `A ok` |
| **B** | back: leave a list | cancel: revert the edit | `B back` · `B cancel` |
| **C** | next page (held ≥ 500 ms: previous page) | cancel the edit, then next page | `C page` |
| **USER/BOOT** | hold > 1 s: screensaver on/off | — | — |

## How to send

- **USB serial** (USB CDC, any baud rate): one command per line, `\n`-terminated. The badge reports USB serial number `df6750955f2a3536` (its flash unique ID, the same number MicroPython reported); find it by that, not by port name.
- **Host tool** `badgectl` (not in this repo): `badgectl a up up a` sends commands and prints the replies; `--grab` adds a badge-cam photo after each; `--tour` reboots the badge and replays the simulator tour, saving framebuffer dumps, badge-cam photos and simulator PNGs side by side; `--monitor` prints the console (for testing the buttons by hand).
- Serial handlers only queue the command (16 deep, extra lines dropped); `App.Update` executes it at the start of the next frame on the UI goroutine ([[concurrency]]).

## Commands

| Command | Does | Same as button | Notes |
|---|---|---|---|
| `up` / `up long` | focus previous row; while editing: increase (`long` = large step) | UP (press; held ≥ 500 ms auto-repeats as `long` every 120 ms) | |
| `down` / `down long` | focus next row; while editing: decrease | DOWN (as UP) | |
| `a` / `a long` | select: start editing the focused row, flip a toggle, enter the list; while editing: commit (multi-gauge: next segment) | A (fires on release; held ≥ 500 ms = `long`) | |
| `b` / `b long` | cancel an edit (revert), or leave the list | B (on release) | |
| `c` / `c long` | next page (`long`: previous page); cancels an edit first | C (on release) | |
| `user <ms>` | report a USER hold of `<ms>`; > 1000 ms toggles the screensaver | USER/BOOT held (measured by `ui.PeekButton`, the loop pauses while it is down) | |
| `page <n>` | jump to page `n` (0 Demos, 1 Dash, 2 Gauge, 3 Ctrl, 4 Image, 5 Flow, 6 List, 7 Log, 8 Life, 9 Water, 10 Fire, 11 Pong, 12 Boids) | — | |
| `led <0-100>` | user LED brightness (0 = off) | — (Ctrl page: User LED, LED level) | |
| `bl <10-100>` | backlight brightness; **100 is DC** (pin driven high, no PWM: no stripes in camera photos), below 100 a 200 kHz PWM; the default is 100 on USB power, 80 on battery | — (Ctrl page: Backlight) | |
| `saver` / `saver 1` / `saver 0` | toggle / switch on / switch off the screensaver (`RandomContext` burn-in shift) | USER hold > 1 s toggles | any button press leaves it |
| `stats` | log `fps`, `draw_us`, `push_us`, page, edit state; `heap_alloc`, `heap_sys`, `mallocs`, `gc` | — | |
| `solid [0-255]` | the whole screen one grey level (255 white) for camera tests; no argument returns to the UI | — | |
| `tiles` | frame budget of each Demos tile: `tile <name> tick_us <µs> draw_us <µs>` (draw summed over the three bands) | — | |
| `dump` | send the framebuffer: line `FB 320 240`, 153 600 bytes big-endian RGB565 (all three bands), line `END` | — | what was drawn, not what the glass shows |
| `help` | list the commands | — | |
| `madctl <n>` | write the panel's MADCTL register (orientation probe; default 176 = `0xB0`) | — | firmware only (`main_tufty2040.go`) |
| `pins` | print raw button levels and VBUS | — | firmware only |
| `reboot` | restart the firmware (`arm.SystemReset`); scripted tours start from it | — | firmware only; the badge re-enumerates with the same serial number |

Every press also counts in the Dash page's bar chart, so serial and physical presses are both visible on screen.
