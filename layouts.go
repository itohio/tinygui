package ui

func LayoutHList(p int16) Layouter {
	return func(ctx Context, w Widget) bool {
		wW, _ := w.Size()
		x, y := ctx.Pos()
		x += int16(wW) + p
		return ctx.SetPos(x, y)
	}
}

func LayoutVList(p int16) Layouter {
	return func(ctx Context, w Widget) bool {
		_, wH := w.Size()
		x, y := ctx.Pos()
		y += int16(wH) + p
		return ctx.SetPos(x, y)
	}
}

func LayoutGrid(px, py int16) Layouter {
	return func(ctx Context, w Widget) bool {
		return true
	}
}
