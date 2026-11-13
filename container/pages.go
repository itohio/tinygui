package container

import ui "github.com/itohio/tinygui"

// PageSpec is a page Pages builds the first time it is shown.
type PageSpec struct {
	Name  string
	Build func() ui.Widget
}

// PagesOption customises NewPages.
type PagesOption func(*pagesConfig)

type pagesConfig struct {
	keep, history int
	onRelease     func(int, ui.Widget)
}

// WithKeep sets how many built pages stay in memory (at least 1: the one
// shown). The default keeps every page: once built a page is never built
// again, so with BuildAll in setup nothing is built (allocated) later — the
// choice for an app that runs without a GC, at the cost of every page
// resident. Lower it to save memory: showing a page beyond n releases the
// least recently shown one, and showing a released page builds it again
// (the app's Build allocates every time; refapp
// TestPagesWithoutGCNeedEveryPageKept), and BuildAll needs n ≥ the page
// count.
func WithKeep(n int) PagesOption {
	return func(c *pagesConfig) { c.keep = max(n, 1) }
}

// WithHistory sets how many pages Back can return through (oldest dropped;
// default 8, at least 1).
func WithHistory(n int) PagesOption {
	return func(c *pagesConfig) { c.history = max(n, 1) }
}

// WithOnRelease sets fn to be called with a page's index and widget when
// Pages drops it, before its reference is cleared: the place to drop the
// app's own references so the page can be collected.
func WithOnRelease(fn func(i int, w ui.Widget)) PagesOption {
	return func(c *pagesConfig) { c.onRelease = fn }
}

// Pages shows one page out of many and manages their memory: each page is
// built on demand from its PageSpec, at most Keep stay built, and a bounded
// history lets Back return to the pages shown before. Pages itself never
// allocates after NewPages (showing, going back, drawing); building a page
// is the page's own work.
type Pages struct {
	ui.WidgetBase
	specs     []PageSpec
	built     []ui.Widget // by index, nil when not built
	recent    []int       // built pages, most recently shown first
	hist      []int       // ring of pages shown before the current one
	histHead  int
	histN     int
	cur       int
	ctx       ui.ContextImpl
	onRelease func(int, ui.Widget)
}

// NewPages makes a w×h page area over specs; nothing is built until Show.
// By default every page is kept once built (WithKeep) and Back returns
// through 8 pages (WithHistory).
func NewPages(w, h uint16, specs []PageSpec, opts ...PagesOption) *Pages {
	cfg := pagesConfig{keep: max(len(specs), 1), history: 8}
	for _, o := range opts {
		o(&cfg)
	}
	return &Pages{
		WidgetBase: ui.NewWidgetBase(w, h),
		specs:      specs,
		built:      make([]ui.Widget, len(specs)),
		recent:     make([]int, 0, cfg.keep),
		hist:       make([]int, cfg.history),
		cur:        -1,
		onRelease:  cfg.onRelease,
	}
}

// Count is the number of pages.
func (p *Pages) Count() int { return len(p.specs) }

// Name is page i's name.
func (p *Pages) Name(i int) string { return p.specs[i].Name }

// Index is the page shown, -1 before the first Show.
func (p *Pages) Index() int { return p.cur }

// Current is the page shown (nil before the first Show).
func (p *Pages) Current() ui.Widget {
	if p.cur < 0 {
		return nil
	}
	return p.built[p.cur]
}

// Page is page i if it is built, else nil.
func (p *Pages) Page(i int) ui.Widget { return p.built[i] }

// Built is how many pages are built now (at most Keep).
func (p *Pages) Built() int { return len(p.recent) }

// BuildAll builds every page now (call it in setup; it needs the default
// Keep or WithKeep at least the page count): showing pages later builds and
// allocates nothing. Pages that are already built stay as they are; nothing
// is shown.
func (p *Pages) BuildAll() {
	if cap(p.recent) < len(p.specs) {
		panic("container.Pages.BuildAll: Keep is smaller than the page count")
	}
	for i := range p.specs {
		if p.built[i] == nil {
			w := p.specs[i].Build()
			w.SetParent(p)
			p.built[i] = w
			p.recent = append(p.recent, i) // within cap: no allocation
		}
	}
}

// Show shows page i, building it if needed, and records the page shown
// before in the history. It returns the page.
func (p *Pages) Show(i int) ui.Widget {
	if i == p.cur {
		return p.built[i]
	}
	if p.cur >= 0 {
		p.hist[p.histHead] = p.cur
		p.histHead = (p.histHead + 1) % len(p.hist)
		p.histN = min(p.histN+1, len(p.hist))
	}
	return p.show(i)
}

// Back shows the page shown before the current one; false when the
// history is empty.
func (p *Pages) Back() bool {
	if p.histN == 0 {
		return false
	}
	p.histHead = (p.histHead - 1 + len(p.hist)) % len(p.hist)
	p.histN--
	p.show(p.hist[p.histHead])
	return true
}

// HistoryLen is how many pages Back can return through.
func (p *Pages) HistoryLen() int { return p.histN }

func (p *Pages) show(i int) ui.Widget {
	p.cur = i
	if p.built[i] == nil {
		if len(p.recent) == cap(p.recent) {
			p.release(p.recent[len(p.recent)-1])
		}
		w := p.specs[i].Build()
		w.SetParent(p)
		p.built[i] = w
		p.recent = append(p.recent, i) // within cap: no allocation
	}
	p.touch(i)
	return p.built[i]
}

// touch moves page i to the front of recent.
func (p *Pages) touch(i int) {
	k := 0
	for k < len(p.recent) && p.recent[k] != i {
		k++
	}
	copy(p.recent[1:k+1], p.recent[:k])
	p.recent[0] = i
}

func (p *Pages) release(i int) {
	if p.onRelease != nil {
		p.onRelease(i, p.built[i])
	}
	p.built[i] = nil
	for k, j := range p.recent {
		if j == i {
			p.recent = append(p.recent[:k], p.recent[k+1:]...)
			break
		}
	}
	// the history may name released pages: they are rebuilt when Back reaches them
}

// Draw draws the page shown.
func (p *Pages) Draw(ctx ui.Context) {
	w := p.Current()
	if w == nil {
		return
	}
	pw, ph := w.Size()
	w.Draw(ui.ChildContext(ctx, &p.ctx, w, pw, ph))
}

// Interact passes cmd to the page shown.
func (p *Pages) Interact(cmd ui.UserCommand) bool {
	if w := p.Current(); w != nil {
		return w.Interact(cmd)
	}
	return false
}
