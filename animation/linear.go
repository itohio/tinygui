package animation

// linear implements Animator using linear interpolation between channels.
type linear struct {
	start      []float32
	end        []float32
	channels   int
	startUS    int64
	durationUS int64
	active     bool
}

// NewLinear returns an Animator that interpolates channels linearly over durationUS microseconds.
func NewLinear(durationUS int64) Animator {
	return &linear{durationUS: durationUS}
}

func (l *linear) Start(start, end []float32, startUnixMicro int64) {
	if len(start) == 0 || len(start) != len(end) {
		l.active = false
		l.channels = 0
		return
	}
	l.start = start
	l.end = end
	l.channels = len(start)
	l.startUS = startUnixMicro
	l.active = true
}

func (l *linear) Update(dst []float32, nowUnixMicro int64) bool {
	if !l.active || l.channels == 0 || len(dst) < l.channels {
		return true
	}

	if l.durationUS <= 0 {
		copy(dst[:l.channels], l.end[:l.channels])
		l.active = false
		return true
	}

	elapsed := nowUnixMicro - l.startUS
	var t float32
	switch {
	case elapsed <= 0:
		t = 0
	case elapsed >= l.durationUS:
		t = 1
	default:
		t = float32(elapsed) / float32(l.durationUS)
	}

	for i := 0; i < l.channels; i++ {
		start := l.start[i]
		dst[i] = start + (l.end[i]-start)*t
	}

	if t >= 1 {
		copy(dst[:l.channels], l.end[:l.channels])
		l.active = false
		return true
	}

	return false
}
