package container_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/stretchr/testify/require"
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
	f.pages = container.NewPages(10, 10, specs, container.PagesConfig{Keep: keep, History: history,
		OnRelease: func(i int, w ui.Widget) {
			if w == nil {
				panic("OnRelease with no widget")
			}
			f.released = append(f.released, i)
		}})
	return f
}

// Pages are built when first shown, once while they stay built, and at
// most Keep stay built: the least recently shown is released, never the
// page on screen.
func TestPagesBuildLazilyAndKeepABoundedNumber(t *testing.T) {
	f := newPagesFixture(6, 3, 8)
	p := f.pages
	require.Equal(t, 0, p.Built())
	require.Nil(t, p.Current())

	for _, i := range []int{0, 1, 2, 1, 0} {
		p.Show(i)
		require.Equal(t, i, p.Index())
	}
	require.Equal(t, []int{1, 1, 1, 0, 0, 0}, f.builds, "each built once while kept")
	require.Equal(t, 3, p.Built())
	require.Empty(t, f.released)

	p.Show(3) // 2 is the least recently shown
	require.Equal(t, []int{2}, f.released)
	require.Nil(t, p.Page(2))
	require.NotNil(t, p.Page(3))
	require.Equal(t, 3, p.Built())

	p.Show(2) // rebuilt; now 1 is the oldest
	require.Equal(t, []int{2, 1}, f.released)
	require.Equal(t, 2, f.builds[2])

	f2 := newPagesFixture(4, 1, 4) // Keep 1: only the page on screen
	for i := range 4 {
		f2.pages.Show(i)
		require.NotNil(t, f2.pages.Current())
		require.Equal(t, 1, f2.pages.Built())
	}
	require.Equal(t, []int{0, 1, 2}, f2.released)
}

// Back returns through the pages shown before, newest first, at most
// History of them; released pages are rebuilt on the way.
func TestPagesBackThroughABoundedHistory(t *testing.T) {
	f := newPagesFixture(6, 2, 3)
	p := f.pages
	for i := range 6 {
		p.Show(i)
	}
	require.Equal(t, 3, p.HistoryLen())
	var got []int
	for p.Back() {
		got = append(got, p.Index())
		require.NotNil(t, p.Current())
	}
	require.Equal(t, []int{4, 3, 2}, got, "the 3 most recent, newest first")
	require.Equal(t, 2, f.builds[2], "2 had been released and is rebuilt")
	require.False(t, p.Back())
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
	require.Zero(t, allocs)
	require.Equal(t, []int{1, 1, 1}, f.builds)
}
