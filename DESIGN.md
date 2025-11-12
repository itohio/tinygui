## TinyGUI Design Overview

TinyGUI is a minimal widget toolkit for TinyGo targets that emphasizes deterministic rendering, limited allocations, and compatibility with microcontroller displays. The library organizes UI construction around three core abstractions: `Widget`, `Context`, and `Container`. Widgets encapsulate drawable entities with a small, testable API. Contexts provide drawing state and access to a `drivers.Displayer`, while containers sequence widgets and handle focus/interaction routing.

### Architectural Goals
- Support resource-constrained devices by avoiding heap churn and reflection-heavy patterns.
- Keep rendering synchronous and deterministic to maintain predictable frame timings.
- Provide composable building blocks (widgets, layouts, containers) that can be extended without modifying core code (Open/Closed principle).
- Allow projects to mix rendering backends that implement `drivers.Displayer` with optional accelerated drawing interfaces (e.g., `RectangleDisplayer`, `LineDisplayer`).
- Enable command-driven interaction (serial, button matrices) that maps to a small set of `UserCommand` values.

### Core Components

**Widget interface (`widget.go`)**
- `Draw(ctx Context)` renders widget content using the cloned context calculated by its parent container.
- `Interact(UserCommand)` allows widgets to react to focused input; defaults provided by `WidgetBase`.
- Metadata: parent pointer, width/height, selection flag. Sizing is static to avoid layout recalculations at runtime.
- Optional capability interfaces keep responsibilities explicit and opt-in only:
  - `Selectable` marks widgets that participate in focus/navigation.
  - `VisibleHandler` reacts to visibility toggles (containers call it only when state changes).
  - `ScrollHandler` receives scroll offset changes.
  - `SelectHandler` notifies widgets when they become (or cease to be) the selected entry.
  - `ExitHandler` fires when the navigator exits an item (e.g. user presses BACK).
  - `EnableState` lets wrappers expose `Enabled()` so navigators can skip disabled entries without extra bookkeeping.

**WidgetBase**
- Convenience struct that implements parent tracking, sizing, and selection state.
- Default `Interact` handles escape by deselecting itself; other commands fall through to the container.

**Container package (`container/`)**
- `Base[T]` embeds `ui.WidgetBase` and handles selection, activation, idle timeouts, and child traversal for any widget slice. It mirrors Fyne’s composable containers but trims allocations for MCU constraints.
- Options (`WithLayout`, `WithChildren`, `WithPadding`, `WithMargin`, `WithTimeout`) configure containers declaratively so constructors stay lean and intent remains explicit.
- `Base` emits opt-in events automatically: `VisibleHandler`, `SelectHandler`, `ExitHandler`, and `ScrollHandler` are invoked only when attached widgets implement them.
- Padding/margin offsets adjust the child context before layouts run so nested containers can respect spacing without hand-rolled coordinate tweaks.
- `Scroll` composes `Base[ui.Widget]` with scroll offsets. It only draws visible children, leaving parent contexts untouched while notifying observers of offset changes.
- `ScrollChange` / `ScrollObserver` let higher-level widgets (e.g., navigable lists) synchronise scrolling with focus changes.
- Tab/pager components will layer on top by composing `Base` and wiring into the navigator.

**Layout package (`layout/`)**
- Provides static `Strategy` functions (`HList`, `VList`, `Grid`, `HFlow`, `VFlow`) that mutate contexts between child draws.
- Layouts are fixed at construction to keep behaviour deterministic; dynamic layouts can compose on top without modifying core data structures.

**Context implementations (`context.go`)**
- `ContextImpl` holds the display handle, dimensions, and drawing origin. `Clone` produces a child context for nested widgets while maintaining absolute display coordinates.
- `RandomContext` periodically shifts the drawing origin within the physical display bounds to mitigate OLED burn-in. Reuses `ContextImpl` cloning logic.

**Drawing helpers (`drawing.go`)**
- Provides fallbacks for drawing lines and rectangles when optimized interfaces are unavailable.
- Integrates PNG decoder via `tinygo.org/x/drivers/image/png` with a callback-based renderer that streams decoded pixels to `BitmapDisplayer`.

### Widget Catalog (`widget/`)
- `Label`, `MultilineLabel`, and `Log` support text rendering via `tinyfont`, using closures for dynamic content.
- `Gauge[T]` covers horizontal/vertical progress displays, binding directly to mutable value pointers without additional callbacks.
- `Icon` wraps `DrawPng` to render embedded images.
- `Separator` renders horizontal or vertical rules based on its dimensions, reusing accelerated displayer paths when available.
- Widget constructors encapsulate size configuration, ensuring deterministic layout footprints.
- Text widgets own their font and color configuration so they remain self-contained and theme-ready.
- Interactive widgets (e.g., toggle/selector) encapsulate their behaviour by accepting getter/setter callbacks, enabling focus-driven state changes without direct hardware coupling.
- `InteractiveLabel` embeds a `Label`, annotates text with ▲/▼ while selected, and edits pointer-backed values using opt-in options (`WithValue`, `WithRange`, `WithSteps`, etc.) without extra allocation.
- `InteractiveIcon` embeds `Icon`, cycling through preloaded PNGs on directional commands while optionally mirroring an external index for deterministic state.
- `InteractiveChoice` embeds a `Label`, rotating through a static string table with optional external index wiring for deterministic selection.
- `HorizontalInteractiveGauge` / `VerticalInteractiveGauge` compose the gauge displays for single values, while multi-value variants wrap `HorizontalMultiGauge` / `VerticalMultiGauge` to provide segment navigation (ENTER to advance, BACK/ESC to revert) and option-driven configuration.
- Phase 2 adds new composites:
  - `Toggle` and future selectors implement `Selectable`/`EnableState` to opt into navigation.
  - `Tab` widgets encapsulate tab headers, track active state, and forward activation commands to associated pages.
  - Scrolling-aware widgets (log, multiline labels) leverage `container.Scroll` to redraw only visible lines and can opt into `ScrollHandler` for fine control.

