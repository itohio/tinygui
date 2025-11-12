package widget

import (
	ui "github.com/itohio/tinygui"
)

// InteractiveIcon renders one of several PNG icons and swaps the active image on user commands.
type InteractiveIcon struct {
	*Icon

	icons    []string
	index    int
	external *int
	onChange func(int)
	enabled  bool
}

// InteractiveIconOption configures optional behaviour for InteractiveIcon.
type InteractiveIconOption func(*InteractiveIcon)

// WithIconIndex wires an external index pointer that is synchronised with the widget.
func WithIconIndex(ptr *int) InteractiveIconOption {
	return func(i *InteractiveIcon) {
		i.external = ptr
		if ptr != nil {
			i.index = *ptr
		}
	}
}

// WithIconChange registers a callback invoked after the icon switches to a new index.
func WithIconChange(fn func(int)) InteractiveIconOption {
	return func(i *InteractiveIcon) {
		i.onChange = fn
	}
}

// WithIconDisabled initialises the icon in a disabled state.
func WithIconDisabled() InteractiveIconOption {
	return func(i *InteractiveIcon) {
		i.enabled = false
	}
}

// NewInteractiveIcon constructs an icon selector that cycles through the provided icons.
func NewInteractiveIcon(width, height uint16, icons []string, opts ...InteractiveIconOption) *InteractiveIcon {
	icon := NewIcon(width, height, func() string { return "" })
	i := &InteractiveIcon{
		Icon:    icon,
		icons:   icons,
		enabled: true,
	}
	for _, opt := range opts {
		opt(i)
	}
	i.load(false)
	return i
}

// Enabled reports whether the icon responds to user commands.
func (i *InteractiveIcon) Enabled() bool {
	return i.enabled
}

// SetEnabled toggles user interaction.
func (i *InteractiveIcon) SetEnabled(v bool) {
	i.enabled = v
	if !v {
		i.Icon.SetSelected(false)
	}
}

// Interact handles rotation commands. Up/Next/Right advance, Down/Prev/Left go backwards.
func (i *InteractiveIcon) Interact(cmd ui.UserCommand) bool {
	if !i.enabled || len(i.icons) == 0 {
		return false
	}

	switch cmd {
	case ui.UP, ui.NEXT, ui.RIGHT, ui.LONG_UP:
		i.shift(1)
		return true
	case ui.DOWN, ui.PREV, ui.LEFT, ui.LONG_DOWN:
		i.shift(-1)
		return true
	case ui.ESC, ui.BACK:
		i.load(false)
		return i.Icon.WidgetBase.Interact(cmd)
	default:
		return i.Icon.WidgetBase.Interact(cmd)
	}
}

// OnSelect keeps behaviour consistent with other interactive widgets.
func (i *InteractiveIcon) OnSelect() {}

// OnDeselect does nothing for icons.
func (i *InteractiveIcon) OnDeselect() {}

// OnExit restores the external index, if any.
func (i *InteractiveIcon) OnExit() {
	i.load(false)
}

func (i *InteractiveIcon) shift(delta int) {
	if len(i.icons) == 0 {
		return
	}
	index := i.currentIndex()
	index = wrapIndex(index+delta, len(i.icons))
	i.applyIndex(index, true)
}

func (i *InteractiveIcon) load(notify bool) {
	if len(i.icons) == 0 {
		i.SetImage("")
		return
	}
	index := i.currentIndex()
	i.applyIndex(wrapIndex(index, len(i.icons)), notify)
}

func (i *InteractiveIcon) currentIndex() int {
	if i.external != nil {
		return *i.external
	}
	return i.index
}

func (i *InteractiveIcon) applyIndex(index int, notify bool) {
	if len(i.icons) == 0 {
		i.SetImage("")
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(i.icons) {
		index = len(i.icons) - 1
	}
	i.index = index
	if i.external != nil {
		*i.external = index
	}
	i.SetImage(i.icons[index])
	if notify && i.onChange != nil {
		i.onChange(index)
	}
}

func wrapIndex(idx, length int) int {
	if length <= 0 {
		return 0
	}
	for idx < 0 {
		idx += length
	}
	for idx >= length {
		idx -= length
	}
	return idx
}
