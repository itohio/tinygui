package container

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
	"github.com/stretchr/testify/require"
)

type dummyWidget struct {
	ui.WidgetBase
}

func newDummyWidget(w, h uint16) *dummyWidget {
	return &dummyWidget{WidgetBase: ui.NewWidgetBase(w, h)}
}

func (d *dummyWidget) Draw(ui.Context) {}

func (d *dummyWidget) Interact(ui.UserCommand) bool { return false }

type recordingObserver struct {
	last ScrollChange
}

func (r *recordingObserver) OnScrollChange(change ScrollChange) {
	r.last = change
}

func TestScrollClampsOffsets(t *testing.T) {
	widget := newDummyWidget(50, 50)
	sc := NewScroll(20, 20, layout.HList(0), widget)

	require.True(t, sc.Scroll(15, 15))
	ox, oy := sc.ScrollOffset()
	require.Equal(t, int16(15), ox)
	require.Equal(t, int16(15), oy)

	require.True(t, sc.Scroll(100, 100))
	ox, oy = sc.ScrollOffset()
	require.Equal(t, int16(30), ox)
	require.Equal(t, int16(30), oy)
}

func TestScrollObservers(t *testing.T) {
	widget := newDummyWidget(100, 20)
	sc := NewScroll(20, 20, layout.HList(0), widget)
	rec := &recordingObserver{}
	sc.AddObserver(rec)

	require.True(t, sc.Scroll(10, 0))
	require.Equal(t, int16(10), rec.last.DX)
	require.Equal(t, int16(0), rec.last.DY)
	require.Equal(t, int16(10), rec.last.OffsetX)
	require.Equal(t, int16(0), rec.last.OffsetY)
}

type drawRecorder struct {
	ui.WidgetBase
	drawnAt [][2]int16
}

func (d *drawRecorder) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	d.drawnAt = append(d.drawnAt, [2]int16{x, y})
}

// Content below the viewport must be drawn once it is scrolled into view.
func TestScrollDrawsContentBeyondViewport(t *testing.T) {
	rows := make([]*drawRecorder, 6)
	widgets := make([]ui.Widget, len(rows))
	for i := range rows {
		rows[i] = &drawRecorder{WidgetBase: ui.NewWidgetBase(20, 10)}
		widgets[i] = rows[i]
	}
	sc := NewScroll(20, 30, layout.VList(0), widgets...) // 3 of 6 rows visible
	require.True(t, sc.Scroll(0, 30))                    // rows 3..5 now visible

	ctx := ui.NewContext(nil, 100, 100, 0, 0)
	sc.Draw(&ctx)

	for i := 3; i < 6; i++ {
		require.Len(t, rows[i].drawnAt, 1, "row %d", i)
		require.Equal(t, [2]int16{0, int16(i-3) * 10}, rows[i].drawnAt[0], "row %d", i)
	}
	for i := 0; i < 3; i++ {
		require.Empty(t, rows[i].drawnAt, "row %d is scrolled out", i)
	}
}
