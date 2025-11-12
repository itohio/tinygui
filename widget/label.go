package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

// Label renders a single line of text using tinyfont.
type Label struct {
	ui.WidgetBase
	font  tinyfont.Fonter
	text  func() string
	color color.RGBA
}

// NewLabel constructs a label of fixed size, font, and colour.
func NewLabel(w, h uint16, font tinyfont.Fonter, text func() string, color color.RGBA) *Label {
	if font == nil {
		font = &tinyfont.TomThumb
	}
	if text == nil {
		text = func() string { return "" }
	}
	return &Label{
		WidgetBase: ui.NewWidgetBase(w, h),
		font:       font,
		text:       text,
		color:      color,
	}
}

func (l *Label) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	tinyfont.WriteLine(ctx.D(), l.font, x, y+int16(l.Height), l.text(), l.color)
}

// SetFont updates the font used to draw the label.
func (l *Label) SetFont(font tinyfont.Fonter) {
	if font != nil {
		l.font = font
	}
}

// SetColor updates the text colour.
func (l *Label) SetColor(color color.RGBA) {
	l.color = color
}

// SetText assigns a static text provider.
func (l *Label) SetText(text string) {
	l.SetTextProvider(func() string { return text })
}

// SetTextProvider swaps the callback used to fetch text during drawing.
func (l *Label) SetTextProvider(fn func() string) {
	if fn != nil {
		l.text = fn
	}
}

// Text returns the currently rendered string.
func (l *Label) Text() string {
	return l.text()
}

// NewLabelArray returns a slice of labels constructed from multiple callbacks.
func NewLabelArray(w, h uint16, font tinyfont.Fonter, color color.RGBA, text ...func() string) []ui.Widget {
	widgets := make([]ui.Widget, len(text))
	for i, fn := range text {
		widgets[i] = NewLabel(w, h, font, fn, color)
	}
	return widgets
}
