package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
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
	check.True(t, ok)
	check.Equal(t, "Second", current)

	check.True(t, selector.Handle(ui.UP))
	check.Equal(t, 2, idx)
	check.Equal(t, 2, notifiedIndex)
	check.Equal(t, "Third", notifiedValue)

	check.True(t, selector.Handle(ui.UP))
	check.Equal(t, 0, idx)
	current, _ = selector.Current()
	check.Equal(t, "First", current)

	check.True(t, selector.Handle(ui.DOWN))
	check.Equal(t, 2, idx)
}

func TestInteractiveSelectorDisabled(t *testing.T) {
	items := []string{"Only"}
	selector := NewInteractiveSelector(items, WithSelectorDisabled[string]())

	check.False(t, selector.Enabled())
	check.False(t, selector.Handle(ui.UP))

	selector.SetEnabled(true)
	check.True(t, selector.Enabled())
	check.True(t, selector.Handle(ui.UP))
}

func TestInteractiveLabelChoiceUpdatesLabel(t *testing.T) {
	items := []string{"Alpha", "Beta"}
	idx := 0
	choice := NewInteractiveLabelChoice(40, 12, items,
		WithLabelChoiceIndex(&idx),
		WithLabelChoiceFont(&tinyfont.TomThumb),
		WithLabelChoiceColor(color.RGBA{255, 255, 255, 255}),
	)

	check.Equal(t, "Alpha", choice.currentText())
	check.True(t, choice.Interact(ui.UP))
	check.Equal(t, "Beta", choice.currentText())

	ctx := ui.NewContext(mockDisplay{}, 40, 12, 0, 0)
	notPanics(t, func() { choice.Draw(&ctx) })
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

	check.Equal(t, "icon0", choice.Image())
	check.True(t, choice.Interact(ui.UP))
	check.Equal(t, "icon1", choice.Image())
	check.Equal(t, 1, notifiedIndex)
	check.Equal(t, "icon1", notifiedValue)
}

func TestInteractiveWidgetChoiceCyclesWidgets(t *testing.T) {
	widgets := []*Label{buildLabel("A"), buildLabel("B")}
	choice := NewInteractiveWidgetChoice[*Label](32, 12, widgets)
	choice.SetSelected(true)

	check.True(t, widgets[0].Selected())
	check.True(t, choice.Interact(ui.UP))
	check.True(t, widgets[1].Selected())
}

// A label choice bound to an index shows that index when something else
// changes it (the Tufty fire's gravity, turned in full screen, kept reading
// "down" on its page).
func TestLabelChoiceFollowsItsBoundIndex(t *testing.T) {
	idx := 0
	c := NewInteractiveLabelChoice(60, 10, []string{"down", "left", "up"}, WithLabelChoiceIndex(&idx))
	check.Equal(t, "down", c.Label.Text())
	idx = 2
	check.Equal(t, "up", c.Label.Text())
	c.Interact(ui.UP) // and editing still writes it back
	check.Equal(t, 0, idx)
	check.Equal(t, "down", c.Label.Text())
}
