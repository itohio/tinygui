package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

// Toggle renders a labelled on/off widget driven by getter/setter callbacks.
type Toggle struct {
	ui.WidgetBase
	font     tinyfont.Fonter
	onLabel  string
	offLabel string
	onColor  color.RGBA
	offColor color.RGBA
	text     color.RGBA
	get      func() bool
	set      func(bool)
}

// NewToggle constructs a toggle widget with explicit labels and colours.
func NewToggle(w, h uint16, font tinyfont.Fonter, text color.RGBA, onLabel, offLabel string, onColor, offColor color.RGBA, getter func() bool, setter func(bool)) *Toggle {
	if getter == nil {
		getter = func() bool { return false }
	}
	if setter == nil {
		setter = func(bool) {}
	}
	return &Toggle{
		WidgetBase: ui.NewWidgetBase(w, h),
		font:       font,
		onLabel:    onLabel,
		offLabel:   offLabel,
		onColor:    onColor,
		offColor:   offColor,
		text:       text,
		get:        getter,
		set:        setter,
	}
}

func (t *Toggle) Draw(ctx ui.Context) {
	d := ctx.D()
	if d == nil {
		return
	}

	x, y := ctx.DisplayPos()
	active := t.get()
	bg := t.offColor
	label := t.offLabel
	if active {
		bg = t.onColor
		label = t.onLabel
	}

	if fast, ok := d.(ui.RectangleDisplayer); ok {
		_ = fast.FillRectangle(x, y, int16(t.Width), int16(t.Height), bg)
	} else {
		for dy := int16(0); dy < int16(t.Height); dy++ {
			ui.HLine(d, x, y+dy, int16(t.Width), bg)
		}
	}

	textY := y + int16(t.Height) - 2
	tinyfont.WriteLine(d, t.font, x+2, textY, label, t.text)
}

func (t *Toggle) Interact(cmd ui.UserCommand) bool {
	switch cmd {
	case ui.ENTER, ui.LEFT, ui.RIGHT:
		t.set(!t.get())
		return true
	default:
		return t.WidgetBase.Interact(cmd)
	}
}
