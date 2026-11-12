package widget

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// A full log keeps its newest capacity entries, oldest first, and appending
// to it allocates nothing (it used to copy the whole history per line).
func TestLogKeepsTheNewestEntriesWithoutAllocating(t *testing.T) {
	l := NewInteractiveLog(100, 10, 4)
	l.SetCapacity(8)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strconv.Itoa(i)
	}
	for _, s := range lines {
		l.Append(s)
	}
	require.Equal(t, lines[12:], l.lines)

	i := 0
	allocs := testing.AllocsPerRun(100, func() {
		l.Append(lines[i%20])
		i++
	})
	require.Zero(t, allocs)
	require.Len(t, l.lines, 8)
}
