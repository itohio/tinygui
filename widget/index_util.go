package widget

func wrapIndex(idx, length int) int {
	if length <= 0 {
		return 0
	}
	for idx < 0 {
		idx += length
	}
	for idx >= length {
		idx -= length
	}
	return idx
}
