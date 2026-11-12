package refapp

import (
	"image/color"
	"io"
	"runtime"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/animation"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

func mallocs() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Mallocs
}

var countProbe []*[16]byte

// needMallocCount skips an allocation test on a runtime that counts no
// allocations: TinyGo's default host GC (Boehm) reports Mallocs = 0, which
// would pass every check. Run TinyGo with `tinygo test -gc=leaking`, which
// counts each one (as the badge's own GC does).
func needMallocCount(t *testing.T) {
	t.Helper()
	m0 := mallocs()
	countProbe = append(countProbe, new([16]byte))
	if mallocs() == m0 {
		t.Skip("this runtime counts no allocations (TinyGo's host Boehm GC): run tinygo test -gc=leaking")
	}
}

// The menu-driven soak allocates nothing after setup: every library path
// (draw, input, navigation, pages, layout, overlays, animation, text) runs
// allocation-free, so the app runs with the garbage collector disabled.
//
// Under TinyGo (`tinygo test ./internal/refapp`, and with `-gc=leaking`) the
// count must be exactly 0. Go's own runtime fills its per-call-site
// type-assertion caches lazily (at random, on misses) and allocates them at
// type assertions anywhere, library code included; so on Go the check is
// testing.AllocsPerRun's: fewer than one allocation per step on average.
func TestSoakDoesNotAllocate(t *testing.T) {
	needMallocCount(t)
	a := New()
	steps := 0
	a.AfterStep = func(int) { steps++ }
	for range 3 { // the first rounds size buffers lazily
		a.Round()
	}
	steps = 0
	var bad, at int
	var last uint64
	if runtime.Compiler != "gc" { // TinyGo: name the first step that allocates
		a.AfterStep = func(s int) {
			steps++
			if m := mallocs(); m != last && bad == 0 {
				bad, at = int(m-last), s
			}
			last = mallocs()
		}
	}
	last = mallocs()
	m0 := last
	for range 20 {
		a.Round()
	}
	n := mallocs() - m0
	if runtime.Compiler != "gc" {
		if n != 0 {
			t.Fatalf("%d allocations in 20 soak rounds after setup (first: %d at step %d)", n, bad, at)
		}
		return
	}
	if n/uint64(steps) != 0 {
		t.Fatalf("%d allocations in %d soak steps after setup", n, steps)
	}
}

// display is a 320×80 band framebuffer drawn as band 0.
func display() (*framebuffer.RGB565, ui.ContextImpl) {
	fb := framebuffer.NewBand(W, H, BandH, make([]uint16, W*BandH), nil)
	return fb, ui.NewContext(fb, W, BandH, 0, 0)
}

