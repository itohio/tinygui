// Package refapp is a reference app over every part of TinyGUI — widgets,
// containers, layouts, navigation, pages, overlays, animation, text — used
// to prove the library allocates nothing after setup (owner: "tinygui itself
// should be 0 allocations (setup does not count)"; it must run with the
// garbage collector disabled). Its tests run under `go test` and under
// `tinygo test -gc=leaking`; cmd/refsoak is the same soak as a program for
// `tinygo build -print-allocs`.
//
// It uses only plain `testing` (no testify) so TinyGo can run it.
package refapp

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/animation"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
	"tinygo.org/x/tinyfont/proggy"
)

const (
	W, H  = 320, 240
	BandH = 80 // drawn in three bands, like the Tufty's framebuffer
)

var (
	font  = &proggy.TinySZ8pt7b
	fg    = color.RGBA{230, 230, 230, 255}
	bg    = color.RGBA{10, 12, 16, 255}
	track = color.RGBA{40, 44, 60, 255}
	good  = color.RGBA{80, 200, 120, 255}
	acc   = color.RGBA{0, 190, 210, 255}
	style = widget.OverlayStyle{Font: font, LineH: 12, FG: fg, BG: bg, Border: track, Accent: acc}
)

// App is the reference UI: a root column of a page area (container.Pages)
// over an Overlay, a navigator per page, animated values.
type App struct {
	FB      *framebuffer.RGB565
	root    *container.Base[ui.Widget]
	overlay *container.Overlay
	pages   *container.Pages
	navs    []*ui.Navigator
	ctx     ui.ContextImpl
	now     int64

	// models the widgets show
	level, mix0         float32
	mix                 []float32
	count               int
	gain                float32
	choice, pick, ddIdx int
	ledOn               bool
	curve               [1]float32
	curveFrom, curveTo  [1]float32
	anim                animation.Animator
	cells               []uint8
	samples             []uint16
	pix16               []uint16
	pix8                []uint8
	dialog              *widget.Dialog
	dropdown            *widget.Dropdown
	log                 *widget.InteractiveLog
	events              int
	lines               [3]string

	// AfterStep, if set, is called after every step of a Round with the
	// step's number (the tests count allocations per step).
	AfterStep func(step int)
	stepN     int
}

type observer struct{ n *int }

func (o observer) OnNavigatorEvent(ui.NavigatorEvent) { *o.n++ }

// New builds the app and draws a frame of every page: everything that
// allocates (construction, first draws, lazily built pages) happens here.
func New() *App {
	a := &App{
		FB:      framebuffer.NewBand(W, H, BandH, make([]uint16, W*BandH), nil),
		level:   40,
		mix:     []float32{20, 30, 25},
		gain:    1.5,
		cells:   make([]uint8, 16*12),
		samples: make([]uint16, 600),
		pix16:   make([]uint16, 32*24),
		pix8:    make([]uint8, 32*24),
	}
	for i := range a.cells {
		a.cells[i] = uint8(i % 4)
	}
	for i := range a.samples {
		a.samples[i] = uint16(i % 50)
	}
	a.anim = animation.NewEaseInOut(1_000_000)
	a.curveTo[0] = 100
	a.anim.Start(a.curveFrom[:], a.curveTo[:], 0)
	a.lines = [3]string{"multiline", "label of", "three lines"}

	specs := []container.PageSpec{
		{Name: "Widgets", Build: a.pageWidgets},
		{Name: "Controls", Build: a.pageControls},
		{Name: "List", Build: a.pageList},
		{Name: "Overlays", Build: a.pageOverlays},
	}
	a.navs = make([]*ui.Navigator, len(specs))
	a.pages = container.NewPages(W, H-14, specs, container.PagesConfig{Keep: len(specs), History: 8})
	header := widget.NewLabel(W, 12, font, func() string { return "TinyGUI reference" }, fg)
	a.root = container.New[ui.Widget](W, H,
		container.WithLayout[ui.Widget](layout.VList(2)),
		container.WithChildren[ui.Widget](header, a.pages))
	a.overlay = container.NewOverlay(a.root)
	a.ctx = ui.NewContext(a.FB, W, H, 0, 0)
	a.dialog = widget.NewDialog(style, "Reset?", []string{"All values return."}, []string{"Yes", "No"}, func(int) {}, func() {})
	for i := range specs { // build every page and draw it once: first draws size their buffers
		a.Show(i)
		a.Frame()
	}
	a.Show(0)
	a.Frame()
	return a
}

