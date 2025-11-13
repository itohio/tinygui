package widget

// Number captures numeric types supported by TinyGUI interactive widgets.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func clamp[T Number](min, max, v T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

const maxGaugeSegments = 4
