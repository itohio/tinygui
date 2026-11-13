package container_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/internal/check"
	"github.com/itohio/tinygui/layout"
)

// spot records where it was drawn (display coordinates).
type spot struct {
	ui.WidgetBase
	at [2]int16
}

func (s *spot) Draw(ctx ui.Context) {
	x, y := ctx.DisplayPos()
	s.at = [2]int16{x, y}
}
func (s *spot) Interact(ui.UserCommand) bool { return false }

func newSpot(w, h uint16) *spot { return &spot{WidgetBase: ui.NewWidgetBase(w, h)} }

// drawAt draws a container of w×h at display (10, 10) and returns where each
// child was drawn.
func drawAt(t *testing.T, w, h uint16, a layout.Arranger, kids ...*spot) [][2]int16 {
	t.Helper()
	ws := make([]ui.Widget, len(kids))
	for i, k := range kids {
		k.at = [2]int16{-999, -999}
		ws[i] = k
	}
	c := container.New[ui.Widget](w, h, container.WithArranger[ui.Widget](a), container.WithChildren[ui.Widget](ws...))
	ctx := ui.NewContext(nil, w, h, 10, 10) // the container's box, at display (10, 10)
	c.Draw(&ctx)
	out := make([][2]int16, len(kids))
	for i, k := range kids {
		out[i] = k.at
	}
	return out
}

// Review R-TG-1 L1: an arranged child whose position lies outside the
// container's box (negative, or beyond it) is drawn at that position, in
// full (nothing clips: findings F-13), not at the previous child's position,
// which is where SetPos's refusal left the cursor.
func TestArrangedChildrenKeepPositionsOutsideTheBox(t *testing.T) {
	t.Run("row centred across a box shorter than its tallest child", func(t *testing.T) {
		tall, short := newSpot(20, 20), newSpot(20, 4) // y = (10-20)/2 = -5
		check.Equal(t, [][2]int16{{10, 5}, {30, 13}},
			drawAt(t, 100, 10, layout.Row(0, layout.Start, layout.Center), tall, short))
	})
	t.Run("corner badge pulled out by a negative anchor offset", func(t *testing.T) {
		badge := newSpot(10, 10) // was drawn at the cursor's start, (10, 10)
		check.Equal(t, [][2]int16{{7, 7}},
			drawAt(t, 100, 40, &layout.Anchor{H: layout.Start, V: layout.Start, DX: -3, DY: -3}, badge))
	})
	t.Run("wrapped block centred in a box it overflows", func(t *testing.T) {
		a, b, c := newSpot(60, 20), newSpot(60, 20), newSpot(60, 20)
		got := drawAt(t, 70, 30, &layout.Wrap{Block: layout.Center}, a, b, c)
		check.Equal(t, [][2]int16{{10, 10 + (30-60)/2}, {10, 10 + (30-60)/2 + 20}, {10, 10 + (30-60)/2 + 40}}, got)
	})
	t.Run("row centred along a box its content overflows", func(t *testing.T) {
		w1, w2 := newSpot(60, 10), newSpot(60, 10)
		check.Equal(t, [][2]int16{{10 + (100-120)/2, 10}, {10 + (100-120)/2 + 60, 10}},
			drawAt(t, 100, 10, layout.Row(0, layout.Center, layout.Start), w1, w2))
	})
}

// A container of zero size takes its arranger's measured size; with an
// arranger that can't measure (Wrap: its height depends on its width) it
// would stay 0×0 and draw nothing, so construction panics with a clear
// message instead (review R-TG-1 L6).
func TestZeroSizeContainerNeedsAMeasuringArranger(t *testing.T) {
	c := container.New[ui.Widget](0, 0, container.WithArranger[ui.Widget](&layout.Table{Cols: 2}),
		container.WithChildren[ui.Widget](newSpot(10, 4), newSpot(30, 6), newSpot(20, 8)))
	w, h := c.Size()
	check.Equal(t, [2]uint16{60, 14}, [2]uint16{w, h})
	check.PanicsWithValue(t,
		"container.New: a zero width or height needs an arranger that can measure its children (layout.Measurer); give layout.Wrap a fixed size",
		func() {
			container.New[ui.Widget](0, 40, container.WithArranger[ui.Widget](&layout.Wrap{}),
				container.WithChildren[ui.Widget](newSpot(10, 4)))
		})
}

// seen records the visibility notifications it gets.
type seen struct {
	ui.WidgetBase
	got []bool
}

func (s *seen) Draw(ui.Context)              {}
func (s *seen) Interact(ui.UserCommand) bool { return false }
func (s *seen) OnVisible(v bool)             { s.got = append(s.got, v) }

// A child replaced in place (same index, same number of children) is told it
// is visible on its first draw, and the child it replaced that it is gone
// (verify TG-1-postcommit R1: the state was per index and only reset when
// the length changed, so the new child never heard OnVisible).
func TestChildReplacedInPlaceGetsOnVisible(t *testing.T) {
	old, other := &seen{WidgetBase: ui.NewWidgetBase(10, 10)}, &seen{WidgetBase: ui.NewWidgetBase(10, 10)}
	c := container.New[ui.Widget](100, 20, container.WithLayout[ui.Widget](layout.HList(2)), container.WithChildren[ui.Widget](old, other))
	ctx := ui.NewContext(nil, 100, 20, 0, 0)
	c.Draw(&ctx)
	check.Equal(t, []bool{true}, old.got)

	repl := &seen{WidgetBase: ui.NewWidgetBase(10, 10)}
	c.Items[0] = repl
	repl.SetParent(c)
	c.Draw(&ctx)
	check.Equal(t, []bool{true}, repl.got, "the new child hears OnVisible(true) on its first draw")
	check.Equal(t, []bool{true, false}, old.got, "the replaced child hears it is gone")
	check.Equal(t, []bool{true}, other.got, "the other child hears nothing new")
	check.Zero(t, testing.AllocsPerRun(20, func() { c.Draw(&ctx) }))
}
