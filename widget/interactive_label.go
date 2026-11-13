package widget

import (
	"image/color"
	"strconv"
	"unicode/utf8"
	"unsafe"

	ui "github.com/itohio/tinygui"
)

// InteractiveLabel composes a Label with numeric editing behaviour.
type InteractiveLabel[T Number] struct {
	*Label

	external  *T
	min       T
	max       T
	stepSmall T
	stepLarge T

	// formatter: nil shows the number (strconv: as fmt's %v prints a plain
	// number). A String method of T isn't used: names come from WithFormatter.
	formatter func(T) string
	onCommit  func(T)
	text      [40]byte // the shown text, rebuilt per draw without allocating
	enabled   bool

	pending   T
	original  T
	selecting bool
}

// NewInteractiveLabel constructs an interactive label using options.
func NewInteractiveLabel[T Number](width, height uint16, opts ...InteractiveOption[T]) *InteractiveLabel[T] {
	l := &InteractiveLabel[T]{
		min:       0,
		max:       1,
		stepSmall: 1,
		stepLarge: 1,
		enabled:   true,
	}

	label := NewLabel(width, height, nil, nil, color.RGBA{})
	label.SetTextProvider(func() string { return l.displayText() })
	l.Label = label

	for _, opt := range opts {
		opt.applyToLabel(l)
	}

	l.resetDefaults()
	l.load()
	return l
}

func (l *InteractiveLabel[T]) resetDefaults() {
	if l.stepSmall == 0 {
		l.stepSmall = 1
	}
	if l.stepLarge == 0 {
		l.stepLarge = l.stepSmall
	}
	if l.min > l.max {
		l.min, l.max = l.max, l.min
	}
}

// Enabled reports whether the label accepts commands.
func (l *InteractiveLabel[T]) Enabled() bool {
	return l.enabled
}

// SetEnabled toggles interaction ability.
func (l *InteractiveLabel[T]) SetEnabled(v bool) {
	l.enabled = v
	if !v {
		l.SetSelected(false)
	}
}

// Interact handles user commands.
func (l *InteractiveLabel[T]) Interact(cmd ui.UserCommand) bool {
	if !l.enabled {
		return false
	}
	switch cmd {
	case ui.UP, ui.NEXT:
		l.adjust(l.stepSmall)
		return true
	case ui.DOWN, ui.PREV:
		l.adjust(-l.stepSmall)
		return true
	case ui.LONG_UP, ui.RIGHT:
		l.adjust(l.stepLarge)
		return true
	case ui.LONG_DOWN, ui.LEFT:
		l.adjust(-l.stepLarge)
		return true
	case ui.ENTER:
		if !l.Selected() {
			return false
		}
		l.commit()
		l.SetSelected(false)
		return true
	case ui.BACK, ui.ESC:
		if !l.Selected() {
			return false
		}
		l.revert()
		return l.Label.WidgetBase.Interact(cmd)
	default:
		return l.Label.WidgetBase.Interact(cmd)
	}
}

// OnSelect prepares for editing.
func (l *InteractiveLabel[T]) OnSelect() {
	if !l.enabled {
		return
	}
	l.selecting = true
	l.load()
}

// OnDeselect clears selection state.
func (l *InteractiveLabel[T]) OnDeselect() {
	l.selecting = false
}

// OnExit reverts pending edits.
func (l *InteractiveLabel[T]) OnExit() {
	l.revert()
	l.selecting = false
}

// SetSelected overrides label selection to hook editing lifecycle.
func (l *InteractiveLabel[T]) SetSelected(sel bool) {
	prev := l.Selected()
	l.Label.SetSelected(sel)
	if sel && !prev {
		l.OnSelect()
	}
	if !sel && prev {
		l.OnDeselect()
	}
}

func (l *InteractiveLabel[T]) displayText() string {
	v := l.pending
	if !l.selecting && l.external != nil {
		// not editing: show the bound value as it is now, even if something
		// else changed it (it was loaded only when editing began)
		v = clamp(l.min, l.max, *l.external)
	}
	marked := l.enabled && l.Selected()
	n := 0
	if marked {
		n = copy(l.text[:], markOpen)
	}
	if l.formatter != nil {
		// copied into the fixed buffer, cut at a rune boundary with room for
		// the closing mark: a longer text never grows (allocates) a buffer
		room := len(l.text) - n
		if marked {
			room -= len(markClose)
		}
		s := l.formatter(v)
		if len(s) > room {
			s = s[:room]
			for len(s) > 0 { // drop a rune cut in half
				if r, size := utf8.DecodeLastRuneInString(s); r != utf8.RuneError || size > 1 {
					break
				}
				s = s[:len(s)-1]
			}
		}
		n += copy(l.text[n:], s)
	} else {
		n = len(appendNumber(l.text[:n], v)) // ≤ 24 bytes: fits beside both marks
	}
	if marked {
		n += copy(l.text[n:], markClose)
	}
	if n == 0 {
		return ""
	}
	// a view of l.text, valid until the next displayText (Label reads it
	// during Draw; Label.Text's doc says so)
	return unsafe.String(&l.text[0], n)
}

const (
	markOpen  = "▲ "
	markClose = " ▼"
)

// appendNumber appends v as fmt's %v prints it, without fmt and without
// converting v to an interface (which heap-allocates a float64 on 32-bit
// TinyGo targets).
func appendNumber[T Number](b []byte, v T) []byte {
	var half T = 1
	half /= 2
	switch {
	case half != 0: // a float type
		bits := 64
		if unsafe.Sizeof(v) == 4 {
			bits = 32
		}
		return strconv.AppendFloat(b, float64(v), 'g', -1, bits)
	case v < 0:
		return strconv.AppendInt(b, int64(v), 10)
	default:
		return strconv.AppendUint(b, uint64(v), 10)
	}
}

func (l *InteractiveLabel[T]) load() {
	var current T
	if l.external != nil {
		current = *l.external
	}
	current = clamp(l.min, l.max, current)
	l.pending = current
	l.original = current
}

func (l *InteractiveLabel[T]) adjust(delta T) {
	if !l.selecting {
		return
	}
	l.pending = clamp(l.min, l.max, l.pending+delta)
}

func (l *InteractiveLabel[T]) commit() {
	l.pending = clamp(l.min, l.max, l.pending)
	l.original = l.pending
	if l.external != nil {
		*l.external = l.pending
	}
	if l.onCommit != nil {
		l.onCommit(l.pending)
	}
}

func (l *InteractiveLabel[T]) revert() {
	l.pending = l.original
}

func (l *InteractiveLabel[T]) setExternal(ptr *T) {
	l.external = ptr
}
