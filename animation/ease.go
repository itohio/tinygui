package animation

// easingAnimator reuses linear behaviour with a custom easing curve.
type easingAnimator struct {
	linear
	ease func(float32) float32
}

// NewEaseIn creates an animator with a quadratic ease-in curve.
func NewEaseIn(durationUS int64) Animator {
	anim := &easingAnimator{
		linear: linear{durationUS: durationUS},
		ease: func(t float32) float32 {
			return t * t
		},
	}
	return anim
}

// NewEaseOut creates an animator with a quadratic ease-out curve.
func NewEaseOut(durationUS int64) Animator {
	anim := &easingAnimator{
		linear: linear{durationUS: durationUS},
		ease: func(t float32) float32 {
			inv := 1 - t
			return 1 - inv*inv
		},
	}
	return anim
}

// NewEaseInOut creates an animator with a smoothstep ease-in-out curve.
func NewEaseInOut(durationUS int64) Animator {
	anim := &easingAnimator{
		linear: linear{durationUS: durationUS},
		ease: func(t float32) float32 {
			return t * t * (3 - 2*t)
		},
	}
	return anim
}

func (e *easingAnimator) Update(dst []float32, nowUnixMicro int64) bool {
	if !e.linear.active || e.linear.channels == 0 || len(dst) < e.linear.channels {
		return true
	}

	if e.linear.durationUS <= 0 {
		copy(dst[:e.linear.channels], e.linear.end[:e.linear.channels])
		e.linear.active = false
		return true
	}

	elapsed := nowUnixMicro - e.linear.startUS
	var t float32
	switch {
	case elapsed <= 0:
		t = 0
	case elapsed >= e.linear.durationUS:
		t = 1
	default:
		t = float32(elapsed) / float32(e.linear.durationUS)
	}

	eased := e.ease(t)
	for i := 0; i < e.linear.channels; i++ {
		start := e.linear.start[i]
		dst[i] = start + (e.linear.end[i]-start)*eased
	}

	if t >= 1 {
		copy(dst[:e.linear.channels], e.linear.end[:e.linear.channels])
		e.linear.active = false
		return true
	}

	return false
}
