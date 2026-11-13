package container

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
	"github.com/itohio/tinygui/layout"
)

// ModalFor blocks only roots under the top layer (review R-TG-postcommit
// T1): the base and lower layers are blocked, the top layer and anything
// inside it are not, and nothing is blocked with no layer open. The check
// runs before every navigator command, so it must not allocate.
func TestOverlayModalForBlocksOnlyRootsUnderTheTopLayer(t *testing.T) {
	base := New[ui.Widget](100, 40, WithLayout[ui.Widget](layout.HList(1)), WithChildren[ui.Widget](newChoiceWidget(10, 10)))
	o := NewOverlay(base)
	check.False(t, o.ModalFor(base), "no layer open")

	inner := New[ui.Widget](20, 10, WithLayout[ui.Widget](layout.HList(1)), WithChildren[ui.Widget](newChoiceWidget(10, 10)))
	panel := New[ui.Widget](60, 30, WithLayout[ui.Widget](layout.VList(1)), WithChildren[ui.Widget](inner))
	check.True(t, o.OpenCentered(panel))
	check.True(t, o.ModalFor(base), "base under the panel")
	check.False(t, o.ModalFor(panel), "the top layer itself")
	check.False(t, o.ModalFor(inner), "inside the top layer")

	popup := newChoiceWidget(30, 10)
	check.True(t, o.OpenCentered(popup))
	check.True(t, o.ModalFor(panel), "the panel is under the popup")
	check.True(t, o.ModalFor(inner))
	check.False(t, o.ModalFor(popup))
	check.Zero(t, testing.AllocsPerRun(50, func() {
		o.ModalFor(inner)
		o.ModalFor(popup)
		o.ModalFor(base)
	}))

	o.Close(popup)
	check.False(t, o.ModalFor(inner), "the panel is the top again")
	o.Close(panel)
	check.False(t, o.ModalFor(base))
}