### Input & Command Handling (`mux.go`)
- `CommandStreamMux` parses newline-delimited commands from an `io.Reader`, dispatching to registered callbacks without extra allocations. Designed for serial command channels or scripting interfaces.
- `SerialReader` adapts `machine.Serialer` to `io.Reader`, reading bytes while respecting buffered availability.

### Tooling (`cmd/`)
- `i2cscan`: simple utility leveraging TinyGo drivers to enumerate I2C devices.
- `png2bin`: converts PNG/JPEG assets into Go source arrays (RGB565) suitable for embedding; reinforces image handling workflow for `Icon` widgets.

### Platform Integration
- Watchdog support is abstracted via build tags (`watchdog_rp2040.go`, `watchdog_esp32.go`), enabling button polling to keep watchdog timers alive without coupling UI logic to specific targets.
- The UI event layer expects button input via `PeekButton`, with long-press detection returning duration thresholds to map into `UserCommand` variants (e.g., `LONG_UP`).

### Rendering Flow
1. Application creates a root `Context` with a displayer instance and desired viewport dimensions.
2. Containers draw child widgets sequentially, cloning contexts to adjust origins.
3. Layout callbacks mutate context positions between draws, establishing simple flow layouts.
4. After draw pass, application calls the underlying display’s `Display()` (outside library) to flush.

### Interaction Flow
1. Input devices map hardware actions to `UserCommand` values.
2. Root container receives `Interact` calls with commands.
3. Inactive containers use `NEXT`/`PREV` to change selection; `ENTER` activates the focused child.
4. Active child widget processes commands; `ESC` or inactivity timeout returns focus to parent.
- Phase 1 adds a dedicated `Navigator` that manages a stack of `Navigable` widgets (containers, tabs, scroll panes). Navigation commands update this stack and emit focus/activation events mirroring SurroundAmp semantics.
- The navigator exposes a device-independent API (`Focus`, `Next`, `Prev`, `Enter`, `Back`, `WalkPath`) so encoders, buttons, or scripted command streams can drive traversal without coupling to specific widgets.
- Selection change events bubble via observer interfaces, enabling backlight control, logging, or persistence of the active menu path.
- Phase 2 integrates scroll commands (`SCROLL_UP`, `SCROLL_DOWN`, etc.) so navigator-aware containers adjust viewports while maintaining predictable focus. Layout negotiation metadata (`MinSize`, `PreferredSize`) will let containers respect widget sizing hints before scrolling.

### Extensibility Points
- Developers can implement new widgets by embedding `WidgetBase` and providing `Draw`/`Interact`.
- Custom layouts: supply a `layout.Strategy` closure to apply grid, flex, or absolute positioning.
- Alternative contexts: embed additional behavior (double-buffering, clipping) by implementing `Context`.
- Rendering acceleration: add interfaces similar to `RectangleDisplayer` to leverage hardware features.

### Current Limitations
- Layout system only offers simple sequential positioning; `layout.Grid` is a stub.
- Focus management is container-centric; nested containers require manual orchestration for complex navigation trees.
- Widget sizing is entirely static—no content measurement or adaptive layout.
- Legacy gauges now replaced by generic pointer-driven `Gauge[T]` to guarantee dynamic updates without closure indirection.
- `PeekButton` sleeps for fixed intervals and busy-waits, which may block cooperative scheduling on some targets.
- Lack of formal theme/styling abstraction; colors/fonts are set per widget.
- No integration with asynchronous data sources or event pumps; applications poll and redraw manually.

### Strategic Direction
- Maintain lightweight footprint compatible with TinyGo and microcontroller constraints.
- Gradually expand layout and widget set while preserving deterministic behavior.
- Introduce configuration patterns (options structs, interfaces) instead of inheritance for future features.
- Encourage asynchronous-safe APIs (non-blocking input, rendering triggers) to match embedded runtime patterns.
- Navigation is device-independent: hardware adapters push abstract `UserCommand`s into the navigator, and focus/activation changes propagate via callback interfaces instead of concrete types.
- Clipping discipline: containers calculate child bounds before draw; if a child lies outside the current viewport it is skipped, with navigator still tracking it for structural completeness.
- Status tracking: navigators emit structured events so applications can persist and restore menu paths, ensuring back-compat for existing TinyGUI apps while unlocking advanced menu flows.
- Layout extensibility: upcoming `LayoutOptions` carry alignment, spacing, and wrapping hints so containers can compose complex grids without bespoke logic.
- Scroll performance: dirty-region tracking combines with viewport calculations to limit redraw to visible content, keeping frame times stable on constrained MCUs.

This document captures the current structure to inform future planning (`PLAN.md`) and ensure subsequent enhancements remain consistent with the library’s guiding principles.

- Option helpers (`InteractiveOption[T]`) provide a shared configuration surface (`WithValue`, `WithRange`, `WithSteps`, `WithForeground`, etc.), keeping constructors minimal and aligned with the opt-in capability philosophy.

