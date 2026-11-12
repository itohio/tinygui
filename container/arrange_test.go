package container

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
	"github.com/stretchr/testify/require"
)

// An arranged container draws each child where the arranger puts it, inside
// margin + padding, and lays out again when a child changes size.
func TestArrangerPlacesChildrenAndFollowsSizeChanges(t *testing.T) {
	a := &drawRecorder{WidgetBase: ui.NewWidgetBase(20, 10)}
	b := &drawRecorder{WidgetBase: ui.NewWidgetBase(20, 10)}
	c := New[ui.Widget](100, 40,
		WithPadding[ui.Widget](4, 2),
		WithArranger[ui.Widget](layout.Row(4, layout.Center, layout.Center)),
		WithChildren[ui.Widget](a, b))
	ctx := ui.NewContext(nil, 100, 40, 10, 20)
	c.Draw(&ctx)
	// inner box 92×36 from (4,2); children 44 wide in total → start at 24
	require.Equal(t, [2]int16{10 + 4 + 24, 20 + 2 + 13}, a.drawnAt[0])
	require.Equal(t, [2]int16{10 + 4 + 48, 20 + 2 + 13}, b.drawnAt[0])

	a.Width, a.Height = 40, 20 // e.g. a tile magnified on focus
	c.Draw(&ctx)
	// 64 wide in total → start at 14; a is centred across: (36-20)/2 = 8
	require.Equal(t, [2]int16{10 + 4 + 14, 20 + 2 + 8}, a.drawnAt[1])
	require.Equal(t, [2]int16{10 + 4 + 58, 20 + 2 + 13}, b.drawnAt[1])
}

func TestArrangerMeasuresAZeroSizedContainer(t *testing.T) {
	c := New[ui.Widget](0, 0,
		WithPadding[ui.Widget](3, 1),
		WithArranger[ui.Widget](layout.Column(2, layout.Start, layout.Start)),
		WithChildren[ui.Widget](
			&drawRecorder{WidgetBase: ui.NewWidgetBase(20, 10)},
			&drawRecorder{WidgetBase: ui.NewWidgetBase(30, 5)}))
	w, h := c.Size()
	require.Equal(t, [2]uint16{30 + 6, 17 + 2}, [2]uint16{w, h})
}

// Drawing a container tree into a plain context allocates nothing: children
// get a context the container keeps (ui.ChildContext), for Base and Scroll.
func TestDrawingContainersDoesNotAllocate(t *testing.T) {
	inner := New[ui.Widget](50, 20, WithLayout[ui.Widget](layout.HList(2)),
		WithChildren[ui.Widget](newDummyWidget(10, 10)))
	scroll := NewScroll(60, 20, layout.VList(1), newDummyWidget(10, 30))
	root := New[ui.Widget](100, 60,
		WithArranger[ui.Widget](layout.Column(2, layout.Start, layout.Start)),
		WithChildren[ui.Widget](inner, scroll))
	ctx := ui.NewContext(nil, 100, 60, 0, 0)
	root.Draw(&ctx)
	if n := testing.AllocsPerRun(20, func() { root.Draw(&ctx) }); n > 0 {
		t.Fatalf("%.0f allocations per draw", n)
	}
}
