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
| `fmt` + `strconv` + `internal/strconv` (pulled in by the demos' `fmt.Sprintf`; at `b140156` also by the library's `NewInteractiveLabel` default formatter — since `e335450` the library formats with `strconv` only) | ≈ 36.5 KB | 1 328 B | `-size full`; the probe without `fmt` has none |
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

None after setup — owner, 2026-11-12: "tinygui itself should be 0 allocations (setup does not count)"; it must run with the garbage collector disabled. App code may allocate; the library may not on the paths its users exercise: draw, input, navigation, pages, layout, overlays, animation, text.

What is checked, all in `internal/refapp` (plain `testing`, so TinyGo runs it too). **How allocations are counted matters:** TinyGo's default host GC (Boehm) reports `runtime.ReadMemStats().Mallocs = 0`, but every TinyGo collector keeps the runtime's own `runtime.mallocs` counter (the one TinyGo's `testing.AllocsPerRun` reads). The tests count with `mallocs()` (`mallocs_tinygo_test.go`: `runtime.mallocs` through `//go:linkname`; `mallocs_gc_test.go` on Go, below), so they run, and count exactly, under plain `tinygo test`, `-gc=leaking` and `go test`. `needMallocCount` fails a test (`t.Fatal`, no skip) if a probe allocation isn't counted. `go test ./internal/refapp` also runs `tinygo test -gc=leaking` itself when `tinygo` is on PATH (`TestTinyGoCountsNoAllocations`), so a TinyGo-only allocation fails a plain `go test`.

Measured (TinyGo 0.43, Go 1.27, 2026-11-13: 10 kept `new([16]byte)` appended to a slice, then `AllocsPerRun` of one `new`):

| runtime | `MemStats.Mallocs` | `MemStats.TotalAlloc` | `testing.AllocsPerRun` |
|---|---:|---:|---:|
| Go | 15 | 408 B | 1 |
| TinyGo, default GC (Boehm) | **0** | 624 B | 1 |
| TinyGo `-gc=leaking` | 14 | 400 B | 1 |

So under TinyGo's default GC `Mallocs` is not usable; `runtime.mallocs` is (10 kept `new([16]byte)`: `Mallocs` +0, `runtime.mallocs` +10 under the default GC; +10 and +10 under `-gc=leaking`). Every check counts the exact total over its whole sequence and wants 0: `AllocsPerRun`'s per-run integer average passes an allocation every few runs (one every 7th mux line passed all the old checks).

On Go, `mallocs()` counts from the heap profile (`runtime.MemProfileRate = 1`; `runtime.GC` publishes it, `runtime.MemProfile` reads it): the allocations with this module's code on the stack, except the runtime's own caches for interface type assertions and type switches (`runtime.buildTypeAssertCache`, `runtime.buildInterfaceSwitchCache` in `runtime/iface.go`). The runtime fills those per-call-site caches at random (about 1 miss in 1000, doubled as they grow): about 180 objects over the first ~1500 soak rounds, a different count each run (`docs/evidence/go-soak-allocs/`). Leaving out exactly those stacks makes the Go count exact and the same every run, from the first round; TinyGo has no such cache.

The root, `layout` and `container` tests use `internal/check` (assertions on plain `testing`), so they run under `tinygo test -gc=leaking` too. testify panics under `tinygo test` (reflection) before any test runs. The `widget` package's tests still use testify and are host checks. The library-wide allocation checks live in `internal/refapp`.

| Check | What it covers | Test |
|---|---|---|
| menu-driven soak | a reference app (every widget, container, overlay, `Pages`, the navigator, animation, text) through 20 rounds of focus moves, edits, list scrolling, dropdown, dialog, page history, each followed by a command-mux replay of an 8-line console session: exactly 0 on Go and on TinyGo (default GC and `-gc=leaking`) | `TestSoakDoesNotAllocate` |
| no GC at all | the soak as a program under the leaking allocator | `tinygo run -gc=leaking ./internal/refapp/cmd/refsoak` → "50 rounds, 0 allocations after setup" |
| per widget | exactly 0 over 50 runs of each listed widget's `Draw` and of a command sequence (icons included: decoded once at construction), the overlays, animation, `Log.Append`, a screensaver frame | `TestWidgetsDrawAndInteractWithoutAllocating` |
| first draw | containers and grids size their buffers when built | `TestFirstDrawDoesNotAllocate` |
| console | the command mux reads and dispatches lines: exactly 0 over 50 replays of an 8-line session | `TestCommandMuxDispatchesWithoutAllocating` |
| pages without a GC | the default `Keep` (every page; `WithKeep(n)` with n ≥ the page count does the same) and `BuildAll` in setup: a walk (`Show` each page, then `Back`, with a `Draw` after each) allocates nothing; with fewer kept (`WithKeep(n)`, saving the memory of the released pages), rebuilds allocate. Runs on Go and under TinyGo with either GC; `container` `TestPagesDefaultKeepsEveryPage` counts the zero exactly with no options | `TestPagesWithoutGCNeedEveryPageKept` |
| static | `tinygo build -print-allocs` on `cmd/refsoak`: each library function with heap sites is listed in `allocs_allowlist.txt` with its number of sites, reviewed as setup-only; a new site, even in a listed function, fails | `TestLibraryAllocatesOnlyInSetup` |

Limits: the static check sees only sites the compiler places; allocations inside runtime helpers (`append` growing a slice, string concatenation, conversions, map hashing) show only in the runtime counts above. Library code that no test exercises is not covered — a widget added later needs its own case in the table.

Fixed on the way (findings F-38…F-43): the visibility map, the navigator's paths, `Log`'s copy per line, `ScrollChoice`'s cursor, `InteractiveLabel`'s `Sprintf` text (`e335450`), `ChildContext`'s `Clone` fallback under scroll offsets (`f35331d`). The per-frame `Clone` and per-event sites the table below found at `b140156` are gone.

Outside the library: tinyfont's `const1bit` fonts allocate per rune inside their `GetGlyph` (verify A-6); `ui.DrawText` with a regular `tinyfont.Font` (e.g. `proggy`, `freemono`) allocates nothing. A `Label`'s text provider and option callbacks are app code: they must not allocate either for a GC-free app (the Tufty showcase uses fixed text buffers).

At `b140156` (before TG-1), static analysis on the demos (build-fixed copy, [[t-0001-m5stickc-examples-build]]) found:

| Class | choice | gauge | Sites |
|---|---|---|---|
| construction (once) | 37 | 34 | option closures, constructors, `container.New`, `NewNavigator`, layout closures |
| **every frame** | 2 | 2 | `context.go:79` (`ContextImpl.Clone` returns `&ret`: once per `Base` per frame, twice per `Scroll.Draw` (`container/scroll.go:68`, `:122`), and once per frame in `InteractiveWidgetChoice.Draw` (`widget/interactive_widget_choice.go:94`)), `container/scroll.go:156` (`offsetContext.Clone`) |
| **every input event** | 3 | 2 | `container/choice.go:166` (`ScrollChoice.measure` builds a context), `navigator.go:168` (`Path` for observers), `widget/label.go:52` (`SetText` closure) |

Host measurements then (`testing.AllocsPerRun`, Go 1.27): a 3-widget scene `Base > ScrollChoice > Base{Gauge, Label, InteractiveLabel}`: 14 allocations per frame; `InteractiveLabel.Draw` alone: 1 unselected (`fmt.Sprintf`), 6 selected; `Navigator.Next+Prev` with one observer: 2 ([[baseline#Host probes]]).

Display driver: the ST7735 `setWindow` makes five heap allocations per call: four `[]uint8{…}` literals passed to `Tx` (`examples/m5stick/st7735/st7735.go:222-225`) and `Command(RAMWR)`, which builds `[]byte{command}` (`:226` → `:416`); all are flagged by `-print-allocs`, and upstream drivers v0.31.0 has the same code. Every `FillRectangle`, bitmap and `SetPixel` goes through `setWindow`, so text, which is drawn pixel by pixel, costs five heap allocations per lit pixel.

## Watch out for

- TinyGo's default GC on these targets is `gc.conservative` (target tags in [[topology]]); allocation per frame means periodic GC pauses inside the UI loop #verify duration on the device.
- `var X = []T{…}` assets land in RAM; tables meant for flash need a form TinyGo keeps in `.rodata` (e.g. `string` constants, or arrays the compiler can prove read-only) #verify which forms TinyGo 0.43 places in flash.
