---
aliases:
  - "Bind input"
  - "Buttons to navigator"
tags:
  - tinygui
  - guides
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# Bind input

From physical buttons to focus moves and value edits. Part of [[guides]]. Reference: [[navigation-and-input]].

## The two modes

TinyGUI keeps two pieces of state per container: which child is **focused** (`Index()`) and whether that child is **active**, i.e. being edited (`Active()`). `nav.Enter()` on a leaf widget activates it by calling `SetActive` on the **container** that holds it (`navigator.go:126`), so "editing" is read from the parent: `root.Active()`. Your input code decides what a button means in each mode:

| Button | Browsing (`!root.Active()`) | Editing (`root.Active()`) |
|---|---|---|
| up / down | `nav.Prev()` / `nav.Next()` | `nav.Current().Interact(ui.UP / ui.DOWN)` |
| ok | `nav.Enter()` | `Interact(ui.ENTER)` (editors commit, `Toggle` flips), then back to browsing |
| back | `nav.Back()` | `Interact(ui.ESC)` (editors revert), then back to browsing |

## Code (used by [[guide-first-screen]], checked on the host)

```go
type button uint8

const (
	none button = iota
	up
	down
	ok
	back
)

// handle maps four buttons onto the navigator and the focused widget.
// Browsing: up/down move focus, ok starts editing.
// Editing (the container is active): up/down change the value,
// ok commits, back reverts; both return to browsing on the same row.
func handle(nav *ui.Navigator, root *container.Base[ui.Widget], b button) {
	if !root.Active() {
		switch b {
		case up:
			nav.Prev()
		case down:
			nav.Next()
		case ok:
			nav.Enter()
		case back:
			nav.Back()
		}
		return
	}
	cur, row := nav.Current(), root.Index()
	switch b {
	case up:
		cur.Interact(ui.UP)
		return
	case down:
		cur.Interact(ui.DOWN)
		return
	case ok:
		cur.Interact(ui.ENTER) // editors commit; a Toggle flips
	case back:
		cur.Interact(ui.ESC) // editors revert
	default:
		return
	}
	nav.Back()     // deactivates the row and clears the focus
	nav.Focus(row) // so focus it again
}
```

## Reading the buttons

- **Debounce and long press** in board code: fire navigation keys on press with auto-repeat while held, fire ok/back on release with "held ≥ 500 ms = long" (`examples/tufty2040/badge_tufty2040.go`, `button.poll`). Long presses map to `ui.LONG_UP`/`LONG_DOWN`, which numeric editors use for the large step.
- **`ui.PeekButton(pin)`** blocks until release and returns the hold time; fine for a rare button (the Tufty demo uses it for USER → screensaver), wrong for navigation keys because the loop stops while the button is down.
- **Serial or scripted input**: `ui.NewCommandStreamMux(reader, map[string]func([]byte))` dispatches text lines. Run it in its own goroutine and only *queue* commands from the handlers; apply them in the UI loop, so the widget tree is touched by one goroutine ([[concurrency]]; `examples/tufty2040/app/commands.go`). Wrap `ui.SerialReader` so it sleeps when nothing is buffered, otherwise the mux goroutine spins (`idleReader` in `examples/tufty2040/main_tufty2040.go`).

## Watch out for

- After `nav.Back()` the container clears its focus (`SetActive(-1)` → `SetIndex(-1)`), hence the `nav.Focus(row)`.
- Nested containers (a `ScrollChoice` list inside a page): `nav.Enter()` on a container pushes it on the navigator stack instead of activating it; `nav.Depth()` tells you which happened ([[example-tufty2040]] list page).
- The container's idle timeout (default 10 s) only fires when you call `root.Interact(ui.IDLE)` (e.g. once per frame, `container/base.go:157-158`, `:260-265`); nothing calls it for you. It counts only `Interact` calls as activity, so focus moved purely through the navigator is cleared after the timeout even while buttons are being pressed ([[findings]] F-24). Workaround: call `root.Interact(ui.USER)` on every press, or `container.WithTimeout` with a long value.
- `ScrollChoice` change callbacks don't fire for navigator moves ([[findings]] F-23); read `list.Index()` instead.
- This glue is the same in every app; a library adapter (`ui.Buttons` + `nav.Handle(cmd)`) is proposed for TG-1b (UX-28 A2).
