package ui

import (
	"image/color"
	"math/rand"
	"time"

	"tinygo.org/x/drivers"
)

var (
	_ Context = (*ContextImpl)(nil)
	_ Context = (*RandomContext)(nil)
)

type BitmapDisplayer interface {
	DrawRGBBitmap(x int16, y int16, data []uint16, w int16, h int16) error
	DrawRGBBitmap8(x int16, y int16, data []uint8, w int16, h int16) error
}

type LineDisplayer interface {
	DrawFastHLine(x0 int16, x1 int16, y int16, c color.RGBA)
	DrawFastVLine(x int16, y0 int16, y1 int16, c color.RGBA)
}

type RectangleDisplayer interface {
	FillRectangle(x int16, y int16, width int16, height int16, c color.RGBA) error
	FillRectangleWithBuffer(x int16, y int16, width int16, height int16, buffer []color.RGBA) error
	FillScreen(c color.RGBA)
}

type Context interface {
	D() drivers.Displayer
	Size() (W, H uint16)
	Start() (X, Y int16)
	Pos() (X, Y int16)
	DisplayPos() (X, Y int16)
	AddPos(dx, dy int16) (X, Y int16)
	SetPos(x, y int16) bool
	Clone(widget Widget, W, H uint16) Context
	Widget() Widget
}

// ContextImpl is a simple Widget drawing context implementation
type ContextImpl struct {
	// D is used to display pixers
	d      drivers.Displayer
	widget Widget
	// Width of the context
	w uint16
	// Height of the context
	h uint16
	// Top Left coordinate of the context
	x, y int16
	// Coordinates to be used for the widget
	posX, posY int16
}

func NewContext(d drivers.Displayer, w, h uint16, x, y int16) ContextImpl {
	return ContextImpl{
		d: d,
		w: w,
		h: h,
		x: x,
		y: y,
	}
}

func (c *ContextImpl) D() drivers.Displayer     { return c.d }
func (c *ContextImpl) Widget() Widget           { return c.widget }
func (c *ContextImpl) Size() (W, H uint16)      { return c.w, c.h }
func (c *ContextImpl) Start() (X, Y int16)      { return c.x, c.y }
func (c *ContextImpl) Pos() (X, Y int16)        { return c.posX, c.posY }
func (c *ContextImpl) DisplayPos() (X, Y int16) { return c.posX + c.x, c.posY + c.y }
func (c *ContextImpl) AddPos(dx, dy int16) (X, Y int16) {
	c.posX += dx
	c.posY += dy
	return c.Pos()
}
func (c *ContextImpl) SetPos(x, y int16) bool {
	if x > int16(c.w) || y > int16(c.h) || x < 0 || y < 0 {
		return false
	}
	c.posX = x
	c.posY = y
	return true
}

func (c *ContextImpl) Clone(widget Widget, W, H uint16) Context {
	x, y := c.DisplayPos()
	ret := NewContext(c.d, W, H, x, y)
	ret.widget = widget
	return &ret
}

// RandomContext implements randomly shifting context.
// It is especially useful for OLED displays to prevent burn-in.
type RandomContext struct {
	ContextImpl
	dW, dH   int16
	lastTime time.Time
	interval time.Duration
}

func NewRandomContext(d drivers.Displayer, interval time.Duration, w, h uint16) RandomContext {
	dW, dH := d.Size()
	return RandomContext{
		ContextImpl: NewContext(d, w, h, 0, 0),
		lastTime:    time.Now(),
		dW:          dW,
		dH:          dH,
		interval:    interval,
	}
}

func (c *RandomContext) Clone(widget Widget, w, h uint16) Context {
	if time.Since(c.lastTime) > c.interval {
		dx := int32(c.dW - int16(c.w))
		dy := int32(c.dH - int16(c.h))
		if dx <= 0 {
			dx = 1
		}
		if dy <= 0 {
			dy = 1
		}

		c.x = int16(rand.Int31n(dx))
		c.y = int16(rand.Int31n(dy))
		c.lastTime = time.Now()
	}

	return c.ContextImpl.Clone(widget, w, h)
}
