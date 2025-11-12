package widget

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

func TestInteractiveIconCyclesAndNotifies(t *testing.T) {
	icons := []string{"icon0", "icon1", "icon2"}
	idx := 1
	changed := -1
	widget := NewInteractiveIcon(16, 16, icons, WithIconIndex(&idx), WithIconChange(func(i int) {
		changed = i
	}))

	require.Equal(t, 1, idx)
	require.Equal(t, "icon1", widget.Image())

	ctx := ui.NewContext(mockDisplay{}, 16, 16, 0, 0)
	require.NotPanics(t, func() { widget.Draw(&ctx) })

	require.True(t, widget.Interact(ui.UP))
	require.Equal(t, 2, idx)
	require.Equal(t, 2, changed)
	require.Equal(t, "icon2", widget.Image())

	require.True(t, widget.Interact(ui.UP))
	require.Equal(t, 0, idx)
	require.Equal(t, 0, changed)
	require.Equal(t, "icon0", widget.Image())

	require.True(t, widget.Interact(ui.DOWN))
	require.Equal(t, 2, idx)
}

func TestInteractiveIconDisabled(t *testing.T) {
	icons := []string{"icon0"}
	widget := NewInteractiveIcon(16, 16, icons, WithIconDisabled())

	require.False(t, widget.Enabled())
	require.False(t, widget.Interact(ui.UP))

	widget.SetEnabled(true)
	require.True(t, widget.Enabled())
	require.True(t, widget.Interact(ui.UP))
}
