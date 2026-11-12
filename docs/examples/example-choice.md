---
aliases:
  - "Choice example"
  - "examples/choice"
tags:
  - tinygui
  - examples
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# examples/choice

A navigator-driven settings screen: pick a feeder mode (text choice) and a status icon (bitmap choice). Part of [[examples]].

## What it shows

| Feature | Widgets / APIs | Source |
|---|---|---|
| Text option cycling bound to an `int` | `widget.NewInteractiveLabelChoice` + `WithLabelChoiceIndex(&modeIndex)` + change callback | `examples/choice/main.go:129-135` |
| Choosing among widgets | `widget.NewInteractiveWidgetChoice[ui.Widget]` over four `Bitmap16` icons | `:137-150` |
| Scrollable focus list | `container.NewScrollChoice(…, layout.VList(12), …)` with change callback | `:155-162` |
| Live text | `widget.NewLabel` with closures over app strings | `:164-165` |
| Root layout | `container.New` with padding 10 and `VList(12)` | `:173-177` |
| Navigation + observer | `ui.NewNavigator(root)`, `AddObserver` updates the status line on focus/activate/deactivate | `:87-109`, `:179-182` |
| Input | two buttons, short / long (≥600 ms) / ultra (≥1.5 s) press on release | `:30-78`, `:191-244` |

Controls: Side short = next item; Side long = enter (or commit + back when editing); Side ultra = cancel + back. Main short/long = up/down while editing; Main ultra = back.

## Target and build

- Target `m5stick-c` with build tag `m5stickc` (`examples/choice/main.go:1`); depends on [[example-m5stick]].
- `GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -tags=m5stickc -o /dev/null ./examples/choice`

## Status at `b140156`

| Check | Result |
|---|---|
| Host | excluded by build tag (nothing to build) |
| TinyGo `m5stick-c` | **fails**: `serial_reader.go:10:12: undefined: machine.Serialer`, then `main.go:84:9` `ui.Displayer` mismatch ([[t-0001-m5stickc-examples-build]]) |
| Size after the fix | 139 132 B flash / 22 252 B RAM; icons take 10 112 B of RAM ([[baseline#Size]]) |
| Allocations after the fix | 37 construction, 2 per frame, 3 per input event ([[memory-and-buffers#Heap allocations]]) |

## Defects visible in the source

- Screen geometry 240×135 (`:22-23`) does not match the configured 160×80 StickC panel ([[findings]] F-02).
- All four bitmaps are sized `AquariumWidth×AquariumHeight` (40×40), but `ThermometerPng` is 16×16 (`examples/choice/icons/thermometer.go:4-5`), so `Bitmap16` refuses to draw the fourth choice (`widget/bitmap.go:52-55`) ([[findings]] F-12).
- `icons/generate.go` points at a non-existent `../media` generator ([[findings]] F-16); regenerate with [[tool-png2bin]].
- Full clear + redraw every 33 ms ([[render-path]]).
