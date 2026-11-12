package app

import "math"

// The boids' world and the screen's box (owner: "the frame is the window
// into boids world ... by moving the walls around we don't move the boids
// since they are as if flying in the free space"). The flock lives in world
// coordinates (1/256 cell); the screen is a boidsW×boidsH box with a centre
// and an angle in that world. Turning or moving the box sweeps its walls
// through the flock; boids away from the walls keep their world heading.
// Buttons turn and shake the box now; an IMU's orientation will drive the
// same pose.

const (
	boxTurnStep = 16       // a button press turns the box 22.5° (256 = 360°)
	boxShake    = 10 * 256 // how far a shake throws the box (world units)
	boxShakeLen = 20       // ticks of a shake (out and back)
	q14         = 1 << 14
)

// sinQ14 is sin(2πi/256) in Q14.
var sinQ14 = func() (t [256]int32) {
	for i := range t {
		t[i] = int32(math.Round(math.Sin(2*math.Pi*float64(i)/256) * q14))
	}
	return
}()

type boxPose struct {
	angle, target int // current and wanted angle, unwrapped (256 per turn)
	spin          bool
	baseX, baseY  int32 // centre without the shake
	cx, cy        int32 // centre now
	shakeT        int   // ticks into a shake, 0 when still
	shakeDX       int32
	shakeDY       int32
	frame         int
}

func newBoxPose() boxPose {
	c := boxPose{baseX: boidsW * boidsFix / 2, baseY: boidsH * boidsFix / 2}
	c.cx, c.cy = c.baseX, c.baseY
	return c
}

func (p *boxPose) sincos() (int32, int32) {
	a := uint8(p.angle)
	return sinQ14[a], sinQ14[uint8(a+64)]
}

// advance moves the box one tick towards its target angle and along a shake.
func (p *boxPose) advance() {
	p.frame++
	if p.spin && p.frame&1 == 0 {
		p.target++
	}
	switch {
	case p.angle < p.target:
		p.angle++
	case p.angle > p.target:
		p.angle--
	}
	p.cx, p.cy = p.baseX, p.baseY
	if p.shakeT > 0 {
		// a triangle: out to the full throw at the middle, back at the end
		k := int32(boxShakeLen/2 - abs(p.shakeT-boxShakeLen/2))
		p.cx += p.shakeDX * k / (boxShakeLen / 2)
		p.cy += p.shakeDY * k / (boxShakeLen / 2)
		if p.shakeT++; p.shakeT > boxShakeLen {
			p.shakeT = 0
		}
	}
}

// shake throws the box in a random direction and back.
func (p *boxPose) shake() {
	a := uint8(rand32())
	p.shakeDX, p.shakeDY = sinQ14[uint8(a+64)]*boxShake/q14, sinQ14[a]*boxShake/q14
	p.shakeT = 1
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// toBox maps a world point into box coordinates (0..boidsW*boidsFix across).
// Division, not >>: the shift rounds negative values away from zero and
// would bias the flock (F-35's cousin in the boids, see step).
func (p *boxPose) toBox(x, y int32) (int32, int32) {
	s, c := p.sincos()
	dx, dy := x-p.cx, y-p.cy
	return (c*dx+s*dy)/q14 + boidsW*boidsFix/2, (-s*dx+c*dy)/q14 + boidsH*boidsFix/2
}

// toWorld maps box coordinates back to the world.
func (p *boxPose) toWorld(bx, by int32) (int32, int32) {
	s, c := p.sincos()
	dx, dy := bx-boidsW*boidsFix/2, by-boidsH*boidsFix/2
	return (c*dx-s*dy)/q14 + p.cx, (s*dx+c*dy)/q14 + p.cy
}

// vecToBox / vecToWorld rotate a velocity between the frames.
func (p *boxPose) vecToBox(vx, vy int32) (int32, int32) {
	s, c := p.sincos()
	return (c*vx + s*vy) / q14, (-s*vx + c*vy) / q14
}

func (p *boxPose) vecToWorld(vx, vy int32) (int32, int32) {
	s, c := p.sincos()
	return (c*vx - s*vy) / q14, (s*vx + c*vy) / q14
}
