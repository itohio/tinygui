package widget

import (
	"image/color"

	"tinygo.org/x/tinyfont"
)

// InteractiveOption configures interactive widgets (labels and gauges).
type InteractiveOption[T Number] struct {
	applyLabel  func(*InteractiveLabel[T])
	applyHGauge func(*HorizontalInteractiveGauge[T])
	applyVGauge func(*VerticalInteractiveGauge[T])
	applyHMulti func(*HorizontalInteractiveMultiGauge[T])
	applyVMulti func(*VerticalInteractiveMultiGauge[T])
}

func (o InteractiveOption[T]) applyToLabel(target *InteractiveLabel[T]) {
	if o.applyLabel != nil {
		o.applyLabel(target)
	}
}

func (o InteractiveOption[T]) applyToHorizontalGauge(target *HorizontalInteractiveGauge[T]) {
	if o.applyHGauge != nil {
		o.applyHGauge(target)
	}
}

func (o InteractiveOption[T]) applyToVerticalGauge(target *VerticalInteractiveGauge[T]) {
	if o.applyVGauge != nil {
		o.applyVGauge(target)
	}
}

func (o InteractiveOption[T]) applyToHorizontalMultiGauge(target *HorizontalInteractiveMultiGauge[T]) {
	if o.applyHMulti != nil {
		o.applyHMulti(target)
	}
}

func (o InteractiveOption[T]) applyToVerticalMultiGauge(target *VerticalInteractiveMultiGauge[T]) {
	if o.applyVMulti != nil {
		o.applyVMulti(target)
	}
}

// WithValue wires a pointer that will be updated on commit. Applies to label and single-value gauges.
func WithValue[T Number](value *T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.setExternal(value)
		},
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.setExternal(value)
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.setExternal(value)
		},
	}
}

// WithValues wires a slice pointer used by multi-value gauges.
func WithValues[T Number](values *[]T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.setValues(values)
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.setValues(values)
		},
	}
}

// WithRange configures min/max clamps.
func WithRange[T Number](min, max T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.min = min
			l.max = max
			if l.min > l.max {
				l.min, l.max = l.max, l.min
			}
		},
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.min = min
			g.max = max
			if g.min > g.max {
				g.min, g.max = g.max, g.min
			}
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.min = min
			g.max = max
			if g.min > g.max {
				g.min, g.max = g.max, g.min
			}
		},
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.min = min
			g.max = max
			if g.min > g.max {
				g.min, g.max = g.max, g.min
			}
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.min = min
			g.max = max
			if g.min > g.max {
				g.min, g.max = g.max, g.min
			}
		},
	}
}

// WithSteps configures small and long-step deltas.
func WithSteps[T Number](small, large T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.stepSmall = small
			l.stepLarge = large
		},
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.stepSmall = small
			g.stepLarge = large
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.stepSmall = small
			g.stepLarge = large
		},
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.stepSmall = small
			g.stepLarge = large
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.stepSmall = small
			g.stepLarge = large
		},
	}
}

// WithFormatter overrides the label formatter.
func WithFormatter[T Number](formatter func(T) string) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.formatter = formatter
		},
	}
}

// WithCommit registers a commit callback for single-value widgets.
func WithCommit[T Number](fn func(T)) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.onCommit = fn
		},
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.onCommit = fn
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.onCommit = fn
		},
	}
}

// WithMultiCommit registers a commit callback for multi-value gauges.
func WithMultiCommit[T Number](fn func([]T)) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.onCommit = fn
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.onCommit = fn
		},
	}
}

// WithDisabled sets the initial enabled state.
func WithDisabled[T Number]() InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.enabled = false
		},
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.enabled = false
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.enabled = false
		},
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.enabled = false
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.enabled = false
		},
	}
}

// WithFont assigns the font used by Interactive labels.
func WithFont[T Number](font tinyfont.Fonter) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.SetFont(font)
		},
	}
}

// WithTextColor assigns the label text colour.
func WithTextColor[T Number](c color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.SetColor(c)
		},
	}
}

// WithForeground sets gauge foreground colour.
func WithForeground[T Number](c color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.Foreground = c
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.Foreground = c
		},
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.Foreground = c
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.Foreground = c
		},
	}
}

// WithBackground sets gauge background colour.
func WithBackground[T Number](c color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyHGauge: func(g *HorizontalInteractiveGauge[T]) {
			g.Background = c
		},
		applyVGauge: func(g *VerticalInteractiveGauge[T]) {
			g.Background = c
		},
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.Background = c
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.Background = c
		},
	}
}

// WithSegmentColors sets per-segment colours for multi-value gauges.
func WithSegmentColors[T Number](colors []color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyHMulti: func(g *HorizontalInteractiveMultiGauge[T]) {
			g.Colors = colors
		},
		applyVMulti: func(g *VerticalInteractiveMultiGauge[T]) {
			g.Colors = colors
		},
	}
}
