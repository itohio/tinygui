package app

import (
	"image/color"
	"time"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/animation"
	"github.com/itohio/tinygui/container"
)

// Example-local widgets. They show how an app extends TinyGUI: embed
// ui.WidgetBase, implement Draw/Interact, opt into capability interfaces.

// static hides a widget from navigation (ui.Selectable).
type static struct{ ui.Widget }

func (static) CanSelect() bool { return false }

func still(w ui.Widget) ui.Widget { return static{w} }

// field is a focusable row: caption on the left, a child widget on the right,
// a focus frame when selected and an accent bar while the child is being
// edited. It forwards every capability hook the containers call.
type field struct {
	ui.WidgetBase
	caption  string
	captionW int16
	child    ui.Widget
	editing  bool
}

const fieldPad = 2

func newField(caption string, captionW int16, width uint16, child ui.Widget) *field {
	_, ch := child.Size()
	h := ch
	if h < 12 {
		h = 12
	}
	f := &field{
		WidgetBase: ui.NewWidgetBase(width, h+2*fieldPad),
		caption:    caption,
		captionW:   captionW,
		child:      child,
	}
	child.SetParent(f)
	return f
}

func (f *field) Name() string { return f.caption }

func (f *field) Draw(ctx ui.Context) {
	d := ctx.D()
	x, y := ctx.DisplayPos()
	w, h := int16(f.Width), int16(f.Height)
	if f.Selected() {
		fillRect(d, x, y, w, h, colPanel)
		strokeRect(d, x, y, w, h, colAccent)
	}
	// Thicker frame + side bar while the child owns the input. Containers keep
	// OnActivate state when focus moves inside an active container (findings
	// F-28), so only the focused row may show it.
	if f.editing && f.Selected() {
		strokeRect(d, x+1, y+1, w-2, h-2, colAccent)
		fillRect(d, x, y, 4, h, colAccent)
	}
	if f.caption != "" {
		c := colMuted
		if f.Selected() {
			c = colFG
		}
		ui.DrawText(d, fontBody, x+6, y+h/2+4, f.caption, c)
	}
	_, ch := f.child.Size()
	dx, dy := f.captionW+fieldPad+4, (h-int16(ch))/2
	ctx.AddPos(dx, dy)
	f.child.Draw(ctx)
	ctx.AddPos(-dx, -dy)
}

func (f *field) Interact(cmd ui.UserCommand) bool { return f.child.Interact(cmd) }

func (f *field) SetSelected(s bool) {
	f.WidgetBase.SetSelected(s)
	f.child.SetSelected(s)
}

func (f *field) OnSelect() {
	if h, ok := f.child.(ui.SelectHandler); ok {
		h.OnSelect()
	}
}

func (f *field) OnDeselect() {
	if h, ok := f.child.(ui.SelectHandler); ok {
		h.OnDeselect()
	}
}

func (f *field) OnActivate() {
	f.editing = true
	if h, ok := f.child.(ui.ActivationHandler); ok {
		h.OnActivate()
	}
}

func (f *field) OnDeactivate() {
	f.editing = false
	if h, ok := f.child.(ui.ActivationHandler); ok {
		h.OnDeactivate()
	}
}

func (f *field) OnExit() {
	if h, ok := f.child.(ui.ExitHandler); ok {
		h.OnExit()
	}
}

func (f *field) CanSelect() bool {
	if s, ok := f.child.(ui.Selectable); ok {
		return s.CanSelect()
	}
	return true
}

func (f *field) Enabled() bool {
	if e, ok := f.child.(ui.EnableState); ok {
		return e.Enabled()
	}
	return true
}

// heading draws text in the scaled title font.
type heading struct {
	ui.WidgetBase
	text  func() string
	color *color.RGBA
}

func newHeading(w, h uint16, text func() string, c *color.RGBA) *heading {
	return &heading{WidgetBase: ui.NewWidgetBase(w, h), text: text, color: c}
}

func (h *heading) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	ui.DrawText(ctx.D(), fontTitle, x, y+int16(h.Height)-4, h.text(), *h.color)
}

// swatch is a filled rectangle, the smallest possible custom widget.
type swatch struct {
	ui.WidgetBase
	c color.RGBA
}

func newSwatch(w, h uint16, c color.RGBA) *swatch {
	return &swatch{WidgetBase: ui.NewWidgetBase(w, h), c: c}
}

func (s *swatch) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	fillRect(ctx.D(), x, y, int16(s.Width), int16(s.Height), s.c)
}

// chip is a rounded-looking tag: background + text, width from the text.
type chip struct {
	ui.WidgetBase
	text string
}

func newChip(text string) *chip {
	return &chip{WidgetBase: ui.NewWidgetBase(uint16(textWidth(fontBody, text)+8), 12), text: text}
}

