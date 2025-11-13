package animation

// Animator describes a time-based interpolator operating on multiple value channels.
type Animator interface {
	// Start begins an animation between start and end channel values at startUnixMicro (microseconds since Unix epoch).
	// The provided slices must remain valid until another Start call.
	Start(start, end []float32, startUnixMicro int64)
	// Update advances the animation to nowUnixMicro, writing interpolated values into dst in place and returning true when finished.
	Update(dst []float32, nowUnixMicro int64) bool
}
