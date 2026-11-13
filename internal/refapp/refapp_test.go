package refapp

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/animation"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

var countProbe []*[16]byte

// needMallocCount fails an allocation test on a runtime whose counter
// misses allocations: every check would pass. mallocs() counts on Go and on
// TinyGo under every collector (the default Boehm one included).
func needMallocCount(t *testing.T) {
	t.Helper()
	m0 := mallocs()
	countProbe = append(countProbe, new([16]byte))
	if mallocs() == m0 {
		t.Fatal("the allocation counter missed an allocation: the allocation checks cannot run")
	}
}

// allocsIn is the exact number of heap allocations made by n calls of f:
// a total, not testing.AllocsPerRun's per-run integer average, which
// passes an allocation every few runs.
func allocsIn(n int, f func()) uint64 {
	m0 := mallocs()
	for range n {
		f()
	}
	return mallocs() - m0
}

// muxLines is a console session for the command mux: commands with and
// without arguments, CRLF, an unknown command, an empty line.
const muxLines = "up\nbl 40\r\nstats\nnope x\n\nup\nbl 7\nstats\n"

// newMux is a command mux over a replayed console: run replays the session
// once; got counts dispatched commands, arg is the last bl argument's length.
func newMux() (run func(), got, arg *int) {
	got, arg = new(int), new(int)
	r := &console{s: muxLines}
	m := ui.NewCommandStreamMux(r, map[string]func([]byte){
		"up": func([]byte) { *got++ }, "stats": func([]byte) { *got++ },
		"bl": func(a []byte) { *got++; *arg = len(a) },
	})
	return func() { r.i = 0; _ = m.Run() }, got, arg
}

// The menu-driven soak allocates nothing after setup: every library path
// (draw, input, navigation, pages, layout, overlays, animation, text, the
// serial command mux) runs allocation-free, so the app runs with the
// garbage collector disabled. The count over 20 rounds must be exactly 0,
// on Go and on TinyGo (any collector).
func TestSoakDoesNotAllocate(t *testing.T) {
	needMallocCount(t)
	a := New()
	mux, _, _ := newMux()
	for range 3 { // the first rounds size buffers lazily
		a.Round()
		mux()
	}
	var bad, at int
	var last, inMux uint64
	a.AfterStep = func(s int) { // name the first step that allocates
		if m := mallocs(); m != last && bad == 0 {
			bad, at = int(m-last), s
		}
		last = mallocs()
	}
	last = mallocs()
	n := allocsIn(20, func() {
		a.Round()
		m := mallocs()
		mux()
		last = mallocs()
		inMux += last - m
	})
	if n != 0 {
		t.Fatalf("%d allocations in 20 soak rounds after setup (%d in the command mux; first step: %d at step %d)", n, inMux, bad, at)
	}
}

// display is a 320×80 band framebuffer drawn as band 0.
func display() (*framebuffer.RGB565, ui.ContextImpl) {
	fb := framebuffer.NewBand(W, H, BandH, make([]uint16, W*BandH), nil)
	return fb, ui.NewContext(fb, W, BandH, 0, 0)
}