// Every widget draws, and handles its commands, without allocating
// (testing.AllocsPerRun == 0 per draw and per interaction), on Go and on
// TinyGo.
func TestWidgetsDrawAndInteractWithoutAllocating(t *testing.T) {
	needMallocCount(t)
	_, ctx := display()
	var (
		f32           float32 = 40
		i32                   = 7
		mix                   = []float32{20, 30, 25}
		idx, pick, dd int
		on            bool
		cells         = make([]uint8, 64)
		samples       = make([]uint16, 300)
		pix16         = make([]uint16, 16*16)
		pix8          = make([]uint8, 16*16)
		lines         = []string{"one", "two", "three"}
	)
	cols := []color.RGBA{fg, acc, good}
	ml := widget.NewMultilineLabel(100, 11, 3, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	ml.SetLines(lines)
	log := widget.NewInteractiveLog(100, 11, 3, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	log.SetCapacity(8)
	sc := widget.NewScope(100, 50, bg, track, track, good, acc)
	sc.Samples, sc.Hi, sc.Trigger, sc.TrigAt = samples, 50, 20, 30
	rows := []ui.Widget{widget.NewLabel(80, 11, font, func() string { return "row" }, fg),
		widget.NewLabel(80, 11, font, func() string { return "row" }, fg)}
	host := container.NewOverlay(container.New[ui.Widget](W, BandH))
	ddw := widget.NewDropdown(60, style, []string{"a", "b"}, &dd, nil)
	host.Base.(*container.Base[ui.Widget]).Items = []ui.Widget{ddw}
	ddw.SetParent(host.Base)
	host.Draw(&ctx)
	dialog := widget.NewDialog(style, "Title", []string{"text"}, []string{"Yes", "No"}, func(int) {}, func() {})
	anim := animation.NewEaseInOut(1000)
	animV, from, to := [1]float32{}, [1]float32{0}, [1]float32{100}
	anim.Start(from[:], to[:], 0)
	var now int64

	cases := []struct {
		name string
		w    ui.Widget
		cmds []ui.UserCommand
	}{
		{"Label", widget.NewLabel(80, 11, font, func() string { return "label" }, fg), nil},
		{"MultilineLabel", ml, nil},
		{"InteractiveLog", log, []ui.UserCommand{ui.UP, ui.DOWN}},
		{"Separator", widget.NewSeparator(80, 3, track), nil},
		{"Gauge", widget.NewGauge[float32](80, 10, &f32, 0, 100, acc, track), nil},
		{"MultiGauge", widget.NewMultiGauge[float32](80, 10, 0, 100, &mix, cols, track, fg), nil},
		{"VolumeGauge", widget.NewVolumeGauge[float32](80, 12, &f32, 0, 100, 10, good, track), nil},
		{"SolidGauge", widget.NewSolidGauge[float32](80, 12, &f32, 0, 100, acc, track), nil},
		{"Bitmap16", widget.NewBitmap16(16, 16, pix16), nil},
		{"Bitmap8", widget.NewBitmap8(16, 16, pix8), nil},
		{"CellGrid", widget.NewCellGrid(8, 8, 4, 1, widget.CellRound, cells, cols, bg), nil},
		{"Scope", sc, nil},
		{"Toggle", widget.NewToggle(60, 12, font, fg, "ON", "OFF", good, track, func() bool { return on }, func(v bool) { on = v }),
			[]ui.UserCommand{ui.ENTER, ui.ENTER}},
		{"InteractiveLabel[int]", widget.NewInteractiveLabel[int](80, 11, widget.WithValue(&i32), widget.WithRange(0, 100),
			widget.WithFont[int](font), widget.WithTextColor[int](fg)), []ui.UserCommand{ui.UP, ui.LONG_UP, ui.DOWN, ui.ESC}},
		{"InteractiveLabel[float32]", widget.NewInteractiveLabel[float32](80, 11, widget.WithValue(&f32), widget.WithRange[float32](0, 100),
			widget.WithSteps[float32](0.5, 5), widget.WithFont[float32](font), widget.WithTextColor[float32](fg)), []ui.UserCommand{ui.UP, ui.DOWN, ui.ESC}},
		{"InteractiveLabelChoice", widget.NewInteractiveLabelChoice(80, 11, []string{"a", "b", "c"}, widget.WithLabelChoiceIndex(&idx),
			widget.WithLabelChoiceFont(font), widget.WithLabelChoiceColor(fg)), []ui.UserCommand{ui.UP, ui.DOWN}},
		{"InteractiveGauge", widget.NewInteractiveGauge[float32](80, 10, widget.WithValue(&f32), widget.WithRange[float32](0, 100),
			widget.WithForeground[float32](good), widget.WithBackground[float32](track)), []ui.UserCommand{ui.UP, ui.DOWN, ui.ESC}},
		{"InteractiveMultiGauge", widget.NewInteractiveMultiGauge[float32](80, 10, widget.WithValues(&mix), widget.WithRange[float32](0, 100),
			widget.WithSegmentColors[float32](cols), widget.WithBackground[float32](track)), []ui.UserCommand{ui.UP, ui.ENTER, ui.DOWN, ui.ESC}},
		{"InteractiveWidgetChoice", widget.NewInteractiveWidgetChoice[ui.Widget](20, 16,
			[]ui.Widget{widget.NewBitmap16(16, 16, pix16), widget.NewBitmap8(16, 16, pix8)}, widget.WithWidgetChoiceIndex[ui.Widget](&pick)),
			[]ui.UserCommand{ui.UP, ui.DOWN}},
		{"Base(HFlow)", container.New[ui.Widget](200, 40, container.WithLayout[ui.Widget](layout.HFlow(2, 200)),
			container.WithChildren[ui.Widget](widget.NewLabel(60, 11, font, func() string { return "a" }, fg), widget.NewSeparator(60, 3, fg))),
			[]ui.UserCommand{ui.NEXT, ui.PREV}},
		{"Base(Wrap)", container.New[ui.Widget](200, 40, container.WithArranger[ui.Widget](&layout.Wrap{Gap: 2}),
			container.WithChildren[ui.Widget](widget.NewSeparator(60, 3, fg), widget.NewSeparator(60, 3, fg))), nil},
		{"ScrollChoice", container.NewScrollChoice(90, 20, layout.VList(1), rows), []ui.UserCommand{ui.DOWN, ui.UP}},
		{"ScrollChoice of containers", container.NewScrollChoice(90, 20, layout.VList(1), []ui.Widget{
			container.New[ui.Widget](80, 12, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](widget.NewSeparator(30, 3, fg))),
			container.New[ui.Widget](80, 12, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](widget.NewSeparator(30, 3, fg))),
		}), []ui.UserCommand{ui.DOWN, ui.UP}},
		{"Overlay+Dropdown", host, nil},
	}
	for _, c := range cases {
		c.w.Draw(&ctx) // first draw sizes the widget's buffers
		if n := testing.AllocsPerRun(50, func() { c.w.Draw(&ctx) }); n != 0 {
			t.Errorf("%s: %.0f allocations per Draw", c.name, n)
		}
		if len(c.cmds) == 0 {
			continue
		}
		c.w.SetSelected(true)
		if n := testing.AllocsPerRun(50, func() {
			for _, cmd := range c.cmds {
				c.w.Interact(cmd)
			}
			c.w.Draw(&ctx)
		}); n != 0 {
			t.Errorf("%s: %.0f allocations per command sequence", c.name, n)
		}
	}
	if n := testing.AllocsPerRun(50, func() {
		ddw.Interact(ui.ENTER)
		host.Interact(ui.DOWN)
		host.DrawLayers(&ctx)
		host.Interact(ui.ENTER)
		host.OpenCentered(dialog)
		host.DrawLayers(&ctx)
		host.Interact(ui.RIGHT)
		host.Interact(ui.ENTER)
	}); n != 0 {
		t.Errorf("overlays (dropdown list, dialog): %.0f allocations", n)
	}
	if n := testing.AllocsPerRun(50, func() {
		now += 100
		if anim.Update(animV[:], now) {
			from, to = to, from
			anim.Start(from[:], to[:], now)
		}
	}); n != 0 {
		t.Errorf("animation: %.0f allocations per update", n)
	}
	if n := testing.AllocsPerRun(50, func() { log.Append("a log line") }); n != 0 {
		t.Errorf("Log.Append: %.0f allocations", n)
	}
	// a screensaver: a RandomContext moved every frame, its text drawn through
	// a reused child context (verify H-2: Clone allocated per frame)
	fb, _ := display()
	saver := ui.NewRandomContext(fb, 0, 100, 12)
	text := widget.NewLabel(100, 12, font, func() string { return "saver" }, fg)
	var sctx ui.ContextImpl
	if n := testing.AllocsPerRun(50, func() {
		saver.Move()
		text.Draw(ui.ChildContext(&saver.ContextImpl, &sctx, text, 100, 12))
	}); n != 0 {
		t.Errorf("screensaver: %.0f allocations per frame", n)
	}
}

