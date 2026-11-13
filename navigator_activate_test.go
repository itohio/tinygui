package ui_test

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/tinyfont"
)

var overlayStyle = widget.OverlayStyle{Font: &tinyfont.TomThumb, LineH: 10}

// Review R-TG-1 L5: with the library's own Navigator, a modal dialog over
// the root stops focus moves, Enter and Back (they used to move the focus
// behind it); once it closes they work again.
func TestModalOverlayBlocksTheNavigator(t *testing.T) {
	a, b := newEventWidget(), newEventWidget()
	root := container.New[ui.Widget](100, 40, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](a, b))
	overlay := container.NewOverlay(root)
	nav := ui.NewNavigator(root)
	require.True(t, nav.Focus(0))

	dialog := widget.NewDialog(overlayStyle, "Sure?", nil, []string{"Yes", "No"}, nil, nil)
	require.True(t, overlay.OpenCentered(dialog))
	require.True(t, overlay.ModalFor(root))
	require.False(t, nav.Next(), "Next behind a modal dialog")
	require.False(t, nav.Enter(), "Enter behind a modal dialog")
	require.False(t, nav.Back(), "Back behind a modal dialog")
	require.True(t, a.focused)
	require.False(t, b.focused)

	overlay.Close(dialog)
	require.False(t, overlay.ModalFor(root))
	require.True(t, nav.Next())
	require.True(t, b.focused)
}

// Enter on an Activator acts at once instead of entering edit mode: a
// Dropdown opens its list in the overlay, a Toggle flips; nothing is left
// active, and observers see NavigatorEventActed.
func TestEnterActivatesDropdownsAndToggles(t *testing.T) {
	pick, on := 0, false
	dd := widget.NewDropdown(60, overlayStyle, []string{"x1", "x2"}, &pick, nil)
	tg := widget.NewToggle(30, 10, &tinyfont.TomThumb, color.RGBA{255, 255, 255, 255}, "ON", "OFF",
		color.RGBA{0, 255, 0, 255}, color.RGBA{60, 60, 60, 255}, func() bool { return on }, func(v bool) { on = v })
	root := container.New[ui.Widget](100, 40, container.WithLayout[ui.Widget](layout.VList(1)), container.WithChildren[ui.Widget](dd, tg))
	overlay := container.NewOverlay(root)
	nav := ui.NewNavigator(root)
	obs := &recordObserver{}
	nav.AddObserver(obs)

	require.True(t, nav.Focus(0))
	require.True(t, nav.Enter())
	require.NotNil(t, overlay.Top(), "the dropdown's list opened")
	require.False(t, root.Active(), "nothing left in edit mode")
	require.Equal(t, ui.NavigatorEventActed, obs.events[len(obs.events)-1].Type)
	overlay.Interact(ui.BACK) // close the list

	require.True(t, nav.Next())
	require.True(t, nav.Enter())
	require.True(t, on, "the toggle flipped")
	require.False(t, root.Active())
	require.True(t, nav.Enter())
	require.False(t, on)
}
