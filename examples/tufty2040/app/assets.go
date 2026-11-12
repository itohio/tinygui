package app

// Procedural bitmaps: no asset files, no RAM beyond the pixel arrays.

// sine256 is a quarter-wave-free full sine table, values 0..255.
var sine256 = func() (t [256]uint8) {
	// parabolic approximation of sin, good enough for a plasma effect
	for i := range 256 {
		x := i % 128
		v := x * (127 - x) * 4 / 127 // 0..127
		if i < 128 {
			t[i] = uint8(128 + v)
		} else {
			t[i] = uint8(127 - v)
		}
	}
	return
}()

func rgb565(r, g, b uint8) uint16 {
	return uint16(r&0xF8)<<8 | uint16(g&0xFC)<<3 | uint16(b)>>3
}

// renderPlasma fills a w×h RGB565 buffer with an animated plasma.
func renderPlasma(dst []uint16, w, h, phase int) {
	for y := range h {
		for x := range w {
			v := int(sine256[(x*4+phase)&255]) + int(sine256[(y*6+phase*2)&255]) + int(sine256[((x+y)*3+phase*3)&255])
			v /= 3
			r := sine256[(v+phase)&255]
			g := sine256[(v+85+phase/2)&255]
			b := sine256[(v+170)&255]
			dst[y*w+x] = rgb565(r, g, b)
		}
	}
}

// renderGradient fills a w×h RGB332 buffer (one byte per pixel) with a hue/value ramp.
func renderGradient(dst []uint8, w, h int) {
	for y := range h {
		for x := range w {
			r := uint8(x * 8 / w)
			g := uint8(y * 8 / h)
			b := uint8((x + y) * 4 / (w + h))
			dst[y*w+x] = r<<5 | g<<2 | b
		}
	}
}
