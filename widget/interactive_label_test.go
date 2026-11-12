package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/tinyfont"
)

func TestInteractiveLabelAdjustCommitAndCancel(t *testing.T) {
	value := float32(1.5)
	var committed float32
	label := NewInteractiveLabel[float32](64, 12,
		WithValue(&value),
		WithRange[float32](0, 5),
		WithSteps[float32](0.1, 0.5),
		WithFont[float32](&tinyfont.TomThumb),
		WithTextColor[float32](color.RGBA{255, 255, 255, 255}),
		WithCommit[float32](func(v float32) { committed = v }),
	)

	label.SetSelected(true)
	require.True(t, label.Selected())
	require.InEpsilon(t, 1.5, label.pending, 0.001)

	require.True(t, label.Interact(ui.UP))
	require.InEpsilon(t, 1.6, label.pending, 0.001)

	require.True(t, label.Interact(ui.LONG_UP))
	require.InEpsilon(t, 2.1, label.pending, 0.001)

	// Commit and ensure setter invoked.
	require.True(t, label.Interact(ui.ENTER))
	require.InEpsilon(t, 2.1, committed, 0.001)
	require.InEpsilon(t, 2.1, value, 0.001)
	require.False(t, label.Selected())

	// Re-select and cancel.
	label.SetSelected(true)
	require.True(t, label.Interact(ui.DOWN))
	require.InEpsilon(t, 2.0, label.pending, 0.001)
	require.True(t, label.Interact(ui.ESC))
	require.InEpsilon(t, 2.1, value, 0.001) // unchanged
}

func TestInteractiveLabelDisable(t *testing.T) {
	value := float32(0)
	label := NewInteractiveLabel[float32](32, 10,
		WithValue(&value),
		WithRange[float32](0, 10),
		WithSteps[float32](1, 2),
		WithFont[float32](&tinyfont.TomThumb),
		WithTextColor[float32](color.RGBA{255, 255, 255, 255}),
		WithDisabled[float32](),
	)

	require.False(t, label.Interact(ui.UP))
	require.Equal(t, float32(0), label.pending)
}

// The label shows its bound value when that value changes elsewhere (a
// serial command, another widget); while it is being edited it shows the
// pending edit instead.
func TestInteractiveLabelFollowsItsBoundValue(t *testing.T) {
	value := 80
	label := NewInteractiveLabel[int](64, 12, WithValue(&value), WithRange[int](10, 100), WithSteps[int](5, 25),
		WithFont[int](&tinyfont.TomThumb))
	value = 60
	require.Equal(t, "60", label.displayText())
	value = 200
	require.Equal(t, "100", label.displayText(), "clamped to the range")

	value = 60
	label.SetSelected(true) // editing
	label.Interact(ui.UP)
	value = 30 // changed elsewhere mid-edit: the edit wins on screen
	require.Contains(t, label.displayText(), "65")
}
