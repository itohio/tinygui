package app

import "testing"

// Over many steps with the largest flock, including scatters, every boid
// stays on the grid and flies between the speed limits.
func TestBoidsStayOnGridWithinSpeed(t *testing.T) {
	b := newBoids()
	b.count = len(boidsN) - 1
	for s := range 2000 {
		if s%500 == 250 {
			b.scatter()
		}
		b.tick()
		for i := range boidsN[b.count] {
			if b.x[i] < 0 || b.y[i] < 0 || b.x[i] >= boidsW*boidsFix || b.y[i] >= boidsH*boidsFix {
				t.Fatalf("step %d boid %d off grid: %d,%d", s, i, b.x[i], b.y[i])
			}
			ax, ay := abs32(b.vx[i]), abs32(b.vy[i])
			mag := max(ax, ay) + min(ax, ay)>>2 + min(ax, ay)>>3
			// ±2 for integer rounding of the scaled components
			if mag > boidsVMax+2 || mag < boidsVMin-2 {
				t.Fatalf("step %d boid %d speed %d outside [%d, %d]", s, i, mag, boidsVMin, boidsVMax)
			}
		}
	}
}

func TestOctantFollowsHeading(t *testing.T) {
	for _, c := range []struct {
		vx, vy int32
		want   uint8
	}{{100, 0, 0}, {70, 70, 1}, {0, 100, 2}, {-70, 70, 3}, {-100, 0, 4}, {-70, -70, 5}, {0, -100, 6}, {70, -70, 7}, {100, -30, 0}, {30, 100, 2}} {
		if got := octant(c.vx, c.vy); got != c.want {
			t.Errorf("octant(%d,%d) = %d, want %d", c.vx, c.vy, got, c.want)
		}
	}
}
