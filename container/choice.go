package container

import (
	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

type scrollChoiceConfig struct {
	selectorOpts []widget.InteractiveSelectorOption[ui.Widget]
	onChange     func(int, ui.Widget)
	enabled      bool
}

// ScrollChoiceOption customises ScrollChoice construction.
type ScrollChoiceOption func(*scrollChoiceConfig)

// WithScrollChoiceIndex binds the selector to an external index pointer.
func WithScrollChoiceIndex(ptr *int) ScrollChoiceOption {
	return func(cfg *scrollChoiceConfig) {
		cfg.selectorOpts = append(cfg.selectorOpts, widget.WithSelectorIndex[ui.Widget](ptr))
	}
}

// WithScrollChoiceChange registers a callback fired when the active widget changes.
func WithScrollChoiceChange(fn func(int, ui.Widget)) ScrollChoiceOption {
	return func(cfg *scrollChoiceConfig) {
		cfg.onChange = fn
	}
}

// WithScrollChoiceDisabled initialises the container in a disabled state.
func WithScrollChoiceDisabled() ScrollChoiceOption {
	return func(cfg *scrollChoiceConfig) {
		cfg.enabled = false
		cfg.selectorOpts = append(cfg.selectorOpts, widget.WithSelectorDisabled[ui.Widget]())
	}
}

// ScrollChoice composes Scroll with InteractiveSelector-driven navigation.
type ScrollChoice struct {
	*Scroll
	selector *widget.InteractiveSelector[ui.Widget]
	config   scrollChoiceConfig
	enabled  bool
}

// NewScrollChoice constructs a scrollable choice container.
func NewScrollChoice(viewportW, viewportH uint16, lay layout.Strategy, widgets []ui.Widget, opts ...ScrollChoiceOption) *ScrollChoice {
	cfg := scrollChoiceConfig{enabled: true}
	for _, opt := range opts {
		opt(&cfg)
	}

	scroll := NewScroll(viewportW, viewportH, lay, widgets...)
	choice := &ScrollChoice{
		Scroll:  scroll,
		config:  cfg,
		enabled: cfg.enabled,
	}

	selectorOpts := append(cfg.selectorOpts, widget.WithSelectorChange(func(i int, value ui.Widget) {
		choice.setIndexInternal(i)
		if cfg.onChange != nil {
			cfg.onChange(i, value)
		}
	}))
	choice.selector = widget.NewInteractiveSelector(widgets, selectorOpts...)
	if !cfg.enabled {
		choice.selector.SetEnabled(false)
	}

	// Align container selection with selector.
	if idx := choice.selector.Index(); idx >= 0 && idx < len(choice.Items) {
		choice.Scroll.SetIndex(idx)
	}
	choice.ensureVisible(choice.Index())

	return choice
}

// Enabled reports whether navigation commands are handled.
func (c *ScrollChoice) Enabled() bool { return c.enabled }

// SetEnabled toggles navigation commands.
func (c *ScrollChoice) SetEnabled(v bool) {
	c.enabled = v
	c.selector.SetEnabled(v)
}

// Interact routes commands through the selector before falling back to Scroll.
func (c *ScrollChoice) Interact(cmd ui.UserCommand) bool {
	if !c.enabled {
		return c.Scroll.Interact(cmd)
	}

	if c.Active() {
		return c.Scroll.Interact(cmd)
	}

	if c.selector.Handle(cmd) {
		c.ensureVisible(c.selector.Index())
		return true
	}

	handled := c.Scroll.Interact(cmd)
	if handled {
		c.selector.SetIndex(c.Index(), false)
		c.ensureVisible(c.Index())
	}
	return handled
}

// SetIndex overrides the base implementation to keep selector and scroll offsets synchronised.
func (c *ScrollChoice) SetIndex(index int) {
	c.setIndexInternal(index)
	c.selector.SetIndex(c.Index(), false)
}

func (c *ScrollChoice) setIndexInternal(index int) {
	c.Scroll.SetIndex(index)
	c.ensureVisible(c.Index())
}

func (c *ScrollChoice) ensureVisible(index int) {
	if index < 0 || index >= len(c.Items) {
		return
	}

	rect, ok := c.measure(index)
	if !ok {
		return
	}

	targetX := c.offsetX
	targetY := c.offsetY
	viewportW, viewportH := c.Size()
	maxX := int16(viewportW)
	maxY := int16(viewportH)

	if rect.x < c.offsetX {
		targetX = rect.x
	} else if rect.x+rect.w > c.offsetX+maxX {
		targetX = rect.x + rect.w - maxX
	}

	if rect.y < c.offsetY {
		targetY = rect.y
	} else if rect.y+rect.h > c.offsetY+maxY {
		targetY = rect.y + rect.h - maxY
	}

	dx := targetX - c.offsetX
	dy := targetY - c.offsetY
	if dx != 0 || dy != 0 {
		c.Scroll.Scroll(dx, dy)
	}
}

type childRect struct {
	x, y int16
	w, h int16
}

func (c *ScrollChoice) measure(target int) (childRect, bool) {
	base := ui.NewContext(nil, 0x7fff, 0x7fff, 0, 0)
	ctx := &base
	ctx.SetPos(c.paddingX, c.paddingY)

	for i, child := range c.Items {
		x, y := ctx.Pos()
		width, height := child.Size()
		if i == target {
			return childRect{x: x, y: y, w: int16(width), h: int16(height)}, true
		}
		if c.layouter != nil && !c.layouter(ctx, child) {
			break
		}
	}
	return childRect{}, false
}
