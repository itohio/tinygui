package widget

import (
	"image/color"

	"tinygo.org/x/tinyfont"
)

// InteractiveOption configures interactive widgets (labels and gauges).
type InteractiveOption[T Number] struct {
	applyLabel func(*InteractiveLabel[T])
	applyGauge func(*InteractiveGauge[T])
	applyMulti func(*InteractiveMultiGauge[T])
}

func (o InteractiveOption[T]) applyToLabel(target *InteractiveLabel[T]) {
	if o.applyLabel != nil {
		o.applyLabel(target)
	}
}

func (o InteractiveOption[T]) applyToGauge(target *InteractiveGauge[T]) {
	if o.applyGauge != nil {
		o.applyGauge(target)
	}
}

func (o InteractiveOption[T]) applyToMulti(target *InteractiveMultiGauge[T]) {
	if o.applyMulti != nil {
		o.applyMulti(target)
	}
}

// WithValue wires a pointer that will be updated on commit. Applies to label and single-value gauges.
func WithValue[T Number](value *T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyLabel: func(l *InteractiveLabel[T]) {
			l.setExternal(value)
		},
		applyGauge: func(g *InteractiveGauge[T]) {
			g.setExternal(value)
		},
	}
}

// WithValues wires a slice pointer used by multi-value gauges.
func WithValues[T Number](values *[]T) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyMulti: func(g *InteractiveMultiGauge[T]) {
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
		applyGauge: func(g *InteractiveGauge[T]) {
			g.min = min
			g.max = max
			if g.min > g.max {
				g.min, g.max = g.max, g.min
			}
		},
		applyMulti: func(g *InteractiveMultiGauge[T]) {
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
		applyGauge: func(g *InteractiveGauge[T]) {
			g.stepSmall = small
			g.stepLarge = large
		},
		applyMulti: func(g *InteractiveMultiGauge[T]) {
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
		applyGauge: func(g *InteractiveGauge[T]) {
			g.onCommit = fn
		},
	}
}

// WithMultiCommit registers a commit callback for multi-value gauges.
func WithMultiCommit[T Number](fn func([]T)) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyMulti: func(g *InteractiveMultiGauge[T]) {
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
		applyGauge: func(g *InteractiveGauge[T]) {
			g.enabled = false
		},
		applyMulti: func(g *InteractiveMultiGauge[T]) {
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
		applyGauge: func(g *InteractiveGauge[T]) {
			g.Foreground = c
		},
		applyMulti: func(g *InteractiveMultiGauge[T]) {
			g.Foreground = c
		},
	}
}

// WithBackground sets gauge background colour.
func WithBackground[T Number](c color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyGauge: func(g *InteractiveGauge[T]) {
			g.Background = c
		},
		applyMulti: func(g *InteractiveMultiGauge[T]) {
			g.Background = c
		},
	}
}

// WithSegmentColors sets per-segment colours for multi-value gauges.
func WithSegmentColors[T Number](colors []color.RGBA) InteractiveOption[T] {
	return InteractiveOption[T]{
		applyMulti: func(g *InteractiveMultiGauge[T]) {
			g.Colors = colors
		},
	}
}
