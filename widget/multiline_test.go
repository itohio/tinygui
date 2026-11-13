package widget

import (
	"runtime"
	"strconv"
	"testing"
)

func mallocs() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Mallocs
}

var countProbe []*[16]byte

// needMallocCount skips an allocation test on a runtime that counts no
// allocations: TinyGo's default host GC (Boehm) reports Mallocs = 0, which
// would pass every check. Run TinyGo with `tinygo test -gc=leaking`.
func needMallocCount(t *testing.T) {
	t.Helper()
	m0 := mallocs()
	countProbe = append(countProbe, new([16]byte))
	if mallocs() == m0 {
		t.Skip("this runtime counts no allocations (TinyGo's host Boehm GC): run tinygo test -gc=leaking")
	}
}

// A log keeps its newest capacity entries, oldest first, and appending
// allocates nothing from the first line on: SetCapacity reserves the room,
// and a full log reuses it (it used to copy the whole history per line).
// Counted exactly over the whole sequence (review TG-1-postcommit R2: an
// average per call rounds a few reallocations down to zero).
func TestLogKeepsTheNewestEntriesWithoutAllocating(t *testing.T) {
	needMallocCount(t)
	l := NewInteractiveLog(100, 10, 4)
	l.SetCapacity(8)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strconv.Itoa(i)
	}
	m0 := mallocs()
	for i := range 200 {
		l.Append(lines[i%20])
	}
	if n := mallocs() - m0; n != 0 {
		t.Fatalf("%d allocations in 200 appends from an empty log, want 0", n)
	}
	// plain checks, no testify: this test runs under TinyGo too
	if len(l.lines) != 8 {
		t.Fatalf("%d entries kept, want 8", len(l.lines))
	}
	for i, s := range l.lines {
		if s != lines[12+i] {
			t.Fatalf("entries %v, want the newest 8, oldest first: %v", l.lines, lines[12:])
		}
	}
}
