package container

import (
	"time"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
)

// Base provides common container behaviour for slices of widgets.
type Base[T ui.Widget] struct {
	ui.WidgetBase
	layouter layout.Strategy
	lastTime time.Time
	index    int
	active   bool

	Timeout time.Duration
	Items   []T
	visible map[ui.Widget]bool

	paddingX int16
	paddingY int16
	marginX  int16
	marginY  int16
}

// Option configures a container at construction time.
type Option[T ui.Widget] func(*Base[T])

// WithLayout applies a layout strategy.
func WithLayout[T ui.Widget](strategy layout.Strategy) Option[T] {
	return func(c *Base[T]) {
		c.layouter = strategy
	}
}

// WithChildren appends child widgets to the container.
func WithChildren[T ui.Widget](children ...T) Option[T] {
	return func(c *Base[T]) {
		for _, child := range children {
			c.Items = append(c.Items, child)
			child.SetParent(c)
		}
	}
}

// WithTimeout overrides the idle timeout.
func WithTimeout[T ui.Widget](timeout time.Duration) Option[T] {
	return func(c *Base[T]) {
		c.Timeout = timeout
	}
}

// WithPadding sets inner padding applied before laying out children.
func WithPadding[T ui.Widget](px, py int16) Option[T] {
	return func(c *Base[T]) {
		c.paddingX = px
		c.paddingY = py
	}
}

// WithMargin sets outer margins applied inside the container bounds.
func WithMargin[T ui.Widget](mx, my int16) Option[T] {
	return func(c *Base[T]) {
		c.marginX = mx
		c.marginY = my
	}
}

func determineSize[T ui.Widget](width, height uint16, l layout.Strategy, widgets []T) (uint16, uint16) {
	ctx := ui.NewContext(nil, 0x7FFF, 0x7FFF, 0, 0)
	var w, h uint16
	for _, widget := range widgets {
		x, y := ctx.DisplayPos()
		wW, wH := widget.Size()

		wCandidate := uint32(x) + uint32(wW)
		hCandidate := uint32(y) + uint32(wH)
		if w < uint16(wCandidate) {
			w = uint16(wCandidate)
		}
		if h < uint16(hCandidate) {
			h = uint16(hCandidate)
		}
		if l != nil {
			l(&ctx, widget)
		}
	}
	if width != 0 {
		w = width
	}
	if height != 0 {
		h = height
	}
	return w, h
}

