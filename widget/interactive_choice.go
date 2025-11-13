package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

// InteractiveLabelChoice renders selectable text options using a label while delegating navigation to InteractiveSelector.
type InteractiveLabelChoice struct {
	Label
	selector *InteractiveSelector[string]
	onChange func(int, string)
}

// InteractiveLabelChoiceOption configures label choice construction.
type InteractiveLabelChoiceOption func(*labelChoiceConfig)

type labelChoiceConfig struct {
	font         tinyfont.Fonter
	color        color.RGBA
	selectorOpts []InteractiveSelectorOption[string]
	onChange     func(int, string)
}

// WithLabelChoiceIndex wires an external index pointer that stays synchronised with the widget.
func WithLabelChoiceIndex(ptr *int) InteractiveLabelChoiceOption {
	return func(cfg *labelChoiceConfig) {
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorIndex[string](ptr))
	}
}

// WithLabelChoiceChange registers a callback invoked whenever the selection changes.
func WithLabelChoiceChange(fn func(int, string)) InteractiveLabelChoiceOption {
	return func(cfg *labelChoiceConfig) {
		cfg.onChange = fn
	}
}

// WithLabelChoiceDisabled initialises the choice in a disabled state.
func WithLabelChoiceDisabled() InteractiveLabelChoiceOption {
	return func(cfg *labelChoiceConfig) {
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorDisabled[string]())
	}
}

// WithLabelChoiceFont assigns the font used for rendering the active item.
func WithLabelChoiceFont(font tinyfont.Fonter) InteractiveLabelChoiceOption {
	return func(cfg *labelChoiceConfig) {
		cfg.font = font
	}
}

// WithLabelChoiceColor assigns the text colour for the active item.
func WithLabelChoiceColor(col color.RGBA) InteractiveLabelChoiceOption {
	return func(cfg *labelChoiceConfig) {
		cfg.color = col
	}
}

// NewInteractiveLabelChoice constructs a label-backed selector that cycles through the provided items.
func NewInteractiveLabelChoice(width, height uint16, items []string, opts ...InteractiveLabelChoiceOption) *InteractiveLabelChoice {
	cfg := labelChoiceConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.font == nil {
		cfg.font = &tinyfont.TomThumb
	}

	selectorOpts := cfg.selectorOpts
	choice := &InteractiveLabelChoice{}
	selectorOpts = append(selectorOpts, WithSelectorChange[string](func(i int, value string) {
		if cfg.onChange != nil {
			cfg.onChange(i, value)
		}
		choice.Label.SetText(value)
	}))

	selector := NewInteractiveSelector(items, selectorOpts...)
	choice.selector = selector

	value, _ := selector.Current()
	choice.Label = *NewLabel(width, height, cfg.font, nil, cfg.color)
	choice.Label.SetText(value)

	return choice
}

// Selector exposes the underlying selector for advanced coordination (commit/cancel flows).
func (c *InteractiveLabelChoice) Selector() *InteractiveSelector[string] {
	return c.selector
}

// Enabled reports whether the selector accepts user interaction.
func (c *InteractiveLabelChoice) Enabled() bool {
	return c.selector.Enabled()
}

// SetEnabled toggles interaction ability.
func (c *InteractiveLabelChoice) SetEnabled(v bool) {
	c.selector.SetEnabled(v)
	if !v {
		c.Label.SetSelected(false)
	}
}

// Interact processes navigation commands and updates the label text accordingly.
func (c *InteractiveLabelChoice) Interact(cmd ui.UserCommand) bool {
	handled := c.selector.Handle(cmd)
	if handled {
		if cmd == ui.ESC || cmd == ui.BACK {
			return c.Label.WidgetBase.Interact(cmd)
		}
		value, _ := c.selector.Current()
		c.Label.SetText(value)
		return true
	}
	return c.Label.WidgetBase.Interact(cmd)
}

// Draw renders the current label.
func (c *InteractiveLabelChoice) Draw(ctx ui.Context) {
	c.Label.Draw(ctx)
}

// currentText returns the selected text for testing or diagnostics.
func (c *InteractiveLabelChoice) currentText() string {
	value, _ := c.selector.Current()
	return value
}
