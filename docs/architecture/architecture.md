---
aliases:
  - "Architecture"
tags:
  - tinygui
  - architecture
type: hub
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Architecture

As-is description of TinyGUI at `b140156` (module `github.com/itohio/tinygui`, Go 1.23.4 directive). Every claim cites `file:line` at that revision. Intent and roadmap from the original design doc: [[design-intent]]. Home: [[index]].

| Note | Covers |
|---|---|
| [[topology]] | package graph, build tags, external modules, display/input drivers |
| [[render-path]] | input → navigator → widget state → draw → flush |
| [[navigation-and-input]] | `UserCommand`, `Navigator`, container activation, command mux, serial reader, buttons |
| [[widget-catalog]] | every widget and container: binding style, input handling, draw primitives |
| [[memory-and-buffers]] | who owns which buffer, where the heap is touched |
| [[concurrency]] | goroutines, scheduler assumptions, watchdog |
| [[animation-design]] | the `animation` package and its design note |

## Packages

| Package | Path | Responsibility |
|---|---|---|
| `ui` (root) | `*.go` | core interfaces (`Widget`, `Context`, `Navigable`, `Container`, capability interfaces), `WidgetBase`, contexts, drawing helpers, `Navigator`, `CommandStreamMux`, `SerialReader`, button/duration helpers, watchdog shims |
| `widget` | `widget/` | concrete widgets: text, gauges, bitmaps/icons, toggle, separator, multiline/log, interactive editors and selectors |
| `container` | `container/` | `Base[T]` (child list, selection, activation, idle timeout, visibility events), `Scroll`, `ScrollChoice` |
| `layout` | `layout/` | `Strategy` closures that advance a `Context` position between children: `HList`, `VList`, `Grid`, `HFlow`, `VFlow`; `Arranger`s that place all children at once: `Row`/`Column` (weights, alignment), `Wrap`, `Table`, `Anchor` ([[layouts-and-overlays]]) |
| `animation` | `animation/` | `Animator` interface, linear, quadratic ease-in/ease-out and smoothstep (cubic) ease-in-out over `float32` channels (`animation/ease.go:10-41`); not used by any widget (no importer in the repo) |
| examples | `examples/` | M5StickC demos and board support: [[examples]] |
| tools | `cmd/` | `i2cscan` (TinyGo), `png2bin` (host): [[examples]] |

## Core types

| Type | Source | Role |
|---|---|---|
| `UserCommand` (`byte`) | `widget.go:4-31` | abstract input: `IDLE`, `UP`…`BACK`, `DEL`, `RESET`, `SAVE`, `LOAD`, `LONG_*`, `USER`=64 for app codes |
| `Widget` | `widget.go:48-61` | `Size`, `Parent/SetParent`, `Draw(Context)`, `Interact(UserCommand) bool`, `SetSelected/Selected` |
| capability interfaces | `widget.go:65-115` | opt-in: `FocusHandler`, `ActivationHandler`, `Selectable`, `VisibleHandler`, `ScrollHandler`, `SelectHandler`, `ExitHandler`, `EnableState`, `Scrollable` |
| `Navigable` | `widget.go:119-128` | `Widget` + `Index/SetIndex/Item/SetActive/Active/ChildCount/Child`; what the `Navigator` stacks |
| `Container` | `widget.go:133-145` | `Navigable` without child enumeration; no implementation references it in the repo |
| `Sizer`, `Padder`, `Marginer` | `widget.go:33-43` | `Padder`/`Marginer` are declared but nothing calls them; padding/margin are container fields (`container/base.go:22-25`) |
| `WidgetBase` | `widget.go:149-178` | parent, fixed `Width/Height` (`uint16`), selected flag; default `Interact` consumes only `ESC` (deselects) |
| `Context` / `ContextImpl` | `context.go:17-82` | displayer handle, viewport size, origin, cursor (`posX/posY`); `Clone` makes a child context at the current cursor |
| `RandomContext` | `context.go` | OLED burn-in shift: re-randomises the origin after `interval`, on `Move()` — called by `ChildContext(&rc, …)` and `Clone`; an app drawing through `&rc.ContextImpl` calls `Move()` once per frame (no allocation either way) |
| `Displayer` and friends | `drawing.go:11-32` | `BitmapDisplayer`, `LineDisplayer`, `RectangleDisplayer`; widgets type-assert for fast paths and fall back to `SetPixel` loops |
| `Navigator` | `navigator.go:40-323` | stack of `Navigable`, focus/enter/back, observers |
| `layout.Strategy` | `layout/layout.go:12` | `func(ctx ui.Context, w ui.Sizer) bool`; returns false to stop drawing further children |

## Design properties (as built)

- **Retained tree, immediate redraw.** Widgets are long-lived objects in containers; each frame the app redraws the whole tree into the display (`examples/choice/main.go:185-189`). No invalidation or dirty tracking exists (no field or method for it in the repo).
- **Static sizing.** Every widget gets a fixed `Width/Height` at construction (`widget.go:157-162`). Containers with a layout and a zero dimension measure children once in `New` (`container/base.go:111-119`).
- **Absolute coordinates.** Contexts carry absolute display offsets; widgets call `ctx.DisplayPos()` and draw directly to `ctx.D()` (e.g. `widget/gauge.go:68`). There is no clipping: a widget can draw outside its context.
- **Pull-based data binding.** Widgets read state at draw time through pointers (`Gauge.Value *T`, `widget/gauge.go:12`) or closures (`Label.text func() string`, `widget/label.go:14`). Editors keep a pending copy and write back on commit (`widget/interactive_label.go:174-183`).
- **Capability interfaces instead of base-class hooks.** Containers and the navigator type-assert for optional interfaces (`container/base.go:322`, `navigator.go:312-321`).
- **Generics** for numeric widgets (`widget/numeric.go:4-8`) and containers (`container/base.go:11`).

## Known gaps in the as-built architecture

Detailed with evidence in [[findings]]: no clipping, no dirty rectangles, layout closures that keep state across frames, per-frame heap allocations on the render path, and the examples not building for the M5StickC ([[t-0001-m5stickc-examples-build]]).
