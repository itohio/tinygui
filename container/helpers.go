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

func intersectsRect(x0, y0, w0, h0, x1, y1, w1, h1 int16) bool {
	r0x1 := x0 + w0
	r0y1 := y0 + h0
	r1x1 := x1 + w1
	r1y1 := y1 + h1

	return x0 < r1x1 && r0x1 > x1 && y0 < r1y1 && r0y1 > y1
}
