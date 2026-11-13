package widget

import (
	"image/color"
	"math"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
	"tinygo.org/x/tinyfont/proggy"
)

var (
	rTrack = color.RGBA{40, 44, 60, 255}
	rFill  = color.RGBA{0, 190, 210, 255}
	rBG    = color.RGBA{10, 12, 16, 255}
	rLabel = color.RGBA{230, 230, 230, 255}
	rUnset = color.RGBA{255, 0, 255, 255} // canvas sentinel: never painted
)

// drawRing draws a w×h ring of value (0..100) on a canvas pre-filled with
// rUnset.
func drawRing(w, h int16, value float32, style RingStyle, opts ...RingOption) *canvas {
	c := newCanvas(w, h)
	for i := range c.px {
		c.px[i] = rUnset
	}
	r := NewRing[float32](uint16(w), uint16(h), &value, 0, 100, style, opts...)
	ctx := ui.NewContext(c, uint16(w), uint16(h), 0, 0)
	r.Draw(&ctx)
	return c
}

// atAngle is the pixel at deg (clockwise from 12 o'clock) and radius rad
// from the centre of a w×h canvas.
func atAngle(c *canvas, deg, rad float64) color.RGBA {
	s, co := math.Sincos(deg * math.Pi / 180)
	x := float64(c.w)/2 + rad*s
	y := float64(c.h)/2 - rad*co
	return c.at(int16(math.Floor(x)), int16(math.Floor(y)))
}

// colours is the set of colours on the canvas, without the sentinel.
func colours(c *canvas) map[color.RGBA]int {
	m := map[color.RGBA]int{}
	for _, p := range c.px {
		if p != rUnset {
			m[p]++
		}
	}
	return m
}

var ringStyle = RingStyle{Track: rTrack, Fill: rFill}

// At 0 % the whole ring is the track, at 100 % the fill: one colour each
// (no BG, so the edges are hard and nothing else is painted).
func TestRingEmptyAndFullAreOneColour(t *testing.T) {
	for _, c := range []struct {
		value float32
		want  color.RGBA
	}{{0, rTrack}, {100, rFill}} {
		got := colours(drawRing(120, 120, c.value, ringStyle))
		check.Len(t, got, 1, "%v%%: colours %v", c.value, got)
		check.Positive(t, got[c.want], "%v%%: the ring is drawn in %v", c.value, c.want)
	}
}

// Values outside min..max are clamped: below min draws as 0 %, above max as
// 100 %; min and max given the wrong way round are swapped.
func TestRingClampsValue(t *testing.T) {
	empty, full := drawRing(60, 60, 0, ringStyle), drawRing(60, 60, 100, ringStyle)
	check.Equal(t, empty.px, drawRing(60, 60, -50, ringStyle).px, "below min")
	check.Equal(t, full.px, drawRing(60, 60, 250, ringStyle).px, "above max")

	v := float32(250)
	r := NewRing[float32](60, 60, &v, 100, 0, ringStyle) // min, max swapped
	c := newCanvas(60, 60)
	ctx := ui.NewContext(c, 60, 60, 0, 0)
	r.Draw(&ctx)
	for i, p := range full.px {
		if p == rUnset {
			full.px[i] = color.RGBA{}
		}
	}
	check.Equal(t, full.px, c.px, "min > max")
}

// 25 % clockwise from 12 o'clock fills the top-right quadrant only; a start
// angle of 90° moves it to the bottom-right, counter-clockwise mirrors it
// to the top-left.
func TestRingQuarterFillsTheRightQuadrant(t *testing.T) {
	const mid = 52 // radius 60, thickness 15: the ring is 45..60
	for _, c := range []struct {
		name string
		opts []RingOption
		fill float64 // centre of the filled quadrant
	}{
		{"clockwise from 0°", nil, 45},
		{"clockwise from 90°", []RingOption{WithStartAngle(90)}, 135},
		{"counter-clockwise from 0°", []RingOption{WithCounterClockwise()}, 315},
		{"counter-clockwise from 450° (= 90°)", []RingOption{WithCounterClockwise(), WithStartAngle(450)}, 45},
		{"clockwise from -90°", []RingOption{WithStartAngle(-90)}, 315},
	} {
		cv := drawRing(120, 120, 25, ringStyle, c.opts...)
		for q := 0.0; q < 360; q += 90 { // the quadrant centred at c.fill+q
			want := rTrack
			if q == 0 {
				want = rFill
			}
			for _, off := range []float64{-40, -20, 0, 20, 40} { // across the quadrant, clear of its edges
				deg := c.fill + q + off
				check.Equal(t, want, atAngle(cv, deg, mid), "%s: at %v°", c.name, math.Mod(deg+360, 360))
			}
		}
	}
}

