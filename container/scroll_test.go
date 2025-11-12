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
