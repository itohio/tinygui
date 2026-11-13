package ui

import (
	"testing"
	"time"
)

// A screensaver moves its RandomContext once per frame and draws through
// ChildContext: no allocation per frame (Clone allocated a context each
// call, verify H-2), and the origin stays on the display.
func TestRandomContextMovesWithoutAllocating(t *testing.T) {
	d := &pixelSet{}
	c := NewRandomContext(d, 0, 150, 40)
	var dst ContextImpl
	allocs := testing.AllocsPerRun(100, func() {
		c.Move()
		ctx := ChildContext(&c.ContextImpl, &dst, nil, 150, 40)
		x, y := ctx.DisplayPos()
		if x < 0 || y < 0 || x+150 > 320 || y+40 > 240 {
			t.Fatalf("origin (%d, %d) off a 320×240 display", x, y)
		}
		time.Sleep(time.Microsecond) // past the zero interval: it moves every call
	})
	if allocs != 0 {
		t.Fatalf("%.0f allocations per move and draw", allocs)
	}
}

// moved is a context whose display position is shifted, like a scroll's
// offset context.
type moved struct {
	*ContextImpl
	dx, dy int16
}

func (m moved) DisplayPos() (int16, int16) {
	x, y := m.ContextImpl.DisplayPos()
	return x - m.dx, y - m.dy
}

// A child context under any parent (here a shifted one, as inside a
// Scroll) starts at the parent's display position and costs no allocation
// (it used to fall back to the parent's Clone, allocating per draw).
func TestChildContextUnderAnyParent(t *testing.T) {
	base := NewContext(&pixelSet{}, 200, 100, 10, 20)
	base.SetPos(5, 6)
	var p Context = &moved{&base, 3, 40}
	var dst ContextImpl
	ctx := ChildContext(p, &dst, nil, 50, 10)
	if x, y := ctx.DisplayPos(); x != 10+5-3 || y != 20+6-40 {
		t.Fatalf("child at (%d, %d), want (12, -14)", x, y)
	}
	if n := testing.AllocsPerRun(100, func() { ChildContext(p, &dst, nil, 50, 10) }); n != 0 {
		t.Fatalf("%.0f allocations per child context", n)
	}
}

// A container drawn into a RandomContext through ChildContext is moved
// around (burn-in protection), as Clone moved it — without allocating
// (verify TG-1-postcommit R12: ChildContext stopped moving it, so the
// protection was silently off).
func TestChildContextMovesARandomContextParent(t *testing.T) {
	d := &pixelSet{}
	c := NewRandomContext(d, 0, 50, 20) // moves on every call
	var dst ContextImpl
	var p Context = &c
	seen := map[[2]int16]bool{}
	for range 30 {
		x, y := ChildContext(p, &dst, nil, 10, 10).DisplayPos()
		seen[[2]int16{x, y}] = true
	}
	if len(seen) < 2 {
		t.Fatalf("the child stayed at %v: the RandomContext parent never moved", seen)
	}
	if n := testing.AllocsPerRun(50, func() { ChildContext(p, &dst, nil, 10, 10) }); n != 0 {
		t.Fatalf("%.0f allocations per child context", n)
	}
}
