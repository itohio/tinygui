//go:build !(esp32 || esp32c3 || rp2040)

package ui

// updateWatchdog is a no-op on boards without watchdog support.
func updateWatchdog() {}
