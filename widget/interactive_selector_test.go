package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/tinyfont"
)

func buildLabel(text string) *Label {
	lbl := NewLabel(32, 12, &tinyfont.TomThumb, nil, color.RGBA{255, 255, 255, 255})
	lbl.SetText(text)
	return lbl
}

func TestInteractiveSelectorCyclesStrings(t *testing.T) {
	items := []string{"First", "Second", "Third"}
	idx := 1
	var notifiedIndex int
	var notifiedValue string

	selector := NewInteractiveSelector(items,
		WithSelectorIndex[string](&idx),
		WithSelectorChange[string](func(i int, value string) {
			notifiedIndex = i
			notifiedValue = value
		}),
	)

	current, ok := selector.Current()
	require.True(t, ok)
	require.Equal(t, "Second", current)

	require.True(t, selector.Handle(ui.UP))
	require.Equal(t, 2, idx)
	require.Equal(t, 2, notifiedIndex)
	require.Equal(t, "Third", notifiedValue)

	require.True(t, selector.Handle(ui.UP))
	require.Equal(t, 0, idx)
	current, _ = selector.Current()
	require.Equal(t, "First", current)

	require.True(t, selector.Handle(ui.DOWN))
	require.Equal(t, 2, idx)
}

func TestInteractiveSelectorDisabled(t *testing.T) {
	items := []string{"Only"}
	selector := NewInteractiveSelector(items, WithSelectorDisabled[string]())

	require.False(t, selector.Enabled())
	require.False(t, selector.Handle(ui.UP))

	selector.SetEnabled(true)
	require.True(t, selector.Enabled())
	require.True(t, selector.Handle(ui.UP))
}

func TestInteractiveLabelChoiceUpdatesLabel(t *testing.T) {
	items := []string{"Alpha", "Beta"}
	idx := 0
	choice := NewInteractiveLabelChoice(40, 12, items,
		WithLabelChoiceIndex(&idx),
		WithLabelChoiceFont(&tinyfont.TomThumb),
		WithLabelChoiceColor(color.RGBA{255, 255, 255, 255}),
	)

	require.Equal(t, "Alpha", choice.currentText())
	require.True(t, choice.Interact(ui.UP))
	require.Equal(t, "Beta", choice.currentText())

	ctx := ui.NewContext(mockDisplay{}, 40, 12, 0, 0)
	require.NotPanics(t, func() { choice.Draw(&ctx) })
}

func TestInteractiveIconChoiceUpdatesImage(t *testing.T) {
	images := []string{"icon0", "icon1"}
	idx := 0
	var notifiedIndex int
	var notifiedValue string

	choice := NewInteractiveIconChoice(16, 16, images,
		WithIconChoiceIndex(&idx),
		WithIconChoiceChange(func(i int, v string) {
			notifiedIndex = i
			notifiedValue = v
		}),
	)

	require.Equal(t, "icon0", choice.Image())
	require.True(t, choice.Interact(ui.UP))
	require.Equal(t, "icon1", choice.Image())
	require.Equal(t, 1, notifiedIndex)
	require.Equal(t, "icon1", notifiedValue)
}

func TestInteractiveWidgetChoiceCyclesWidgets(t *testing.T) {
	widgets := []*Label{buildLabel("A"), buildLabel("B")}
	choice := NewInteractiveWidgetChoice[*Label](32, 12, widgets)
	choice.SetSelected(true)

	require.True(t, widgets[0].Selected())
	require.True(t, choice.Interact(ui.UP))
	require.True(t, widgets[1].Selected())
}
