package widget

import (
	ui "github.com/itohio/tinygui"
)

// Dropdown (a combo box) shows the chosen option; ENTER opens the list of
// options in the nearest ui.OverlayHost, under the dropdown or above it when
// there is no room below. In the list UP/DOWN move, ENTER chooses, ESC/BACK
// close it unchanged.
type Dropdown struct {
	ui.WidgetBase
	style    OverlayStyle
	options  []string
	index    *int
	onChange func(int, string)
	x, y     int16 // where it was last drawn, to open the list next to it
	list     dropdownList
}

// NewDropdown binds the choice to *index; onChange may be nil.
func NewDropdown(w uint16, style OverlayStyle, options []string, index *int, onChange func(int, string)) *Dropdown {
	d := &Dropdown{style: style, options: options, index: index, onChange: onChange}
	d.WidgetBase = ui.NewWidgetBase(w, uint16(style.LineH+2))
	d.list.d = d
	d.list.WidgetBase = ui.NewWidgetBase(w, uint16(int16(len(options))*style.LineH+2))
	return d
}

func (d *Dropdown) Draw(ctx ui.Context) {
	disp := ctx.D()
	d.x, d.y = ctx.DisplayPos()
	w, h := int16(d.Width), int16(d.Height)
	s := &d.style
	border := s.Border
	if d.Selected() {
		border = s.Accent
	}
	fillRect(disp, d.x, d.y, w, h, s.BG)
	frame(disp, d.x, d.y, w, h, border)
	if i := *d.index; i >= 0 && i < len(d.options) {
		ui.DrawText(disp, s.Font, d.x+4, s.baseline(d.y+1), d.options[i], s.FG)
	}
	// a small down-pointing triangle at the right
	ax, ay := d.x+w-10, d.y+h/2-2
	for r := int16(0); r < 4; r++ {
		fillRect(disp, ax+r, ay+r, 7-2*r, 1, s.FG)
	}
}

// Open shows the list. It returns false without an overlay host or with no
// options; the list starts on the bound index, clamped to the options.
func (d *Dropdown) Open() bool {
	if len(d.options) == 0 {
		return false
	}
	h := ui.FindOverlayHost(d)
	if h == nil {
		return false
	}
	d.list.sel = min(max(*d.index, 0), len(d.options)-1)
	y := d.y + int16(d.Height)
	_, lh := d.list.Size()
	if sz, ok := h.(ui.Sizer); ok {
		if _, hh := sz.Size(); y+int16(lh) > int16(hh) {
			y = d.y - int16(lh) // no room below: open upwards
		}
	}
	return h.Open(&d.list, d.x, y)
}

func (d *Dropdown) Interact(cmd ui.UserCommand) bool {
	if cmd == ui.ENTER {
		return d.Open()
	}
	return false
}

// dropdownList is the open list; it lives in the overlay until closed.
type dropdownList struct {
	ui.WidgetBase
	d   *Dropdown
	sel int
}

func (l *dropdownList) Interact(cmd ui.UserCommand) bool {
	d := l.d
	switch cmd {
	case ui.UP, ui.PREV:
		l.sel = max(l.sel-1, 0)
	case ui.DOWN, ui.NEXT:
		l.sel = min(l.sel+1, len(d.options)-1)
	case ui.ENTER:
		l.close()
		*d.index = l.sel
		if d.onChange != nil {
			d.onChange(l.sel, d.options[l.sel])
		}
	case ui.ESC, ui.BACK:
		l.close()
	default:
		return false
	}
	return true
}

func (l *dropdownList) close() {
	if h := ui.FindOverlayHost(l); h != nil {
		h.Close(l)
	}
}

func (l *dropdownList) Draw(ctx ui.Context) {
	disp := ctx.D()
	x, y := ctx.DisplayPos()
	s := &l.d.style
	w, h := int16(l.Width), int16(l.Height)
	fillRect(disp, x, y, w, h, s.BG)
	frame(disp, x, y, w, h, s.Accent)
	for i, o := range l.d.options {
		ry := y + 1 + int16(i)*s.LineH
		fg := s.FG
		if i == l.sel {
			fillRect(disp, x+1, ry, w-2, s.LineH, s.Accent)
			fg = s.BG
		}
		ui.DrawText(disp, s.Font, x+4, s.baseline(ry), o, fg)
	}
}