// New constructs a container configured by options.
func New[T ui.Widget](width, height uint16, opts ...Option[T]) *Base[T] {
	c := &Base[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		lastTime:   time.Now(),
		index:      -1,
		Timeout:    10 * time.Second,
		visible:    make(map[ui.Widget]bool),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.layouter != nil && (width == 0 || height == 0) {
		w, h := determineSize(width, height, c.layouter, c.Items)
		if width == 0 {
			c.Width = w
		}
		if height == 0 {
			c.Height = h
		}
	}
	return c
}

// Draw renders the container and its children using the configured layout.
func (c *Base[T]) Draw(ctx ui.Context) {
	innerW := int16(c.Width) - 2*c.marginX
	innerH := int16(c.Height) - 2*c.marginY
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	localCtx := ctx.Clone(c, uint16(innerW), uint16(innerH))
	localCtx.SetPos(c.marginX+c.paddingX, c.marginY+c.paddingY)

	for _, item := range c.Items {
		visible := childVisible(localCtx, item)
		c.setVisibility(item, visible)
		if !visible {
			if c.layouter != nil && !c.layouter(localCtx, item) {
				return
			}
			continue
		}
		item.Draw(localCtx)
		if c.layouter == nil {
			continue
		}
		if !c.layouter(localCtx, item) {
			return
		}
	}
}

// Interact dispatches navigation commands to children or handles focus changes.
func (c *Base[T]) Interact(cmd ui.UserCommand) bool {
	if cmd == ui.IDLE {
		return c.handleIDLE()
	}
	c.lastTime = time.Now()

	if !c.active {
		return c.handleInactive(cmd)
	}

	if c.index < 0 || c.index >= len(c.Items) {
		return c.handleInactive(cmd)
	}

	return c.Items[c.index].Interact(cmd)
}

// SetIndex selects a child by index and adjusts focus.
func (c *Base[T]) SetIndex(a int) {
	prev := c.currentItem()
	target := c.clampIndex(a)
	if c.index == target {
		return
	}

	for i, item := range c.Items {
		selected := i == target
		item.SetSelected(selected)
		c.notifySelection(item, selected)
	}

	c.index = target
	if target < 0 {
		c.active = false
	}
	focusTransition(prev, c.currentItem())
}

// Index returns the currently selected child index.
func (c *Base[T]) Index() int {
	return c.index
}

// SetActive toggles active state on the selected child.
func (c *Base[T]) SetActive(a int) {
	prev := c.currentItem()
	wasActive := c.active && prev != nil

	if a < 0 {
		if wasActive {
			activationTransition(prev, false)
			c.notifyExit(prev)
		}
		c.active = false
		c.SetIndex(-1)
		return
	}

	c.SetIndex(a)
	next := c.currentItem()
	if next == nil {
		if wasActive {
			activationTransition(prev, false)
			c.notifyExit(prev)
		}
		c.active = false
		return
	}

	if wasActive && prev != nil && prev != next {
		activationTransition(prev, false)
		c.notifyExit(prev)
	}

	c.active = true
	activationTransition(next, true)
}

// Active reports whether the selected child is currently active.
func (c *Base[T]) Active() bool {
	return c.active
}

// Item returns the currently selected child or nil.
func (c *Base[T]) Item() ui.Widget {
	if c.index < 0 || c.index >= len(c.Items) {
		return nil
	}
	return c.Items[c.index]
}

// ChildCount returns the number of children.
func (c *Base[T]) ChildCount() int {
	return len(c.Items)
}

// Child returns the child at a specific index or nil if out of range.
func (c *Base[T]) Child(index int) ui.Widget {
	if index < 0 || index >= len(c.Items) {
		return nil
	}
	return c.Items[index]
}

func (c *Base[T]) handleIDLE() bool {
	if time.Since(c.lastTime) > c.Timeout {
		c.SetIndex(-1)
	}
	return false
}

func (c *Base[T]) handleInactive(cmd ui.UserCommand) bool {
	switch cmd {
	case ui.PREV:
		if c.index > 0 {
			c.SetIndex(c.index - 1)
		}
		return true
	case ui.NEXT:
		if c.index+1 >= len(c.Items) {
			return true
		}
		c.SetIndex(c.index + 1)
		return true
	case ui.ENTER:
		if c.index < 0 || c.index >= len(c.Items) {
			return false
		}
		c.active = true
	default:
		return c.WidgetBase.Interact(cmd)
	}

	return false
}

func (c *Base[T]) currentItem() ui.Widget {
	if c.index < 0 || c.index >= len(c.Items) {
		return nil
	}
	return c.Items[c.index]
}

func (c *Base[T]) clampIndex(a int) int {
	if len(c.Items) == 0 {
		return -1
	}
	if a < -1 {
		return -1
	}
	maxIndex := len(c.Items) - 1
	if a > maxIndex {
		return maxIndex
	}
	return a
}

func (c *Base[T]) setVisibility(item ui.Widget, visible bool) {
	if c.visible == nil {
		c.visible = make(map[ui.Widget]bool)
	}
	prev, ok := c.visible[item]
	if ok && prev == visible {
		return
	}
	c.visible[item] = visible
	if handler, ok := item.(ui.VisibleHandler); ok {
		handler.OnVisible(visible)
	}
}

func (c *Base[T]) notifySelection(item ui.Widget, selected bool) {
	if handler, ok := item.(ui.SelectHandler); ok {
		if selected {
			handler.OnSelect()
		} else {
			handler.OnDeselect()
		}
	}
}

func (c *Base[T]) notifyExit(item ui.Widget) {
	if handler, ok := item.(ui.ExitHandler); ok {
		handler.OnExit()
	}
}