func (a *App) newPage(l layout.Strategy, children ...ui.Widget) *container.Base[ui.Widget] {
	return container.New[ui.Widget](W, H-14, container.WithPadding[ui.Widget](4, 4),
		container.WithLayout[ui.Widget](l), container.WithChildren[ui.Widget](children...))
}

func (a *App) pageWidgets() ui.Widget {
	ml := widget.NewMultilineLabel(100, 11, 3, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	ml.SetLines(a.lines[:])
	p := a.newPage(layout.HFlow(4, W-8),
		widget.NewLabel(100, 11, font, func() string { return "Label" }, fg),
		ml,
		widget.NewSeparator(100, 3, track),
		widget.NewGauge[float32](100, 10, &a.curve[0], 0, 100, acc, track),
		widget.NewVerticalGauge[float32](10, 40, &a.level, 0, 100, good, track),
		widget.NewMultiGauge[float32](100, 10, 0, 100, &a.mix, []color.RGBA{acc, good, fg}, track, fg),
		widget.NewVolumeGauge[float32](100, 12, &a.level, 0, 100, 10, good, track),
		widget.NewSolidGauge[float32](100, 12, &a.curve[0], 0, 100, acc, track),
		widget.NewBitmap16(32, 24, a.pix16),
		widget.NewBitmap8(32, 24, a.pix8),
		widget.NewCellGrid(16, 12, 4, 1, widget.CellRound, a.cells, []color.RGBA{bg, acc, good, fg}, track),
		scope(a.samples),
	)
	a.navs[0] = a.nav(p)
	return p
}

func scope(samples []uint16) ui.Widget {
	s := widget.NewScope(150, 60, bg, track, track, good, acc)
	s.Samples, s.Lo, s.Hi, s.Trigger, s.TrigAt = samples, 0, 50, 25, 60
	return s
}

func (a *App) pageControls() ui.Widget {
	toggle := widget.NewToggle(60, 12, font, fg, "ON", "OFF", good, track,
		func() bool { return a.ledOn }, func(v bool) { a.ledOn = v })
	intLabel := widget.NewInteractiveLabel[int](80, 11,
		widget.WithValue(&a.count), widget.WithRange(0, 1000), widget.WithSteps(1, 10),
		widget.WithFont[int](font), widget.WithTextColor[int](fg)) // default formatting
	floatLabel := widget.NewInteractiveLabel[float32](80, 11,
		widget.WithValue(&a.gain), widget.WithRange[float32](-12, 12), widget.WithSteps[float32](0.5, 3),
		widget.WithFont[float32](font), widget.WithTextColor[float32](fg))
	choice := widget.NewInteractiveLabelChoice(80, 11, []string{"red", "green", "blue"},
		widget.WithLabelChoiceIndex(&a.choice), widget.WithLabelChoiceFont(font), widget.WithLabelChoiceColor(fg))
	gauge := widget.NewInteractiveGauge[float32](100, 10,
		widget.WithValue(&a.level), widget.WithRange[float32](0, 100), widget.WithSteps[float32](2, 10),
		widget.WithForeground[float32](good), widget.WithBackground[float32](track))
	multi := widget.NewInteractiveMultiGauge[float32](100, 10,
		widget.WithValues(&a.mix), widget.WithRange[float32](0, 100), widget.WithSteps[float32](2, 10),
		widget.WithSegmentColors[float32]([]color.RGBA{acc, good, fg}), widget.WithBackground[float32](track))
	pick := widget.NewInteractiveWidgetChoice[ui.Widget](40, 24,
		[]ui.Widget{widget.NewBitmap16(32, 24, a.pix16), widget.NewBitmap8(32, 24, a.pix8)},
		widget.WithWidgetChoiceIndex[ui.Widget](&a.pick))
	p := a.newPage(layout.VList(3), toggle, intLabel, floatLabel, choice, gauge, multi, pick)
	a.navs[1] = a.nav(p)
	return p
}

func (a *App) pageList() ui.Widget {
	rows := make([]ui.Widget, 20)
	for i := range rows {
		label := widget.NewLabel(80, 12, font, func() string { return "a list row" }, fg)
		if i%2 == 0 {
			rows[i] = label
			continue
		}
		// a row that is itself a container: drawn through the scroll's offset context
		rows[i] = container.New[ui.Widget](140, 14, container.WithLayout[ui.Widget](layout.HList(4)),
			container.WithChildren[ui.Widget](label, widget.NewSeparator(40, 3, acc)))
	}
	list := container.NewScrollChoice(150, 120, layout.VList(2), rows)
	a.log = widget.NewInteractiveLog(150, 11, 8, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	a.log.SetCapacity(32)
	p := a.newPage(layout.HList(6), list, a.log)
	a.navs[2] = a.nav(p)
	return p
}

func (a *App) pageOverlays() ui.Widget {
	a.dropdown = widget.NewDropdown(80, style, []string{"x1", "x2", "x4"}, &a.ddIdx, func(int, string) {})
	p := a.newPage(layout.VList(4), a.dropdown,
		widget.NewLabel(150, 11, font, func() string { return "dialogs open over this" }, fg))
	a.navs[3] = a.nav(p)
	return p
}

func (a *App) nav(root *container.Base[ui.Widget]) *ui.Navigator {
	n := ui.NewNavigator(root)
	n.AddObserver(observer{&a.events})
	n.Focus(0)
	return n
}

// Show shows page i.
func (a *App) Show(i int) { a.pages.Show(i) }

// Frame advances the animation and draws the screen in three bands.
func (a *App) Frame() {
	a.now += 33_000
	if a.anim.Update(a.curve[:], a.now) {
		a.curveFrom, a.curveTo = a.curveTo, a.curveFrom
		a.anim.Start(a.curveFrom[:], a.curveTo[:], a.now)
	}
	for b := range H / BandH {
		a.FB.SetY0(int16(b * BandH))
		a.FB.FillScreen(bg)
		a.root.Draw(&a.ctx)
		a.overlay.DrawLayers(&a.ctx)
	}
}

// done ends a step of a round: a frame, then the AfterStep hook.
func (a *App) done() {
	a.Frame()
	if a.AfterStep != nil {
		a.AfterStep(a.stepN)
	}
	a.stepN++
}

// Round is one round of the soak: every page by "menu" (Pages.Show), with
// the interactions a person makes there, a frame after each step. Plain
// calls, no closures: a closure capturing variables is itself a heap
// allocation (of the caller, not of the library).
func (a *App) Round() {
	a.stepN = 0
	// Widgets: focus moves over static widgets
	a.Show(0)
	a.done()
	a.navs[0].Next()
	a.done()
	a.navs[0].Prev()
	a.done()
	// Controls: toggle, edit the int, float and choice labels, the gauges, the widget choice
	a.Show(1)
	a.done()
	n := a.navs[1]
	n.Focus(0)
	n.Current().Interact(ui.ENTER)
	a.done()
	for i := 1; i <= 6; i++ {
		n.Focus(i)
		n.Enter()
		a.done()
		n.Current().Interact(ui.UP)
		a.done()
		n.Current().Interact(ui.LONG_UP)
		a.done()
		n.Current().Interact(ui.DOWN)
		a.done()
		n.Current().Interact(ui.ENTER)
		a.done()
		n.Back()
		a.done()
	}
	n.Focus(1) // edit, then revert with ESC
	n.Enter()
	n.Current().Interact(ui.UP)
	n.Current().Interact(ui.ESC)
	n.Back()
	a.done()
	// List: into the list, scroll down and back, out; log lines
	a.Show(2)
	a.done()
	l := a.navs[2]
	l.Focus(0)
	l.Enter()
	a.done()
	for range 15 {
		l.Next()
		a.done()
	}
	for range 15 {
		l.Prev()
		a.done()
	}
	l.Back()
	a.done()
	a.log.Append(a.lines[0])
	a.log.Append(a.lines[1])
	a.done()
	l.Focus(1)
	a.log.Interact(ui.UP)
	a.log.Interact(ui.DOWN)
	a.done()
	// Overlays: the dropdown's list and a dialog
	a.Show(3)
	a.done()
	a.dropdown.Interact(ui.ENTER)
	a.done()
	a.overlay.Interact(ui.DOWN)
	a.done()
	a.overlay.Interact(ui.ENTER)
	a.done()
	a.overlay.OpenCentered(a.dialog)
	a.done()
	a.overlay.Interact(ui.RIGHT)
	a.done()
	a.overlay.Interact(ui.ENTER)
	a.done()
	// Back through the page history
	a.pages.Back()
	a.done()
	a.pages.Back()
	a.done()
}
