---
aliases:
  - "Gauge example"
  - "examples/gauge"
  - "TinyWater dashboard demo"
tags:
  - tinygui
  - examples
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# examples/gauge

A three-zone irrigation dashboard modelled on the author's TinyWater device: read-only moisture gauges plus editable thresholds and timings per zone. Part of [[examples]].

## What it shows

| Feature | Widgets / APIs | Source |
|---|---|---|
| Read-only bar bound to a `float32` | `widget.NewGauge[float32](…, &zone.moisture, 0, 100, …)` | `examples/gauge/main.go:98-100` |
| Editable values with range, steps, colours | `widget.NewInteractiveGauge[float32]` + `WithValue`, `WithRange`, `WithSteps`, `WithForeground`, `WithBackground` | `:102-132` |
| Nested containers | per-zone `container.New` (`VList(6)`) inside a `ScrollChoice` (`VList(8)`) inside the root (`VList(10)`) | `:134-138`, `:173-174`, `:192-196` |
| Formatted live text | `Label` closures calling `fmt.Sprintf` every frame | `:157-159`, `:177-188` |
| Two-level navigation | `navigator.Focus(1)` focuses the zone list; `Enter` pushes the `ScrollChoice` (it is `Navigable`), so `Next` then moves between **zones**, not across one zone's gauges; `Back` twice leaves | `:198-199`, `:240-267`; `navigator.go:131-140` |

Controls as coded: Side short = next zone at the top level; inside the list it sends `ENTER` to the current zone container and then `Next`, which moves to the next zone; Side long = enter the list, or commit + leave; Main short/long send `UP`/`DOWN` to the current zone container (which forwards them only while it is active). The gauges inside a zone are reachable only by entering the zone container a second time, which the demo never does.

## Target and build

- Target `m5stick-c`, build tag `m5stickc` (`examples/gauge/main.go:1`); depends on [[example-m5stick]].
- `GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -tags=m5stickc -o /dev/null ./examples/gauge`

## Status at `b140156`

| Check | Result |
|---|---|
| Host | excluded by build tag |
| TinyGo `m5stick-c` | **fails**: `undefined: machine.Serialer`, then `main.go:77:9` `ui.Displayer` mismatch ([[t-0001-m5stickc-examples-build]]) |
| Size after the fix | 131 888 B flash / 12 140 B RAM; ≈ 36.5 KB of the flash is `fmt`/`strconv` for the label `Sprintf`s ([[baseline#Size]]) |
| Allocations after the fix | 34 construction, 2 per frame, 2 per input event inside the library; the label closures' `fmt.Sprintf` calls allocate every frame on top (outside the `-print-allocs` filter) |

## Defects visible in the source

- 240×135 geometry vs the configured 160×80 StickC panel ([[findings]] F-02).
- `–` (U+2013) in status and name strings (`:158`, `:214-225`) has no glyph in `TomThumb`; it renders as a gap ([[findings]] F-10).
- Vertical budget does not fit even the intended 240×135 panel: root padding 10 + readouts 70 + gap 10 + zones 110 + gap 10 + status 14 + gap 10 + summary 14 = 248 px of content against 135 px (`:163`, `:173`, `:177-196`). After the zones block the cursor is at y = 210 > 135, `VList`'s `SetPos` returns false and `Base.Draw` stops (`layout/layout.go:28-30`, `context.go:66-69`, `container/base.go:149-151`), so the status and summary labels are never drawn. Host probe with the same geometry: readouts drawn at (10,10), zones at (10,90), status and summary not drawn ([[baseline#Host probes]]). Inside the 110 px `ScrollChoice` each zone is 100 px with an 8 px gap: zone 1 fits, zone 2 starts at y = 108 (2 px visible), zone 3 at y = 216 is never drawn at `b140156` (verify's reading; the Scroll layout defect F-22 also stopped scrolled-in rows from drawing).
- The summary text fits: the longest zone's line is 212 px wide in `TomThumb` (`tinyfont.LineWidth`, same probe) against a 220 px label.
