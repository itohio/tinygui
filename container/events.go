package container

// ScrollChange captures how a scrollable widget changed between frames. Offsets
// are absolute after the adjustment while DX/DY represent the incremental move.
type ScrollChange struct {
	DX      int16
	DY      int16
	OffsetX int16
	OffsetY int16
}

// ScrollObserver can be registered with scrollable components to receive
// notifications whenever the scroll offset changes.
type ScrollObserver interface {
	OnScrollChange(ScrollChange)
}
