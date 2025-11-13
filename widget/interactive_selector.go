package widget

import ui "github.com/itohio/tinygui"

// InteractiveSelector manages index navigation over a slice of items while supporting external synchronisation.
type InteractiveSelector[T any] struct {
	items    []T
	index    int
	external *int
	enabled  bool
	onChange func(int, T)
}

// InteractiveSelectorOption mutates selector construction.
type InteractiveSelectorOption[T any] func(*InteractiveSelector[T])

// WithSelectorIndex binds the selector to an external index pointer.
func WithSelectorIndex[T any](ptr *int) InteractiveSelectorOption[T] {
	return func(s *InteractiveSelector[T]) {
		s.external = ptr
		if ptr != nil {
			s.index = *ptr
		}
	}
}

// WithSelectorChange registers a callback invoked whenever the active index changes.
func WithSelectorChange[T any](fn func(int, T)) InteractiveSelectorOption[T] {
	return func(s *InteractiveSelector[T]) {
		s.onChange = fn
	}
}

// WithSelectorDisabled initialises the selector in a disabled state.
func WithSelectorDisabled[T any]() InteractiveSelectorOption[T] {
	return func(s *InteractiveSelector[T]) {
		s.enabled = false
	}
}

// NewInteractiveSelector constructs a selector over the supplied items.
func NewInteractiveSelector[T any](items []T, opts ...InteractiveSelectorOption[T]) *InteractiveSelector[T] {
	selector := &InteractiveSelector[T]{
		items:   items,
		enabled: true,
	}
	for _, opt := range opts {
		opt(selector)
	}
	selector.clampIndex(false)
	return selector
}

// Enabled reports whether navigation is permitted.
func (s *InteractiveSelector[T]) Enabled() bool {
	return s.enabled
}

// SetEnabled toggles navigation capability.
func (s *InteractiveSelector[T]) SetEnabled(v bool) {
	s.enabled = v
}

// Current returns the currently selected value if any.
func (s *InteractiveSelector[T]) Current() (T, bool) {
	var zero T
	if len(s.items) == 0 || s.index < 0 || s.index >= len(s.items) {
		return zero, false
	}
	return s.items[s.index], true
}

// Handle processes navigation commands, returning true if the selector consumed the command.
func (s *InteractiveSelector[T]) Handle(cmd ui.UserCommand) bool {
	if !s.enabled || len(s.items) == 0 {
		return false
	}

	switch cmd {
	case ui.UP, ui.NEXT, ui.RIGHT, ui.LONG_UP:
		s.shift(1)
		return true
	case ui.DOWN, ui.PREV, ui.LEFT, ui.LONG_DOWN:
		s.shift(-1)
		return true
	case ui.ESC, ui.BACK:
		s.Reset(false)
		return true
	default:
		return false
	}
}

// Reset aligns the selector with the bound external index (if any). notify controls whether callbacks are triggered.
func (s *InteractiveSelector[T]) Reset(notify bool) {
	if len(s.items) == 0 {
		s.index = 0
		return
	}
	if s.external != nil {
		s.applyIndex(clamp(0, len(s.items), *s.external), notify)
		return
	}
	s.applyIndex(clamp(0, len(s.items), s.index), notify)
}

// shift advances the index by delta, wrapping the boundaries.
func (s *InteractiveSelector[T]) shift(delta int) {
	if len(s.items) == 0 {
		return
	}
	index := wrapIndex(s.index+delta, len(s.items))
	s.applyIndex(index, true)
}

func (s *InteractiveSelector[T]) applyIndex(index int, notify bool) {
	if len(s.items) == 0 {
		s.index = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(s.items) {
		index = len(s.items) - 1
	}
	if index == s.index {
		if s.external != nil {
			*s.external = index
		}
		if notify && s.onChange != nil {
			value, ok := s.Current()
			if ok {
				s.onChange(index, value)
			}
		}
		return
	}

	s.index = index
	if s.external != nil {
		*s.external = index
	}
	if notify && s.onChange != nil {
		value, ok := s.Current()
		if ok {
			s.onChange(index, value)
		}
	}
}

func (s *InteractiveSelector[T]) clampIndex(notify bool) {
	if len(s.items) == 0 {
		s.index = 0
		return
	}
	s.applyIndex(clamp(0, len(s.items), s.index), notify)
}

// Index returns the current selector position.
func (s *InteractiveSelector[T]) Index() int {
	return s.index
}

// SetIndex updates the selector position optionally firing change callbacks.
func (s *InteractiveSelector[T]) SetIndex(index int, notify bool) {
	s.applyIndex(index, notify)
}
