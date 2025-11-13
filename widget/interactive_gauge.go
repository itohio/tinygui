package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

// InteractiveGauge provides editable control over a Gauge value.
type InteractiveGauge[T Number] struct {
	*Gauge[T]
	external  *T
	min       T
	max       T
	stepSmall T
	stepLarge T
	onCommit  func(T)
	enabled   bool
	selecting bool
	pending   T
	original  T
}

// InteractiveGaugeOption configures the interactive gauge.
type InteractiveGaugeOption[T Number] func(*InteractiveGauge[T])

// NewInteractiveGauge constructs an interactive gauge using layout-driven orientation.
func NewInteractiveGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveGauge[T] {
	g := &InteractiveGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}
	g.Gauge = NewGauge(width, height, &g.pending, g.min, g.max, color.RGBA{255, 255, 255, 255}, color.RGBA{})

	for _, opt := range opts {
		opt.applyToGauge(g)
	}

	g.resetDefaults()
	g.load()
	return g
}

// HorizontalInteractiveGauge is preserved for backwards compatibility.
type HorizontalInteractiveGauge[T Number] = InteractiveGauge[T]

// VerticalInteractiveGauge is preserved for backwards compatibility.
type VerticalInteractiveGauge[T Number] = InteractiveGauge[T]

// NewHorizontalInteractiveGauge wraps NewInteractiveGauge for compatibility.
func NewHorizontalInteractiveGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveGauge[T] {
	return NewInteractiveGauge(width, height, opts...)
}

// NewVerticalInteractiveGauge wraps NewInteractiveGauge for compatibility.
func NewVerticalInteractiveGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveGauge[T] {
	return NewInteractiveGauge(width, height, opts...)
}

func (g *InteractiveGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
	if g.Gauge != nil {
		g.Min = g.min
		g.Max = g.max
	}
}

func (g *InteractiveGauge[T]) load() {
	current := g.min
	if g.external != nil {
		current = *g.external
	}
	current = clamp(g.min, g.max, current)
	g.pending = current
	g.original = current
	if g.Gauge != nil {
		g.Min = g.min
		g.Max = g.max
		g.Value = &g.pending
	}
}

// Enabled reports whether the gauge accepts commands.
func (g *InteractiveGauge[T]) Enabled() bool { return g.enabled }

