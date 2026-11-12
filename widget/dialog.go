package widget

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
)

// OverlayStyle is how dialogs and dropdown lists look.
type OverlayStyle struct {
	Font   tinyfont.Fonter
	LineH  int16 // height of a text row
	FG, BG color.RGBA
	Border color.RGBA
	Accent color.RGBA // title, selected option
}

func (s *OverlayStyle) baseline(y int16) int16 { return y + s.LineH - 3 }

const dialogPad = 6

// Dialog is a box shown in an overlay (container.Overlay): a title, lines of
// text and options to choose from.
//
//   - popup: no options; A, B or ESC close it (onCancel).
//   - modal dialog: a row of options such as OK / Cancel; LEFT/RIGHT or
//     UP/DOWN move, ENTER chooses (onChoose(i)), ESC/BACK cancel.
//   - choice question: more than two options, listed one per row.
//
// The dialog closes itself (through its ui.OverlayHost) before calling the
// callback, so a callback can open the next dialog.
type Dialog struct {
	ui.WidgetBase
	style    OverlayStyle
	title    string
	lines    []string
	options  []string
	vertical bool
	sel      int
	onChoose func(int)
	onCancel func()
}

// NewDialog sizes the dialog to its text. onChoose and onCancel may be nil.
func NewDialog(style OverlayStyle, title string, lines, options []string, onChoose func(int), onCancel func()) *Dialog {
	d := &Dialog{style: style, title: title, lines: lines, options: options, onChoose: onChoose, onCancel: onCancel}
	d.vertical = len(options) > 2
	w := ui.TextWidth(style.Font, title)
	for _, l := range lines {
		w = max(w, ui.TextWidth(style.Font, l))
	}
	rows := 1 + len(lines)
	if d.vertical {
		for _, o := range options {
			w = max(w, ui.TextWidth(style.Font, o)+8)
		}
		rows += len(options)
	} else if len(options) > 0 {
		row := int16(0)
		for _, o := range options {
			row += ui.TextWidth(style.Font, o) + 12 + dialogPad
		}
		w = max(w, row-dialogPad)
		rows++
	}
	d.WidgetBase = ui.NewWidgetBase(uint16(w+2*dialogPad), uint16(int16(rows)*style.LineH+2*dialogPad+4))
	return d
}

// Selected option index.
func (d *Dialog) Choice() int { return d.sel }

// close leaves the overlay and resets the selection, so a dialog kept and
// opened again starts at its first option (no allocation per opening).
func (d *Dialog) close() {
	d.sel = 0
	if h := ui.FindOverlayHost(d); h != nil {
		h.Close(d)
	}
}

func (d *Dialog) Interact(cmd ui.UserCommand) bool {
	switch cmd {
	case ui.UP, ui.LEFT, ui.PREV:
		d.sel = max(d.sel-1, 0)
	case ui.DOWN, ui.RIGHT, ui.NEXT:
		d.sel = min(d.sel+1, max(len(d.options)-1, 0))
	case ui.ENTER:
		sel := d.sel
		d.close()
		if len(d.options) == 0 {
			if d.onCancel != nil {
				d.onCancel()
			}
		} else if d.onChoose != nil {
			d.onChoose(sel)
		}
	case ui.ESC, ui.BACK:
		d.close()
		if d.onCancel != nil {
			d.onCancel()
		}
	default:
		return false
	}
	return true
}

func (d *Dialog) Draw(ctx ui.Context) {
	disp := ctx.D()
	x, y := ctx.DisplayPos()
	w, h := int16(d.Width), int16(d.Height)
	s := &d.style
	fillRect(disp, x, y, w, h, s.BG)
	frame(disp, x, y, w, h, s.Border)
	tx, ty := x+dialogPad, y+dialogPad
	ui.DrawText(disp, s.Font, tx, s.baseline(ty), d.title, s.Accent)
	ty += s.LineH
	fillRect(disp, tx, ty, w-2*dialogPad, 1, s.Border)
	ty += 3
	for _, l := range d.lines {
		ui.DrawText(disp, s.Font, tx, s.baseline(ty), l, s.FG)
		ty += s.LineH
	}
	for i, o := range d.options {
		ow := ui.TextWidth(s.Font, o) + 12
		if d.vertical {
			ow = w - 2*dialogPad
		}
		fg := s.FG
		if i == d.sel {
			fillRect(disp, tx, ty, ow, s.LineH, s.Accent)
			fg = s.BG
		} else {
			frame(disp, tx, ty, ow, s.LineH, s.Border)
		}
		ui.DrawText(disp, s.Font, tx+6, s.baseline(ty), o, fg)
		if d.vertical {
			ty += s.LineH
		} else {
			tx += ow + dialogPad
		}
	}
}

// frame draws a 1-pixel rectangle outline.
func frame(d drivers.Displayer, x, y, w, h int16, c color.RGBA) {
	fillRect(d, x, y, w, 1, c)
	fillRect(d, x, y+h-1, w, 1, c)
	fillRect(d, x, y, 1, h, c)
	fillRect(d, x+w-1, y, 1, h, c)
}
