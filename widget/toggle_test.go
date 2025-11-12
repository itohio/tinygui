package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

func TestToggleInteract(t *testing.T) {
	state := false
	toggle := NewToggle(
		20,
		12,
		nil,
		color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		"ON",
		"OFF",
		color.RGBA{G: 0x80},
		color.RGBA{R: 0x80},
		func() bool { return state },
		func(v bool) { state = v },
	)

	require.False(t, state)
	require.True(t, toggle.Interact(ui.ENTER))
	require.True(t, state)
	require.True(t, toggle.Interact(ui.RIGHT))
	require.False(t, state)
	require.True(t, toggle.Interact(ui.ESC))
}

func TestToggleDrawHandlesNilDisplay(t *testing.T) {
	toggle := NewToggle(
		16,
		10,
		nil,
		color.RGBA{A: 0xFF},
		"ON",
		"OFF",
		color.RGBA{G: 0x40},
		color.RGBA{R: 0x40},
		nil,
		nil,
	)

	ctx := ui.NewContext(nil, 16, 10, 0, 0)
	require.NotPanics(t, func() { toggle.Draw(&ctx) })
}
