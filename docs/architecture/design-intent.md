---
aliases:
  - "Design intent"
  - "DESIGN.md"
  - "TinyGUI design overview"
tags:
  - tinygui
  - architecture
  - design
type: reference
created: 2025-11-12
updated: 2026-11-10
revision: b140156
---

# Design intent

The author's design overview, folded from the root `DESIGN.md` (moved here on 2026-11-10 so the repo has one documentation home). Goals and direction are kept verbatim; the as-built description now lives in [[architecture]] and its notes, and the claims that no longer match the code are listed under [[#Drift from the code]]. Part of [[architecture]].

## Summary

TinyGUI is a minimal widget toolkit for TinyGo targets that emphasizes deterministic rendering, limited allocations, and compatibility with microcontroller displays. The library organizes UI construction around three core abstractions: `Widget`, `Context`, and `Container`. Widgets encapsulate drawable entities with a small, testable API. Contexts provide drawing state and access to a `drivers.Displayer`, while containers sequence widgets and handle focus/interaction routing.

## Architectural goals

- Support resource-constrained devices by avoiding heap churn and reflection-heavy patterns.
- Keep rendering synchronous and deterministic to maintain predictable frame timings.
- Provide composable building blocks (widgets, layouts, containers) that can be extended without modifying core code (Open/Closed principle).
- Allow projects to mix rendering backends that implement `drivers.Displayer` with optional accelerated drawing interfaces (e.g., `RectangleDisplayer`, `LineDisplayer`).
- Enable command-driven interaction (serial, button matrices) that maps to a small set of `UserCommand` values.

## Planned (not built at `b140156`)

- Tab/pager components layered on `Base` and wired into the navigator.
- `Tab` widgets with headers, active state and page activation.
- Scroll commands (`SCROLL_UP`, `SCROLL_DOWN`, …) integrated into the navigator.
- Layout negotiation metadata (`MinSize`, `PreferredSize`) and `LayoutOptions` (alignment, spacing, wrapping hints).
- Dirty-region tracking combined with viewport calculation to limit redraw to visible content.
- Persisting and restoring menu paths from navigator events.
- A theme/styling abstraction (listed as a limitation).
- Asynchronous-safe APIs: non-blocking input and rendering triggers.

## Strategic direction

- Maintain lightweight footprint compatible with TinyGo and microcontroller constraints.
- Gradually expand layout and widget set while preserving deterministic behavior.
- Introduce configuration patterns (options structs, interfaces) instead of inheritance for future features.
- Encourage asynchronous-safe APIs (non-blocking input, rendering triggers) to match embedded runtime patterns.
- Navigation is device-independent: hardware adapters push abstract `UserCommand`s into the navigator, and focus/activation changes propagate via callback interfaces instead of concrete types.
- Clipping discipline: containers calculate child bounds before draw; if a child lies outside the current viewport it is skipped, with navigator still tracking it for structural completeness.
- Status tracking: navigators emit structured events so applications can persist and restore menu paths, ensuring back-compat for existing TinyGUI apps while unlocking advanced menu flows.
- Layout extensibility: upcoming `LayoutOptions` carry alignment, spacing, and wrapping hints so containers can compose complex grids without bespoke logic.
- Scroll performance: dirty-region tracking combines with viewport calculations to limit redraw to visible content, keeping frame times stable on constrained MCUs.

## Drift from the code

Claims of the former `DESIGN.md` (and two doc comments) that differ from `b140156`:

| Former claim | Code | Source |
|---|---|---|
| "`layout.Grid` is a stub" | `Grid` is implemented (row wrap by context width) | `layout/layout.go:35-69` |
| `ScrollChoice` is "Planned" | implemented and tested | `container/choice.go:41-181`, `container/choice_test.go` |
| Navigator API includes `WalkPath` | the method is `Walk` | `navigator.go:185` |
| Doc comment on `ActivationHandler`: "The navigator toggles them on ENTER/BACK" (`widget.go:70-72`, not `DESIGN.md`) | containers toggle them in `SetActive`; the navigator only calls `SetActive` | `container/base.go:200-232`, `navigator.go:126`, `:153` |
| `InteractiveLabel` edits "without extra allocation" | formats with `fmt.Sprintf` and string concatenation on every draw | `widget/interactive_label.go:37`, `:149-155` |
| `Separator` reuses "accelerated displayer paths" | always `SetPixel` loops via `HLine/VLine` | `widget/separator.go:31-34`, `drawing.go:35-50` |
| "Provides fallbacks for drawing lines and rectangles" | line fallbacks exist; there is no rectangle fallback helper in `drawing.go` (widgets inline their own loops) | `drawing.go`, `widget/volume_gauge.go:179-190` |
| `png2bin` converts "PNG/JPEG" | decodes PNG only; `.jpg` passes the extension check and then fails in `png.Decode` | `cmd/png2bin/png2bin.go:39-56` |
| Watchdog abstraction "enabling button polling to keep watchdog timers alive" | ESP32 implementation is a no-op | `watchdog_esp32.go:5` |
| Refers to `PLAN.md` | no such file in the repo | `git ls-files` |
| `Log` and multiline labels "leverage `container.Scroll`" | they keep their own `viewStart` and do not use `Scroll` | `widget/multiline.go:212-226` |
| `ExitHandler` "fires when the navigator exits an item" | called by containers from `SetActive` when the active child changes or is deactivated | `container/base.go:200-232` |
| `CommandStreamMux` dispatches "without extra allocations" | `string(line)` builds the map key from the line | `mux.go:90-93` |
| `PeekButton` long presses map to `LONG_*` | `PeekButton` returns a duration; nothing maps it to `LONG_*` | `ui.go:12-20` |
| interactive widgets "accept getter/setter callbacks" | only `Toggle` does; the others bind pointers | `widget/toggle.go:24`, `widget/interactive_options.go:35-44` |
| `Base` emits `ScrollHandler` | only `Scroll` does | `container/scroll.go:77-86` |
| "Focus management is container-centric; nested containers require manual orchestration" | the `Navigator` keeps a stack of `Navigable` containers and descends with `Enter` | `navigator.go:40-55`, `:121-144` |
| `layout` package doc: strategies respect `Marginer`/`Padder` | no strategy reads them | `layout/layout.go:1-6`, `:14-151` |
