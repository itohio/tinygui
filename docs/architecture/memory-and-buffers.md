---
aliases:
  - "Memory and buffers"
  - "Allocations"
tags:
  - tinygui
  - architecture
  - memory
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Memory and buffers

Who owns which memory and where the heap is touched, at `b140156`. Numbers from [[baseline]] (TinyGo 0.43, `m5stick-c`). Part of [[architecture]].

## Footprint

| Item | Flash | RAM | Source |
|---|---|---|---|
| TinyGUI packages (`ui`+`container`+`layout`+`widget`), choice demo | 9 574 B code | 0 B static (the 1 536 B PNG line buffer `drawing.go:52` is linked only when `DrawPng`/`Icon` is used; 1 536 B `bss` in the Tufty build) | `-size full`, [[baseline#Size]] |
| same, minimal probe (label, toggle, container, navigator) | 2 054 B code | 0 B static | probe on `pico` |
| `tinyfont` incl. `TomThumb` | 6 402 B | **5 845 B** (`data`: glyph tables are `var` slices, copied to RAM) | `-size full` |
| choice demo icons (`[]uint16` vars) | 10 112 B | **10 112 B** (`data`) | `-size full` |
| `fmt` + `strconv` + `internal/strconv` (pulled in by the demos' `fmt.Sprintf` and by the library: `NewInteractiveLabel`'s default formatter, `widget/interactive_label.go:37`) | ≈ 36.5 KB | 1 328 B | `-size full`; the probe without `fmt` has none |
| whole choice demo | 139 132 B | 22 252 B | `-size short` totals |

The library has no static RAM except the 1 536 B PNG line buffer, linked only with `DrawPng`/`Icon`. RAM goes to fonts and images declared as package-level slice variables, which TinyGo places in `.data` (copied to RAM at boot) rather than flash.

## Buffer ownership

| Buffer | Owner | Lifetime | Source |
|---|---|---|---|
| Framebuffer | none: TinyGUI never allocates one; widgets write straight to the displayer | — | `widget/gauge.go:74-88` |
| PNG decode line buffer `[3*256]uint16` (1.5 KB) | package global in `ui`, shared by every `DrawPng` call | program | `drawing.go:52` |
| `Bitmap16/8` pixels | caller; widget keeps the slice reference, never copies | caller's | `widget/bitmap.go:8-29` |
| Label text | caller's closure returns a string each draw | per draw | `widget/label.go:14`, `:36` |
| `MultilineBase.lines` | widget; `SetLines` copies the slice of string headers (`append(m.lines[:0], …)`) | widget | `widget/multiline.go:71-73` |
| `Log.lines` | widget; re-allocated on every append at capacity | widget | `widget/multiline.go:201-209` |
| Multi-gauge pending/original | widget; fixed `[4]T` arrays | widget | `widget/interactive_gauge.go:209-210` |
| Container visibility map | container, `map[ui.Widget]bool`, one entry per child | container | `container/base.go:20`, `:313-325` |
| Command line buffer | mux, `make([]byte, 0, 128)`, grows with `append` beyond 128 | mux | `mux.go:23`, `:55`, `:69` |
| Navigator stack/observers | navigator slices, grow with depth/observers | navigator | `navigator.go:40-59` |

## Heap allocations

Static analysis with `tinygo build -print-allocs=github.com/itohio/tinygui` on the demos (build-fixed copy, [[t-0001-m5stickc-examples-build]]):

| Class | choice | gauge | Sites |
|---|---|---|---|
| construction (once) | 37 | 34 | option closures, constructors, `container.New`, `NewNavigator`, layout closures |
| **every frame** | 2 | 2 | `context.go:79` (`ContextImpl.Clone` returns `&ret`: once per `Base` per frame, twice per `Scroll.Draw` (`container/scroll.go:68`, `:122`), and once per frame in `InteractiveWidgetChoice.Draw` (`widget/interactive_widget_choice.go:94`)), `container/scroll.go:156` (`offsetContext.Clone`) |
| **every input event** | 3 | 2 | `container/choice.go:166` (`ScrollChoice.measure` builds a context), `navigator.go:168` (`Path` for observers), `widget/label.go:52` (`SetText` closure) |

Not covered by that filter: allocations inside other packages called from the render path. Host measurements (`testing.AllocsPerRun`, Go 1.27 escape analysis, indicative only) on a 3-widget scene `Base > ScrollChoice > Base{Gauge, Label, InteractiveLabel}`: **14 allocations per frame**; a single `Base{Gauge}`: 1; `InteractiveLabel.Draw` alone: 1 unselected (`fmt.Sprintf`), 6 selected (format + `"▲ "+text+" ▼"`); `Navigator.Next+Prev` with one observer: 2 ([[baseline#Host probes]]).

Display driver: the ST7735 `setWindow` makes five heap allocations per call: four `[]uint8{…}` literals passed to `Tx` (`examples/m5stick/st7735/st7735.go:222-225`) and `Command(RAMWR)`, which builds `[]byte{command}` (`:226` → `:416`); all are flagged by `-print-allocs`, and upstream drivers v0.31.0 has the same code. Every `FillRectangle`, bitmap and `SetPixel` goes through `setWindow`, so text, which is drawn pixel by pixel, costs five heap allocations per lit pixel.

## Watch out for

- TinyGo's default GC on these targets is `gc.conservative` (target tags in [[topology]]); allocation per frame means periodic GC pauses inside the UI loop #verify duration on the device.
- `var X = []T{…}` assets land in RAM; tables meant for flash need a form TinyGo keeps in `.rodata` (e.g. `string` constants, or arrays the compiler can prove read-only) #verify which forms TinyGo 0.43 places in flash.
