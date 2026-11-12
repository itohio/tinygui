package app

import (
	"testing"
	"time"

	"github.com/itohio/tinygui/examples/framebuffer"
)

type testHW struct{}

func (testHW) Light() uint16            { return 1000 }
func (testHW) Battery() (float32, bool) { return 3.9, false }
func (testHW) SetLED(int)               {}
func (testHW) SetBacklight(int)         {}

// Every page draws its frames (three 80-row bands, as on the Tufty) without
// allocating: per-frame garbage fills the badge's 142 KB heap and the
// collections that follow cost frames (findings F-34).
func TestPagesDrawFramesWithoutAllocating(t *testing.T) {
	fb := framebuffer.New(ScreenW, 80, make([]uint16, ScreenW*80), nil)
	id := func(d, x time.Duration) time.Duration { return d }
	now := time.Now()
	a := New(fb, testHW{}, Config{DurationAdd: id, DurationSub: id}, now)
	frame := func() {
		now = now.Add(33 * time.Millisecond)
		a.Update(now)
		for b := range 3 {
			fb.SetY0(int16(b * 80))
			a.Draw()
		}
	}
	for i := range a.pages {
		a.showPage(i)
		for range 40 {
			frame()
		}
		if n := testing.AllocsPerRun(30, frame); n != 0 {
			t.Errorf("page %s: %.0f allocations per frame", a.pages[i].name, n)
		}
	}
}