// console is a reader replaying command lines, then io.EOF.
type console struct {
	s string
	i int
}

func (c *console) Read(b []byte) (int, error) {
	if c.i >= len(c.s) {
		return 0, io.EOF
	}
	n := copy(b, c.s[c.i:])
	c.i += n
	return n, nil
}

// The serial command mux reads and dispatches lines without allocating
// (it converted every line to a string for the lookup).
func TestCommandMuxDispatchesWithoutAllocating(t *testing.T) {
	needMallocCount(t)
	var got, arg int
	r := &console{s: "up\nbl 40\r\nstats\nnope x\n\n"}
	m := ui.NewCommandStreamMux(r, map[string]func([]byte){
		"up": func([]byte) { got++ }, "stats": func([]byte) { got++ },
		"bl": func(a []byte) { got++; arg = len(a) },
	})
	run := func() { r.i = 0; _ = m.Run() }
	run()
	if got != 3 || arg != 2 {
		t.Fatalf("dispatched %d commands (bl arg length %d), want 3 (2)", got, arg)
	}
	if n := testing.AllocsPerRun(50, run); n != 0 {
		t.Errorf("command mux: %.0f allocations per 5 lines", n)
	}
}

// A widget's first draw (a page shown for the first time) allocates
// nothing: containers and grids size their buffers when built (owner: no
// allocations after setup; first draws used to size them). 20 fresh copies
// of each, built first, then each drawn once: fewer than one allocation per
// first draw (exactly none on TinyGo).
func TestFirstDrawDoesNotAllocate(t *testing.T) {
	needMallocCount(t)
	_, ctx := display()
	cells := make([]uint8, 64)
	makers := []struct {
		name string
		make func() ui.Widget
	}{
		{"Base(Wrap)", func() ui.Widget {
			return container.New[ui.Widget](200, 40, container.WithArranger[ui.Widget](&layout.Wrap{Gap: 2}),
				container.WithChildren[ui.Widget](widget.NewSeparator(60, 3, fg), widget.NewSeparator(60, 3, fg)))
		}},
		{"Base(HFlow)", func() ui.Widget {
			return container.New[ui.Widget](200, 40, container.WithLayout[ui.Widget](layout.HFlow(2, 200)),
				container.WithChildren[ui.Widget](widget.NewSeparator(60, 3, fg), widget.NewSeparator(60, 3, fg)))
		}},
		{"CellGrid", func() ui.Widget {
			return widget.NewCellGrid(8, 8, 4, 1, widget.CellRound, cells, []color.RGBA{fg, bg}, bg)
		}},
	}
	const copies = 20
	for _, mk := range makers {
		ws := make([]ui.Widget, copies)
		for i := range ws {
			ws[i] = mk.make()
		}
		m0 := mallocs()
		for _, w := range ws {
			w.Draw(&ctx)
		}
		n := mallocs() - m0
		if n != 0 && runtime.Compiler != "gc" || n >= copies {
			t.Errorf("%s: %d allocations in %d first draws", mk.name, n, copies)
		}
	}
}
