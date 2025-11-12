package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

func TestHorizontalInteractiveMultiGaugeCommitAndNavigate(t *testing.T) {
	values := []float32{0.2, 0.8}
	bar := NewHorizontalInteractiveMultiGauge[float32](
		60, 8,
		WithValues(&values),
		WithRange[float32](0, 1),
		WithSteps[float32](0.1, 0.25),
		WithBackground[float32](color.RGBA{128, 128, 128, 255}),
		WithSegmentColors[float32](
			[]color.RGBA{
				{255, 0, 0, 255},
				{0, 255, 0, 255},
			},
		),
	)

	bar.SetSelected(true)
	require.Equal(t, 0, bar.active)
	require.True(t, bar.Interact(ui.UP))
	require.InEpsilon(t, 0.3, bar.pending[0], 0.001)

	require.True(t, bar.Interact(ui.ENTER))
	require.Equal(t, 1, bar.active)
	require.True(t, bar.Interact(ui.DOWN))
	require.InEpsilon(t, 0.7, bar.pending[1], 0.001)

	require.True(t, bar.Interact(ui.LEFT))
	require.Equal(t, 0, bar.active)

	require.True(t, bar.Interact(ui.RIGHT))
	require.Equal(t, 1, bar.active)

	require.True(t, bar.Interact(ui.ENTER))
	require.False(t, bar.Selected())
	require.InEpsilon(t, 0.3, values[0], 0.001)
	require.InEpsilon(t, 0.7, values[1], 0.001)
}

func TestHorizontalInteractiveGaugeCancel(t *testing.T) {
	value := float32(0.5)
	gauge := NewHorizontalInteractiveGauge[float32](
		40, 8,
		WithValue(&value),
		WithRange[float32](0, 1),
		WithSteps[float32](0.1, 0.25),
		WithForeground[float32](color.RGBA{255, 255, 255, 255}),
	)

	gauge.SetSelected(true)
	require.True(t, gauge.Interact(ui.UP))
	require.InEpsilon(t, 0.6, gauge.pending, 0.001)
	require.True(t, gauge.Interact(ui.ESC))
	require.InEpsilon(t, 0.5, value, 0.001)
}

func TestVerticalInteractiveGaugeDisable(t *testing.T) {
	value := float32(0.4)
	gauge := NewVerticalInteractiveGauge[float32](
		12, 60,
		WithValue(&value),
		WithRange[float32](0, 1),
		WithSteps[float32](0.1, 0.2),
		WithDisabled[float32](),
	)

	require.False(t, gauge.Interact(ui.UP))
}
