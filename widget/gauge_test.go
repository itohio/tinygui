package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

func TestHorizontalGaugeUsesValuePointer(t *testing.T) {
	val := float32(0.5)
	g := NewHorizontalGauge[float32](20, 6, &val, 0, 1, color.RGBA{255, 0, 0, 255}, color.RGBA{})

	require.NotNil(t, g)
	val = 0.75
	ctx := ui.NewContext(nil, 20, 6, 0, 0)
	require.NotPanics(t, func() { g.Draw(&ctx) })
}

func TestVerticalGaugeClampsValue(t *testing.T) {
	val := int16(200)
	g := NewVerticalGauge[int16](8, 40, &val, 0, 100, color.RGBA{0, 255, 0, 255}, color.RGBA{})

	ctx := ui.NewContext(nil, 8, 40, 0, 0)
	require.NotPanics(t, func() { g.Draw(&ctx) })
}
