package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/tinyfont"
)

func TestInteractiveChoiceCyclesAndUpdatesIndex(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	idx := 1
	notified := -1

	choice := NewInteractiveChoice(64, 12, items,
		WithChoiceIndex(&idx),
		WithChoiceChange(func(i int) { notified = i }),
		WithChoiceFont(&tinyfont.TomThumb),
		WithChoiceColor(color.RGBA{255, 255, 255, 255}),
	)

	require.Equal(t, "Second", choice.currentText())
	require.Equal(t, 1, idx)

	require.True(t, choice.Interact(ui.UP))
	require.Equal(t, 2, idx)
	require.Equal(t, 2, notified)
	require.Equal(t, "Third", choice.currentText())

	require.True(t, choice.Interact(ui.UP))
	require.Equal(t, 0, idx)
	require.Equal(t, "First", choice.currentText())

	require.True(t, choice.Interact(ui.DOWN))
	require.Equal(t, 2, idx)
	require.Equal(t, "Third", choice.currentText())
}

func TestInteractiveChoiceDisabled(t *testing.T) {
	items := []string{"Only"}
	choice := NewInteractiveChoice(32, 10, items, WithChoiceDisabled())

	require.False(t, choice.Enabled())
	require.False(t, choice.Interact(ui.UP))

	choice.SetEnabled(true)
	require.True(t, choice.Enabled())
	require.True(t, choice.Interact(ui.UP))
}
