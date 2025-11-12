package ui

// UserCommand enumerates high-level inputs processed by widgets and containers.
type UserCommand byte

const (
	IDLE UserCommand = iota
	UP
	DOWN
	LEFT
	RIGHT
	NEXT
	PREV
	ENTER
	ESC
	BACK
	DEL
	RESET
	SAVE
	LOAD
	LONG_UP
	LONG_DOWN
	LONG_LEFT
	LONG_RIGHT
	LONG_ENTER
	LONG_ESC
	LONG_BACK
	LONG_DEL
	LONG_RESET
	USER UserCommand = 64
)

// Widget describes the minimal contract every drawable component must fulfill.
// Implementations render themselves using the provided Context and may opt into
// additional behaviours such as selection or scrolling via separate interfaces.
type Widget interface {
	// Parent returns the parent widget of the widget.
	Parent() Widget
	SetParent(Widget)
	// Draw draws the widget into the context
	Draw(ctx Context)
	// Interact receives user command and does internal state changes.
	// Interact returns true if the interaction was handled.
	Interact(c UserCommand) bool
	// Size returns Width and Height of the widget
	Size() (W uint16, H uint16)
	SetSelected(s bool)
	Selected() bool
}

// FocusHandler may be implemented by widgets that need to be notified whenever
// they gain or lose focus within a navigable structure.
type FocusHandler interface {
	OnFocus()
	OnBlur()
}

// ActivationHandler may be implemented by widgets that perform work only while
// they are active (for example, value editors). The navigator toggles them on
// ENTER/BACK transitions.
type ActivationHandler interface {
	OnActivate()
	OnDeactivate()
}

// Selectable marks widgets that should be considered during navigation.
// Non-selectable widgets are skipped by focus logic automatically.
type Selectable interface {
	CanSelect() bool
}

// VisibleHandler can be implemented by widgets that react to visibility changes.
type VisibleHandler interface {
	OnVisible(visible bool)
}

// ScrollHandler allows widgets to respond to scroll offset updates.
type ScrollHandler interface {
	OnScroll(offsetX, offsetY int16)
}

// SelectHandler is notified when a widget becomes (or stops being) selected.
type SelectHandler interface {
	OnSelect()
	OnDeselect()
}

// ExitHandler is invoked when a widget is exited (e.g. user backs out).
type ExitHandler interface {
	OnExit()
}

// EnableState allows navigation logic to skip disabled widgets.
type EnableState interface {
	Enabled() bool
}

// Scrollable declares support for manual scroll offset adjustments.
// Widgets that do not implement this interface remain static in their context.
type Scrollable interface {
	Scroll(dx, dy int16) bool
	ScrollOffset() (x, y int16)
}

// Navigable combines Widget behaviour with indexed child traversal so the
// Navigator can manage selection/activation for hierarchical layouts.
type Navigable interface {
	Widget
	Index() int
	SetIndex(index int)
	Item() Widget
	SetActive(index int)
	Active() bool
	ChildCount() int
	Child(index int) Widget
}

// Container extends Widget with facilities for selecting and activating child
// widgets. Concrete implementations are free to decide how children are laid
// out and how indices map to widgets.
type Container interface {
	Widget
	// SetIndex sets selected item by index.
	SetIndex(index int)
	// Index returns currently selected item index.
	Index() int
	// Item returns currently selected item. If no item selected, returns nil.
	Item() Widget
	// SetActive selects an item and makes it active. If active < 0, then it deselects and makes all inactive.
	SetActive(active int)
	// Active means that selected item is active and is processing all interactions.
	Active() bool
}

// WidgetBase implements common Widget bookkeeping (parent reference, size,
// selection). Embedding it keeps widgets lightweight without forcing extras.
type WidgetBase struct {
	parent   Widget
	Width    uint16
	Height   uint16
	selected bool
}

// NewWidgetBase constructs a WidgetBase with fixed width/height metadata.
func NewWidgetBase(width, height uint16) WidgetBase {
	return WidgetBase{
		Width:  width,
		Height: height,
	}
}

func (c *WidgetBase) Parent() Widget          { return c.parent }
func (c *WidgetBase) SetParent(widget Widget) { c.parent = widget }
func (c *WidgetBase) SetSelected(s bool)      { c.selected = s }
func (c *WidgetBase) Selected() bool          { return c.selected }
func (c *WidgetBase) Size() (uint16, uint16)  { return c.Width, c.Height }
func (c *WidgetBase) Interact(cmd UserCommand) bool {
	if cmd != ESC {
		return false
	}

	if c.selected {
		c.selected = false
	}
	return true
}
