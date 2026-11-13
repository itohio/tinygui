package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
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
	check.Equal(t, 0, bar.active)
	check.True(t, bar.Interact(ui.UP))
	inEpsilon(t, 0.3, bar.pending[0], 0.001)

	check.True(t, bar.Interact(ui.ENTER))
	check.Equal(t, 1, bar.active)
	check.True(t, bar.Interact(ui.DOWN))
	inEpsilon(t, 0.7, bar.pending[1], 0.001)

	check.True(t, bar.Interact(ui.LEFT))
	check.Equal(t, 0, bar.active)

	check.True(t, bar.Interact(ui.RIGHT))
	check.Equal(t, 1, bar.active)

	check.True(t, bar.Interact(ui.ENTER))
	check.False(t, bar.Selected())
	inEpsilon(t, 0.3, values[0], 0.001)
	inEpsilon(t, 0.7, values[1], 0.001)
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
	check.True(t, gauge.Interact(ui.UP))
	inEpsilon(t, 0.6, gauge.pending, 0.001)
	check.True(t, gauge.Interact(ui.ESC))
	inEpsilon(t, 0.5, value, 0.001)
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

	check.False(t, gauge.Interact(ui.UP))
}