func (c *chip) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	d := ctx.D()
	fillRect(d, x+1, y, int16(c.Width)-2, 12, colTrack)
	fillRect(d, x, y+1, int16(c.Width), 10, colTrack)
	ui.DrawText(d, fontBody, x+4, y+9, c.text, colFG)
}

// sparkline is a line graph over a fixed ring of samples (no allocation).
type sparkline struct {
	ui.WidgetBase
	samples [128]float32
	n, head int
	min     float32 // fixed range when max > min, else auto-scale
	max     float32
}

func newSparkline(w, h uint16, min, max float32) *sparkline {
	return &sparkline{WidgetBase: ui.NewWidgetBase(w, h), min: min, max: max}
}

func (s *sparkline) Push(v float32) {
	s.samples[s.head] = v
	s.head = (s.head + 1) % len(s.samples)
	if s.n < len(s.samples) {
		s.n++
	}
}

func (s *sparkline) at(i int) float32 { // i = 0 oldest
	return s.samples[(s.head-s.n+i+len(s.samples))%len(s.samples)]
}

func (s *sparkline) Draw(ctx ui.Context) {
	d := ctx.D()
	x, y := ctx.DisplayPos()
	w, h := int16(s.Width), int16(s.Height)
	fillRect(d, x, y, w, h, colPanel)
	for gy := int16(1); gy < 4; gy++ { // dotted grid
		for gx := int16(0); gx < w; gx += 4 {
			d.SetPixel(x+gx, y+gy*h/4, colTrack)
		}
	}
	if s.n < 2 {
		return
	}
	lo, hi := s.min, s.max
	if hi <= lo {
		lo, hi = s.at(0), s.at(0)
		for i := 1; i < s.n; i++ {
			v := s.at(i)
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
		if hi-lo < 1e-3 {
			hi = lo + 1
		}
	}
	// plain function, not a closure: a capturing closure would put x, y, h,
	// lo and hi on the heap on every frame
	// one sample per pixel column, newest on the right
	count := s.n
	if count > int(w) {
		count = int(w)
	}
	first := s.n - count
	px := x + w - int16(count)
	prev := sparkY(s.at(first), lo, hi, y, h)
	fill := colAccent
	fill.R, fill.G, fill.B = fill.R/4, fill.G/4, fill.B/4
	for i := 0; i < count; i++ {
		cy := sparkY(s.at(first+i), lo, hi, y, h)
		fillRect(d, px, cy, 1, y+h-cy, fill)
		a, b := prev, cy
		if a > b {
			a, b = b, a
		}
		fillRect(d, px, a, 1, b-a+1, colAccent)
		prev = cy
		px++
	}
}

// sparkY maps v in [lo,hi] to a screen row inside [y, y+h).
func sparkY(v, lo, hi float32, y, h int16) int16 {
	f := (v - lo) / (hi - lo)
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return y + h - 1 - int16(f*float32(h-1))
}

// bars is a labelled bar chart over caller-owned counters.
type bars struct {
	ui.WidgetBase
	values *[6]int
	labels [6]string
}

func newBars(w, h uint16, values *[6]int, labels [6]string) *bars {
	return &bars{WidgetBase: ui.NewWidgetBase(w, h), values: values, labels: labels}
}

func (b *bars) Draw(ctx ui.Context) {
	d := ctx.D()
	x, y := ctx.DisplayPos()
	w, h := int16(b.Width), int16(b.Height)
	fillRect(d, x, y, w, h, colPanel)
	max := 1
	for _, v := range b.values {
		if v > max {
			max = v
		}
	}
	slot := w / int16(len(b.values))
	chartH := h - 12
	for i, v := range b.values {
		bh := int16(v * int(chartH-2) / max)
		bx := x + int16(i)*slot + 3
		fillRect(d, bx, y+chartH-bh, slot-6, bh, colAccent)
		ui.DrawText(d, fontBody, bx+(slot-6)/2-3, y+h-2, b.labels[i], colMuted)
	}
}

// scrollbar follows a Scroll container through container.ScrollObserver.
type scrollbar struct {
	ui.WidgetBase
	offset   int16
	content  int16
	viewport int16
}

func newScrollbar(h uint16, content, viewport int16) *scrollbar {
	return &scrollbar{WidgetBase: ui.NewWidgetBase(4, h), content: content, viewport: viewport}
}

func (s *scrollbar) OnScrollChange(c container.ScrollChange) { s.offset = c.OffsetY }

func (s *scrollbar) Draw(ctx ui.Context) {
	d := ctx.D()
	x, y := ctx.DisplayPos()
	h := int16(s.Height)
	fillRect(d, x, y, 4, h, colTrack)
	if s.content <= s.viewport {
		return
	}
	thumb := h * s.viewport / s.content
	top := h * s.offset / s.content
	fillRect(d, x, y+top, 4, thumb, colAccent)
}

// durationLabel edits a time.Duration with magnitude-scaled steps
// (ui.DurationAdd on the board, injected so the app also builds on the host).
type durationLabel struct {
	ui.WidgetBase
	value    *time.Duration
	pending  time.Duration
	add, sub func(d, delta time.Duration) time.Duration
	min, max time.Duration
	buf      textBuf
}

func newDurationLabel(w uint16, v *time.Duration, add, sub func(d, delta time.Duration) time.Duration) *durationLabel {
	return &durationLabel{WidgetBase: ui.NewWidgetBase(w, 12), value: v, pending: *v, add: add, sub: sub,
		min: 10 * time.Millisecond, max: 10 * time.Second}
}

func (l *durationLabel) OnSelect()     { l.pending = *l.value }
func (l *durationLabel) OnDeselect()   {}
func (l *durationLabel) OnExit()       { l.pending = *l.value }
func (l *durationLabel) Enabled() bool { return true }

func (l *durationLabel) Interact(cmd ui.UserCommand) bool {
	switch cmd {
	case ui.UP, ui.LONG_UP:
		l.pending = l.add(l.pending, 1)
	case ui.DOWN, ui.LONG_DOWN:
		l.pending = l.sub(l.pending, 1)
	case ui.ENTER:
		*l.value = l.pending
		l.SetSelected(false) // library editors end editing on commit
		return true
	case ui.ESC, ui.BACK:
		l.pending = *l.value
		return true
	default:
		return false
	}
	if l.pending < l.min {
		l.pending = l.min
	}
	if l.pending > l.max {
		l.pending = l.max
	}
	return true
}

func (l *durationLabel) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	ms := int(l.pending / time.Millisecond)
	l.buf.reset()
	if ms >= 1000 {
		l.buf.fixed(float32(ms)/1000, 2).str(" s")
	} else {
		l.buf.int(ms).str(" ms")
	}
	ui.DrawText(ctx.D(), fontBody, x, y+10, l.buf.String(), colFG)
}

