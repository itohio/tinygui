package widget

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
)

func TestInteractiveIconCyclesAndNotifies(t *testing.T) {
	icons := []string{"icon0", "icon1", "icon2"}
	idx := 1
	changed := -1
	widget := NewInteractiveIcon(16, 16, icons, WithIconIndex(&idx), WithIconChange(func(i int) {
		changed = i
	}))

	check.Equal(t, 1, idx)
	check.Equal(t, "icon1", widget.Image())

	ctx := ui.NewContext(mockDisplay{}, 16, 16, 0, 0)
	notPanics(t, func() { widget.Draw(&ctx) })

	check.True(t, widget.Interact(ui.UP))
	check.Equal(t, 2, idx)
	check.Equal(t, 2, changed)
	check.Equal(t, "icon2", widget.Image())

	check.True(t, widget.Interact(ui.UP))
	check.Equal(t, 0, idx)
	check.Equal(t, 0, changed)
	check.Equal(t, "icon0", widget.Image())

	check.True(t, widget.Interact(ui.DOWN))
	check.Equal(t, 2, idx)
}

func TestInteractiveIconDisabled(t *testing.T) {
	icons := []string{"icon0"}
	widget := NewInteractiveIcon(16, 16, icons, WithIconDisabled())

	check.False(t, widget.Enabled())
	check.False(t, widget.Interact(ui.UP))

	widget.SetEnabled(true)
	check.True(t, widget.Enabled())
	check.True(t, widget.Interact(ui.UP))
}
