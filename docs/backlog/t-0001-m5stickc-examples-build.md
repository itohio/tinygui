---
aliases:
  - "M5StickC examples don't build"
tags:
  - tinygui
  - ticket
  - proposed
type: ticket
status: backlog
priority: p1
kind: bug
area: build
plan:
idea:
discovery: []
blocked-by: []
estimate: 1h
assignee: agent
created: 2026-11-10
closed:
updated: 2026-11-10
---

# T-0001: M5StickC examples don't build with TinyGo 0.43

Part of [[backlog]]. Proposed by TG-1a; the owner picks the display option below before work starts. Finding F-01 in [[findings]]. Ranked behind [[t-0002-tufty2040-board-package]]: the owner's device is a Tufty 2040, not an M5StickC. Part A still matters on its own: any app importing `ui` fails to build for classic ESP32 targets.

## Goal

`examples/choice` and `examples/gauge` build for `-target=m5stick-c -tags=m5stickc` with TinyGo 0.43.0 and drivers v0.31.0, and every widget they use can draw on the configured panel.

## Context and links

Three errors appear one after another ([[baseline#TinyGo builds]]):

1. `serial_reader.go:10:12: undefined: machine.Serialer`. Cause: TinyGo 0.43 declares `machine.Serialer` only in its USB files (`src/machine/usb.go:45`, `machine_esp32c3_usb.go:43`, `machine_esp32c6_usb.go:42`, `machine_esp32xx_usb.go:43`); classic ESP32 (`m5stick-c`, tags `esp32 … serial.uart`) has no USB, but `serial_reader.go` is built for every tag in its hand-written list, including `esp32`. The library half of the bug: any app importing `ui` fails on that target.
2. `examples/choice/main.go:84:9` (gauge `:77:9`): `st7735.Device` (a value) does not implement `ui.Displayer`: its methods have pointer receivers.
3. With `&display`: `*st7735.Device does not implement ui.Displayer (missing method DrawRGBBitmap)`. Upstream drivers v0.31.0 `st7735` has only `DrawRGBBitmap8` and `DrawBitmap(pixel.Image)`; TinyGUI's `BitmapDisplayer` wants `DrawRGBBitmap([]uint16)` (`drawing.go:11-14`).

## Proposed change

**Part A, library (smallest fix, no API break):** drop the build tags and the `machine` import from `serial_reader.go`, declare the package-local interface `serialer { Buffered() int; ReadByte() (byte, error) }` there, delete `serial_reader_stub.go`. `machine.Serialer`, `*machine.UART` and USB CDC all satisfy it structurally, so callers keep compiling, and the reader now works on every target instead of being a no-op off the list.

```diff
-//go:build stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid
-
 package ui
 
-import "machine"
+type serialer interface {
+	Buffered() int
+	ReadByte() (byte, error)
+}
 
 type SerialReader struct {
-	s machine.Serialer
+	s serialer
 }
 
-func NewSerialReader(s machine.Serialer) *SerialReader {
+func NewSerialReader(s serialer) *SerialReader {
```

**Part B, board package (owner picks one):**

| Option | Change | Result | Trade-off |
|---|---|---|---|
| B1 (recommended) | `examples/m5stick/cfg_m5stickc.go`: import the vendored `github.com/itohio/tinygui/examples/m5stick/st7735` (already in the repo, import commented out at `:9`), return `*st7735.Device` (`return &display`) | builds; full `ui.Displayer`, so `Bitmap16` icons draw | keeps a vendored driver copy; it allocates 5 slices per `setWindow` ([[findings]] F-21), same as upstream |
| B2 | keep upstream `drivers/st7735`; demos' `configureDisplay` returns `*st7735.Device` | builds | `Bitmap16`/`Icon` silently draw nothing (type assertion fails, `widget/bitmap.go:47-50`); the choice demo loses its icon picker |
| B3 | change `ui.BitmapDisplayer` to the drivers' `DrawBitmap(x, y, pixel.Image[…])` API | builds against upstream, no vendored driver | API change for every display adapter; belongs in TG-1b's display-contract work, not a build fix |

Measured on scratch copies: A+B1 → choice 139 364 B flash / 22 252 B RAM, gauge 131 568 / 12 140; A+B2 → 139 132 / 22 252 and 131 888 / 12 140 ([[baseline#Size]]). Host tests still pass with A.

Out of scope here (separate findings): panel geometry (F-02), layout overflow in the gauge demo (F-20), missing glyphs (F-10), the 16×16 thermometer icon (F-12).

## Acceptance criteria

- [ ] `GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -tags=m5stickc -o /dev/null ./examples/choice` exits 0.
- [ ] Same for `./examples/gauge`.
- [ ] `GOTOOLCHAIN=local GOPROXY=off go test ./ ./container ./layout ./widget` passes.
- [ ] Red first: a minimal program calling `ui.NewSerialReader(machine.Serial)` fails `tinygo build -target=m5stick-c` before Part A (`undefined: machine.Serialer`, reproduced by verify) and builds after it.
- [ ] Regression (passes before and after): the same program builds for `-target=pico`, so the reader still accepts the machine type on a USB target.
- [ ] [[examples]] and [[baseline]] show the new build status.

## Verification

The commands above, output pasted into the log. On-device check (owner present): the choice demo shows the label choice and the icon picker.

## Log

- 2026-11-10: created by TG-1a from measured build failures; scratch-copy builds succeed with A+B1 and A+B2.
- 2026-11-12: priority p0 → p1 after the owner's device turned out to be a Tufty 2040 (UX-28 D2); T-0002 goes first.
