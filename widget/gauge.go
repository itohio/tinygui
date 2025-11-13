package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// Gauge renders a progress indicator whose orientation is inferred from its geometry.
type Gauge[T Number] struct {
	ui.WidgetBase
	Value      *T
	Min        T
	Max        T
	Foreground color.RGBA
	Background color.RGBA
}

// NewGauge constructs a gauge with the provided geometry and colours.
func NewGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *Gauge[T] {
	if min > max {
		min, max = max, min
	}
	return &Gauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Value:      value,
		Min:        min,
		Max:        max,
		Foreground: fg,
		Background: bg,
	}
}

// HorizontalGauge is kept for backwards compatibility and aliases Gauge.
type HorizontalGauge[T Number] = Gauge[T]

// VerticalGauge is kept for backwards compatibility and aliases Gauge.
type VerticalGauge[T Number] = Gauge[T]

// NewHorizontalGauge constructs a gauge using horizontal defaults.
func NewHorizontalGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *Gauge[T] {
	return NewGauge(width, height, value, min, max, fg, bg)
}

// NewVerticalGauge constructs a gauge using vertical defaults.
func NewVerticalGauge[T Number](width, height uint16, value *T, min, max T, fg, bg color.RGBA) *Gauge[T] {
	return NewGauge(width, height, value, min, max, fg, bg)
}

// Draw renders the gauge using horizontal or vertical form based on widget dimensions.
func (g *Gauge[T]) Draw(ctx ui.Context) {
	if g == nil || g.Value == nil {
		return
	}
	if g.Height == 0 || g.Width >= g.Height {
		g.drawHorizontal(ctx)
		return
	}
	g.drawVertical(ctx)
}

func (g *Gauge[T]) drawHorizontal(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
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

func (g *Gauge[T]) drawVertical(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
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

// MultiGauge renders a gauge with multiple values represented as coloured segments.
type MultiGauge[T Number] struct {
	ui.WidgetBase
	Values     *[]T
	Min        T
	Max        T
	Colors     []color.RGBA
	Background color.RGBA
	Foreground color.RGBA
}

// NewMultiGauge constructs a multivalue gauge.
func NewMultiGauge[T Number](width, height uint16, min, max T, values *[]T, colors []color.RGBA, background, foreground color.RGBA) *MultiGauge[T] {
	if min > max {
		min, max = max, min
	}
	return &MultiGauge[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		Values:     values,
		Min:        min,
		Max:        max,
		Colors:     colors,
		Background: background,
		Foreground: foreground,
	}
}

// HorizontalMultiGauge and VerticalMultiGauge are retained as aliases for compatibility.
type HorizontalMultiGauge[T Number] = MultiGauge[T]
type VerticalMultiGauge[T Number] = MultiGauge[T]

// NewHorizontalMultiGauge constructs a multivalue gauge sized for horizontal layouts.
func NewHorizontalMultiGauge[T Number](width, height uint16, min, max T, values *[]T, colors []color.RGBA, background, foreground color.RGBA) *MultiGauge[T] {
	return NewMultiGauge(width, height, min, max, values, colors, background, foreground)
}

// NewVerticalMultiGauge constructs a multivalue gauge sized for vertical layouts.
func NewVerticalMultiGauge[T Number](width, height uint16, min, max T, values *[]T, colors []color.RGBA, background, foreground color.RGBA) *MultiGauge[T] {
	return NewMultiGauge(width, height, min, max, values, colors, background, foreground)
}

// Draw renders the multivalue gauge based on orientation.
func (g *MultiGauge[T]) Draw(ctx ui.Context) {
	if g == nil || g.Values == nil || *g.Values == nil {
		return
	}
	if g.Height == 0 || g.Width >= g.Height {
		g.drawHorizontal(ctx)
		return
	}
	g.drawVertical(ctx)
}

func (g *MultiGauge[T]) drawHorizontal(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
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

func (g *MultiGauge[T]) drawVertical(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
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
	for i, v := range vals {
		segVal := clamp(g.Min, g.Max, v)
		segHeight := position(g.Min, g.Max, segVal, height-2)
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

func isZeroColor(c color.RGBA) bool {
	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0
}

func position[T Number](min, max, value T, span int16) int16 {
	if span <= 0 {
		return 0
	}
	if max <= min {
		return span
	}
	fraction := (float32(value) - float32(min)) / (float32(max) - float32(min))
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return int16(fraction * float32(span))
}
