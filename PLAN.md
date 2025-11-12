## TinyGUI Improvement Plan

### Vision
Deliver feature parity with the proven SurroundAmp C++ menu stack while simplifying integration for TinyGo projects. TinyGUI should become the primary UI toolkit for embedded audio gear: efficient, composable, and capable of handling deep navigation, focus, and redraw control without bespoke firmware code.

### Reference Insights (SurroundAmp/include)
- **Hierarchical Menus:** `MenuItemBase` manages parent/child stacks, active selection, and enter/leave transitions.
- **Widget Containers & Layout:** `WidgetContainer` arranges child widgets horizontally, vertically, or relatively, propagating redraw states.
- **Partial Redraw Strategy:** Widgets track border/background/content dirtiness to limit drawing to dirty regions.
- **Input Abstraction:** Rotary encoder events translate into navigation, activation, and value adjustment logic.
- **Stateful Widgets:** Menus remember the current selection and support specialized items (back item, toggle, value editor).

These patterns inform the staged plan below.

Important: 
- conform to Go best practices.
- use testify for tests
- use mockery to mock interfaces
- test as much functionality without actual devices

### Phase 0 – Foundation Alignment
- **Bug Fix:** Replace legacy gauge constructors with generic pointer-based variants and add regression tests.
- **Documentation Sync:** Extend `DESIGN.md` with insights from SurroundAmp (hierarchy, redraw flags) before coding any new features.
- **Testing Harness:** Configure test workflow; add smoke tests for contexts and widget draw ordering.
- **Layout Kickstart:** Implement additional layouters (improved vertical/horizontal spacing, basic grid) to unblock nesting work in later phases.
- **Widget Essentials:** Add at least one interactive widget (e.g., toggle or back-item equivalent) validated with unit tests to exercise navigation plumbing.
- **Label Ownership:** Document and verify that label widgets manage their own font/color configuration, paving the way for theme support.

### Phase 1 – Navigation & Focus Infrastructure
- **Active State Model:** Introduce `FocusHandler`, `ActivationHandler`, and `Selectable` so widgets can declare how they participate in focus/selection.
- **Navigator Abstraction:** Implement a device-independent `Navigator` that operates on the generic `Navigable` interface (menus, tabs, scrollable panes) with stack-based traversal and observer notifications.
- **Menu Walking API:** Provide traversal helpers to inspect the active path, iterate children, and support state restoration.
- **Event Hooks:** Surface focus/activation events through observers so applications can react to navigation changes (backlight, logging, persistence).

### Phase 2 – Layout System Overhaul
- **Flexible Layout Strategies:** Extend layout strategies with a `LayoutOptions` struct (alignment, spacing, wrap policy) and expose negotiation helpers so containers can compute min/preferred sizes before placement.
- **Scrollable Containers:** Add a `container.Scroll` composite that implements `Navigable`, applies clipping, and translates scroll commands (`SCROLL_UP/DOWN/LEFT/RIGHT`, acceleration) into offset adjustments while emitting scroll events.
- **Size Negotiation:** Define optional measurement hooks (`MinSize`, `PreferredSize`) so widgets can hint layout requirements; containers use cached measurements to avoid runtime allocations.
- **Tab / Pager Widget:** Introduce a `TabContainer`/`TabBar` pairing that cooperates with the navigator (focusable headers, activated pages) and supports lazy content drawing for hidden tabs.

### Phase 3 – Rendering Efficiency
- **Dirty Region Tracking:** Integrate border/background/content dirty flags inspired by SurroundAmp into TinyGUI widgets, enabling draw diffing inside `Context`.
- **Visibility Culling:** Extend containers to skip drawing non-visible children (outside viewport or inactive branch).
- **Partial Display Updates:** Add helper to batch `Display()` calls only when necessary, with hooks for double-buffered drivers.

### Phase 4 – Widget Catalog Expansion
- **Menu Widgets:** Port rotary encoder-style selector, toggle, numeric adjuster, and back button equivalents in TinyGo.
- **Scrollable Lists:** Provide list widget with item virtualization for large data sets.
- **Tab Pages & Dialogs:** Create modal and non-modal containers with focus locking; document navigation patterns.
- **Focus Indicators:** Offer theme-aware focus rendering (border color, highlight, animation) configurable per widget.

### Phase 5 – Asynchronous & Input Integration
- **Input Drivers:** Build non-blocking button/encoder drivers emitting `UserCommand` over channels; support debounce and long-press mapping.
- **Scheduler Helpers:** Introduce lightweight goroutine-based event loop coordinating input, redraw, and idle handlers with context cancellation.
- **State Persistence:** Provide patterns for saving/restoring menu paths and widget state (mirroring SurroundAmp’s settings integration).

### Phase 6 – Migration & Samples
- **SurroundAmp Demo:** Recreate the existing menu system using TinyGUI to validate parity; capture performance metrics (navigation latency, redraw cost).
- **Documentation & Guides:** Publish migration guide comparing C++ patterns to TinyGUI equivalents; ensure GoDoc references new features.
- **Board Examples:** Ship RP2040/ESP32 sample apps demonstrating nested menus, scrolling, tabs, and partial redraw.

### Cross-Cutting Themes
- Follow repository rules: prefer composition, dependency injection, asynchronous I/O, and explicit context usage.
- Keep function sizes manageable (<30 lines) by extracting cohesive helpers.
- Maintain zero allocations in hot paths where feasible; document any exceptions.
- Update `DESIGN.md` ahead of each phase to reflect architectural decisions before implementation starts.

### Open Questions
- How to express dirty rectangles within TinyGo display drivers lacking clipping support?
 - if we calculate widget window outside of current context - do not draw/update it - we need to handle clipping.
- Can layout negotiation remain deterministic without dynamic memory?
 - layout negotiation is static now, but can be invoked if it is possible to avoid dynamic memory
- What minimal interface is needed for third-party input sources to drive the navigator?
 - navigator should have device-independent API
 - navigation should be device-independent

All implementation steps must be preceded by documentation updates. This plan should be revisited after each phase to incorporate findings and ensure alignment with the SurroundAmp parity goal.

status tracking should be in place.
backwards compatibility should be maintained as much as possible.

