package widget

import ui "github.com/itohio/tinygui"

type widgetChoiceConfig[T ui.Widget] struct {
	selectorOpts []InteractiveSelectorOption[T]
	onChange     func(int, T)
	enabled      bool
}

// InteractiveWidgetChoice rotates through concrete widgets while delegating navigation to InteractiveSelector.
type InteractiveWidgetChoice[T ui.Widget] struct {
	ui.WidgetBase
	items        []T
	selector     *InteractiveSelector[T]
	config       widgetChoiceConfig[T]
	currentIndex int
}

// InteractiveWidgetChoiceOption configures widget choices.
type InteractiveWidgetChoiceOption[T ui.Widget] func(*widgetChoiceConfig[T])

// WithWidgetChoiceIndex binds the choice to an external index pointer.
func WithWidgetChoiceIndex[T ui.Widget](ptr *int) InteractiveWidgetChoiceOption[T] {
	return func(cfg *widgetChoiceConfig[T]) {
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorIndex[T](ptr))
	}
}

// WithWidgetChoiceChange registers a callback invoked when current widget changes.
func WithWidgetChoiceChange[T ui.Widget](fn func(int, T)) InteractiveWidgetChoiceOption[T] {
	return func(cfg *widgetChoiceConfig[T]) {
		cfg.onChange = fn
	}
}

// WithWidgetChoiceDisabled initialises the choice in a disabled state.
func WithWidgetChoiceDisabled[T ui.Widget]() InteractiveWidgetChoiceOption[T] {
	return func(cfg *widgetChoiceConfig[T]) {
		cfg.enabled = false
		cfg.selectorOpts = append(cfg.selectorOpts, WithSelectorDisabled[T]())
	}
}

// NewInteractiveWidgetChoice constructs a widget-backed selector that draws the active child.
func NewInteractiveWidgetChoice[T ui.Widget](width, height uint16, items []T, opts ...InteractiveWidgetChoiceOption[T]) *InteractiveWidgetChoice[T] {
	cfg := widgetChoiceConfig[T]{enabled: true}
	for _, opt := range opts {
		opt(&cfg)
	}

	choice := &InteractiveWidgetChoice[T]{
		WidgetBase:   ui.NewWidgetBase(width, height),
		items:        items,
		config:       cfg,
		currentIndex: -1,
	}

	selectorOpts := append(cfg.selectorOpts, WithSelectorChange[T](func(i int, value T) {
		choice.apply(i, value)
	}))
	choice.selector = NewInteractiveSelector(items, selectorOpts...)
	if !cfg.enabled {
		choice.selector.SetEnabled(false)
	}

	if value, ok := choice.selector.Current(); ok {
		choice.apply(choice.selector.Index(), value)
	}
	return choice
}

// Enabled reports whether the choice is interactable.
func (c *InteractiveWidgetChoice[T]) Enabled() bool {
	return c.config.enabled
}

// SetEnabled toggles interaction for the choice.
func (c *InteractiveWidgetChoice[T]) SetEnabled(v bool) {
	c.config.enabled = v
	c.selector.SetEnabled(v)
	if !v {
		c.SetSelected(false)
	}
}

// Draw renders the currently selected widget.
func (c *InteractiveWidgetChoice[T]) Draw(ctx ui.Context) {
	widget := c.currentWidget()
	if widget == nil {
		return
	}
	w, h := widget.Size()
	childCtx := ctx.Clone(widget, w, h)
	widget.Draw(childCtx)
}

// Interact processes navigation commands or delegates to the active child.
func (c *InteractiveWidgetChoice[T]) Interact(cmd ui.UserCommand) bool {
	if !c.config.enabled {
		return false
	}

	if c.selector.Handle(cmd) {
		return true
	}

	widget := c.currentWidget()
	if widget == nil {
		return false
	}
	return widget.Interact(cmd)
}

// SetSelected forwards selection state to the underlying widget.
func (c *InteractiveWidgetChoice[T]) SetSelected(sel bool) {
	prev := c.Selected()
	c.WidgetBase.SetSelected(sel)
	widget := c.currentWidget()
	if widget == nil {
		return
	}
	widget.SetSelected(sel)
	if handler, ok := widget.(ui.SelectHandler); ok {
		if sel && !prev {
			handler.OnSelect()
		}
		if !sel && prev {
			handler.OnDeselect()
		}
	}
}

// Selector exposes the underlying selector for advanced control flows.
func (c *InteractiveWidgetChoice[T]) Selector() *InteractiveSelector[T] {
	return c.selector
}

func (c *InteractiveWidgetChoice[T]) currentWidget() ui.Widget {
	if c.currentIndex < 0 || c.currentIndex >= len(c.items) {
		value, ok := c.selector.Current()
		if !ok {
			return nil
		}
		return ui.Widget(value)
	}
	value := c.items[c.currentIndex]
	return ui.Widget(value)
}

func (c *InteractiveWidgetChoice[T]) apply(index int, value T) {
	if index == c.currentIndex {
		if c.config.onChange != nil {
			c.config.onChange(index, value)
		}
		return
	}

	if index < 0 || index >= len(c.items) {
		c.currentIndex = -1
		return
	}

	if c.currentIndex >= 0 && c.currentIndex < len(c.items) {
		prev := ui.Widget(c.items[c.currentIndex])
		if prev != nil {
			prev.SetSelected(false)
			if handler, ok := prev.(ui.SelectHandler); ok {
				handler.OnDeselect()
			}
		}
	}

	c.currentIndex = index
	widget := ui.Widget(value)
	if widget != nil {
		widget.SetParent(c)
		widget.SetSelected(c.Selected())
		if handler, ok := widget.(ui.SelectHandler); ok && c.Selected() {
			handler.OnSelect()
		}
	}

	if c.config.onChange != nil {
		c.config.onChange(index, value)
	}
}
