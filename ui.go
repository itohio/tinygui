package ui

import (
	"machine"
	"time"
)

type UserCommand byte

const (
	IDLE UserCommand = iota
	UP
	DOWN
	LEFT
	RIGHT
	NEXT
	PREV
	ENTER
	ESC
	BACK
	DEL
	RESET
	SAVE
	LOAD
	LONG_UP
	LONG_DOWN
	LONG_LEFT
	LONG_RIGHT
	LONG_ENTER
	LONG_ESC
	LONG_BACK
	LONG_DEL
	LONG_RESET
	USER UserCommand = 64
)

// PeekButton checks the state of the button depending on the time it was pressed.
func PeekButton(p machine.Pin) time.Duration {
	now := time.Now()
	time.Sleep(time.Millisecond * 10)
	for !p.Get() {
		time.Sleep(time.Millisecond)
		machine.Watchdog.Update()
	}
	return time.Since(now)
}

// DurationAdd is a helper function to increase duration in convenient way.
func DurationAdd(dur, delta time.Duration) time.Duration {
	l := pseudo_pow10(pseudo_log10(dur) - 1)

	return dur + time.Duration(delta)*l
}

// DurationSub is a helper function to increase duration in convenient way.
func DurationSub(dur, delta time.Duration) time.Duration {
	l := pseudo_pow10(pseudo_log10(dur) - 1)

	return dur - time.Duration(delta)*l
}

// pseudo_pow10 computes 10^exp, where exp is given as a time.Duration.
// It treats the duration (in nanoseconds) as the integer exponent.
// If exp is negative, it returns 0 since 10^negative is undefined for integers.
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

// pseudo_log10 calculates the integer base-10 logarithm of n, where n is a time.Duration.
// It returns the largest time.Duration `x` such that 10^x <= n.
// If n <= 0, it returns an error.
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
