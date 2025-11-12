package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
)

type HorizontalInteractiveGauge[T Number] struct {
	HorizontalGauge[T]
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

type VerticalInteractiveGauge[T Number] struct {
	VerticalGauge[T]
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

type HorizontalInteractiveMultiGauge[T Number] struct {
	HorizontalMultiGauge[T]
	values      *[]T
	min         T
	max         T
	stepSmall   T
	stepLarge   T
	onCommit    func([]T)
	enabled     bool
	selecting   bool
	active      int
	count       int
	pending     [maxGaugeSegments]T
	original    [maxGaugeSegments]T
	pendingView []T
}

type VerticalInteractiveMultiGauge[T Number] struct {
	VerticalMultiGauge[T]
	values      *[]T
	min         T
	max         T
	stepSmall   T
	stepLarge   T
	onCommit    func([]T)
	enabled     bool
	selecting   bool
	active      int
	count       int
	pending     [maxGaugeSegments]T
	original    [maxGaugeSegments]T
	pendingView []T
}

func NewHorizontalInteractiveGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *HorizontalInteractiveGauge[T] {
	g := &HorizontalInteractiveGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}
	g.pending = 0
	g.HorizontalGauge = *NewHorizontalGauge(width, height, &g.pending, g.min, g.max, color.RGBA{255, 255, 255, 255}, color.RGBA{})

	for _, opt := range opts {
		opt.applyToHorizontalGauge(g)
	}
	g.resetDefaults()
	g.load()
	return g
}

func NewVerticalInteractiveGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *VerticalInteractiveGauge[T] {
	g := &VerticalInteractiveGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}
	g.pending = 0
	g.VerticalGauge = *NewVerticalGauge(width, height, &g.pending, g.min, g.max, color.RGBA{255, 255, 255, 255}, color.RGBA{})

	for _, opt := range opts {
		opt.applyToVerticalGauge(g)
	}
	g.resetDefaults()
	g.load()
	return g
}

func NewHorizontalInteractiveMultiGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *HorizontalInteractiveMultiGauge[T] {
	g := &HorizontalInteractiveMultiGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}
	g.pendingView = g.pending[:0]
	g.HorizontalMultiGauge = *NewHorizontalMultiGauge(width, height, g.min, g.max, &g.pendingView, nil, color.RGBA{}, color.RGBA{255, 255, 255, 255})

	for _, opt := range opts {
		opt.applyToHorizontalMultiGauge(g)
	}
	g.resetDefaults()
	g.load()
	return g
}

func NewVerticalInteractiveMultiGauge[T Number](width, height uint16, opts ...InteractiveOption[T]) *VerticalInteractiveMultiGauge[T] {
	g := &VerticalInteractiveMultiGauge[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}
	g.pendingView = g.pending[:0]
	g.VerticalMultiGauge = *NewVerticalMultiGauge(width, height, g.min, g.max, &g.pendingView, nil, color.RGBA{}, color.RGBA{255, 255, 255, 255})

	for _, opt := range opts {
		opt.applyToVerticalMultiGauge(g)
	}
	g.resetDefaults()
	g.load()
	return g
}

func (g *HorizontalInteractiveGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
	g.HorizontalGauge.Min = g.min
	g.HorizontalGauge.Max = g.max
	g.HorizontalGauge.Value = &g.pending
}

func (g *VerticalInteractiveGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
	g.VerticalGauge.Min = g.min
	g.VerticalGauge.Max = g.max
	g.VerticalGauge.Value = &g.pending
}

func (g *HorizontalInteractiveMultiGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
	g.HorizontalMultiGauge.Min = g.min
	g.HorizontalMultiGauge.Max = g.max
}

func (g *VerticalInteractiveMultiGauge[T]) resetDefaults() {
	if g.stepSmall == 0 {
		g.stepSmall = 1
	}
	if g.stepLarge == 0 {
		g.stepLarge = g.stepSmall
	}
	if g.min > g.max {
		g.min, g.max = g.max, g.min
	}
	g.VerticalMultiGauge.Min = g.min
	g.VerticalMultiGauge.Max = g.max
}

func (g *HorizontalInteractiveGauge[T]) load() {
	var current T
	if g.external != nil {
		current = *g.external
	}
	current = clamp(g.min, g.max, current)
	g.pending = current
	g.original = current
}

func (g *VerticalInteractiveGauge[T]) load() {
	var current T
	if g.external != nil {
		current = *g.external
	}
	current = clamp(g.min, g.max, current)
	g.pending = current
	g.original = current
}

