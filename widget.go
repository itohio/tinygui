package ui

// Widget describes basic widget behavior
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

// WidgetBase implements basic Widget functions
type WidgetBase struct {
	parent   Widget
	Width    uint16
	Height   uint16
	selected bool
}

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
