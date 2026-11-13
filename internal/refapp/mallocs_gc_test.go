//go:build !tinygo

package refapp

import (
	"runtime"
	"strings"
)

// Every allocation goes into the heap profile (rate 1: exact counts).
func init() { runtime.MemProfileRate = 1 }

var (
	profile []runtime.MemProfileRecord
	counted = map[[32]uintptr]bool{}
)

// mallocs counts the heap allocations made with module code on the stack,
// except the Go runtime's own type-assertion and type-switch caches.
//
// runtime.MemStats.Mallocs can't be compared exactly on Go: the runtime
// fills those per-call-site caches lazily — on a random ~1/1024 of misses —
// from inside any type assertion (library code included), so a count over
// a sequence would vary from run to run. Exact counts per call stack come
// from the heap profile instead; runtime.GC publishes it up to now.
// Allocations made here (the record buffer, the cache) are not counted.
func mallocs() uint64 {
	runtime.GC()
	n, ok := runtime.MemProfile(profile, true)
	for !ok {
		profile = make([]runtime.MemProfileRecord, n+n/2+64)
		n, ok = runtime.MemProfile(profile, true)
	}
	var total uint64
	for i := range profile[:n] {
		r := &profile[i]
		count, seen := counted[r.Stack0]
		if !seen {
			count = libraryAllocation(r.Stack())
			counted[r.Stack0] = count
		}
		if count {
			total += uint64(r.AllocObjects)
		}
	}
	return total
}

// libraryAllocation reports whether an allocation's stack runs through the
// module's code, and not through a runtime cache fill or mallocs itself.
func libraryAllocation(stack []uintptr) bool {
	module := false
	frames := runtime.CallersFrames(stack)
	for {
		f, more := frames.Next()
		switch f.Function {
		case "runtime.buildTypeAssertCache", "runtime.buildInterfaceSwitchCache",
			"github.com/itohio/tinygui/internal/refapp.mallocs":
			return false
		}
		if strings.HasPrefix(f.Function, "github.com/itohio/tinygui") {
			module = true
		}
		if !more {
			return module
		}
	}
}
