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
