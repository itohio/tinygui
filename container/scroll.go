package container

import (
	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/layout"
)

// Scroll composes Base with scroll offsets while respecting parent contexts.
type Scroll struct {
	*Base[ui.Widget]

	offsetX   int16
	offsetY   int16
	contentW  uint16
	contentH  uint16
	observers []ScrollObserver
}

// NewScroll returns a scroll-enabled container sized by viewportW/H.
func NewScroll(viewportW, viewportH uint16, lay layout.Strategy, widgets ...ui.Widget) *Scroll {
	options := []Option[ui.Widget]{
		WithLayout[ui.Widget](lay),
	}
	if len(widgets) > 0 {
		options = append(options, WithChildren[ui.Widget](widgets...))
	}
	c := &Scroll{
		Base: New[ui.Widget](viewportW, viewportH, options...),
	}
	c.refreshContentSize()
	return c
}

// Scroll adjusts internal offsets; returns true when a visible change occurred.
func (s *Scroll) Scroll(dx, dy int16) bool {
	if dx == 0 && dy == 0 {
		return false
	}
	prevX, prevY := s.offsetX, s.offsetY
	s.offsetX += dx
	s.offsetY += dy
	s.clampOffsets()
	if s.offsetX == prevX && s.offsetY == prevY {
		return false
	}
	s.notify(ScrollChange{
		DX:      s.offsetX - prevX,
		DY:      s.offsetY - prevY,
		OffsetX: s.offsetX,
		OffsetY: s.offsetY,
	})
	return true
}

// ScrollOffset reports the current scroll offsets.
func (s *Scroll) ScrollOffset() (int16, int16) {
	return s.offsetX, s.offsetY
}

// AddObserver registers a ScrollObserver to receive offset changes.
func (s *Scroll) AddObserver(observer ScrollObserver) {
	s.observers = append(s.observers, observer)
}

// Draw renders only children that intersect the visible area.
func (s *Scroll) Draw(ctx ui.Context) {
	w, h := s.Size()
	base := ctx.Clone(s, w, h)
	offsetCtx := &offsetContext{
		Context: base,
		dx:      s.offsetX,
		dy:      s.offsetY,
	}
	s.drawVisible(offsetCtx, int16(w), int16(h))
}

func (s *Scroll) notify(change ScrollChange) {
	for _, obs := range s.observers {
		obs.OnScrollChange(change)
	}
	for _, item := range s.Items {
		if handler, ok := item.(ui.ScrollHandler); ok {
			handler.OnScroll(s.offsetX, s.offsetY)
		}
	}
}

func (s *Scroll) refreshContentSize() {
	if s.layouter == nil {
		s.contentW = s.Width
		s.contentH = s.Height
		return
	}
	w, h := determineSize[ui.Widget](0, 0, s.layouter, s.Items)
	s.contentW = w
	s.contentH = h
	s.clampOffsets()
}

func (s *Scroll) clampOffsets() {
	s.offsetX = clamp16(s.offsetX, 0, s.maxOffsetX())
	s.offsetY = clamp16(s.offsetY, 0, s.maxOffsetY())
}

func (s *Scroll) maxOffsetX() int16 {
	viewportW, _ := s.Size()
	if s.contentW <= viewportW {
		return 0
	}
	return int16(s.contentW - viewportW)
}

func (s *Scroll) maxOffsetY() int16 {
	_, viewportH := s.Size()
	if s.contentH <= viewportH {
		return 0
	}
	return int16(s.contentH - viewportH)
}

func (s *Scroll) drawVisible(ctx ui.Context, viewportW, viewportH int16) {
	localCtx := ctx.Clone(s, uint16(viewportW), uint16(viewportH))
	originX, originY := localCtx.Start()
	for _, item := range s.Items {
		itemW, itemH := item.Size()
		displayX, displayY := localCtx.DisplayPos()
		visible := intersectsRect(displayX, displayY, int16(itemW), int16(itemH), originX, originY, viewportW, viewportH)
		s.setVisibility(item, visible)
		if visible {
			item.Draw(localCtx)
		}
		if s.layouter == nil {
			continue
		}
		if !s.layouter(localCtx, item) {
			return
		}
	}
}

type offsetContext struct {
	ui.Context
	dx int16
	dy int16
}

func (o *offsetContext) DisplayPos() (int16, int16) {
	x, y := o.Context.DisplayPos()
	return x - o.dx, y - o.dy
}

func (o *offsetContext) Clone(widget ui.Widget, W, H uint16) ui.Context {
	return &offsetContext{
		Context: o.Context.Clone(widget, W, H),
		dx:      o.dx,
		dy:      o.dy,
	}
}

func clamp16(v, min, max int16) int16 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
