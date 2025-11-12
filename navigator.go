package ui

// PathSegment represents a single step within the navigator stack.
type PathSegment struct {
	Widget Widget
	Index  int
}

// Path is an ordered collection of PathSegments from root to current focus.
type Path []PathSegment

// Current returns the widget referenced by the last segment in the path.
func (p Path) Current() Widget {
	if len(p) == 0 {
		return nil
	}
	return p[len(p)-1].Widget
}

// NavigatorEventType identifies the kind of navigation change.
type NavigatorEventType byte

const (
	NavigatorEventFocusChanged NavigatorEventType = iota
	NavigatorEventActivated
	NavigatorEventDeactivated
)

// NavigatorEvent encapsulates navigation changes for observers.
type NavigatorEvent struct {
	Type NavigatorEventType
	Path Path
}

// NavigatorObserver consumes navigator events.
type NavigatorObserver interface {
	OnNavigatorEvent(NavigatorEvent)
}

type Navigator struct {
	stack     []Navigable
	observers []NavigatorObserver
}

// NewNavigator creates a navigator rooted at the provided Navigable.
func NewNavigator(root Navigable) *Navigator {
	if root == nil {
		panic("navigator requires root container")
	}
	return &Navigator{
		stack: []Navigable{
			root,
		},
	}
}

func (n *Navigator) AddObserver(obs NavigatorObserver) {
	n.observers = append(n.observers, obs)
}

// Depth reports how many navigable containers are currently on the stack.
func (n *Navigator) Depth() int {
	return len(n.stack)
}

// Current yields the currently focused widget or the active container when no
// child has focus.
func (n *Navigator) Current() Widget {
	item := n.currentContainer().Item()
	if item != nil {
		return item
	}
	return n.currentContainer()
}

// Next advances focus to the next selectable widget in the current container.
func (n *Navigator) Next() bool {
	container := n.currentContainer()
	target := n.findSelectable(container, container.Index()+1, 1)
	if target < 0 {
		return false
	}
	return n.focusExact(target)
}

func (n *Navigator) Prev() bool {
	container := n.currentContainer()
	start := container.Index()
	if start < 0 {
		start = container.ChildCount() - 1
	} else {
		start--
	}
	target := n.findSelectable(container, start, -1)
	if target < 0 {
		return false
	}
	return n.focusExact(target)
}

// Focus explicitly sets the selection index inside the current container.
func (n *Navigator) Focus(index int) bool {
	container := n.currentContainer()
	if index < 0 {
		prev := container.Index()
		container.SetIndex(-1)
		if container.Index() != prev {
			n.notify(NavigatorEventFocusChanged)
		}
		return true
	}
	target := n.findSelectable(container, index, 1)
	if target < 0 {
		return false
	}
	return n.focusExact(target)
}

// Enter activates the selected widget. Navigable children are pushed onto the
// stack, other widgets receive activation events.
func (n *Navigator) Enter() bool {
	if !n.ensureSelection() {
		return false
	}
	container := n.currentContainer()
	container.SetActive(container.Index())
	item := container.Item()
	if item == nil {
		return false
	}
	if child, ok := item.(Navigable); ok {
		n.stack = append(n.stack, child)
		if child.Index() < 0 && child.ChildCount() > 0 {
			if first := n.findSelectable(child, 0, 1); first >= 0 {
				child.SetIndex(first)
			}
		}
		child.SetActive(child.Index())
		n.notify(NavigatorEventFocusChanged)
		return true
	}
	n.notify(NavigatorEventActivated)
	return true
}

// Back deactivates the current widget and unwinds to the parent when possible.
func (n *Navigator) Back() bool {
	if len(n.stack) == 0 {
		return false
	}
	container := n.currentContainer()
	wasActive := container.Active()
	container.SetActive(-1)
	n.notify(NavigatorEventDeactivated)
	if wasActive {
		return true
	}
	if len(n.stack) == 1 {
		return true
	}
	n.stack = n.stack[:len(n.stack)-1]
	n.notify(NavigatorEventFocusChanged)
	return true
}