func (g *HorizontalInteractiveMultiGauge[T]) load() {
	g.count = 0
	if g.values != nil && *g.values != nil {
		src := *g.values
		if len(src) > maxGaugeSegments {
			src = src[:maxGaugeSegments]
		}
		for i := range src {
			g.pending[i] = clamp(g.min, g.max, src[i])
			g.original[i] = g.pending[i]
		}
		g.count = len(src)
	}
	g.pendingView = g.pending[:g.count]
	g.HorizontalMultiGauge.Values = &g.pendingView
	g.active = 0
}

func (g *VerticalInteractiveMultiGauge[T]) load() {
	g.count = 0
	if g.values != nil && *g.values != nil {
		src := *g.values
		if len(src) > maxGaugeSegments {
			src = src[:maxGaugeSegments]
		}
		for i := range src {
			g.pending[i] = clamp(g.min, g.max, src[i])
			g.original[i] = g.pending[i]
		}
		g.count = len(src)
	}
	g.pendingView = g.pending[:g.count]
	g.VerticalMultiGauge.Values = &g.pendingView
	g.active = 0
}

func (g *HorizontalInteractiveGauge[T]) Enabled() bool { return g.enabled }
func (g *VerticalInteractiveGauge[T]) Enabled() bool   { return g.enabled }
func (g *HorizontalInteractiveMultiGauge[T]) Enabled() bool {
	return g.enabled
}
func (g *VerticalInteractiveMultiGauge[T]) Enabled() bool {
	return g.enabled
}