// Every widget draws, and handles its commands, without allocating: exactly
// 0 allocations over 50 draws and over 50 command sequences, on Go and on
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
	icons := []string{pngOf(8, 8, acc), pngOf(8, 8, good), pngOf(8, 8, fg)}
	ml := widget.NewMultilineLabel(100, 11, 3, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	ml.SetLines(lines)
	log := widget.NewInteractiveLog(100, 11, 3, widget.WithMultilineFont(font), widget.WithMultilineColor(fg))
	log.SetCapacity(8)
	sc := widget.NewScope(100, 50, widget.ScopeStyle{BG: bg, Grid: track, Axis: track, Trace: good, Marker: acc})
	sc.Samples, sc.Hi, sc.Trigger, sc.TrigAt = samples, 50, 20, 30
	sb := widget.NewStatBand(162, 26, 3, statStyle, widget.WithBandZone(1))
	for i := range 200 { // wrapped: Draw reads across the ring buffer's end
		sb.Push(float32(i%7-3)/2, float32(i%5)/2, uint8(i%3))
	}
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
		{"Ring", widget.NewRing[float32](60, 60, &f32, 0, 100, ringStyle, widget.WithPercentLabel()), nil},
		{"Ring(round, ccw, custom label)", widget.NewRing[float32](50, 40, &f32, 0, 100, ringStyle,
			widget.WithCap(widget.RingCapRound), widget.WithCounterClockwise(), widget.WithStartAngle(225),
			widget.WithThickness(9), widget.WithLabel(func() string { return "cal" })), nil},
		{"Bitmap16", widget.NewBitmap16(16, 16, pix16), nil},
		{"Bitmap8", widget.NewBitmap8(16, 16, pix8), nil},
		{"CellGrid", widget.NewCellGrid(8, 8, 4, 1, widget.CellRound, cells, cols, bg), nil},
		{"Scope", sc, nil},
		{"StatBand", sb, nil},
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
		// icons decoded at construction: a press switches, a draw blits
		// (verify TG-1-postcommit R7: re-decoding per press, 32–48 allocations)
		{"InteractiveIcon", widget.NewInteractiveIcon(8, 8, icons), []ui.UserCommand{ui.UP, ui.DOWN, ui.UP}},
		{"InteractiveIconChoice", widget.NewInteractiveIconChoice(8, 8, icons), []ui.UserCommand{ui.UP, ui.DOWN, ui.UP}},
		{"Base(Wrap)", container.New[ui.Widget](200, 40, container.WithArranger[ui.Widget](&layout.Wrap{Gap: 2}),
			container.WithChildren[ui.Widget](widget.NewSeparator(60, 3, fg), widget.NewSeparator(60, 3, fg))), nil},
		{"ScrollChoice", container.NewScrollChoice(90, 20, layout.VList(1), rows), []ui.UserCommand{ui.DOWN, ui.UP}},
		{"ScrollChoice of containers", container.NewScrollChoice(90, 20, layout.VList(1), []ui.Widget{
			container.New[ui.Widget](80, 12, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](widget.NewSeparator(30, 3, fg))),
			container.New[ui.Widget](80, 12, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](widget.NewSeparator(30, 3, fg))),
		}), []ui.UserCommand{ui.DOWN, ui.UP}},
		{"Overlay+Dropdown", host, nil},
	}
	const runs = 50
	for _, c := range cases {
		c.w.Draw(&ctx) // first draw sizes the widget's buffers
		if n := allocsIn(runs, func() { c.w.Draw(&ctx) }); n != 0 {
			t.Errorf("%s: %d allocations in %d draws", c.name, n, runs)
		}
		if len(c.cmds) == 0 {
			continue
		}
		c.w.SetSelected(true)
		seq := func() {
			for _, cmd := range c.cmds {
				c.w.Interact(cmd)
			}
			c.w.Draw(&ctx)
		}
		seq()
		if n := allocsIn(runs, seq); n != 0 {
			t.Errorf("%s: %d allocations in %d command sequences", c.name, n, runs)
		}
	}
	overlays := func() {
		ddw.Interact(ui.ENTER)
		host.Interact(ui.DOWN)
		host.DrawLayers(&ctx)
		host.Interact(ui.ENTER)
		host.OpenCentered(dialog)
		host.DrawLayers(&ctx)
		host.Interact(ui.RIGHT)
		host.Interact(ui.ENTER)
	}
	overlays()
	if n := allocsIn(runs, overlays); n != 0 {
		t.Errorf("overlays (dropdown list, dialog): %d allocations in %d runs", n, runs)
	}
	step := func() {
		now += 100
		if anim.Update(animV[:], now) {
			from, to = to, from
			anim.Start(from[:], to[:], now)
		}
	}
	step()
	if n := allocsIn(runs, step); n != 0 {
		t.Errorf("animation: %d allocations in %d updates", n, runs)
	}
	pushes := 0
	push := func() { // a column per frame, as the calibration page does
		pushes++
		sb.Push(float32(pushes%9-4)/2, float32(pushes%4)/3, uint8(pushes%3))
		sb.Draw(&ctx)
	}
	push()
	if n := allocsIn(runs, push); n != 0 {
		t.Errorf("StatBand: %d allocations in %d pushes and draws", n, runs)
	}
	sb.Reset()
	if n := allocsIn(runs, push); n != 0 {
		t.Errorf("StatBand after Reset: %d allocations in %d pushes and draws", n, runs)
	}
	appendLine := func() { log.Append("a log line") }
	appendLine()
	if n := allocsIn(runs, appendLine); n != 0 {
		t.Errorf("Log.Append: %d allocations in %d appends", n, runs)
	}
	// a ring whose value changes every draw: the arc's geometry is
	// recomputed (float32 sine/cosine), the percent label reformatted
	ringV := float32(0)
	ring := widget.NewRing[float32](60, 60, &ringV, 0, 100, ringStyle, widget.WithPercentLabel(), widget.WithCap(widget.RingCapRound))
	fillRing := func() {
		ringV += 7.3
		if ringV > 110 {
			ringV = -10 // through both clamps
		}
		ring.Draw(&ctx)
	}
	fillRing()
	if n := allocsIn(runs, fillRing); n != 0 {
		t.Errorf("Ring with a changing value: %d allocations in %d draws", n, runs)
	}
	// a screensaver: a RandomContext moved every frame, its text drawn through
	// a reused child context (verify H-2: Clone allocated per frame)
	fb, _ := display()
	saver := ui.NewRandomContext(fb, 0, 100, 12)
	text := widget.NewLabel(100, 12, font, func() string { return "saver" }, fg)
	var sctx ui.ContextImpl
	frame := func() {
		saver.Move()
		text.Draw(ui.ChildContext(&saver.ContextImpl, &sctx, text, 100, 12))
	}
	frame()
	if n := allocsIn(runs, frame); n != 0 {
		t.Errorf("screensaver: %d allocations in %d frames", n, runs)
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
// (it converted every line to a string for the lookup): exactly 0
// allocations over 50 replays of an 8-line session.
func TestCommandMuxDispatchesWithoutAllocating(t *testing.T) {
	needMallocCount(t)
	run, got, arg := newMux()
	run()
	if *got != 6 || *arg != 1 {
		t.Fatalf("dispatched %d commands (last bl arg length %d), want 6 (1)", *got, *arg)
	}
	if n := allocsIn(50, run); n != 0 {
		t.Errorf("command mux: %d allocations in 50 replays of %q", n, muxLines)
	}
}

// A widget's first draw (a page shown for the first time) allocates
// nothing: containers and grids size their buffers when built (owner: no
// allocations after setup; first draws used to size them). 20 fresh copies
// of each, built first, then each drawn once: exactly 0 allocations.
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
		{"StatBand", func() ui.Widget {
			return widget.NewStatBand(162, 26, 3, statStyle, widget.WithBandZone(1))
		}},
	}
	const copies = 20
	for _, mk := range makers {
		ws := make([]ui.Widget, copies)
		for i := range ws {
			ws[i] = mk.make()
		}
		i := 0
		if n := allocsIn(copies, func() { ws[i].Draw(&ctx); i++ }); n != 0 {
			t.Errorf("%s: %d allocations in %d first draws", mk.name, n, copies)
		}
	}
}

