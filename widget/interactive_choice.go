package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

// InteractiveChoice embeds a Label and cycles through predefined string items on user commands.
type InteractiveChoice struct {
	*Label

	items    []string
	index    int
	external *int
	onChange func(int)
	enabled  bool
}

// InteractiveChoiceOption configures optional behaviour for InteractiveChoice.
type InteractiveChoiceOption func(*InteractiveChoice)

// WithChoiceIndex wires an external index pointer that stays synchronised with the widget.
func WithChoiceIndex(ptr *int) InteractiveChoiceOption {
	return func(c *InteractiveChoice) {
		c.external = ptr
		if ptr != nil {
			c.index = *ptr
		}
	}
}

// WithChoiceChange registers a callback invoked whenever the selection changes.
func WithChoiceChange(fn func(int)) InteractiveChoiceOption {
	return func(c *InteractiveChoice) {
		c.onChange = fn
	}
}

// WithChoiceDisabled initialises the choice in a disabled state.
func WithChoiceDisabled() InteractiveChoiceOption {
	return func(c *InteractiveChoice) {
		c.enabled = false
	}
}

// WithChoiceFont assigns the font used for rendering the active item.
func WithChoiceFont(font tinyfont.Fonter) InteractiveChoiceOption {
	return func(c *InteractiveChoice) {
		c.SetFont(font)
	}
}

// WithChoiceColor assigns the text colour for the active item.
func WithChoiceColor(col color.RGBA) InteractiveChoiceOption {
	return func(c *InteractiveChoice) {
		c.SetColor(col)
	}
}

// NewInteractiveChoice constructs a label-backed selector that cycles through the provided items.
func NewInteractiveChoice(width, height uint16, items []string, opts ...InteractiveChoiceOption) *InteractiveChoice {
	c := &InteractiveChoice{
		items:   items,
		enabled: true,
	}

	label := NewLabel(width, height, nil, nil, color.RGBA{})
	label.SetTextProvider(func() string { return c.currentText() })
	c.Label = label

	for _, opt := range opts {
		opt(c)
	}

	c.load(false)
	return c
}

// Enabled reports whether the selector accepts user interaction.
func (c *InteractiveChoice) Enabled() bool {
	return c.enabled
}

// SetEnabled toggles user interaction.
func (c *InteractiveChoice) SetEnabled(v bool) {
	c.enabled = v
	if !v {
		c.SetSelected(false)
	}
}

// Interact processes navigation commands and rotates the selection accordingly.
func (c *InteractiveChoice) Interact(cmd ui.UserCommand) bool {
	if !c.enabled || len(c.items) == 0 {
		return false
	}

	switch cmd {
	case ui.UP, ui.NEXT, ui.RIGHT, ui.LONG_UP:
		c.shift(1)
		return true
	case ui.DOWN, ui.PREV, ui.LEFT, ui.LONG_DOWN:
		c.shift(-1)
		return true
	case ui.ESC, ui.BACK:
		c.load(false)
		return c.Label.WidgetBase.Interact(cmd)
	default:
		return c.Label.WidgetBase.Interact(cmd)
	}
}

// OnSelect is part of the navigation lifecycle; no-op for choices.
func (c *InteractiveChoice) OnSelect() {}

// OnDeselect is part of the navigation lifecycle; no-op for choices.
func (c *InteractiveChoice) OnDeselect() {}

// OnExit restores the external index (if any) when leaving the widget.
func (c *InteractiveChoice) OnExit() {
	c.load(false)
}

func (c *InteractiveChoice) shift(delta int) {
	if len(c.items) == 0 {
		return
	}
	index := wrapIndex(c.index+delta, len(c.items))
	c.applyIndex(index, true)
}

func (c *InteractiveChoice) load(notify bool) {
	if len(c.items) == 0 {
		c.index = 0
		return
	}
	index := c.currentIndex()
	c.applyIndex(wrapIndex(index, len(c.items)), notify)
}

func (c *InteractiveChoice) currentIndex() int {
	if c.external != nil {
		return *c.external
	}
	return c.index
}

func (c *InteractiveChoice) applyIndex(index int, notify bool) {
	if len(c.items) == 0 {
		c.index = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(c.items) {
		index = len(c.items) - 1
	}
	c.index = index
	if c.external != nil {
		*c.external = index
	}
	if notify && c.onChange != nil {
		c.onChange(index)
	}
}

func (c *InteractiveChoice) currentText() string {
	if len(c.items) == 0 {
		return ""
	}
	idx := c.index
	if idx < 0 || idx >= len(c.items) {
		idx = wrapIndex(idx, len(c.items))
	}
	return c.items[idx]
}
