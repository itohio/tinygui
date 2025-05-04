package ui

import (
	"time"
)

// Layouter callback that shifts context coordinates
// in preparation to the next widget. Layouter returns false if rendering should stop.
type Layouter func(ctx Context, w Widget) bool

type ContainerBase[T Widget] struct {
	WidgetBase
	layouter Layouter
	lastTime time.Time
	index    int
	active   bool

	Timeout time.Duration
	Items   []T
}

func determineSize[T Widget](width, height uint16, l Layouter, widgets []T) (uint16, uint16) {
	ctx := NewContext(nil, 0x7FFF, 0x7FFF, 0, 0)
	var w, h uint16
	for _, widget := range widgets {
		l(&ctx, widget)

		if w < uint16(ctx.posX) {
			w = uint16(ctx.posX)
		}
		if h < uint16(ctx.posY) {
			h = uint16(ctx.posY)
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

func NewContainer[T Widget](w, h uint16, l Layouter, widgets ...T) *ContainerBase[T] {
	if l != nil && (w == 0 || h == 0) {
		w, h = determineSize(w, h, l, widgets)
	}

	ret := &ContainerBase[T]{
		WidgetBase: NewWidgetBase(w, h),
		lastTime:   time.Now(),
		layouter:   l,
		index:      -1,
		Items:      widgets,
		Timeout:    time.Second * 10,
	}

	for _, w := range widgets {
		w.SetParent(ret)
	}

	return ret
}

func (c *ContainerBase[T]) Draw(ctx Context) {
	w, h := c.Size()
	localCtx := ctx.Clone(c, w, h)

	for _, item := range c.Items {
		item.Draw(localCtx)
		if c.layouter == nil {
			continue
		}
		if !c.layouter(localCtx, item) {
			return
		}
	}
}

func (c *ContainerBase[T]) Interact(cmd UserCommand) bool {
	if cmd == IDLE {
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

func (c *ContainerBase[T]) SetIndex(a int) {
	for i, item := range c.Items {
		item.SetSelected(i == a)
	}
	c.index = a
	if a < 0 {
		c.active = false
	}
}

func (c *ContainerBase[T]) Index() int {
	return c.index
}

func (c *ContainerBase[T]) SetActive(a int) {
	for i, item := range c.Items {
		item.SetSelected(i == a)
	}
	c.index = a
	if a < 0 {
		c.active = false
	}
}

func (c *ContainerBase[T]) Active() bool {
	return c.active
}

func (c *ContainerBase[T]) Item() Widget {
	if c.index < 0 || c.index >= len(c.Items) {
		return nil
	}
	return c.Items[c.index]
}

func (c *ContainerBase[T]) handleIDLE() bool {
	if time.Since(c.lastTime) > c.Timeout {
		c.SetIndex(-1)
	}
	return false
}

func (c *ContainerBase[T]) handleInactive(cmd UserCommand) bool {
	switch cmd {
	case PREV:
		if c.index > 0 {
			c.SetIndex(c.index - 1)
		}
		return true
	case NEXT:
		if c.index+1 >= len(c.Items) {
			return true
		}
		c.SetIndex(c.index + 1)
		return true
	case ENTER:
		if c.index < 0 || c.index >= len(c.Items) {
			return false
		}
		c.active = true
	default:
		return c.WidgetBase.Interact(cmd)
	}

	return false
}
