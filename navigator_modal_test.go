package ui_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/internal/check"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

// Review R-TG-postcommit T1: a Navigator rooted inside an overlay layer (a
// panel opened as a layer, navigated by its own Navigator) works while that
// layer is the top; the main root's Navigator stays blocked; a dialog opened
// over the panel blocks the panel's Navigator until it closes.
func TestNavigatorRootedInTopLayerIsNotBlocked(t *testing.T) {
	a, b := newEventWidget(), newEventWidget()
	root := container.New[ui.Widget](100, 40, container.WithLayout[ui.Widget](layout.HList(1)), container.WithChildren[ui.Widget](a, b))
	overlay := container.NewOverlay(root)
	mainNav := ui.NewNavigator(root)
	check.True(t, mainNav.Focus(0))

	c, d := newEventWidget(), newEventWidget()
	panel := container.New[ui.Widget](60, 30, container.WithLayout[ui.Widget](layout.VList(1)), container.WithChildren[ui.Widget](c, d))
	check.True(t, overlay.OpenCentered(panel))
	panelNav := ui.NewNavigator(panel)

	check.True(t, panelNav.Focus(0), "Focus in the top layer")
	check.True(t, c.focused)
	check.True(t, panelNav.Next(), "Next in the top layer")
	check.True(t, d.focused)
	check.True(t, panelNav.Enter(), "Enter in the top layer")
	check.True(t, d.activated)
	check.True(t, panelNav.Back(), "Back in the top layer")
	check.False(t, d.activated)

	check.True(t, panelNav.Focus(1), "Back cleared the focus; refocus")
	check.True(t, d.focused)
	check.False(t, mainNav.Next(), "the main root is under the panel")
	check.True(t, a.focused)
	check.False(t, b.focused)

	dialog := widget.NewDialog(overlayStyle, "Sure?", nil, []string{"Yes", "No"}, nil, nil)
	check.True(t, overlay.OpenCentered(dialog))
	check.False(t, panelNav.Prev(), "the panel is under the dialog")
	check.True(t, d.focused)
	check.False(t, mainNav.Next())

	overlay.Close(dialog)
	check.True(t, panelNav.Prev(), "the panel is the top again")
	check.True(t, c.focused)
	check.False(t, mainNav.Next())

	overlay.Close(panel)
	check.True(t, mainNav.Next(), "no layer open")
	check.True(t, b.focused)
}
