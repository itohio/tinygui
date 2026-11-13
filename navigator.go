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
	// NavigatorEventActed: Enter on an Activator, which acted at once (a list
	// opened, a toggle flipped); nothing was left active
	NavigatorEventActed
)

// Activator is a widget that acts as soon as it is entered, instead of
// taking the focus's commands (edit mode): a Dropdown opens its list, a
// Toggle flips. Activate reports whether it acted.
type Activator interface {
	Activate() bool
}

// Modal is an overlay host that reports a modal layer open (container.
// Overlay with a dialog or list open): while it is, a Navigator below it
// doesn't move the focus, enter or go back.
type Modal interface {
	Modal() bool
}

// NavigatorEvent encapsulates navigation changes for observers. Path is
// the navigator's own buffer: valid during the callback, overwritten by the
// next navigation (copy it to keep it).
type NavigatorEvent struct {
	Type NavigatorEventType
	Path Path
}

// NavigatorObserver consumes navigator events.
type NavigatorObserver interface {
	OnNavigatorEvent(NavigatorEvent)
}

// MaxNavDepth bounds how many navigable containers a Navigator stacks: Enter
// into a container deeper than this is refused. The stack and the paths it
// reports live in fixed arrays, so navigating never allocates.
const MaxNavDepth = 8

type Navigator struct {
	stack     [MaxNavDepth]Navigable
	depth     int
	path      [MaxNavDepth + 1]PathSegment // Path's and Walk's buffer
	walk      [2*MaxNavDepth + 2]PathSegment
	observers []NavigatorObserver
}

// NewNavigator creates a navigator rooted at the provided Navigable.
func NewNavigator(root Navigable) *Navigator {
	if root == nil {
		panic("navigator requires root container")
	}
	n := &Navigator{depth: 1}
	n.stack[0] = root
	return n
}

func (n *Navigator) AddObserver(obs NavigatorObserver) {
	n.observers = append(n.observers, obs)
}

// Depth reports how many navigable containers are currently on the stack.
func (n *Navigator) Depth() int {
	return n.depth
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
	if n.blocked() {
		return false
	}
	container := n.currentContainer()
	target := n.findSelectable(container, container.Index()+1, 1)
	if target < 0 {
		return false
	}
	return n.focusExact(target)
}

func (n *Navigator) Prev() bool {
	if n.blocked() {
		return false
	}
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
	if n.blocked() {
		return false
	}
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
// stack (refused, returning false, at MaxNavDepth), other widgets receive
// activation events.
func (n *Navigator) Enter() bool {
	if n.blocked() || !n.ensureSelection() {
		return false
	}
	container := n.currentContainer()
	if act, ok := container.Item().(Activator); ok {
		if !act.Activate() {
			return false
		}
		n.notify(NavigatorEventActed)
		return true
	}
	container.SetActive(container.Index())
	item := container.Item()
	if item == nil {
		return false
	}
	if child, ok := item.(Navigable); ok {
		if n.depth == MaxNavDepth {
			return false
		}
		n.stack[n.depth] = child
		n.depth++
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
	if n.blocked() {
		return false
	}
	container := n.currentContainer()
	wasActive := container.Active()
	container.SetActive(-1)
	n.notify(NavigatorEventDeactivated)
	if wasActive {
		return true
	}
	if n.depth == 1 {
		return true
	}
	n.depth--
	n.stack[n.depth] = nil
	n.notify(NavigatorEventFocusChanged)
	return true
}

// Path returns the navigation stack and current item as a Path: the
// navigator's own buffer, valid until the next navigation (no allocation).
func (n *Navigator) Path() Path {
	path := n.path[:0]
	for _, container := range n.stack[:n.depth] {
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

// Walk traverses the entire navigable hierarchy depth-first invoking fn with
// the path to each node (the navigator's buffer, valid during the call).
// Containers deeper than MaxNavDepth are not descended into.
func (n *Navigator) Walk(fn func(Path) bool) {
	n.walkContainer(n.stack[0], n.walk[:0], fn)
}

// walkContainer traverses container depth-first, building paths in the
// fixed buffer and invoking fn at each node. Returning false aborts traversal.
func (n *Navigator) walkContainer(container Navigable, path Path, fn func(Path) bool) bool {
	currentPath := append(path, PathSegment{
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
		childPath := append(currentPath, PathSegment{
			Widget: w,
			Index:  i,
		})
		if !fn(childPath) {
			return false
		}
		if child, ok := w.(Navigable); ok && len(childPath) < cap(n.walk)-1 {
			if !n.walkContainer(child, childPath, fn) {
				return false
			}
		}
	}
	return true
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
// blocked: a modal overlay (a dialog, an open list) is over the root, so
// focus moves, Enter and Back are refused (review R-TG-1 L5).
func (n *Navigator) blocked() bool {
	m, ok := FindOverlayHost(n.stack[0]).(Modal)
	return ok && m.Modal()
}

func (n *Navigator) currentContainer() Navigable {
	return n.stack[n.depth-1]
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