// Path returns the navigation stack and current item as a Path.
func (n *Navigator) Path() Path {
	path := make(Path, 0, len(n.stack)+1)
	for _, container := range n.stack {
		path = append(path, PathSegment{
			Widget: container,
			Index:  container.Index(),
		})
	}
	if item := n.currentContainer().Item(); item != nil {
		path = append(path, PathSegment{
			Widget: item,
			Index:  -1,
		})
	}
	return path
}

// Walk traverses the entire navigable hierarchy depth-first invoking fn.
func (n *Navigator) Walk(fn func(Path) bool) {
	if len(n.stack) == 0 {
		return
	}
	walkContainer(n.stack[0], nil, fn)
}

// walkContainer traverses container depth-first, building paths and invoking fn
// at each node. Returning false aborts traversal.
func walkContainer(container Navigable, path Path, fn func(Path) bool) bool {
	currentPath := appendPath(path, PathSegment{
		Widget: container,
		Index:  container.Index(),
	})
	if !fn(currentPath) {
		return false
	}
	for i := 0; i < container.ChildCount(); i++ {
		w := container.Child(i)
		if w == nil {
			continue
		}
		childPath := appendPath(currentPath, PathSegment{
			Widget: w,
			Index:  i,
		})
		if !fn(childPath) {
			return false
		}
		if child, ok := w.(Navigable); ok {
			if !walkContainer(child, childPath, fn) {
				return false
			}
		}
	}
	return true
}

// appendPath produces a new Path containing segment appended to path.
func appendPath(path Path, segment PathSegment) Path {
	next := append(Path(nil), path...)
	return append(next, segment)
}

// ensureSelection guarantees that the current container points at a selectable
// child before navigation actions are performed.
func (n *Navigator) ensureSelection() bool {
	container := n.currentContainer()
	if container.ChildCount() == 0 {
		return false
	}
	if container.Index() >= 0 && isSelectable(container.Item()) {
		return true
	}
	target := n.findSelectable(container, 0, 1)
	if target < 0 {
		return false
	}
	container.SetIndex(target)
	n.notify(NavigatorEventFocusChanged)
	return true
}

// currentContainer returns the Navigable instance at the top of the stack.
func (n *Navigator) currentContainer() Navigable {
	return n.stack[len(n.stack)-1]
}

// notify emits a NavigatorEvent of type t to every observer.
func (n *Navigator) notify(t NavigatorEventType) {
	if len(n.observers) == 0 {
		return
	}
	event := NavigatorEvent{
		Type: t,
		Path: n.Path(),
	}
	for _, obs := range n.observers {
		obs.OnNavigatorEvent(event)
	}
}

// findSelectable searches forward/backward from start for the next selectable
// child index. Returns -1 when no matching widget is found.
func (n *Navigator) findSelectable(container Navigable, start int, direction int) int {
	count := container.ChildCount()
	if count == 0 {
		return -1
	}
	if direction >= 0 {
		if start < 0 {
			start = 0
		}
		for i := start; i < count; i++ {
			if w := container.Child(i); isSelectable(w) {
				return i
			}
		}
		return -1
	}
	if start >= count {
		start = count - 1
	}
	for i := start; i >= 0; i-- {
		if w := container.Child(i); isSelectable(w) {
			return i
		}
	}
	return -1
}

// focusExact applies an index change without additional traversal logic.
func (n *Navigator) focusExact(index int) bool {
	container := n.currentContainer()
	prev := container.Index()
	container.SetIndex(index)
	if container.Index() != prev {
		n.notify(NavigatorEventFocusChanged)
	}
	return container.Index() == index
}

// isSelectable reports whether a widget participates in navigation.
func isSelectable(w Widget) bool {
	if w == nil {
		return false
	}
	if selectable, ok := w.(Selectable); ok {
		if !selectable.CanSelect() {
			return false
		}
	}
	if enabler, ok := w.(EnableState); ok {
		if !enabler.Enabled() {
			return false
		}
	}
	return true
}
