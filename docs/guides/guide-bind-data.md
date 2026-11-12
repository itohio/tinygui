---
aliases:
  - "Bind data"
  - "Data binding"
tags:
  - tinygui
  - guides
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# Bind data

How widgets get your values, and how edits come back. Part of [[guides]]. Reference: [[widget-catalog]] (Binding column), [[architecture#Design properties (as built)]].

## Three binding styles

| Style | Widgets | You write | Redraw |
|---|---|---|---|
| **Pointer** | `Gauge`, `MultiGauge`, `VolumeGauge`, `SolidGauge` | `widget.NewGauge[float32](w, h, &temp, 0, 100, fg, bg)` | next frame reads `*ptr`; just assign `temp = 23.5` |
| **Closure** | `Label`, `Icon`, `Toggle` (getter/setter) | `widget.NewLabel(w, h, font, func() string { return status }, c)` | called on every draw |
| **Pending copy + commit** | `InteractiveLabel`, `InteractiveGauge`, `InteractiveMultiGauge` | `widget.WithValue(&level)`, optional `widget.WithCommit(func(v int){ … })` | edits change a private copy; `ENTER` writes it to `*ptr` and calls the commit callback; `ESC` restores |

Selectors (`InteractiveLabelChoice`, `InteractiveIconChoice`, `InteractiveWidgetChoice`, `ScrollChoice`) bind an index with `With…Index(&i)` and call `With…Change(func(i, v))` on every move.

## Rules that keep it simple

- **The bound pointer is the source of truth.** Use change/commit callbacks for side effects (switch an LED, restart an animation), not to copy the value somewhere else (UX-28 A4).
- **One goroutine owns the values the UI reads.** Sensors running in other goroutines hand values over (a channel or a queue drained in the UI loop), as the Tufty demo does for serial commands ([[concurrency]]).
- **Sample slow sources on a timer**, not inside a closure that runs every frame (`App.Update` in `examples/tufty2040/app/app.go` reads the ADCs every `samplePeriod`).

## Text without allocating

A label closure runs every frame, so `fmt.Sprintf` there allocates every frame and links ≈ 34 KB of `fmt`/`strconv` ([[findings]] F-26). Format into a fixed buffer owned by the widget's data instead:

```go
var buf textBuf // examples/tufty2040/app/text.go: fixed [48]byte, int/fixed/pad2 writers
label := widget.NewLabel(170, 11, font, func() string {
	return buf.reset().str("Battery ").fixed(vbat, 2).str(" V").String()
}, fg)
```

The returned string aliases the buffer; it's valid until the next `reset`, which is the next draw. The same pattern works as an `InteractiveLabel` formatter (`widget.WithFormatter`).

## Watch out for

- `NewInteractiveLabel` without `WithFormatter` uses `fmt.Sprintf("%v")` and decorates the selected value with `▲ ▼`, glyphs most tinyfont fonts don't have ([[findings]] F-10).
- `InteractiveMultiGauge` holds at most 4 values (`widget/numeric.go:20`).
- `Label.SetText` and `Icon.SetImage` allocate a new closure per call; call them on change, not per frame.
