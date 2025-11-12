//go:build rp2040

package ui

import "machine"

func updateWatchdog() {
	machine.Watchdog.Update()
}
