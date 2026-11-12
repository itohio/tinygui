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

// A pool burns for a while (not out at once, not forever), and when the
// fuel is gone a new splash drops in.
func TestFireBurnsItsFuelThenRefuels(t *testing.T) {
	f := newFire()
	start := f.count()
	steps := 0
	for f.count() > 0 && steps < 20000 {
		f.step()
		steps++
	}
	if steps < 600 || steps >= 20000 {
		t.Fatalf("a %d-cell pool burnt out after %d steps, want 600..20000 (~25 s .. 13 min)", start, steps)
	}
	for range fireRefuel + 2 {
		f.step()
	}
	if f.count() == 0 {
		t.Fatal("no new fuel after the pool burnt out")
	}
}
