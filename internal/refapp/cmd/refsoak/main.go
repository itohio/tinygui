// Command refsoak runs the reference app's menu-driven soak: after setup it
// must not allocate. Run it with the garbage collector disabled,
//
//	tinygo run -gc=leaking ./internal/refapp/cmd/refsoak
//
// and list the library's heap allocation sites with
//
//	tinygo build -o /dev/null -print-allocs=github.com/itohio/tinygui ./internal/refapp/cmd/refsoak
package main

import (
	"os"
	"runtime"

	"github.com/itohio/tinygui/internal/refapp"
)

func main() {
	a := refapp.New()
	for range 3 {
		a.Round()
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	m0 := ms.Mallocs
	for range 50 {
		a.Round()
	}
	runtime.ReadMemStats(&ms)
	if n := ms.Mallocs - m0; n != 0 {
		println("refsoak:", n, "allocations in 50 rounds after setup")
		os.Exit(1)
	}
	println("refsoak: 50 rounds, 0 allocations after setup")
}
