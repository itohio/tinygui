//go:build !tinygo

package container_test

import "runtime"

// mallocCount is the Go runtime's cumulative allocation count.
func mallocCount() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Mallocs
}