// tabBar shows page names; the underline glides to the active tab with an
// ease-out animation (animation package, two channels: x and width).
type tabBar struct {
	ui.WidgetBase
	names  []string
	active int
	first  int // first visible tab
	anim   animation.Animator
	from   [2]float32
	to     [2]float32
	cur    [2]float32
	moving bool
}

func newTabBar(w, h uint16, names []string) *tabBar {
	t := &tabBar{WidgetBase: ui.NewWidgetBase(w, h), names: names, anim: animation.NewEaseOut(220_000)}
	t.cur = t.slot(0)
	return t
}

// tabsVisible tabs fit the bar; with more, the bar scrolls so the active tab
// stays in view (two tabs of context on its left where possible).
const tabsVisible = 6

func (t *tabBar) visible() int { return min(len(t.names), tabsVisible) }

func (t *tabBar) slot(i int) [2]float32 {
	sw := float32(t.Width) / float32(t.visible())
	return [2]float32{sw * float32(i-t.first), sw}
}

func (t *tabBar) setActive(i int, nowUS int64) {
	t.active = i
	t.first = min(max(i-2, 0), len(t.names)-t.visible())
	t.from = t.cur
	t.to = t.slot(i)
	t.anim.Start(t.from[:], t.to[:], nowUS)
	t.moving = true
}

func (t *tabBar) tick(nowUS int64) {
	if t.moving && t.anim.Update(t.cur[:], nowUS) {
		t.moving = false
	}
}

func (t *tabBar) Draw(ctx ui.Context) {
	d := ctx.D()
	x, y := ctx.DisplayPos()
	w, h := int16(t.Width), int16(t.Height)
	fillRect(d, x, y, w, h, colPanel)
	vis := t.visible()
	sw := w / int16(vis)
	for i := t.first; i < t.first+vis; i++ {
		c := colMuted
		if i == t.active {
			c = colFG
		}
		n := t.names[i]
		tw := textWidth(fontBody, n)
		ui.DrawText(d, fontBody, x+int16(i-t.first)*sw+(sw-tw)/2, y+h-8, n, c)
	}
	// more tabs beyond the edges
	if t.first > 0 {
		ui.DrawText(d, fontBody, x+1, y+h-8, "<", colMuted)
	}
	if t.first+vis < len(t.names) {
		ui.DrawText(d, fontBody, x+w-6, y+h-8, ">", colMuted)
	}
	fillRect(d, x+int16(t.cur[0])+4, y+h-3, int16(t.cur[1])-8, 3, colAccent)
}
