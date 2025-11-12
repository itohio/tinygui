package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// HorizontalGauge renders a horizontal progress indicator.
type HorizontalGauge[T Number] struct {
	ui.WidgetBase
	Value      *T
	Min        T
	Max        T
	Foreground color.RGBA
	Background color.RGBA
}

// VerticalGauge renders a vertical progress indicator.
type VerticalGauge[T Number] struct {
	ui.WidgetBase
	Value      *T
	Min        T
	Max        T
	Foreground color.RGBA
	Background color.RGBA
}

// NewHorizontalGauge constructs a horizontal gauge.
func NewHorizontalGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *HorizontalGauge[T] {
	if min > max {
		min, max = max, min
	}
	return &HorizontalGauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Value:      value,
		Min:        min,
		Max:        max,
		Foreground: fg,
		Background: bg,
	}
}

// NewVerticalGauge constructs a vertical gauge.
func NewVerticalGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *VerticalGauge[T] {
	if min > max {
		min, max = max, min
	}
	return &VerticalGauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Value:      value,
		Min:        min,
		Max:        max,
		Foreground: fg,
		Background: bg,
	}
}

// Draw renders the horizontal gauge based on the current value.
func (g *HorizontalGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Value == nil {
		return
	}
	val := clamp(g.Min, g.Max, *g.Value)
	x, y := ctx.DisplayPos()
	width := int16(g.Width)
	height := int16(g.Height)

	fill := position(g.Min, g.Max, val, width-2)
	if !isZeroColor(g.Background) {
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x, y, width, height, g.Background)
		} else {
			for dy := int16(0); dy < height; dy++ {
				ui.HLine(d, x, y+dy, width, g.Background)
			}
		}
	}
	if fast, ok := d.(ui.RectangleDisplayer); ok {
		_ = fast.FillRectangle(x+1, y+1, fill, height-2, g.Foreground)
	} else {
		for dy := int16(0); dy < height-2; dy++ {
			ui.HLine(d, x+1, y+1+dy, fill, g.Foreground)
		}
	}
}

// Draw renders the vertical gauge based on the current value.
func (g *VerticalGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Value == nil {
		return
	}
	val := clamp(g.Min, g.Max, *g.Value)
	x, y := ctx.DisplayPos()
	width := int16(g.Width)
	height := int16(g.Height)

	fill := position(g.Min, g.Max, val, height-2)
	if !isZeroColor(g.Background) {
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x, y, width, height, g.Background)
		} else {
			for dx := int16(0); dx < width; dx++ {
				ui.VLine(d, x+dx, y, height, g.Background)
			}
		}
	}
	if fast, ok := d.(ui.RectangleDisplayer); ok {
		_ = fast.FillRectangle(x+1, y+height-1-fill, width-2, fill, g.Foreground)
	} else {
		for dx := int16(0); dx < width-2; dx++ {
			ui.VLine(d, x+1+dx, y+height-1-fill, fill, g.Foreground)
		}
	}
}

func isZeroColor(c color.RGBA) bool {
	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0
}

// HorizontalMultiGauge draws a horizontal gauge with multiple values represented as colored segments.
// T is the value type.
type HorizontalMultiGauge[T Number] struct {
	ui.WidgetBase
	Values     *[]T         // slice of values, each representing a segment
	Min        T            // minimum value for scale
	Max        T            // maximum value for scale
	Colors     []color.RGBA // segment colors, fallback to Foreground if not enough provided
	Background color.RGBA
	Foreground color.RGBA
}

// NewHorizontalMultiGauge creates a horizontal gauge for multiple values.
func NewHorizontalMultiGauge[T Number](width, height uint16, min, max T, values *[]T, colors []color.RGBA, background, foreground color.RGBA) *HorizontalMultiGauge[T] {
	return &HorizontalMultiGauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Values:     values,
		Min:        min,
		Max:        max,
		Colors:     colors,
		Background: background,
		Foreground: foreground,
	}
}

// Draw renders the horizontal multivalue gauge.
func (g *HorizontalMultiGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Values == nil || *g.Values == nil {
		return
	}
	x, y := ctx.DisplayPos()
	width := int16(g.Width)
	height := int16(g.Height)

	if !isZeroColor(g.Background) {
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x, y, width, height, g.Background)
		} else {
			for dy := int16(0); dy < height; dy++ {
				ui.HLine(d, x, y+dy, width, g.Background)
			}
		}
	}

	vals := *g.Values
	cumulative := int16(0)
	for i, v := range vals {
		segVal := clamp(g.Min, g.Max, v)
		segWidth := position(g.Min, g.Max, segVal, width-2)
		if segWidth <= 0 {
			continue
		}
		segColor := g.Foreground
		if i < len(g.Colors) {
			segColor = g.Colors[i]
		}
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x+1+cumulative, y+1, segWidth, height-2, segColor)
		} else {
			for dy := int16(0); dy < height-2; dy++ {
				ui.HLine(d, x+1+cumulative, y+1+dy, segWidth, segColor)
			}
		}
		cumulative += segWidth
		if cumulative >= width-2 {
			break
		}
	}
}

// VerticalMultiGauge draws a vertical gauge with multiple values represented as colored segments.
// T is the value type.
type VerticalMultiGauge[T Number] struct {
	ui.WidgetBase
	Values     *[]T
	Min        T
	Max        T
	Colors     []color.RGBA
	Background color.RGBA
	Foreground color.RGBA
}

// NewVerticalMultiGauge creates a vertical gauge for multiple values.
func NewVerticalMultiGauge[T Number](width, height uint16, min, max T, values *[]T, colors []color.RGBA, background, foreground color.RGBA) *VerticalMultiGauge[T] {
	return &VerticalMultiGauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Values:     values,
		Min:        min,
		Max:        max,
		Colors:     colors,
		Background: background,
		Foreground: foreground,
	}
}

// Draw renders the vertical multivalue gauge.
func (g *VerticalMultiGauge[T]) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil || g.Values == nil || *g.Values == nil {
		return
	}
	x, y := ctx.DisplayPos()
	width := int16(g.Width)
	height := int16(g.Height)

	if !isZeroColor(g.Background) {
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x, y, width, height, g.Background)
		} else {
			for dx := int16(0); dx < width; dx++ {
				ui.VLine(d, x+dx, y, height, g.Background)
			}
		}
	}

	vals := *g.Values
	cumulative := int16(0)
	for i := len(vals) - 1; i >= 0; i-- { // Draw from bottom up
		v := clamp(g.Min, g.Max, vals[i])
		segHeight := position(g.Min, g.Max, v, height-2)
		if segHeight <= 0 {
			continue
		}
		segColor := g.Foreground
		if i < len(g.Colors) {
			segColor = g.Colors[i]
		}
		if fast, ok := d.(ui.RectangleDisplayer); ok {
			_ = fast.FillRectangle(x+1, y+height-1-cumulative-segHeight, width-2, segHeight, segColor)
		} else {
			for dx := int16(0); dx < width-2; dx++ {
				ui.VLine(d, x+1+dx, y+height-1-cumulative-segHeight, segHeight, segColor)
			}
		}
		cumulative += segHeight
		if cumulative >= height-2 {
			break
		}
	}
}
