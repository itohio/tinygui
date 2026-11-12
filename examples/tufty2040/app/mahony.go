package app

import "math"

// mahony is Mahony's complementary filter for orientation from a 6-axis
// (gyro + accelerometer) or 9-axis (+ magnetometer) IMU: the gyro
// integrates the orientation; the measured gravity (and magnetic field)
// pull it back with a proportional-integral feedback, so gyro drift is
// cancelled for roll and pitch (and for yaw when there is a magnetometer).
//
// q is the rotation from the sensor frame to the world frame (w, x, y, z).
// World frame: z up; with a magnetometer x points to magnetic north.
type mahony struct {
	q      [4]float32
	kp, ki float32
	ix     [3]float32 // integral feedback (gyro bias estimate)
}

func newMahony() *mahony {
	return &mahony{q: [4]float32{1, 0, 0, 0}, kp: 2, ki: 0.005}
}

func invSqrt(x float32) float32 { return 1 / float32(math.Sqrt(float64(x))) }

// update advances by dt seconds with gyro g (rad/s), accelerometer a (any
// unit; only its direction is used) and magnetometer m (nil without one).
func (f *mahony) update(g, a [3]float32, m *[3]float32, dt float32) {
	q0, q1, q2, q3 := f.q[0], f.q[1], f.q[2], f.q[3]
	var ex, ey, ez float32
	if n := a[0]*a[0] + a[1]*a[1] + a[2]*a[2]; n > 0 {
		r := invSqrt(n)
		ax, ay, az := a[0]*r, a[1]*r, a[2]*r
		// gravity (world up) seen in the sensor frame by the current estimate
		vx := 2 * (q1*q3 - q0*q2)
		vy := 2 * (q0*q1 + q2*q3)
		vz := q0*q0 - q1*q1 - q2*q2 + q3*q3
		ex, ey, ez = ay*vz-az*vy, az*vx-ax*vz, ax*vy-ay*vx
		if m != nil {
			if n := m[0]*m[0] + m[1]*m[1] + m[2]*m[2]; n > 0 {
				r := invSqrt(n)
				mx, my, mz := m[0]*r, m[1]*r, m[2]*r
				// the field in the world frame, its horizontal part on x
				hx := 2 * (mx*(0.5-q2*q2-q3*q3) + my*(q1*q2-q0*q3) + mz*(q1*q3+q0*q2))
				hy := 2 * (mx*(q1*q2+q0*q3) + my*(0.5-q1*q1-q3*q3) + mz*(q2*q3-q0*q1))
				bx := float32(math.Sqrt(float64(hx*hx + hy*hy)))
				bz := 2 * (mx*(q1*q3-q0*q2) + my*(q2*q3+q0*q1) + mz*(0.5-q1*q1-q2*q2))
				// that reference seen in the sensor frame
				wx := 2 * (bx*(0.5-q2*q2-q3*q3) + bz*(q1*q3-q0*q2))
				wy := 2 * (bx*(q1*q2-q0*q3) + bz*(q0*q1+q2*q3))
				wz := 2 * (bx*(q0*q2+q1*q3) + bz*(0.5-q1*q1-q2*q2))
				ex += my*wz - mz*wy
				ey += mz*wx - mx*wz
				ez += mx*wy - my*wx
			}
		}
		f.ix[0] += f.ki * ex * dt
		f.ix[1] += f.ki * ey * dt
		f.ix[2] += f.ki * ez * dt
	}
	gx := g[0] + f.kp*ex + f.ix[0]
	gy := g[1] + f.kp*ey + f.ix[1]
	gz := g[2] + f.kp*ez + f.ix[2]
	h := 0.5 * dt
	q0, q1, q2, q3 = q0+(-q1*gx-q2*gy-q3*gz)*h,
		q1+(q0*gx+q2*gz-q3*gy)*h,
		q2+(q0*gy-q1*gz+q3*gx)*h,
		q3+(q0*gz+q1*gy-q2*gx)*h
	r := invSqrt(q0*q0 + q1*q1 + q2*q2 + q3*q3)
	f.q = [4]float32{q0 * r, q1 * r, q2 * r, q3 * r}
}

// toWorld rotates a sensor-frame vector into the world frame.
func (f *mahony) toWorld(v [3]float32) [3]float32 {
	q0, q1, q2, q3 := f.q[0], f.q[1], f.q[2], f.q[3]
	return [3]float32{
		(1-2*(q2*q2+q3*q3))*v[0] + 2*(q1*q2-q0*q3)*v[1] + 2*(q1*q3+q0*q2)*v[2],
		2*(q1*q2+q0*q3)*v[0] + (1-2*(q1*q1+q3*q3))*v[1] + 2*(q2*q3-q0*q1)*v[2],
		2*(q1*q3-q0*q2)*v[0] + 2*(q2*q3+q0*q1)*v[1] + (1-2*(q1*q1+q2*q2))*v[2],
	}
}

// yaw is the heading in radians (rotation about world z).
func (f *mahony) yaw() float32 {
	q0, q1, q2, q3 := f.q[0], f.q[1], f.q[2], f.q[3]
	return float32(math.Atan2(float64(2*(q0*q3+q1*q2)), float64(1-2*(q2*q2+q3*q3))))
}
