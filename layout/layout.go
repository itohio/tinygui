// Package layout provides generic strategies to lay out objects that implement the Sizer interface.
// Layouter is a general-purpose arrangement utility that composes objects in a desired order.
// Layout strategies work with any object providing a Size method, enabling use beyond widgets.
// Layouts operate on a ui.Context to manage child origins and sizes; if the context or the laid-out
// object implements Marginer or Padder, the strategy will respect margin and padding values.
// This decouples spatial arrangement from specific widget implementations, allowing reusable flow/grid patterns.
package layout

import ui "github.com/itohio/tinygui"

// Strategy adjusts a drawing context between child render calls.
type Strategy func(ctx ui.Context, w ui.Sizer) bool

// HList arranges widgets horizontally with padding p between entries.
func HList(p int16) Strategy {
	return func(ctx ui.Context, w ui.Sizer) bool {
		wW, _ := w.Size()
		x, y := ctx.Pos()
		x += int16(wW) + p
		return ctx.SetPos(x, y)
	}
}

// VList arranges widgets vertically with padding p between entries.
func VList(p int16) Strategy {
	return func(ctx ui.Context, w ui.Sizer) bool {
		_, wH := w.Size()
		x, y := ctx.Pos()
		y += int16(wH) + p
		return ctx.SetPos(x, y)
	}
}

// Grid wraps widgets into rows and columns separated by px/py padding.
func Grid(px, py int16) Strategy {
	var (
		startX      int16
		rowHeight   int16
		initialized bool
	)

	return func(ctx ui.Context, w ui.Sizer) bool {
		if !initialized {
			x, _ := ctx.Pos()
			startX = x
			initialized = true
			rowHeight = 0
		}

		wW, wH := w.Size()
		x, y := ctx.Pos()
		ctxW, _ := ctx.Size()

		if h := int16(wH); h > rowHeight {
			rowHeight = h
		}

		nextX := x + int16(wW) + px
		nextY := y

		if ctxW != 0 && nextX-startX > int16(ctxW) {
			nextY = y + rowHeight + py
			nextX = startX
			rowHeight = 0
		}

		return ctx.SetPos(nextX, nextY)
	}
}

// HFlow lays out widgets left-to-right, wrapping to a new line when needed.
func HFlow(spacing int16, maxWidth uint16) Strategy {
	var (
		rowHeight int16
		startX    int16
		ready     bool
	)

	return func(ctx ui.Context, w ui.Sizer) bool {
		if !ready {
			x, _ := ctx.Pos()
			startX = x
			ready = true
			rowHeight = 0
		}

		wW, wH := w.Size()
		x, y := ctx.Pos()
		limit := int16(maxWidth)
		if limit == 0 {
			ctxW, _ := ctx.Size()
			limit = int16(ctxW)
		}

		if h := int16(wH); h > rowHeight {
			rowHeight = h
		}

		nextX := x + int16(wW) + spacing
		nextY := y

		if limit > 0 && nextX-startX >= limit {
			nextY = y + rowHeight + spacing
			nextX = startX
			rowHeight = 0
		}

		return ctx.SetPos(nextX, nextY)
	}
}

// VFlow lays out widgets top-to-bottom, wrapping to a new column when needed.
func VFlow(spacing int16, maxHeight uint16) Strategy {
	var (
		columnWidth int16
		startY      int16
		ready       bool
	)

	return func(ctx ui.Context, w ui.Sizer) bool {
		if !ready {
			_, y := ctx.Pos()
			startY = y
			ready = true
			columnWidth = 0
		}

		wW, wH := w.Size()
		x, y := ctx.Pos()
		limit := int16(maxHeight)
		if limit == 0 {
			_, ctxH := ctx.Size()
			limit = int16(ctxH)
		}

		if w := int16(wW); w > columnWidth {
			columnWidth = w
		}

		nextX := x
		nextY := y + int16(wH) + spacing

		if limit > 0 && nextY-startY >= limit {
			nextX = x + columnWidth + spacing
			nextY = startY
			columnWidth = 0
		}

		return ctx.SetPos(nextX, nextY)
	}
}
