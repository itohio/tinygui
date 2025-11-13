package layout

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

type testSizer struct {
	w, h uint16
}

func newTestSizer(w, h uint16) *testSizer {
	return &testSizer{w: w, h: h}
}

func (t *testSizer) Size() (uint16, uint16) {
	return t.w, t.h
}

func TestGridWrapsByWidth(t *testing.T) {
	strategy := Grid(2, 1)
	ctx := ui.NewContext(nil, 10, 10, 0, 0)

	first := newTestSizer(4, 3)
	require.True(t, strategy(&ctx, first))
	x, y := ctx.Pos()
	require.Equal(t, int16(6), x)
	require.Equal(t, int16(0), y)

	second := newTestSizer(4, 3)
	require.True(t, strategy(&ctx, second))
	x, y = ctx.Pos()
	require.Equal(t, int16(0), x)
	require.Equal(t, int16(4), y)
}

func TestHFlowRespectsWidthLimit(t *testing.T) {
	strategy := HFlow(1, 8)
	ctx := ui.NewContext(nil, 20, 10, 0, 0)

	item := newTestSizer(3, 2)
	require.True(t, strategy(&ctx, item))
	x, y := ctx.Pos()
	require.Equal(t, int16(4), x)
	require.Equal(t, int16(0), y)

	require.True(t, strategy(&ctx, item))
	x, y = ctx.Pos()
	require.Equal(t, int16(0), x)
	require.Equal(t, int16(3), y)
}

func TestVFlowRespectsHeightLimit(t *testing.T) {
	strategy := VFlow(1, 5)
	ctx := ui.NewContext(nil, 10, 20, 0, 0)

	item := newTestSizer(2, 2)
	require.True(t, strategy(&ctx, item))
	x, y := ctx.Pos()
	require.Equal(t, int16(0), x)
	require.Equal(t, int16(3), y)

	require.True(t, strategy(&ctx, item))
	x, y = ctx.Pos()
	require.Equal(t, int16(3), x)
	require.Equal(t, int16(0), y)
}