// WithThickness sets the ring's radial extent: along the radius at 12
// o'clock exactly that many pixels are painted, then the hole; the default
// is diameter/8.
func TestRingThickness(t *testing.T) {
	for _, c := range []struct {
		opts []RingOption
		want int16
	}{{nil, 7}, {[]RingOption{WithThickness(10)}, 10}, {[]RingOption{WithThickness(3)}, 3}, {[]RingOption{WithThickness(0)}, 7}} {
		cv := drawRing(60, 60, 0, ringStyle, c.opts...)
		var n int16
		for y := int16(0); y < 30; y++ {
			if p := cv.at(30, y); p != rUnset {
				check.Equal(t, rTrack, p)
				check.Equal(t, n, y, "the ring is one run from the top edge")
				n++
			}
		}
		check.Equal(t, c.want, n, "thickness")
	}
	// past the radius: a full disc
	cv := drawRing(60, 60, 0, ringStyle, WithThickness(100))
	check.Equal(t, rTrack, cv.at(30, 30), "the centre of a full disc")
}

// A round cap paints a half disc past each end of the arc, where a butt
// cap leaves the track; at 0 % and 100 % there are no caps.
func TestRingRoundCapPaintsPastTheEnds(t *testing.T) {
	const mid = 52.5                // cap radius 7.5 px
	past := 5 / mid * 180 / math.Pi // 5 px along the mid radius, in degrees
	butt := drawRing(120, 120, 25, ringStyle)
	round := drawRing(120, 120, 25, ringStyle, WithCap(RingCapRound))
	for _, deg := range []float64{90 + past, -past} {
		check.Equal(t, rTrack, atAngle(butt, deg, mid), "butt, %.1f°", deg)
		check.Equal(t, rFill, atAngle(round, deg, mid), "round, %.1f°", deg)
	}
	check.Equal(t, rTrack, atAngle(round, 90+3*past, mid), "the cap ends")
	check.Equal(t, drawRing(120, 120, 0, ringStyle).px, drawRing(120, 120, 0, ringStyle, WithCap(RingCapRound)).px, "0 %: no caps")
	check.Equal(t, drawRing(120, 120, 100, ringStyle).px, drawRing(120, 120, 100, ringStyle, WithCap(RingCapRound)).px, "100 %: no caps")
}

// BG paints the box's corners and the ring's hole, and the edges blend into
// it; with no BG they are left alone.
func TestRingBackground(t *testing.T) {
	plain := drawRing(100, 60, 50, ringStyle)
	check.Equal(t, rUnset, plain.at(0, 0), "corner, no BG")
	check.Equal(t, rUnset, plain.at(50, 30), "hole, no BG")
	check.Equal(t, rUnset, plain.at(19, 30), "beside the ring, no BG")
	check.Equal(t, rTrack, plain.at(20, 30), "the ring is centred in its box: 20 px each side")
	check.Equal(t, rUnset, plain.at(80, 30), "right of the ring")
	check.Equal(t, rFill, plain.at(79, 30))

	style := ringStyle
	style.BG = rBG
	bg := drawRing(100, 60, 50, style)
	for _, p := range [][2]int16{{0, 0}, {50, 30}, {19, 30}, {99, 59}, {80, 30}} {
		check.Equal(t, rBG, bg.at(p[0], p[1]), "BG at %v", p)
	}
	_, unset := colours(bg)[rUnset]
	check.False(t, unset)
	check.Greater(t, len(colours(bg)), 3, "edge pixels are blends")
}

