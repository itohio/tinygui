//go:build tinygo

package refapp

import _ "unsafe" // go:linkname

// mallocs reads TinyGo's own allocation counter, which every collector
// (Boehm, leaking, conservative, precise) increments per heap allocation.
// runtime.MemStats.Mallocs is not usable: the default host GC (Boehm)
// reports it as 0. TinyGo's testing.AllocsPerRun reads the same counter.
//
//go:linkname mallocs runtime.mallocs
func mallocs() uint64
