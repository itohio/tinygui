package app

import (
	"math"
	"testing"
)

// rotX gives a sensor reading of a world vector when the sensor is rolled
// by angle about x (sensor = R^T · world).
func rotX(angle float64, v [3]float32) [3]float32 {
	c, s := float32(math.Cos(angle)), float32(math.Sin(angle))
	return [3]float32{v[0], c*v[1] + s*v[2], -s*v[1] + c*v[2]}
}

func near(t *testing.T, what string, got, want [3]float32, tol float32) {
	t.Helper()
	for i := range got {
		if d := got[i] - want[i]; d > tol || d < -tol {
			t.Fatalf("%s = %v, want %v (±%.2f)", what, got, want, tol)
		}
	}
}

// A badge held still at a 40° roll: from the identity guess the filter must
// settle so the sensor's measured gravity maps back to world up.
func TestMahonyFindsTiltFromGravity(t *testing.T) {
	f := newMahony()
	a := rotX(40*math.Pi/180, [3]float32{0, 0, 1})
	for range 2000 { // 20 s at 100 Hz
		f.update([3]float32{}, a, nil, 0.01)
	}
	near(t, "gravity in world", f.toWorld(a), [3]float32{0, 0, 1}, 0.02)
}

// A gyro with a constant bias on a still badge: the integral feedback
// learns the roll/pitch bias, so the orientation doesn't drift away.
func TestMahonyCancelsGyroBias(t *testing.T) {
	f := newMahony()
	a := [3]float32{0, 0, 1}
	bias := [3]float32{0.02, -0.03, 0}
	for range 30000 { // 5 min at 100 Hz
		f.update(bias, a, nil, 0.01)
	}
	near(t, "gravity in world", f.toWorld(a), [3]float32{0, 0, 1}, 0.02)
}

// Turning at 90°/s for 1 s about the vertical turns the heading by 90°.
func TestMahonyIntegratesYawRate(t *testing.T) {
	f := newMahony()
	for range 100 {
		f.update([3]float32{0, 0, math.Pi / 2}, [3]float32{0, 0, 1}, nil, 0.01)
	}
	if y := f.yaw() * 180 / math.Pi; y < 88 || y > 92 {
		t.Fatalf("yaw %.1f°, want 90°", y)
	}
}

// With a magnetometer the heading converges to magnetic north even when the
// gyro says nothing: the badge faces 60° east of north, the field (with a
// downward dip) arrives rotated by -60° in the sensor frame.
func TestMahonyHeadingFromMagnetometer(t *testing.T) {
	f := newMahony()
	h := -60 * math.Pi / 180
	field := [3]float32{float32(0.4 * math.Cos(h)), float32(0.4 * math.Sin(h)), -0.3}
	for range 3000 {
		f.update([3]float32{}, [3]float32{0, 0, 1}, &field, 0.01)
	}
	if y := f.yaw() * 180 / math.Pi; y < 58 || y > 62 {
		t.Fatalf("yaw %.1f°, want 60°", y)
	}
}
