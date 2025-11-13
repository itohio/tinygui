package widget

import ui "github.com/itohio/tinygui"

// InteractiveIconChoice renders selectable image data via an Icon widget while using InteractiveSelector for navigation.
type InteractiveIconChoice struct {
	*Icon
	selector *InteractiveSelector[string]
	onChange func(int, string)
	current  string
}

// InteractiveIconChoiceOption configures icon choice construction.
type InteractiveIconChoiceOption func(*iconChoiceConfig)

type iconChoiceConfig struct {
	selectorOpts []InteractiveSelectorOption[string]
	onChange     func(int, string)
}

// WithIconChoiceIndex wires an external index pointer for the icon selector.
func WithIconChoiceIndex(ptr *int) InteractiveIconChoiceOption {
	return func(cfg *iconChoiceConfig) {
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorIndex[string](ptr))
	}
}

// WithIconChoiceChange registers a callback invoked whenever the icon changes.
func WithIconChoiceChange(fn func(int, string)) InteractiveIconChoiceOption {
	return func(cfg *iconChoiceConfig) {
		cfg.onChange = fn
	}
}

// WithIconChoiceDisabled initialises the icon choice in a disabled state.
func WithIconChoiceDisabled() InteractiveIconChoiceOption {
	return func(cfg *iconChoiceConfig) {
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorDisabled[string]())
	}
}

// NewInteractiveIconChoice constructs an icon-backed selector.
func NewInteractiveIconChoice(width, height uint16, images []string, opts ...InteractiveIconChoiceOption) *InteractiveIconChoice {
	cfg := iconChoiceConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	choice := &InteractiveIconChoice{}
	selectorOpts := cfg.selectorOpts
	selectorOpts = append(selectorOpts, WithSelectorChange[string](func(i int, image string) {
		choice.current = image
		choice.Icon.SetImage(image)
		if cfg.onChange != nil {
			cfg.onChange(i, image)
		}
	}))

	selector := NewInteractiveSelector(images, selectorOpts...)
	choice.selector = selector
	current, _ := selector.Current()
	choice.current = current

	icon := NewIcon(width, height, func() string { return choice.current })
	choice.Icon = icon
	choice.Icon.SetImage(current)
	return choice
}

// Enabled reports whether the choice reacts to input.
func (c *InteractiveIconChoice) Enabled() bool {
	return c.selector.Enabled()
}

// SetEnabled toggles input handling.
func (c *InteractiveIconChoice) SetEnabled(v bool) {
	c.selector.SetEnabled(v)
	if !v {
		c.Icon.SetSelected(false)
	}
}

// Interact processes navigation commands and updates the active icon image.
func (c *InteractiveIconChoice) Interact(cmd ui.UserCommand) bool {
	handled := c.selector.Handle(cmd)
	if handled {
		if cmd == ui.ESC || cmd == ui.BACK {
			return c.Icon.WidgetBase.Interact(cmd)
		}
		image, _ := c.selector.Current()
		c.current = image
		c.Icon.SetImage(image)
		return true
	}
	return c.Icon.WidgetBase.Interact(cmd)
}

// Selector provides access to the underlying selector for advanced coordination.
func (c *InteractiveIconChoice) Selector() *InteractiveSelector[string] {
	return c.selector
}
