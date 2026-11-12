package app

import "unsafe"

// textBuf formats numbers into a fixed buffer and exposes the result as a
// string without allocating. The string aliases the buffer: use it before the
// next reset (labels read it during Draw, which is exactly that).
type textBuf struct {
	b [48]byte
	n int
}

func (t *textBuf) reset() *textBuf { t.n = 0; return t }

func (t *textBuf) str(s string) *textBuf {
	t.n += copy(t.b[t.n:], s)
	return t
}

func (t *textBuf) int(v int) *textBuf {
	if v < 0 {
		t.str("-")
		v = -v
	}
	var tmp [12]byte
	i := len(tmp)
	for {
		i--
		tmp[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	return t.str(unsafe.String(&tmp[i], len(tmp)-i))
}

// pad2 writes v with at least two digits.
func (t *textBuf) pad2(v int) *textBuf {
	if v < 10 {
		t.str("0")
	}
	return t.int(v)
}

// fixed writes v with the given number of decimals (rounded).
func (t *textBuf) fixed(v float32, decimals int) *textBuf {
	scale := 1
	for range decimals {
		scale *= 10
	}
	neg := v < 0
	if neg {
		v = -v
	}
	n := int(v*float32(scale) + 0.5)
	if neg && n != 0 {
		t.str("-")
	}
	t.int(n / scale)
	if decimals > 0 {
		t.str(".")
		frac := n % scale
		for d := scale / 10; d > 0; d /= 10 {
			t.b[t.n] = byte('0' + frac/d%10)
			t.n++
		}
	}
	return t
}

func (t *textBuf) String() string {
	if t.n == 0 {
		return ""
	}
	return unsafe.String(&t.b[0], t.n)
}
