---
aliases:
  - "Widget catalog"
  - "Widgets"
tags:
  - tinygui
  - architecture
  - widgets
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Widget catalog

Every widget and container at `b140156`. Part of [[architecture]]. "Binding" = how the widget gets its data at draw time. "Draw path" = displayer calls it makes (fast path → fallback).

## Display widgets

| Widget | Source | Binding | Input | Draw path | Tests |
|---|---|---|---|---|---|
| `Label` | `widget/label.go:11-75` | `func() string` provider; `SetText` wraps a string in a new closure (`:52-54`) | none (ESC deselect) | `tinyfont.WriteLine` at baseline `y+Height` (`:36`) → `SetPixel` per pixel | via `InteractiveLabelChoice` |
| `MultilineLabel` | `widget/multiline.go:161-168` | `SetLines([]string)` copies into own slice (`:71-73`) | none | `WriteLine` per visible line (`:113-132`) | none |
| `Log` | `widget/multiline.go:171-209` | `Append(string)`; capacity cap re-allocates the slice on every append once full (`:201-209`) | none | as above | none |
| `Separator` | `widget/separator.go:10-36` | colour field | none | `HLine`/`VLine` via `SetPixel` loop (`drawing.go:35-50`); does not use `LineDisplayer` | none |
| `Gauge[T]` | `widget/gauge.go:10-118` | `*T` + `Min/Max`; orientation from geometry (`:55`) | none | `FillRectangle` background + fill → `HLine/VLine` loops | `gauge_test.go` |
| `MultiGauge[T]` | `widget/gauge.go:121-261` | `*[]T` stacked segments, colours per segment | none | as `Gauge` | via interactive |
| `VolumeGauge[T]` | `widget/volume_gauge.go:15-106` | `*T`; 12 bars default, green→yellow→red, dimmed above value | none | `fillRect` per bar and gap (`:179-190`); `math.Round` on `float64` per bar (`:86`) | `volume_gauge_test.go` |
| `SolidGauge[T]` | `widget/volume_gauge.go:35-131` | `*T` | none | two `fillRect` | `volume_gauge_test.go` |
| `Bitmap16` / `Bitmap8` | `widget/bitmap.go:32-97` | caller-owned `[]uint16` / `[]uint8`, not copied | none | `DrawRGBBitmap` / `DrawRGBBitmap8`; silently draws nothing if the displayer lacks it or the slice is short (`:47-55`) | `bitmap_test.go` |
| `Icon` | `widget/icon.go:8-55` | `func() string` holding PNG bytes | none | `DrawPng` decodes the PNG on every draw into a shared 1.5 KB line buffer (`drawing.go:52-76`) | via interactive |

## Interactive widgets

| Widget | Source | Edits | Commands | Commit model |
|---|---|---|---|---|
| `Toggle` | `widget/toggle.go:11-79` | `getter/setter func(bool)` | `ENTER/LEFT/RIGHT` flip immediately | immediate; draws background + label via `FillRectangle` |
| `InteractiveLabel[T]` | `widget/interactive_label.go:11-191` | `*T` via `WithValue` | `UP/NEXT ±small`, `LONG_UP/RIGHT ±large`, `ENTER` commit, `ESC/BACK` revert | pending copy; text via `fmt.Sprintf("%v")` and `"▲ "+text+" ▼"` while selected (`:149-155`) |
| `InteractiveGauge[T]` | `widget/interactive_gauge.go:10-195` | `*T` | as label | pending copy drawn by the embedded `Gauge` |
| `InteractiveMultiGauge[T]` | `widget/interactive_gauge.go:198-412` | `*[]T`, max 4 segments (`widget/numeric.go:20`) | `UP/DOWN` value, `LEFT/RIGHT` segment, `ENTER` next segment then commit | fixed arrays `[4]T` for pending/original (`:209-210`) |
| `InteractiveIcon` | `widget/interactive_icon.go:8-149` | `[]string` PNGs, external `*int` | `UP/NEXT/RIGHT` forward, `DOWN/PREV/LEFT` back, wraps | immediate; `SetImage` allocates a closure per change (`widget/icon.go:21-23`) |
| `InteractiveSelector[T]` | `widget/interactive_selector.go:6-168` | `[]T`, external `*int` | as icon; `ESC/BACK` reset to external | helper, not a widget (no `Draw`) |
| `InteractiveLabelChoice` | `widget/interactive_choice.go:11-132` | `[]string` | selector | label text replaced on change |
| `InteractiveIconChoice` | `widget/interactive_icon_choice.go:6-101` | `[]string` PNGs | selector | icon image replaced on change |
| `InteractiveWidgetChoice[T]` | `widget/interactive_widget_choice.go:12-187` | `[]T` widgets | selector, else forwards to current child | draws only the current child |

Options: interactive numeric widgets share `InteractiveOption[T]` (`widget/interactive_options.go:10-194`), a struct of three per-type closures. Unsupported combinations are silently ignored (e.g. `WithFormatter` on a gauge has no `applyGauge`, `:101-107`).

## Containers

| Container | Source | Behaviour |
|---|---|---|
| `Base[T]` | `container/base.go:11-341` | child slice, layout strategy, padding/margin, selection/activation, idle timeout, visibility map (`map[ui.Widget]bool`, `:20`) |
| `Scroll` | `container/scroll.go:9-171` | `Base[ui.Widget]` + offsets; content size measured once at construction (`:88-98`); draws only intersecting children (`:121-142`); observers |
| `ScrollChoice` | `container/choice.go:41-181` | `Scroll` + `InteractiveSelector`; keeps the focused child visible by re-measuring the layout from the first child (`:165-181`) |

## Layouts

`layout/layout.go`: `HList(p)`, `VList(p)` (stateless), `Grid(px,py)`, `HFlow(spacing,max)`, `VFlow(spacing,max)` (stateful closures; state survives across frames, measured defect in [[findings]]).

## Missing compared with mature toolkits (inventory only)

No button, checkbox/radio group, slider with thumb, spinner, text input/keyboard, list with item templates, table, tabs/pager (planned in [[design-intent]]), dialog/message box, chart (line, bar, sparkline, scatter), arc/needle meter, image with transparency, theming/styles, clipping. Gap matrix vs LVGL and Fyne is TG-1b.