// The percent label shows the percentage rounded down ("100%" only with a
// full ring), centred in the ring: its
// pixels are those of the same text drawn by WithLabel, and the box of its
// pixels is centred within half a pixel.
func TestRingPercentLabel(t *testing.T) {
	style := RingStyle{Track: rTrack, Fill: rFill, Font: &proggy.TinySZ8pt7b, Label: rLabel}
	for _, c := range []struct {
		value float32
		text  string
	}{{37.4, "37%"}, {37.6, "37%"}, {38, "38%"}, {29, "29%"}, {5, "5%"}, {0, "0%"}, {0.4, "0%"}, {99.6, "99%"}, {99.99, "99%"}, {100, "100%"}} {
		got := drawRing(120, 120, c.value, style, WithPercentLabel())
		want := drawRing(120, 120, c.value, style, WithLabel(func() string { return c.text }))
		check.Equal(t, want.px, got.px, "%v → %q", c.value, c.text)

		x0, y0, x1, y1 := int16(120), int16(120), int16(-1), int16(-1)
		for y := range int16(120) {
			for x := range int16(120) {
				if got.at(x, y) == rLabel {
					x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x+1), max(y1, y+1)
				}
			}
		}
		check.Positive(t, x1-x0, "%q is drawn", c.text)
		check.LessOrEqual(t, abs16(x0+x1-120), 1, "%q centred across: %d..%d", c.text, x0, x1)
		check.LessOrEqual(t, abs16(y0+y1-120), 1, "%q centred down: %d..%d", c.text, y0, y1)
	}
	// WithLabel overrides WithPercentLabel, whichever comes first
	a := drawRing(120, 120, 50, style, WithLabel(func() string { return "ok" }), WithPercentLabel())
	b := drawRing(120, 120, 50, style, WithPercentLabel(), WithLabel(func() string { return "ok" }))
	check.Equal(t, a.px, b.px)
	// no font: no label
	style.Font = nil
	_, drawn := colours(drawRing(120, 120, 50, style, WithPercentLabel()))[rLabel]
	check.False(t, drawn, "no font, no label")
}

func abs16(v int16) int16 { return max(v, -v) }

// On a clipping display (one band of a framebuffer) the ring draws only the
// rows inside the clip, and those as a full draw does.
func TestRingDrawsOnlyClippedRows(t *testing.T) {
	style := RingStyle{Track: rTrack, Fill: rFill, BG: rBG, Font: &proggy.TinySZ8pt7b, Label: rLabel}
	full := drawRing(120, 120, 63, style, WithPercentLabel(), WithCap(RingCapRound))
	v := float32(63)
	r := NewRing[float32](120, 120, &v, 0, 100, style, WithPercentLabel(), WithCap(RingCapRound))
	for _, band := range [][2]int16{{0, 40}, {40, 40}, {55, 10}, {80, 40}} {
		c := &clippedCanvas{canvas: newCanvas(120, 120), cy: band[0], ch: band[1]}
		for i := range c.px {
			c.px[i] = rUnset
		}
		ctx := ui.NewContext(c, 120, 120, 0, 0)
		r.Draw(&ctx)
		for y := range int16(120) {
			for x := range int16(120) {
				want := rUnset
				if y >= band[0] && y < band[0]+band[1] {
					want = full.at(x, y)
				}
				// the label is drawn whole when it overlaps the band (the
				// framebuffer clips it; this canvas doesn't)
				if got := c.at(x, y); got != want && (want != rUnset || got != rLabel) {
					t.Fatalf("band %v: pixel (%d, %d) = %v, want %v", band, x, y, got, want)
				}
			}
		}
	}
}

// The value is read on every draw: changing it moves the arc.
func TestRingFollowsTheValue(t *testing.T) {
	c := newCanvas(60, 60)
	v := float32(0)
	r := NewRing[float32](60, 60, &v, 0, 100, ringStyle)
	ctx := ui.NewContext(c, 60, 60, 0, 0)
	r.Draw(&ctx)
	check.Equal(t, rTrack, atAngle(c, 135, 26))
	v = 50
	r.Draw(&ctx)
	check.Equal(t, rFill, atAngle(c, 135, 26))
	check.Equal(t, rTrack, atAngle(c, 225, 26))
}

// A Ring is a widget like the others: sized, drawable, not interactive.
func TestRingIsAWidget(t *testing.T) {
	v := 3
	var w ui.Widget = NewRing(40, 30, &v, 0, 10, ringStyle)
	width, height := w.Size()
	check.Equal(t, [2]uint16{40, 30}, [2]uint16{width, height})
	check.False(t, w.Interact(ui.ENTER))
	ctx := ui.NewContext(nil, 40, 30, 0, 0)
	w.Draw(&ctx) // no display: nothing to do
	var nilValue *int
	ctx = ui.NewContext(newCanvas(40, 30), 40, 30, 0, 0)
	NewRing(40, 30, nilValue, 0, 10, ringStyle).Draw(&ctx)
}
