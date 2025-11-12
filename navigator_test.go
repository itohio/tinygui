package ui_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/layout"
	"github.com/stretchr/testify/require"
)

type eventWidget struct {
	ui.WidgetBase
	focused    bool
	activated  bool
	focusCount int
}

func newEventWidget() *eventWidget {
	return &eventWidget{
		WidgetBase: ui.NewWidgetBase(10, 10),
	}
}

func (w *eventWidget) Draw(ui.Context) {}

func (w *eventWidget) Interact(ui.UserCommand) bool { return false }

func (w *eventWidget) OnFocus() {
	w.focused = true
	w.focusCount++
}

func (w *eventWidget) OnBlur() {
	w.focused = false
}

func (w *eventWidget) OnActivate() {
	w.activated = true
}

func (w *eventWidget) OnDeactivate() {
	w.activated = false
}

type staticWidget struct {
	ui.WidgetBase
}

func newStaticWidget() *staticWidget {
	return &staticWidget{WidgetBase: ui.NewWidgetBase(8, 8)}
}

func (w *staticWidget) Draw(ui.Context) {}

func (w *staticWidget) Interact(ui.UserCommand) bool { return false }

func (w *staticWidget) CanSelect() bool { return false }

type recordObserver struct {
	events []ui.NavigatorEvent
}

func (r *recordObserver) OnNavigatorEvent(e ui.NavigatorEvent) {
	r.events = append(r.events, e)
}

func TestNavigatorFocusTransitions(t *testing.T) {
	a := newEventWidget()
	b := newEventWidget()
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(1)),
		container.WithChildren[ui.Widget](a, b),
	)
	nav := ui.NewNavigator(root)

	require.True(t, nav.Focus(0))
	require.True(t, a.focused)
	require.False(t, b.focused)

	require.True(t, nav.Next())
	require.False(t, a.focused)
	require.True(t, b.focused)

	require.False(t, nav.Next())
	require.True(t, nav.Prev())
	require.True(t, a.focused)
	require.False(t, b.focused)
}

func TestNavigatorActivationAndBack(t *testing.T) {
	childA := newEventWidget()
	childB := newEventWidget()
	childContainer := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(1)),
		container.WithChildren[ui.Widget](childA, childB),
	)
	require.Implements(t, (*ui.Navigable)(nil), childContainer)

	parentWidgets := []ui.Widget{childContainer, newEventWidget()}
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(1)),
		container.WithChildren[ui.Widget](parentWidgets...),
	)
	nav := ui.NewNavigator(root)

	require.True(t, nav.Focus(0))
	require.Equal(t, 0, root.Index())
	_, navigable := root.Item().(ui.Navigable)
	require.True(t, navigable)
	require.True(t, nav.Enter())
	require.Equal(t, 2, nav.Depth())
	require.True(t, childContainer.Active())
	require.True(t, childContainer.Index() >= 0)

	require.True(t, nav.Next())
	require.True(t, childB.focused)
	require.False(t, childA.focused)

	require.True(t, nav.Enter())
	require.True(t, childB.activated)

	require.True(t, nav.Back())
	require.False(t, childB.activated)
	require.Equal(t, 2, nav.Depth())

	require.True(t, nav.Back())
	require.Equal(t, 1, nav.Depth())
}

func TestNavigatorObserversReceiveEvents(t *testing.T) {
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(0)),
		container.WithChildren[ui.Widget](newEventWidget(), newEventWidget()),
	)
	nav := ui.NewNavigator(root)
	rec := &recordObserver{}
	nav.AddObserver(rec)

	require.True(t, nav.Focus(0))
	require.True(t, nav.Next())
	require.True(t, nav.Enter())

	require.NotEmpty(t, rec.events)
	types := make([]ui.NavigatorEventType, 0, len(rec.events))
	for _, event := range rec.events {
		types = append(types, event.Type)
		require.NotNil(t, event.Path.Current())
	}
	require.Contains(t, types, ui.NavigatorEventFocusChanged)
	require.Contains(t, types, ui.NavigatorEventActivated)
}

func TestNavigatorWalk(t *testing.T) {
	child := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(0)),
		container.WithChildren[ui.Widget](newEventWidget()),
	)
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(0)),
		container.WithChildren[ui.Widget](child),
	)
	nav := ui.NewNavigator(root)

	var visited int
	nav.Walk(func(path ui.Path) bool {
		if len(path) > 0 {
			visited++
		}
		return true
	})

	require.GreaterOrEqual(t, visited, 3)
}

func TestNavigatorSkipsNonSelectableWidgets(t *testing.T) {
	static := newStaticWidget()
	dynamic := newEventWidget()
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(0)),
		container.WithChildren[ui.Widget](static, dynamic),
	)
	nav := ui.NewNavigator(root)

	require.True(t, nav.Focus(0))
	require.True(t, dynamic.focused)
	require.Equal(t, 1, root.Index())

	require.False(t, nav.Prev())
	require.True(t, dynamic.focused)
}
