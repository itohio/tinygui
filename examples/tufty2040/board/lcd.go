//go:build tufty2040

// Package board drives the Pimoroni Tufty 2040's panel: an ST7789 320×240 on
// an 8-bit parallel bus (8080 write cycles), bit-banged through the RP2040 SIO
// set/clear registers. Init sequence and orientation follow Pimoroni's driver
// (pimoroni-pico drivers/st7789, DISPLAY_TUFTY_2040: ROTATE_180 because the
// panel is mounted upside down).
package board

import (
	"device/arm"
	"device/rp"
	"machine"
	"time"
)

const (
	Width  = 320
	Height = 240

	lcdDataShift = 14 // LCD_DB0..DB7 = GPIO14..21
	lcdDataMask  = uint32(0xFF) << lcdDataShift
	lcdWRMask    = uint32(1) << 12 // LCD_WR = GPIO12

	// MADCTLRotate180 is ROW_ORDER | SWAP_XY | SCAN_ORDER (landscape, upside down).
	MADCTLRotate180 = 0x80 | 0x20 | 0x10
)

// LCD is the Tufty's ST7789 panel.
type LCD struct{}

// Configure sets up the bus pins and runs the panel init sequence.
func (l LCD) Configure() {
	for _, p := range []machine.Pin{
		machine.LCD_CS, machine.LCD_DC, machine.LCD_WR, machine.LCD_RD,
		machine.LCD_DB0, machine.LCD_DB1, machine.LCD_DB2, machine.LCD_DB3,
		machine.LCD_DB4, machine.LCD_DB5, machine.LCD_DB6, machine.LCD_DB7,
	} {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	machine.LCD_CS.High()
	machine.LCD_RD.High() // never read
	machine.LCD_WR.High()

	l.Command(0x01) // SWRESET
	time.Sleep(150 * time.Millisecond)
	l.Command(0x35)       // TEON
	l.Command(0x3A, 0x05) // COLMOD: 16 bpp
	l.Command(0xB2, 0x0C, 0x0C, 0x00, 0x33, 0x33)
	l.Command(0xC0, 0x2C)
	l.Command(0xC2, 0x01)
	l.Command(0xC3, 0x12)
	l.Command(0xC4, 0x20)
	l.Command(0xD0, 0xA4, 0xA1)
	l.Command(0xC6, 0x0F)
	l.Command(0xB0, 0x00, 0xC0) // RAMCTRL
	l.Command(0xB7, 0x35)       // GCTRL (320×240)
	l.Command(0xBB, 0x1F)       // VCOMS (320×240)
	l.Command(0xE0, 0xD0, 0x08, 0x11, 0x08, 0x0C, 0x15, 0x39, 0x33, 0x50, 0x36, 0x13, 0x14, 0x29, 0x2D)
	l.Command(0xE1, 0xD0, 0x08, 0x10, 0x08, 0x06, 0x06, 0x39, 0x44, 0x51, 0x0B, 0x16, 0x14, 0x2F, 0x31)
	l.Command(0x21) // INVON
	l.Command(0x11) // SLPOUT
	l.Command(0x29) // DISPON
	time.Sleep(100 * time.Millisecond)
	l.SetMADCTL(MADCTLRotate180)
}

// SetMADCTL sets the memory access control (orientation) register.
func (l LCD) SetMADCTL(v uint8) { l.Command(0x36, v) }

// write clocks one byte out: data and WR low, then WR high latches it.
// ST7789 8080 write timing: WR low ≥ 15 ns, WR high ≥ 15 ns, cycle ≥ 66 ns.
// TinyGo 0.43 runs the RP2040 at 200 MHz (machine.CPUFrequency, measured), so
// one cycle is 5 ns: with two nops per phase a fill loop measured 42 ns per
// byte and the WR-low pulse was ~15 ns — out of spec. Six nops per phase
// measured 83 ns per byte (WR low ≈ 35 ns).
func write(b uint8) {
	rp.SIO.GPIO_OUT_CLR.Set(lcdDataMask | lcdWRMask)
	rp.SIO.GPIO_OUT_SET.Set(uint32(b) << lcdDataShift)
	arm.Asm("nop\nnop\nnop\nnop\nnop\nnop")
	rp.SIO.GPIO_OUT_SET.Set(lcdWRMask)
	arm.Asm("nop\nnop\nnop\nnop\nnop\nnop")
}

// Command sends a command byte followed by its parameters.
func (l LCD) Command(cmd uint8, args ...uint8) {
	machine.LCD_CS.Low()
	machine.LCD_DC.Low()
	write(cmd)
	machine.LCD_DC.High()
	for _, a := range args {
		write(a)
	}
	machine.LCD_CS.High()
}

// window selects full-width rows y0…y1 and starts a RAM write.
func (l LCD) window(y0, y1 int) {
	l.Command(0x2A, 0x00, 0x00, 0x01, 0x3F)                           // CASET 0..319
	l.Command(0x2B, uint8(y0>>8), uint8(y0), uint8(y1>>8), uint8(y1)) // RASET
	machine.LCD_CS.Low()
	machine.LCD_DC.Low()
	write(0x2C) // RAMWR
	machine.LCD_DC.High()
}

// Push sends full-width rows starting at y0 (RGB565 pixels, MSB first).
func (l LCD) Push(y0 int16, pix []uint16) {
	l.window(int(y0), int(y0)+len(pix)/Width-1)
	for _, p := range pix {
		write(uint8(p >> 8))
		write(uint8(p))
	}
	machine.LCD_CS.High()
}

// Fill paints the whole panel one RGB565 colour without a buffer.
func (l LCD) Fill(c uint16) {
	l.window(0, Height-1)
	for range Width * Height {
		write(uint8(c >> 8))
		write(uint8(c))
	}
	machine.LCD_CS.High()
}
