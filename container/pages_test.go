package container_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/internal/check"
)

type pageWidget struct {
	ui.WidgetBase
	drawn int
}

func (w *pageWidget) Draw(ui.Context)              { w.drawn++ }
func (w *pageWidget) Interact(ui.UserCommand) bool { return true }

type pagesFixture struct {
	pages    *container.Pages
	builds   []int
	released []int
}

func newPagesFixture(n, keep, history int) *pagesFixture {
	f := &pagesFixture{builds: make([]int, n)}
	specs := make([]container.PageSpec, n)
	for i := range specs {
		specs[i] = container.PageSpec{Name: string(rune('A' + i)), Build: func() ui.Widget {
			f.builds[i]++
			return &pageWidget{WidgetBase: ui.NewWidgetBase(10, 10)}
		}}
	}
	f.pages = container.NewPages(10, 10, specs, container.WithKeep(keep), container.WithHistory(history),
		container.WithOnRelease(func(i int, w ui.Widget) {
			if w == nil {
				panic("OnRelease with no widget")
			}
			f.released = append(f.released, i)
		}))
	return f
}

// Pages are built when first shown, once while they stay built, and at
// most Keep stay built: the least recently shown is released, never the
// page on screen.
func TestPagesBuildLazilyAndKeepABoundedNumber(t *testing.T) {
	f := newPagesFixture(6, 3, 8)
	p := f.pages
	check.Equal(t, 0, p.Built())
	check.Nil(t, p.Current())

	for _, i := range []int{0, 1, 2, 1, 0} {
		p.Show(i)
		check.Equal(t, i, p.Index())
	}
	check.Equal(t, []int{1, 1, 1, 0, 0, 0}, f.builds, "each built once while kept")
	check.Equal(t, 3, p.Built())
	check.Empty(t, f.released)

	p.Show(3) // 2 is the least recently shown
	check.Equal(t, []int{2}, f.released)
	check.Nil(t, p.Page(2))
	check.NotNil(t, p.Page(3))
	check.Equal(t, 3, p.Built())

	p.Show(2) // rebuilt; now 1 is the oldest
	check.Equal(t, []int{2, 1}, f.released)
	check.Equal(t, 2, f.builds[2])

	f2 := newPagesFixture(4, 1, 4) // Keep 1: only the page on screen
	for i := range 4 {
		f2.pages.Show(i)
		check.NotNil(t, f2.pages.Current())
		check.Equal(t, 1, f2.pages.Built())
	}
	check.Equal(t, []int{0, 1, 2}, f2.released)
}

// Back returns through the pages shown before, newest first, at most
// History of them; released pages are rebuilt on the way.
func TestPagesBackThroughABoundedHistory(t *testing.T) {
	f := newPagesFixture(6, 2, 3)
	p := f.pages
	for i := range 6 {
		p.Show(i)
	}
	check.Equal(t, 3, p.HistoryLen())
	var got []int
	for p.Back() {
		got = append(got, p.Index())
		check.NotNil(t, p.Current())
	}
	check.Equal(t, []int{4, 3, 2}, got, "the 3 most recent, newest first")
	check.Equal(t, 2, f.builds[2], "2 had been released and is rebuilt")
	check.False(t, p.Back())
}

// Showing built pages, going back and drawing allocate nothing in Pages.
func TestPagesNavigateWithoutAllocating(t *testing.T) {
	f := newPagesFixture(3, 3, 4)
	p := f.pages
	for i := range 3 {
		p.Show(i)
	}
	ctx := ui.NewContext(nil, 10, 10, 0, 0)
	allocs := testing.AllocsPerRun(50, func() {
		p.Show(0)
		p.Show(1)
		p.Draw(&ctx)
		p.Interact(ui.UP)
		p.Back()
		p.Show(2)
	})
	check.Zero(t, allocs)
	check.Equal(t, []int{1, 1, 1}, f.builds)
}

// BuildAll builds every page up front (in setup): showing pages afterwards,
// back and forth, builds nothing again, releases nothing and allocates
// nothing (owner: no allocations after setup).
func TestPagesBuildAllUpFront(t *testing.T) {
	f := newPagesFixture(5, 5, 4)
	p := f.pages
	p.BuildAll()
	check.Equal(t, []int{1, 1, 1, 1, 1}, f.builds)
	check.Equal(t, 5, p.Built())
	check.Equal(t, -1, p.Index(), "nothing shown yet")
	walk := func() {
		for i := range 5 {
			p.Show(i)
		}
		p.Back()
		p.Back()
	}
	walk()
	check.Equal(t, []int{1, 1, 1, 1, 1}, f.builds)
	check.Empty(t, f.released)
	check.Zero(t, testing.AllocsPerRun(20, walk))
}

// BuildAll with Keep smaller than the page count would build pages only to
// release them: it panics instead (a setup mistake).
func TestPagesBuildAllNeedsRoomForEveryPage(t *testing.T) {
	check.Panics(t, func() { newPagesFixture(5, 3, 4).pages.BuildAll() })
}

// With no options Pages keeps every page (the GC-free default; WithKeep
// lowers it): BuildAll succeeds, and walking every page twice afterwards,
// drawing each, builds nothing again and allocates nothing — counted
// exactly from right after BuildAll, not averaged.
func TestPagesDefaultKeepsEveryPage(t *testing.T) {
	const n = 5
	builds := make([]int, n)
	specs := make([]container.PageSpec, n)
	for i := range specs {
		specs[i] = container.PageSpec{Name: string(rune('A' + i)), Build: func() ui.Widget {
			builds[i]++
			return &pageWidget{WidgetBase: ui.NewWidgetBase(10, 10)}
		}}
	}
	p := container.NewPages(10, 10, specs)
	p.BuildAll() // panics when the default keeps fewer than every page
	check.Equal(t, n, p.Built())
	ctx := ui.NewContext(nil, 10, 10, 0, 0)
	m0 := mallocCount()
	for range 2 {
		for i := range n {
			p.Show(i)
			p.Draw(&ctx)
		}
	}
	allocs := mallocCount() - m0
	check.Equal(t, uint64(0), allocs, "allocations walking built pages")
	check.Equal(t, []int{1, 1, 1, 1, 1}, builds, "each page built once")
	check.Equal(t, n, p.Built())
}
