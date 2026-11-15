---
aliases:
  - "First screen"
  - "Getting started"
tags:
  - tinygui
  - guides
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# First screen

An editable gauge and an on/off toggle, driven by four buttons. Part of [[guides]].

## The five pieces

1. **Display**: anything that implements `ui.Displayer` (a panel driver, or `examples/framebuffer` for a RAM frame). See [[guide-port-to-board]].
2. **Widgets**: fixed-size objects that read your variables when they draw ([[guide-bind-data]]).
3. **Container**: lays children out (`layout.VList`) and tracks which child is focused and whether it is being edited.
4. **Navigator**: moves focus and enters/leaves editing ([[navigation-and-input]]).
5. **Loop**: read input, clear, draw, `Display()`, sleep. TinyGUI has no loop of its own ([[render-path]]).

## Code

`main` is 25 lines; `handle` is in [[guide-bind-input]], `newDisplay` and `readButton` are your board's ([[guide-port-to-board]]).

```go
package main

import (
	"image/color"
	"time"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
	"tinygo.org/x/tinyfont/proggy"
)

var (
	white = color.RGBA{255, 255, 255, 255}
	teal  = color.RGBA{0, 190, 200, 255}
	grey  = color.RGBA{40, 44, 58, 255}
	black = color.RGBA{0, 0, 0, 255}
)

func main() {
	display := newDisplay() // your board's ui.Displayer (see "port to a board")
	level, pump := 40, false
	font := &proggy.TinySZ8pt7b

	root := container.New[ui.Widget](240, 135,
		container.WithPadding[ui.Widget](8, 8),
		container.WithLayout[ui.Widget](layout.VList(6)),
		container.WithChildren[ui.Widget](
			widget.NewInteractiveGauge[int](200, 12, widget.WithValue(&level),
				widget.WithRange(0, 100), widget.WithSteps(5, 20),
				widget.WithForeground[int](teal), widget.WithBackground[int](grey)),
			widget.NewToggle(60, 12, font, white, "ON", "OFF", teal, grey,
				func() bool { return pump }, func(v bool) { pump = v }),
		))
	nav := ui.NewNavigator(root)
	nav.Focus(0)
	ctx := ui.NewContext(display, 240, 135, 0, 0)

	for {
		handle(nav, root, readButton()) // see "bind input"
		display.FillScreen(black)
		root.Draw(&ctx)
		_ = display.Display()
		time.Sleep(33 * time.Millisecond)
	}
}
```

Checked on the host with a framebuffer display and the button script `ok up up ok down ok ok`: the gauge goes 40 → 45 → 50 and stays at 50 after the commit; the toggle shows ON after the second `ok`.

## Watch out for

- **Every size is in pixels and fixed** (`200, 12`): widgets don't measure their content; text that doesn't fit is drawn past the edge (no clipping, [[findings]] F-13).
- **Content taller than the container is dropped silently**: `VList` stops drawing at the bottom edge ([[findings]] F-20). Add up the heights.
- **Most built-in widgets draw no focus marker.** Focus and edit state exist (`Selected()`, the container's `Active()`); only `InteractiveLabel` shows anything (`▲ value ▼` while selected, if the font has those glyphs, [[guide-bind-data]]). Gauges, toggles, icons and bitmaps look the same focused or not: wrap rows in a widget that draws a frame when selected, as the Tufty demo's `field` does (`tinytufty2040/app/widgets.go`).
- **The type parameter `[ui.Widget]` is needed on every container option** (`container.WithPadding[ui.Widget]`); Go can't infer it from the arguments. A shorter entry point is proposed for TG-1b (UX-28 A1).
- **Full redraw every frame.** On an unbuffered SPI/parallel panel the clear is visible as flicker; draw into a RAM framebuffer and push once ([[guide-port-to-board]]).
- **No host preview in the library itself** yet; `examples/framebuffer` + the Tufty simulator are the way to see a screen without hardware today ([[example-tufty2040]]).
