---
aliases:
  - "Baseline"
  - "Baseline measurements"
tags:
  - tinygui
  - analysis
  - measurement
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Baseline

Measured facts at `b140156`, 2026-11-10, darwin/arm64, Go 1.27.2, TinyGo 0.43.0 (LLVM 22.1.4), drivers v0.31.0, tinyfont v0.6.0. Every Go/TinyGo command ran with `GOTOOLCHAIN=local GOPROXY=off`. Part of [[analysis]].

## Host tests

| Command | Result |
|---|---|
| `go test -count=1 ./...` | `ui`, `container`, `layout`, `widget` pass; `examples/m5stick/i2c`, `examples/m5stick/st7735` fail setup: `package machine is not in std` (TinyGo-only packages without build tags) |
| `go test -count=1 -race ./...` | same; no races reported |
| `go vet ./...` | fails on the same `machine` import; clean on the host-buildable packages |
| `go test -cover` | `ui` 49.6 %, `container` 43.5 %, `layout` 80.8 %, `widget` 51.5 %, `animation` 0 % (no tests) |

Host coverage of the root package excludes `ui.go` (`PeekButton`, `DurationAdd/Sub`): its build tags exclude the host ([[topology]]).

## TinyGo builds

| Build | Result |
|---|---|
| `tinygo build -target=m5stick-c -tags=m5stickc ./examples/choice` (and `gauge`) | **fails**: `serial_reader.go:10:12: undefined: machine.Serialer` |
| same after the `serial_reader.go` fix only | **fails**: `examples/choice/main.go:84:9: cannot use display (variable of struct type st7735.Device) as ui.Displayer value in return statement: … (method Display has pointer receiver)` (gauge: `main.go:77:9`) |
| same + `return &display` | **fails**: `*st7735.Device does not implement ui.Displayer (missing method DrawRGBBitmap)` |
| fix + demos return `*st7735.Device` (upstream driver) | builds; `Bitmap16` then draws nothing (no `BitmapDisplayer`) |
| fix + board package uses the vendored `examples/m5stick/st7735` | builds; full `ui.Displayer` |
| `tinygo build -target=m5stick-c ./cmd/i2cscan` | builds |
| `go build ./cmd/png2bin` | builds (host tool); not run: no PNG in the repo |

Exact diffs: [[t-0001-m5stickc-examples-build]].

## Size

`tinygo build -target=m5stick-c -tags=m5stickc -size full` on the build-fixed copies:

| Build | Flash | RAM | TinyGUI code | Largest RAM users |
|---|---|---|---|---|
| choice, upstream driver | 139 132 | 22 252 | 9 574 | icons 10 112 (`data`), tinyfont 5 845, C stack 4 096 |
| gauge, upstream driver | 131 888 | 12 140 | 9 915 | tinyfont 5 845, C stack 4 096 |
| choice, vendored driver | 139 364 | 22 252 | 9 629 | as above |
| gauge, vendored driver | 131 568 | 12 140 | 9 914 | as above |
| minimal probe (label, toggle, container, navigator, null display), `pico` | 47 748 | 12 296 | 2 054 | tinyfont 5 845, C stack 4 096 |
| same, `m5stick-c` | 52 876 | 10 772 | — | — |

TinyGUI packages have 0 B rodata/data/bss in every build. `fmt` + `strconv` + `internal/strconv` ≈ 36.5 KB flash in both demos (their `fmt.Sprintf` calls); absent from the probe.

## Allocations

`tinygo build -target=m5stick-c -tags=m5stickc -print-allocs=github.com/itohio/tinygui`:

| Build | Sites | Construction | Per frame | Per input |
|---|---|---|---|---|
| choice | 42 | 37 | 2 | 3 |
| gauge | 38 | 34 | 2 | 2 |
| probe | 10 | 8 | 1 | 1 |
| choice, vendored driver | 50 | 40 | 7 (+5 in `setWindow`: `st7735.go:222-225` and `Command` at `:416`, per draw primitive) | 3 |

Site list and classification: [[memory-and-buffers#Heap allocations]].

## Host probes

Throwaway tests in a separate module (with `replace github.com/itohio/tinygui` pointing at this repo), `go test -count=1 -v ./...`:

```
TestGridStableAcrossFrames   FAIL  frame1=[0 10] frame2=[0 30] for the 4th child (Grid(0,0), width 25, children 10x10,10x10,10x10,10x30)
                                    3rd child drawn at [20 0] in a 25 px wide grid
TestAllocsPerFrame                 14.0 allocs/frame  Base > ScrollChoice > Base{Gauge, Label, InteractiveLabel}
TestAllocsPerFrameGaugeOnly         1.0 allocs/frame  Base{Gauge}
TestAllocsInteractiveLabelDraw      1.0 unselected, 6.0 selected per Draw
TestAllocsNavigatorEvent            0.0 without observers, 2.0 with one observer per Next+Prev
TestLogAppendAllocs                 2.0 per Append at capacity 16
TestGridPaddingDropsChild          Grid(0,0), width 25, padding 5, three 10x10 children: drawn at [5 0], [15 0], third never drawn
TestGaugeExampleVerticalBudget     gauge-demo root geometry with boxes: readouts at [10 10], zones at [10 90], status and summary never drawn;
                                    longest summary line 212 px in TomThumb (tinyfont.LineWidth)
```

Host Go and TinyGo escape analysis differ; host numbers show where allocations are, TinyGo `-print-allocs` is the reference for the target.

## Other facts used in the notes

| Fact | Command / source |
|---|---|
| `m5stick-c` target tags: `xtensa baremetal … esp32 … scheduler.tasks … serial.uart tinygo.unicore gc.conservative` | `tinygo info -target=m5stick-c` |
| `xiao-esp32s3` tags lack `esp32` (has `esp32s3`, `serial.usb`) | `tinygo info -target=xiao-esp32s3` |
| TinyGo 0.43 offers `-scheduler=cores` with `numCPU = 2` on RP2040/RP2350 | `tinygo help`; `src/runtime/runtime_rp2.go:21` |
| `machine.Serialer` defined only in USB files (`usb.go:45`, `machine_esp32c3_usb.go:43`, `machine_esp32c6_usb.go:42`, `machine_esp32xx_usb.go:43`) | `grep -rn Serialer $(tinygo env TINYGOROOT)/src/machine` |
| `TomThumb` last glyph rune 235 (U+00EB) | `tinyfont@v0.6.0/tomthumb.go` |
| no goroutines, channels, `sync` or interrupts in library code | `grep -rnE '\bgo [a-zA-Z(]|chan |"sync"|sync/atomic|SetInterrupt' --include=*.go .` → comments only |
