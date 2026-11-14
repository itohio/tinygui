package container

import (
	ui "github.com/itohio/tinygui"
)

func focusTransition(from, to ui.Widget) {
	if from != nil {
		if handler, ok := from.(ui.FocusHandler); ok {
			handler.OnBlur()
		}
	}
	if to != nil {
		if handler, ok := to.(ui.FocusHandler); ok {
			handler.OnFocus()
		}
	}
}

func activationTransition(w ui.Widget, active bool) {
	if w == nil {
		return
	}
	handler, ok := w.(ui.ActivationHandler)
	if !ok {
		return
	}
	if active {
		handler.OnActivate()
		return
	}
	handler.OnDeactivate()
}

func childVisible(ctx ui.Context, child ui.Widget) bool {
	startX, startY := ctx.Start()
	width, height := ctx.Size()
	childW, childH := child.Size()
	displayX, displayY := ctx.DisplayPos()

	return intersectsRect(displayX, displayY, int16(childW), int16(childH), startX, startY, int16(width), int16(height))
}

// clipSlack is how far outside its box a child may paint (a focus frame)
// and still be drawn in a band it doesn't reach itself.
const clipSlack = 4

// inClip reports whether child, at ctx's position, reaches the display's
// clip (ui.Clipper: the band a banded framebuffer holds). Only drawing is
// skipped outside it: visibility stays the container's box (childVisible),
// so a band changes no visibility. Without a Clipper everything is in.
func inClip(ctx ui.Context, child ui.Widget) bool {
	c, ok := ctx.D().(ui.Clipper)
	if !ok {
		return true
	}
	cx, cy, cw, ch := c.Clip()
	w, h := child.Size()
	x, y := ctx.DisplayPos()
	return intersectsRect(x-clipSlack, y-clipSlack, int16(w)+2*clipSlack, int16(h)+2*clipSlack, cx, cy, cw, ch)
}

func intersectsRect(x0, y0, w0, h0, x1, y1, w1, h1 int16) bool {
	r0x1 := x0 + w0
	r0y1 := y0 + h0
	r1x1 := x1 + w1
	r1y1 := y1 + h1

	return x0 < r1x1 && r0x1 > x1 && y0 < r1y1 && r0y1 > y1
}
