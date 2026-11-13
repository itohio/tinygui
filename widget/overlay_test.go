package widget_test

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/internal/check"
	"github.com/itohio/tinygui/widget"
	"tinygo.org/x/tinyfont"
)

// nopDisplay accepts drawing and records nothing.
type nopDisplay struct{ w, h int16 }

func (d nopDisplay) Size() (int16, int16)                                                 { return d.w, d.h }
func (nopDisplay) SetPixel(int16, int16, color.RGBA)                                      {}
func (nopDisplay) Display() error                                                         { return nil }
func (nopDisplay) FillRectangle(int16, int16, int16, int16, color.RGBA) error             { return nil }
func (nopDisplay) FillRectangleWithBuffer(int16, int16, int16, int16, []color.RGBA) error { return nil }
func (nopDisplay) FillScreen(color.RGBA)                                                  {}

var style = widget.OverlayStyle{Font: &tinyfont.TomThumb, LineH: 10}

// recorder is a base widget that records the commands it receives.
type recorder struct {
	ui.WidgetBase
	got []ui.UserCommand
}

func (r *recorder) Draw(ui.Context) {}
func (r *recorder) Interact(c ui.UserCommand) bool {
	r.got = append(r.got, c)
	return true
}

func newHost() (*container.Overlay, *recorder, ui.ContextImpl) {
	base := &recorder{WidgetBase: ui.NewWidgetBase(200, 100)}
	o := container.NewOverlay(base)
	return o, base, ui.NewContext(nopDisplay{200, 100}, 200, 100, 0, 0)
}

// A dialog is modal: while it is open the base gets no commands; choosing
// closes it and reports the option.
func TestDialogIsModalAndReportsTheChoice(t *testing.T) {
	o, base, _ := newHost()
	chosen, cancelled := -1, false
	d := widget.NewDialog(style, "Reset?", []string{"All demos restart."}, []string{"Yes", "No"},
		func(i int) { chosen = i }, func() { cancelled = true })
	check.True(t, o.OpenCentered(d))
	o.Interact(ui.RIGHT)
	check.Empty(t, base.got)
	o.Interact(ui.ENTER)
	check.Equal(t, 1, chosen)
	check.False(t, cancelled)
	check.Nil(t, o.Top())
	o.Interact(ui.DOWN) // closed: commands reach the base again
	check.Equal(t, []ui.UserCommand{ui.DOWN}, base.got)
}

func TestDialogCancelAndPopup(t *testing.T) {
	o, _, _ := newHost()
	cancelled := 0
	choice := widget.NewDialog(style, "Pattern", nil, []string{"A", "B", "C"}, func(int) { t.Fatal("chosen") }, func() { cancelled++ })
	o.OpenCentered(choice)
	o.Interact(ui.DOWN)
	o.Interact(ui.BACK)
	check.Equal(t, 1, cancelled)
	check.Nil(t, o.Top())

	popup := widget.NewDialog(style, "About", []string{"tinygui"}, nil, nil, func() { cancelled++ })
	o.OpenCentered(popup)
	o.Interact(ui.ENTER)
	check.Equal(t, 2, cancelled)
	check.Nil(t, o.Top())
}

// A callback may open the next dialog: the first one is closed by then.
func TestDialogCallbackOpensAnotherDialog(t *testing.T) {
	o, _, _ := newHost()
	second := widget.NewDialog(style, "Done", nil, nil, nil, nil)
	first := widget.NewDialog(style, "Go?", nil, []string{"OK"}, func(int) { o.OpenCentered(second) }, nil)
	o.OpenCentered(first)
	o.Interact(ui.ENTER)
	check.Equal(t, ui.Widget(second), o.Top())
}

func TestOverlayCloseDropsLayersAbove(t *testing.T) {
	o, _, _ := newHost()
	a := widget.NewDialog(style, "a", nil, nil, nil, nil)
	b := widget.NewDialog(style, "b", nil, nil, nil, nil)
	o.Open(a, 0, 0)
	o.Open(b, 0, 0)
	o.Close(a)
	check.Nil(t, o.Top())
	for range container.MaxOverlays {
		check.True(t, o.Open(widget.NewDialog(style, "x", nil, nil, nil, nil), 0, 0))
	}
	check.False(t, o.Open(a, 0, 0))
}

// The dropdown opens its list under itself, or above when the host has no
// room below; ENTER in the list sets the index, BACK leaves it.
func TestDropdownOpensNextToItselfAndChooses(t *testing.T) {
	idx, changed := 0, ""
	dd := widget.NewDropdown(60, style, []string{"x1", "x2", "x4"}, &idx, func(_ int, s string) { changed = s })
	page := container.New[ui.Widget](200, 100, container.WithChildren[ui.Widget](dd))
	o := container.NewOverlay(page)
	ctx := ui.NewContext(nopDisplay{200, 100}, 200, 100, 0, 0)

	check.False(t, widget.NewDropdown(60, style, []string{"a"}, &idx, nil).Open(), "no host")

	page.Draw(&ctx) // the dropdown learns where it is: (0, 0), 12 tall
	check.True(t, dd.Open())
	o.Interact(ui.DOWN)
	o.Interact(ui.DOWN)
	o.Interact(ui.ENTER)
	check.Equal(t, 2, idx)
	check.Equal(t, "x4", changed)
	check.Nil(t, o.Top())

	dd.Open()
	o.Interact(ui.UP)
	o.Interact(ui.BACK)
	check.Equal(t, 2, idx, "BACK keeps the choice")

	// near the bottom (y 80, 12 tall) the 32-px list doesn't fit below:
	// it opens upwards, at y = 80 - 32 = 48
	rec := &fillRecorder{nopDisplay: nopDisplay{200, 100}}
	low := ui.NewContext(rec, 200, 100, 0, 80)
	page.Draw(&low)
	check.True(t, dd.Open())
	rec.fills = nil
	at := ui.NewContext(rec, 200, 100, 0, 0)
	o.Draw(&at)
	check.Contains(t, rec.fills, [4]int16{0, 48, 60, 32}, "list background")
}

