package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
)

func TestHorizontalGaugeUsesValuePointer(t *testing.T) {
	val := float32(0.5)
	g := NewHorizontalGauge[float32](20, 6, &val, 0, 1, color.RGBA{255, 0, 0, 255}, color.RGBA{})

	check.NotNil(t, g)
	val = 0.75
	ctx := ui.NewContext(nil, 20, 6, 0, 0)
	notPanics(t, func() { g.Draw(&ctx) })
}

func TestVerticalGaugeClampsValue(t *testing.T) {
	val := int16(200)
	g := NewVerticalGauge[int16](8, 40, &val, 0, 100, color.RGBA{0, 255, 0, 255}, color.RGBA{})

	ctx := ui.NewContext(nil, 8, 40, 0, 0)
	notPanics(t, func() { g.Draw(&ctx) })
}

// rectLog records the rectangles a widget fills.
type rectLog struct {
	mockDisplay
	rects [][4]int16
}

func (r *rectLog) FillRectangle(x, y, w, h int16, _ color.RGBA) error {
	r.rects = append(r.rects, [4]int16{x, y, w, h})
	return nil
}

// A MultiGauge never draws outside itself: values that add up past the range
// are cut at the bar's end (they ran past it), and with Shares the segments
// are each value's share of the total, end to end, filling the bar exactly.
func TestMultiGaugeStaysInsideAndSharesFill(t *testing.T) {
	colors := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}}
	for _, c := range []struct {
		name   string
		w, h   uint16
		shares bool
		vals   []float32
		want   []int16 // segment lengths along the bar (inner length 98)
	}{
		{"range, sum past max", 100, 10, false, []float32{60, 30, 50}, []int16{58, 29, 11}},
		{"shares", 100, 10, true, []float32{60, 30, 50}, []int16{42, 21, 35}},
		{"shares, negative counts 0", 100, 10, true, []float32{1, -5, 1}, []int16{49, 49}},
		{"shares, vertical", 10, 100, true, []float32{25, 25, 50}, []int16{25, 24, 49}},
	} {
		vals := c.vals
		g := NewMultiGauge[float32](c.w, c.h, 0, 100, &vals, colors, color.RGBA{}, color.RGBA{255, 255, 255, 255})
		g.Shares = c.shares
		d := &rectLog{}
		ctx := ui.NewContext(d, c.w, c.h, 0, 0)
		g.Draw(&ctx)
		var got []int16
		end := int16(1)
		for _, r := range d.rects {
			start, n := r[0], r[2]
			if c.h > c.w { // vertical: segments stack up from the bottom
				start, n = int16(c.h)-r[1]-r[3], r[3]
			}
			if start != end {
				t.Fatalf("%s: a segment starts at %d, want %d (end to end): %v", c.name, start, end, d.rects)
			}
			end += n
			got = append(got, n)
		}
		check.Equal(t, c.want, got, c.name)
		if limit := int16(max(c.w, c.h)) - 1; end > limit {
			t.Fatalf("%s: segments end at %d, past the bar's inside (%d)", c.name, end, limit)
		}
	}
}