// pngOf encodes a w×h PNG of one colour (setup for the icon cases).
func pngOf(w, h int, c color.RGBA) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		panic(err)
	}
	return b.String()
}

// Running without a GC needs Pages to keep every page (Keep ≥ the number of
// pages, then BuildAll in setup): with fewer kept, showing a released page
// builds it again — the app's Build allocates every time, so a GC-free heap
// grows for ever. Walking (Show, Back, Draw) through 5 pages: Keep 2 → the
// rebuilds allocate; every page kept and built in setup → nothing.
func TestPagesWithoutGCNeedEveryPageKept(t *testing.T) {
	needMallocCount(t)
	_, ctx := display()
	specs := make([]container.PageSpec, 5)
	for i := range specs {
		specs[i] = container.PageSpec{Name: "p", Build: func() ui.Widget {
			return container.New[ui.Widget](100, 40, container.WithLayout[ui.Widget](layout.VList(1)),
				container.WithChildren[ui.Widget](widget.NewSeparator(80, 3, fg)))
		}}
	}
	walk := func(p *container.Pages) func() {
		return func() {
			for i := range specs {
				p.Show(i)
				p.Draw(&ctx)
			}
			p.Back()
			p.Draw(&ctx)
		}
	}
	few := container.NewPages(100, 40, specs, container.WithKeep(2), container.WithHistory(4))
	walkFew := walk(few)
	walkFew()
	m0 := mallocs()
	walkFew()
	if n := mallocs() - m0; n == 0 {
		t.Fatal("Keep 2 of 5: a walk rebuilt pages without allocating?")
	}

	all := container.NewPages(100, 40, specs, container.WithHistory(4))
	all.BuildAll()
	walkAll := walk(all)
	walkAll()
	m0 = mallocs()
	for range 10 {
		walkAll()
	}
	if n := mallocs() - m0; n != 0 {
		t.Fatalf("every page kept and built in setup: %d allocations in 10 walks", n)
	}
}