func (g *HorizontalInteractiveGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

func (g *VerticalInteractiveGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

func (g *HorizontalInteractiveMultiGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

func (g *VerticalInteractiveMultiGauge[T]) SetEnabled(v bool) {
	g.enabled = v
	if !v {
		g.SetSelected(false)
	}
}

func (g *HorizontalInteractiveGauge[T]) Interact(cmd ui.UserCommand) bool {
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
		return g.HorizontalGauge.WidgetBase.Interact(cmd)
	default:
		return g.HorizontalGauge.WidgetBase.Interact(cmd)
	}
}

func (g *VerticalInteractiveGauge[T]) Interact(cmd ui.UserCommand) bool {
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
		return g.VerticalGauge.WidgetBase.Interact(cmd)
	default:
		return g.VerticalGauge.WidgetBase.Interact(cmd)
	}
}

func (g *HorizontalInteractiveMultiGauge[T]) Interact(cmd ui.UserCommand) bool {
	if !g.enabled {
		return false
	}
	if g.count == 0 {
		return g.HorizontalMultiGauge.WidgetBase.Interact(cmd)
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
		if g.active+1 < g.count {
			g.active++
			return true
		}
		g.adjust(g.stepLarge)
		return true
	case ui.LEFT:
		if g.active > 0 {
			g.active--
			return true
		}
		g.adjust(-g.stepLarge)
		return true
	case ui.ENTER:
		if !g.Selected() {
			return false
		}
		if g.active+1 < g.count {
			g.active++
			return true
		}
		g.commit()
		g.SetSelected(false)
		return true
	case ui.BACK:
		if !g.Selected() {
			return false
		}
		if g.active > 0 {
			g.active--
			return true
		}
		g.revert()
		return true
	case ui.ESC:
		if !g.Selected() {
			return false
		}
		g.revert()
		return g.HorizontalMultiGauge.WidgetBase.Interact(cmd)
	default:
		return g.HorizontalMultiGauge.WidgetBase.Interact(cmd)
	}
}

func (g *VerticalInteractiveMultiGauge[T]) Interact(cmd ui.UserCommand) bool {
	if !g.enabled {
		return false
	}
	if g.count == 0 {
		return g.VerticalMultiGauge.WidgetBase.Interact(cmd)
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
		if g.active+1 < g.count {
			g.active++
			return true
		}
		g.adjust(g.stepLarge)
		return true
	case ui.LEFT:
		if g.active > 0 {
			g.active--
			return true
		}
		g.adjust(-g.stepLarge)
		return true
	case ui.ENTER:
		if !g.Selected() {
			return false
		}
		if g.active+1 < g.count {
			g.active++
			return true
		}
		g.commit()
		g.SetSelected(false)
		return true
	case ui.BACK:
		if !g.Selected() {
			return false
		}
		if g.active > 0 {
			g.active--
			return true
		}
		g.revert()
		return true
	case ui.ESC:
		if !g.Selected() {
			return false
		}
		g.revert()
		return g.VerticalMultiGauge.WidgetBase.Interact(cmd)
	default:
		return g.VerticalMultiGauge.WidgetBase.Interact(cmd)
	}
}

func (g *HorizontalInteractiveGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.load()
}

func (g *VerticalInteractiveGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.load()
}

func (g *HorizontalInteractiveMultiGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.load()
}

func (g *VerticalInteractiveMultiGauge[T]) OnSelect() {
	if !g.enabled {
		return
	}
	g.selecting = true
	g.load()
}

func (g *HorizontalInteractiveGauge[T]) OnDeselect()      { g.selecting = false }
func (g *VerticalInteractiveGauge[T]) OnDeselect()        { g.selecting = false }
func (g *HorizontalInteractiveMultiGauge[T]) OnDeselect() { g.selecting = false }
func (g *VerticalInteractiveMultiGauge[T]) OnDeselect()   { g.selecting = false }

func (g *HorizontalInteractiveGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

func (g *VerticalInteractiveGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

func (g *HorizontalInteractiveMultiGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

func (g *VerticalInteractiveMultiGauge[T]) OnExit() {
	g.revert()
	g.selecting = false
}

func (g *HorizontalInteractiveGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.HorizontalGauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *VerticalInteractiveGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.VerticalGauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *HorizontalInteractiveMultiGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.HorizontalMultiGauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *VerticalInteractiveMultiGauge[T]) SetSelected(sel bool) {
	prev := g.Selected()
	g.VerticalMultiGauge.SetSelected(sel)
	if sel && !prev {
		g.OnSelect()
	}
	if !sel && prev {
		g.OnDeselect()
	}
}

func (g *HorizontalInteractiveGauge[T]) adjust(delta T) {
	if !g.selecting {
		return
	}
	g.pending = clamp(g.min, g.max, g.pending+delta)
}

func (g *VerticalInteractiveGauge[T]) adjust(delta T) {
	if !g.selecting {
		return
	}
	g.pending = clamp(g.min, g.max, g.pending+delta)
}

func (g *HorizontalInteractiveMultiGauge[T]) adjust(delta T) {
	if !g.selecting || g.count == 0 {
		return
	}
	value := clamp(g.min, g.max, g.pending[g.active]+delta)
	g.pending[g.active] = value
}

func (g *VerticalInteractiveMultiGauge[T]) adjust(delta T) {
	if !g.selecting || g.count == 0 {
		return
	}
	value := clamp(g.min, g.max, g.pending[g.active]+delta)
	g.pending[g.active] = value
}

func (g *HorizontalInteractiveGauge[T]) commit() {
	g.pending = clamp(g.min, g.max, g.pending)
	g.original = g.pending
	if g.external != nil {
		*g.external = g.pending
	}
	if g.onCommit != nil {
		g.onCommit(g.pending)
	}
}

func (g *VerticalInteractiveGauge[T]) commit() {
	g.pending = clamp(g.min, g.max, g.pending)
	g.original = g.pending
	if g.external != nil {
		*g.external = g.pending
	}
	if g.onCommit != nil {
		g.onCommit(g.pending)
	}
}

func (g *HorizontalInteractiveMultiGauge[T]) commit() {
	for i := 0; i < g.count; i++ {
		value := clamp(g.min, g.max, g.pending[i])
		g.pending[i] = value
		g.original[i] = value
		if g.values != nil && *g.values != nil && i < len(*g.values) {
			(*g.values)[i] = value
		}
	}
	if g.onCommit != nil {
		g.onCommit(g.pendingView)
	}
	g.active = 0
}

func (g *VerticalInteractiveMultiGauge[T]) commit() {
	for i := 0; i < g.count; i++ {
		value := clamp(g.min, g.max, g.pending[i])
		g.pending[i] = value
		g.original[i] = value
		if g.values != nil && *g.values != nil && i < len(*g.values) {
			(*g.values)[i] = value
		}
	}
	if g.onCommit != nil {
		g.onCommit(g.pendingView)
	}
	g.active = 0
}

func (g *HorizontalInteractiveGauge[T]) revert() {
	g.pending = g.original
}

func (g *VerticalInteractiveGauge[T]) revert() {
	g.pending = g.original
}

func (g *HorizontalInteractiveMultiGauge[T]) revert() {
	for i := 0; i < g.count; i++ {
		g.pending[i] = g.original[i]
	}
	g.active = 0
}

func (g *VerticalInteractiveMultiGauge[T]) revert() {
	for i := 0; i < g.count; i++ {
		g.pending[i] = g.original[i]
	}
	g.active = 0
}

func (g *HorizontalInteractiveGauge[T]) setExternal(ptr *T) {
	g.external = ptr
}

func (g *VerticalInteractiveGauge[T]) setExternal(ptr *T) {
	g.external = ptr
}

func (g *HorizontalInteractiveMultiGauge[T]) setValues(values *[]T) {
	g.values = values
}

func (g *VerticalInteractiveMultiGauge[T]) setValues(values *[]T) {
	g.values = values
}

func position[T Number](min, max, value T, span int16) int16 {
	if span <= 0 {
		return 0
	}
	val := clamp(min, max, value)
	num := float64(val - min)
	den := float64(max - min)
	if den == 0 {
		return 0
	}
	ratio := num / den
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return int16(ratio*float64(span) + 0.5)
}