// fillRecorder records filled rectangles (x, y, w, h).
type fillRecorder struct {
	nopDisplay
	fills [][4]int16
}

func (r *fillRecorder) FillRectangle(x, y, w, h int16, _ color.RGBA) error {
	r.fills = append(r.fills, [4]int16{x, y, w, h})
	return nil
}

// A bound index past the end opens on the last option and chooses it; a
// dropdown with no options doesn't open (verify B-F3: these panicked).
func TestDropdownClampsItsIndexAndNeedsOptions(t *testing.T) {
	ctx := ui.NewContext(nopDisplay{200, 100}, 200, 100, 0, 0)
	idx := 7
	dd := widget.NewDropdown(60, style, []string{"a", "b"}, &idx, nil)
	page := container.New[ui.Widget](200, 100, container.WithChildren[ui.Widget](dd))
	o := container.NewOverlay(page)
	page.Draw(&ctx)
	check.True(t, dd.Open())
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()
		o.Interact(ui.ENTER)
	}()
	check.Equal(t, 1, idx)

	empty := 0
	none := widget.NewDropdown(60, style, nil, &empty, nil)
	page2 := container.New[ui.Widget](200, 100, container.WithChildren[ui.Widget](none))
	o2 := container.NewOverlay(page2)
	page2.Draw(&ctx)
	check.False(t, none.Open())
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()
		none.Interact(ui.ENTER)
		o2.Interact(ui.DOWN)
		o2.Interact(ui.ENTER)
	}()
	check.Nil(t, o2.Top())
}

// A list that fits neither below nor above the dropdown stays on screen
// (verify B-F4: it opened above at y = -42).
func TestDropdownListStaysOnScreen(t *testing.T) {
	idx := 0
	dd := widget.NewDropdown(60, style, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, &idx, nil) // 92 px list
	page := container.New[ui.Widget](200, 100, container.WithChildren[ui.Widget](dd))
	o := container.NewOverlay(page)
	rec := &fillRecorder{nopDisplay: nopDisplay{200, 100}}
	mid := ui.NewContext(rec, 200, 100, 0, 50) // the dropdown at y 50, 12 tall
	page.Draw(&mid)
	check.True(t, dd.Open())
	rec.fills = nil
	at := ui.NewContext(rec, 200, 100, 0, 0)
	o.Draw(&at)
	found := false
	for _, f := range rec.fills {
		if f[2] == 60 && f[3] == 92 { // the list's background
			found = true
			check.GreaterOrEqual(t, f[1], int16(0))
			check.LessOrEqual(t, f[1]+f[3], int16(100))
		}
	}
	check.True(t, found, "list background drawn")
}

// A dialog kept and opened again (no allocation per opening) starts at its
// first option each time, whatever was chosen or cancelled last time.
func TestReopenedDialogStartsAtTheFirstOption(t *testing.T) {
	o, _, _ := newHost()
	var chosen []int
	d := widget.NewDialog(style, "Demos", nil, []string{"Pause", "Reset", "About", "Close"}, func(i int) { chosen = append(chosen, i) }, nil)
	o.OpenCentered(d)
	o.Interact(ui.DOWN)
	o.Interact(ui.DOWN)
	o.Interact(ui.ENTER)
	o.OpenCentered(d)
	o.Interact(ui.DOWN)
	o.Interact(ui.BACK)
	o.OpenCentered(d)
	check.Equal(t, 0, d.Choice())
	o.Interact(ui.ENTER)
	check.Equal(t, []int{2, 0}, chosen)
}

// A dialog closed by its host, not by its own keys (Overlay.Close on it, or
// on a layer below it), still reopens at its first option.
func TestDialogClosedByHostReopensAtTheFirstOption(t *testing.T) {
	o, _, _ := newHost()
	var chosen []int
	d := widget.NewDialog(style, "Demos", nil, []string{"Pause", "Reset", "About"}, func(i int) { chosen = append(chosen, i) }, nil)
	o.OpenCentered(d)
	o.Interact(ui.DOWN)
	o.Close(d)
	o.OpenCentered(d)
	check.Equal(t, 0, d.Choice(), "after Overlay.Close(d)")

	below := widget.NewDialog(style, "Menu", nil, []string{"OK"}, nil, nil)
	o.Close(d)
	o.OpenCentered(below)
	o.OpenCentered(d)
	o.Interact(ui.DOWN)
	o.Interact(ui.DOWN)
	o.Close(below)
	o.OpenCentered(d)
	check.Equal(t, 0, d.Choice(), "after a layer below closed")
	o.Interact(ui.ENTER)
	check.Equal(t, []int{0}, chosen)
}

// drawSpy records its name when drawn.
type drawSpy struct {
	ui.WidgetBase
	name  string
	order *[]string
}

func newDrawSpy(name string, order *[]string) *drawSpy {
	return &drawSpy{WidgetBase: ui.NewWidgetBase(10, 10), name: name, order: order}
}

func (s *drawSpy) Draw(ui.Context) { *s.order = append(*s.order, s.name) }

