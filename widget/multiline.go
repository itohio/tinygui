package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/tinyfont"
)

type MultilineLabel struct {
	ui.WidgetBase
	font     tinyfont.Fonter
	text     func() []string
	color    color.RGBA
	maxLines int
}

func NewMultilineLabel(w, h uint16, maxLines int, font tinyfont.Fonter, text func() []string, color color.RGBA) *MultilineLabel {
	return &MultilineLabel{
		WidgetBase: ui.NewWidgetBase(w, h*uint16(maxLines)),
		font:       font,
		text:       text,
		color:      color,
		maxLines:   maxLines,
	}
}

func (l *MultilineLabel) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	lines := l.text()
	if len(lines) == 0 {
		return
	}
	H := int16(l.Height / uint16(l.maxLines))
	for _, line := range lines {
		y += H
		tinyfont.WriteLine(ctx.D(), l.font, x, y, line, l.color)
	}
}

type Log struct {
	*MultilineLabel
	lines    []string
	numLines int
}

func NewLog(w, h uint16, maxLines int, font tinyfont.Fonter, color color.RGBA) *Log {
	ret := &Log{
		lines:    make([]string, maxLines),
		numLines: 0,
	}
	ret.MultilineLabel = NewMultilineLabel(w, h, maxLines, font, func() []string { return ret.lines[:ret.numLines] }, color)
	return ret
}

func (l *Log) Log(ctx ui.Context, text string) {
	copy(l.lines[1:], l.lines)
	l.lines[0] = text
	if l.numLines < len(l.lines) {
		l.numLines++
	}

	if ctx != nil {
		localCtx := ctx.Clone(l, l.Width, l.Height)
		l.Draw(localCtx)
		ctx.D().Display()
	}
}
