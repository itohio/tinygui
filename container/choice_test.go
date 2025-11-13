package container

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
	"github.com/stretchr/testify/require"
)

type choiceWidget struct {
	ui.WidgetBase
}

func newChoiceWidget(w, h uint16) *choiceWidget {
	return &choiceWidget{WidgetBase: ui.NewWidgetBase(w, h)}
}

func (w *choiceWidget) Draw(ui.Context) {}

func (w *choiceWidget) Interact(ui.UserCommand) bool { return false }

func TestScrollChoiceSelectorSync(t *testing.T) {
	widgets := []ui.Widget{
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
	}

	sc := NewScrollChoice(20, 18, layout.VList(2), widgets)

	require.Equal(t, 0, sc.Index())

	// Move selection down via selector commands.
	require.True(t, sc.Interact(ui.DOWN))
	require.Equal(t, 3, sc.Index())
	// DOWN should wrap because selector wraps by default.
	require.True(t, sc.Interact(ui.DOWN))
	require.Equal(t, 2, sc.Index())

	// External SetIndex keeps selector aligned.
	sc.SetIndex(1)
	require.Equal(t, 1, sc.selector.Index())
}

func TestScrollChoiceEnsureVisible(t *testing.T) {
	widgets := []ui.Widget{
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
		newChoiceWidget(20, 10),
	}

	sc := NewScrollChoice(20, 12, layout.VList(2), widgets)

	// Select last item; list should scroll to expose it.
	sc.SetIndex(3)
	_, oy := sc.ScrollOffset()
	require.Greater(t, oy, int16(0))

	// Move back to top, offset should shrink.
	sc.SetIndex(0)
	_, oy = sc.ScrollOffset()
	require.Equal(t, int16(0), oy)
}
