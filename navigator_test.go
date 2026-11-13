package ui_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/internal/check"
	"github.com/itohio/tinygui/layout"
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

	check.True(t, nav.Focus(0))
	check.True(t, a.focused)
	check.False(t, b.focused)

	check.True(t, nav.Next())
	check.False(t, a.focused)
	check.True(t, b.focused)

	check.False(t, nav.Next())
	check.True(t, nav.Prev())
	check.True(t, a.focused)
	check.False(t, b.focused)
}

func TestNavigatorActivationAndBack(t *testing.T) {
	childA := newEventWidget()
	childB := newEventWidget()
	childContainer := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(1)),
		container.WithChildren[ui.Widget](childA, childB),
	)
	_, isNav := any(childContainer).(ui.Navigable)
	check.True(t, isNav, "the child container is a ui.Navigable")

	parentWidgets := []ui.Widget{childContainer, newEventWidget()}
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(1)),
		container.WithChildren[ui.Widget](parentWidgets...),
	)
	nav := ui.NewNavigator(root)

	check.True(t, nav.Focus(0))
	check.Equal(t, 0, root.Index())
	_, navigable := root.Item().(ui.Navigable)
	check.True(t, navigable)
	check.True(t, nav.Enter())
	check.Equal(t, 2, nav.Depth())
	check.True(t, childContainer.Active())
	check.True(t, childContainer.Index() >= 0)

	check.True(t, nav.Next())
	check.True(t, childB.focused)
	check.False(t, childA.focused)

	check.True(t, nav.Enter())
	check.True(t, childB.activated)

	check.True(t, nav.Back())
	check.False(t, childB.activated)
	check.Equal(t, 2, nav.Depth())

	check.True(t, nav.Back())
	check.Equal(t, 1, nav.Depth())
}

func TestNavigatorObserversReceiveEvents(t *testing.T) {
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(0)),
		container.WithChildren[ui.Widget](newEventWidget(), newEventWidget()),
	)
	nav := ui.NewNavigator(root)
	rec := &recordObserver{}
	nav.AddObserver(rec)

	check.True(t, nav.Focus(0))
	check.True(t, nav.Next())
	check.True(t, nav.Enter())

	check.NotEmpty(t, rec.events)
	types := make([]ui.NavigatorEventType, 0, len(rec.events))
	for _, event := range rec.events {
		types = append(types, event.Type)
		check.NotNil(t, event.Path.Current())
	}
	check.Contains(t, types, ui.NavigatorEventFocusChanged)
	check.Contains(t, types, ui.NavigatorEventActivated)
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

	check.GreaterOrEqual(t, visited, 3)
}

func TestNavigatorSkipsNonSelectableWidgets(t *testing.T) {
	static := newStaticWidget()
	dynamic := newEventWidget()
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.HList(0)),
		container.WithChildren[ui.Widget](static, dynamic),
	)
	nav := ui.NewNavigator(root)

	check.True(t, nav.Focus(0))
	check.True(t, dynamic.focused)
	check.Equal(t, 1, root.Index())

	check.False(t, nav.Prev())
	check.True(t, dynamic.focused)
}

type countObserver struct{ n int }

func (c *countObserver) OnNavigatorEvent(e ui.NavigatorEvent) {
	if e.Path.Current() != nil {
		c.n++
	}
}

// Navigating (focus, enter, back) with an observer attached allocates
// nothing: the stack and the reported paths live in fixed buffers.
func TestNavigatorNavigatesWithoutAllocating(t *testing.T) {
	inner := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(0)),
		container.WithChildren[ui.Widget](newEventWidget(), newEventWidget()),
	)
	root := container.New[ui.Widget](0, 0,
		container.WithLayout[ui.Widget](layout.VList(0)),
		container.WithChildren[ui.Widget](newEventWidget(), inner, newEventWidget()),
	)
	nav := ui.NewNavigator(root)
	obs := &countObserver{}
	nav.AddObserver(obs)
	nav.Focus(0)
	allocs := testing.AllocsPerRun(50, func() {
		nav.Next() // onto the inner container
		nav.Enter()
		nav.Next()
		nav.Prev()
		nav.Back()
		nav.Back()
		nav.Prev()
		_ = nav.Path()
		nav.Walk(func(ui.Path) bool { return true })
	})
	check.Zero(t, allocs)
	check.Positive(t, obs.n)
}

// The stack is bounded: entering containers nested deeper than MaxNavDepth
// is refused, and Back unwinds to the root.
func TestNavigatorStackIsBounded(t *testing.T) {
	var w ui.Widget = newEventWidget()
	for range ui.MaxNavDepth + 2 {
		w = container.New[ui.Widget](0, 0,
			container.WithLayout[ui.Widget](layout.VList(0)),
			container.WithChildren[ui.Widget](w),
		)
	}
	nav := ui.NewNavigator(w.(ui.Navigable))
	for nav.Enter() {
	}
	check.Equal(t, ui.MaxNavDepth, nav.Depth())
	for range ui.MaxNavDepth {
		nav.Back()
		nav.Back()
	}
	check.Equal(t, 1, nav.Depth())
}
