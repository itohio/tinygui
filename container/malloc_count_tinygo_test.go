//go:build tinygo

package container_test

import _ "unsafe" // go:linkname

// mallocCount reads TinyGo's own allocation counter, which every collector
// increments per heap allocation (runtime.MemStats.Mallocs is 0 under the
// default host GC, Boehm). TinyGo's testing.AllocsPerRun reads the same
// counter.
//
//go:linkname mallocCount runtime.mallocs
func mallocCount() uint64
