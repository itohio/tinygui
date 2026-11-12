package app

import "testing"

// centroid of the cells where pick(i) holds.
func centroid(pick func(i int) bool) (float64, float64, bool) {
	var sx, sy, n float64
	for i := range fireW * fireH {
		if pick(i) {
			sx, sy, n = sx+float64(i%fireW), sy+float64(i/fireW), n+1
		}
	}
	return sx / n, sy / n, n > 0
}

// The flames rise against gravity, whichever way it points: tilting the
// badge tilts the fire.
func TestFireFlamesRiseAgainstGravity(t *testing.T) {
	for _, d := range []int{0, 2, 4, 6} {
		f := newFire()
		f.dir = d
		for range 120 {
			f.step()
		}
		fx, fy, okF := centroid(func(i int) bool { return f.state[i]&liquidBit != 0 })
		hx, hy, okH := centroid(func(i int) bool { return f.heat[i] > fireMax/2 })
		if !okF || !okH {
			t.Fatalf("gravity %s: fuel %v flames %v", gravityNames[d], okF, okH)
		}
		g := gravityVec[d]
		if along := (hx-fx)*float64(g[0]) + (hy-fy)*float64(g[1]); along > -2 {
			t.Errorf("gravity %s: flames are %.1f cells along gravity from the fuel, want above it (< -2)", gravityNames[d], along)
		}
	}
}

// The fuel is never used up (owner: "liquid never consumed"): it only
// moves, through gravity changes too, and splashes add to it up to a cap.
func TestFireFuelIsNeverConsumed(t *testing.T) {
	f := newFire()
	n := f.count()
	for s := range 3000 {
		if s%500 == 250 {
			f.turn(1)
		}
		f.step()
	}
	if got := f.count(); got != n {
		t.Fatalf("fuel %d after 3000 steps, want %d", got, n)
	}
	for range 20 {
		f.splash()
	}
	if got := f.count(); got > fireMaxFuel {
		t.Fatalf("splashes grew the fuel to %d, over the %d cap", got, fireMaxFuel)
	}
}
