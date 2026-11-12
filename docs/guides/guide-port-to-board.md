---
aliases:
  - "Port to a board"
  - "Board support"
tags:
  - tinygui
  - guides
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# Port to a board

What a board must provide, where the code goes, and how to see the screen before flashing. Part of [[guides]]. Worked example: [[example-tufty2040]]. Rules: [[0001-library-scope-and-layers]] (board code lives in `examples/`, selected by build tags).

## 1. A display

TinyGUI draws into a `drivers.Displayer` (`Size`, `SetPixel`, `Display`) and uses faster paths when the value also has them ([[topology#Display contract]]):

| Interface | Methods | Without it |
|---|---|---|
| `ui.RectangleDisplayer` | `FillRectangle`, `FillRectangleWithBuffer`, `FillScreen` | gauges and toggles fall back to `SetPixel` loops |
| `ui.BitmapDisplayer` | `DrawRGBBitmap([]uint16)`, `DrawRGBBitmap8` | `Bitmap16`/`Bitmap8`/`Icon` draw **nothing** |
| `ui.LineDisplayer` | `DrawFastHLine`, `DrawFastVLine` | unused by the library today |

Two ways to wire a panel:

- **RAM framebuffer (recommended when RAM allows).** `examples/framebuffer.New(w, h, pix, flush)` implements all of the above with clipping; your `flush` pushes `f.Pix` to the panel once per frame. No flicker, every draw is a memory write. A full 320×240 frame is 153 600 B; on the Tufty that left too little heap and the app ran out of memory ([[findings]] F-30). `framebuffer.NewBand(w, screenH, bandH, pix, flush)` holds only `bandH` rows: draw the UI once per band (`SetY0`, `Draw`, `Display`); the Tufty uses three 80-row bands = 51 200 B (`tinytufty2040/main_tufty2040.go`), each extra band costing one more draw pass. Draw code must then paint the same frame on every pass: keep per-frame state changes out of `Draw`.
- **Direct driver.** Use a `tinygo.org/x/drivers` panel driver as the displayer. Check its method set: drivers v0.31 `st7735`/`st7789` lack `DrawRGBBitmap([]uint16)` and have pointer receivers, so pass `&dev` and expect bitmaps to be skipped ([[t-0001-m5stickc-examples-build]]). Every glyph pixel becomes a 1×1 SPI transfer ([[findings]] F-05).

## 2. Input and peripherals

Buttons, ADCs, PWM: plain `machine` code in board files, exposed to the app through a small interface (`app.Hardware` in `tinytufty2040/app/app.go`) so the UI package never imports `machine`. Input mapping: [[guide-bind-input]].

## 3. Files and build tags

```
examples/<board>/
  app/                  UI only: TinyGUI + your widgets; builds everywhere
  main_<board>.go       //go:build <target tag>   frame loop, wiring
  lcd_<board>.go        //go:build <target tag>   panel driver
  badge_<board>.go      //go:build <target tag>   buttons, sensors, PWM
  sim/main.go           //go:build !tinygo        host simulator
```

Pick the tag TinyGo sets for the target (`tinygo info -target=tufty2040` → `… rp2040 … tufty2040 …`). Prefer capability tags (`serial.usb`, `serial.uart`) over chip lists for optional features ([[topology#Build tags and per-target files]]).

## 4. See it before flashing

The simulator builds the same `app` package with fake hardware and writes PNGs:

```sh
GOTOOLCHAIN=local GOPROXY=off go run ./tinytufty2040/sim -out /tmp/tufty -script "c,a,up,up,shot:edit"
```

Script steps: `up`, `down`, `a`, `b`, `c` (`+` = long), `user:<ms>`, `wait:<frames>`, `shot:<name>`. This host framebuffer → PNG path is the first TG-1c deliverable (UX-28 D3); today it lives in the example.

## 5. Build, size, flash

```sh
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=tufty2040 -size full -o /dev/null ./tinytufty2040
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=tufty2040 -print-allocs='github.com/itohio/tinygui' -o /dev/null ./tinytufty2040
tinygo flash -target=tufty2040 ./tinytufty2040        # owner present
```

Budget items that surprise ([[memory-and-buffers]]): fonts and `var` image tables live in RAM; `widget.Icon` links a PNG decoder with a 34 KB inflate window; `fmt` anywhere ≈ 34 KB flash.
