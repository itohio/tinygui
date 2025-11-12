//go:build stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid || xiao

package ui

import (
	"machine"
	"time"
)

// PeekButton waits for a pin press to finish and returns how long the button
// was held down. Hardware watchdogs are serviced while waiting.
func PeekButton(p machine.Pin) time.Duration {
	now := time.Now()
	time.Sleep(time.Millisecond * 10)
	for !p.Get() {
		time.Sleep(time.Millisecond)
		updateWatchdog()
	}
	return time.Since(now)
}

// DurationAdd increases a duration by a delta scaled to the order of magnitude
// of the duration, providing a simple adjustment mechanism for timers.
func DurationAdd(dur, delta time.Duration) time.Duration {
	l := pseudo_pow10(pseudo_log10(dur) - 1)

	return dur + time.Duration(delta)*l
}

// DurationSub decreases a duration by a delta scaled to the order of magnitude
// of the duration, mirroring DurationAdd.
func DurationSub(dur, delta time.Duration) time.Duration {
	l := pseudo_pow10(pseudo_log10(dur) - 1)

	return dur - time.Duration(delta)*l
}

// pseudo_pow10 computes 10^exp where exp is the integer value of a duration.
func pseudo_pow10(exp time.Duration) time.Duration {
	if exp < 0 {
		return 0
	}

	result := time.Duration(1)
	for i := time.Duration(0); i < exp; i++ {
		result *= 10
	}
	return result
}

// pseudo_log10 returns the integer base-10 logarithm for positive durations.
func pseudo_log10(n time.Duration) time.Duration {
	if n <= 0 {
		return 0
	}

	log := time.Duration(0)
	for n >= 10 {
		n /= 10
		log++
	}
	return log
}
