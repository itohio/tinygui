package widget

import ui "github.com/itohio/tinygui"

// InteractiveIconChoice renders one of several PNG images, chosen with an
// InteractiveSelector. The images are decoded once, at construction (w·h·2
// bytes each); a press only moves the selection.
type InteractiveIconChoice struct {
	ui.WidgetBase
	selector *InteractiveSelector[string]
	sheet    iconSheet
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

	choice := &InteractiveIconChoice{
		WidgetBase: ui.NewWidgetBase(width, height),
		sheet:      newIconSheet(width, height, images),
	}
	selectorOpts := cfg.selectorOpts
	if cfg.onChange != nil {
		selectorOpts = append(selectorOpts, WithSelectorChange[string](cfg.onChange))
	}
	choice.selector = NewInteractiveSelector(images, selectorOpts...)
	return choice
}

// Draw blits the selected image (decoded at construction).
func (c *InteractiveIconChoice) Draw(ctx ui.Context) { c.sheet.draw(ctx, c.selector.Index()) }

// Image is the selected image's PNG payload.
func (c *InteractiveIconChoice) Image() string {
	img, _ := c.selector.Current()
	return img
}

// Enabled reports whether the choice reacts to input.
func (c *InteractiveIconChoice) Enabled() bool {
	return c.selector.Enabled()
}

// SetEnabled toggles input handling.
func (c *InteractiveIconChoice) SetEnabled(v bool) {
	c.selector.SetEnabled(v)
	if !v {
		c.SetSelected(false)
	}
}

// Interact processes navigation commands and updates the active icon image.
func (c *InteractiveIconChoice) Interact(cmd ui.UserCommand) bool {
	if c.selector.Handle(cmd) {
		if cmd == ui.ESC || cmd == ui.BACK {
			return c.WidgetBase.Interact(cmd)
		}
		return true
	}
	return c.WidgetBase.Interact(cmd)
}

// Selector provides access to the underlying selector for advanced coordination.
func (c *InteractiveIconChoice) Selector() *InteractiveSelector[string] {
	return c.selector
}
