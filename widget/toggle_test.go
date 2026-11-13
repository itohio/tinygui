package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
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

	check.False(t, state)
	check.True(t, toggle.Interact(ui.ENTER))
	check.True(t, state)
	check.True(t, toggle.Interact(ui.RIGHT))
	check.False(t, state)
	check.True(t, toggle.Interact(ui.ESC))
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
	notPanics(t, func() { toggle.Draw(&ctx) })
}
