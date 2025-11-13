package widget

import (
	ui "github.com/itohio/tinygui"
)

// BitmapBase stores common bitmap metadata and pixel buffer for specialised bitmap widgets.
type BitmapBase[T any] struct {
	ui.WidgetBase
	pixels []T
}

// NewBitmapBase constructs the shared bitmap state with fixed dimensions and backing pixel slice.
func NewBitmapBase[T any](width, height uint16, pixels []T) *BitmapBase[T] {
	return &BitmapBase[T]{
		WidgetBase: ui.NewWidgetBase(width, height),
		pixels:     pixels,
	}
}

// Pixels returns the underlying pixel buffer without copying.
func (b *BitmapBase[T]) Pixels() []T {
	return b.pixels
}

// SetPixels swaps the backing pixel buffer reference without additional allocation.
func (b *BitmapBase[T]) SetPixels(pixels []T) {
	b.pixels = pixels
}

// Bitmap16 renders 16-bit (RGB565) bitmap data using DrawRGBBitmap.
type Bitmap16 struct {
	*BitmapBase[uint16]
}

// NewBitmap16 constructs a 16-bit bitmap widget.
func NewBitmap16(width, height uint16, pixels []uint16) *Bitmap16 {
	return &Bitmap16{BitmapBase: NewBitmapBase(width, height, pixels)}
}

// Draw renders the RGB565 bitmap when the displayer exposes DrawRGBBitmap.
func (b *Bitmap16) Draw(ctx ui.Context) {
	data := b.Pixels()
	if len(data) == 0 {
		return
	}
	disp, ok := ctx.D().(ui.BitmapDisplayer)
	if !ok {
		return
	}
	w, h := b.Size()
	needed := int(w) * int(h)
	if len(data) < needed {
		return
	}
	x, y := ctx.DisplayPos()
	_ = disp.DrawRGBBitmap(x, y, data[:needed], int16(w), int16(h))
}

// SetPixels replaces the bitmap buffer reference without reallocating.
func (b *Bitmap16) SetPixels(pixels []uint16) {
	b.BitmapBase.SetPixels(pixels)
}

// Bitmap8 renders 8-bit bitmap data (e.g., grayscale or indexed) using DrawRGBBitmap8.
type Bitmap8 struct {
	*BitmapBase[uint8]
}

// NewBitmap8 constructs an 8-bit bitmap widget.
func NewBitmap8(width, height uint16, pixels []uint8) *Bitmap8 {
	return &Bitmap8{BitmapBase: NewBitmapBase(width, height, pixels)}
}

// Draw renders the 8-bit bitmap when the displayer exposes DrawRGBBitmap8.
func (b *Bitmap8) Draw(ctx ui.Context) {
	data := b.Pixels()
	if len(data) == 0 {
		return
	}
	disp, ok := ctx.D().(ui.BitmapDisplayer)
	if !ok {
		return
	}
	w, h := b.Size()
	needed := int(w) * int(h)
	if len(data) < needed {
		return
	}
	x, y := ctx.DisplayPos()
	_ = disp.DrawRGBBitmap8(x, y, data[:needed], int16(w), int16(h))
}

// SetPixels replaces the bitmap buffer reference without reallocating.
func (b *Bitmap8) SetPixels(pixels []uint8) {
	b.BitmapBase.SetPixels(pixels)
}