// SetEnabled toggles interaction ability.
func (g *InteractiveGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

// Interact handles user commands.
func (g *InteractiveGauge[T]) Interact(cmd ui.UserCommand) bool {
	if !g.enabled {
		return false
	}
	switch cmd {
	case ui.UP, ui.NEXT:
		g.adjust(g.stepSmall)
		return true
	case ui.DOWN, ui.PREV:
		g.adjust(-g.stepSmall)
		return true
	case ui.LONG_UP, ui.RIGHT:
		g.adjust(g.stepLarge)
		return true
	case ui.LONG_DOWN, ui.LEFT:
		g.adjust(-g.stepLarge)
		return true
	case ui.ENTER:
		if !g.Selected() {
			return false
		}
		g.commit()
		g.SetSelected(false)
		return true
	case ui.BACK, ui.ESC:
		if !g.Selected() {
			return false
		}
		g.revert()
		return g.Gauge.WidgetBase.Interact(cmd)
	default:
		return g.Gauge.WidgetBase.Interact(cmd)
	}
}

// OnSelect prepares for editing.
func (g *InteractiveGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.load()
}

// OnDeselect clears selection state.
func (g *InteractiveGauge[T]) OnDeselect() { g.selecting = false }

// OnExit reverts pending edits.
func (g *InteractiveGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

// SetSelected overrides selection to hook editing lifecycle.
func (g *InteractiveGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.Gauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *InteractiveGauge[T]) adjust(delta T) {
	if !g.selecting {
		return
	}
	g.pending = clamp(g.min, g.max, g.pending+delta)
}

func (g *InteractiveGauge[T]) commit() {
	g.pending = clamp(g.min, g.max, g.pending)
	g.original = g.pending
	if g.external != nil {
		*g.external = g.pending
	}
	if g.onCommit != nil {
		g.onCommit(g.pending)
	}
}

func (g *InteractiveGauge[T]) revert() {
	g.pending = g.original
}

func (g *InteractiveGauge[T]) setExternal(ptr *T) {
	g.external = ptr
}

// InteractiveMultiGauge provides editable control over MultiGauge values.
type InteractiveMultiGauge[T Number] struct {
	*MultiGauge[T]
	values      *[]T
	min         T
	max         T
	stepSmall   T
	stepLarge   T
	onCommit    func([]T)
	enabled     bool
	selecting   bool
	active      int
	pending     [maxGaugeSegments]T
	original    [maxGaugeSegments]T
	pendingView []T
}

// InteractiveMultiGaugeOption customises the multivalue gauge.
type InteractiveMultiGaugeOption[T Number] func(*InteractiveMultiGauge[T])

// NewInteractiveMultiGauge constructs a multivalue gauge.
func NewInteractiveMultiGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveMultiGauge[T] {
	g := &InteractiveMultiGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
		active:    0,
	}
	g.pendingView = g.pending[:0]
	g.MultiGauge = NewMultiGauge(width, height, g.min, g.max, &g.pendingView, nil, color.RGBA{}, color.RGBA{255, 255, 255, 255})

	for _, opt := range opts {
		opt.applyToMulti(g)
	}

	g.resetDefaults()
	g.load()
	return g
}

// HorizontalInteractiveMultiGauge is an alias for compatibility.
type HorizontalInteractiveMultiGauge[T Number] = InteractiveMultiGauge[T]

// VerticalInteractiveMultiGauge is an alias for compatibility.
type VerticalInteractiveMultiGauge[T Number] = InteractiveMultiGauge[T]

// NewHorizontalInteractiveMultiGauge wraps the new constructor.
func NewHorizontalInteractiveMultiGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveMultiGauge[T] {
	return NewInteractiveMultiGauge(width, height, opts...)
}

// NewVerticalInteractiveMultiGauge wraps the new constructor.
func NewVerticalInteractiveMultiGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveMultiGauge[T] {
	return NewInteractiveMultiGauge(width, height, opts...)
}

func (g *InteractiveMultiGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
}

func (g *InteractiveMultiGauge[T]) load() {
	n := 0
	if g.values != nil {
		for ; n < len(*g.values) && n < len(g.pending); n++ {
			v := clamp(g.min, g.max, (*g.values)[n])
			g.pending[n] = v
			g.original[n] = v
		}
	}
	g.pendingView = g.pending[:n]
	g.Values = &g.pendingView
	g.Min = g.min
	g.Max = g.max
}

// Enabled reports whether the gauge accepts commands.
func (g *InteractiveMultiGauge[T]) Enabled() bool { return g.enabled }

// SetEnabled toggles interaction ability.
func (g *InteractiveMultiGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

// Interact handles user commands across segments.
func (g *InteractiveMultiGauge[T]) Interact(cmd ui.UserCommand) bool {
	if !g.enabled {
		return false
	}
	switch cmd {
	case ui.UP, ui.NEXT:
		g.adjust(g.stepSmall)
		return true
	case ui.DOWN, ui.PREV:
		g.adjust(-g.stepSmall)
		return true
	case ui.LONG_UP:
		g.adjust(g.stepLarge)
		return true
	case ui.LONG_DOWN:
		g.adjust(-g.stepLarge)
		return true
	case ui.RIGHT:
		return g.moveActive(1)
	case ui.LEFT:
		return g.moveActive(-1)
	case ui.ENTER:
		if !g.Selected() {
			return false
		}
		if len(g.pendingView) == 0 {
			return false
		}
		g.active++
		if g.active >= len(g.pendingView) {
			g.commit()
			g.SetSelected(false)
		}
		return true
	case ui.BACK, ui.ESC:
		if !g.Selected() {
			return false
		}
		g.revert()
		g.SetSelected(false)
		return true
	default:
		return g.MultiGauge.WidgetBase.Interact(cmd)
	}
}

// OnSelect prepares for editing.
func (g *InteractiveMultiGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.active = 0
	g.load()
}

// OnDeselect clears selection state.
func (g *InteractiveMultiGauge[T]) OnDeselect() { g.selecting = false }

// OnExit reverts pending edits.
func (g *InteractiveMultiGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

// SetSelected overrides selection to hook editing lifecycle.
func (g *InteractiveMultiGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.MultiGauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *InteractiveMultiGauge[T]) adjust(delta T) {
	if !g.selecting || len(g.pendingView) == 0 {
		return
	}
	idx := clampInt(g.active, 0, len(g.pendingView)-1)
	g.pendingView[idx] = clamp(g.min, g.max, g.pendingView[idx]+delta)
}

func (g *InteractiveMultiGauge[T]) moveActive(delta int) bool {
	if !g.selecting || len(g.pendingView) == 0 {
		return false
	}
	next := (g.active + delta + len(g.pendingView)) % len(g.pendingView)
	if next == g.active {
		return false
	}
	g.active = next
	return true
}

func (g *InteractiveMultiGauge[T]) commit() {
	if g.values != nil {
		for i := range g.pendingView {
			(*g.values)[i] = g.pendingView[i]
		}
	}
	if g.onCommit != nil {
		g.onCommit(g.pendingView)
	}
	copy(g.original[:], g.pending[:])
	g.selecting = false
}

func (g *InteractiveMultiGauge[T]) revert() {
	copy(g.pending[:], g.original[:])
	n := len(g.pendingView)
	g.pendingView = g.pending[:n]
}

func (g *InteractiveMultiGauge[T]) setValues(values *[]T) {
	g.values = values
}
