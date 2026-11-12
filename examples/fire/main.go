// Command fire is a small TinyGUI example: the classic "Doom fire" drawn by
// widget.CellGrid as an LED matrix of round dots, rendered into an RGB565
// framebuffer and saved as an animated GIF. No board needed:
//
//	go run ./examples/fire -out fire.gif
//
// The same grid widget drives the badge demos in itohio/tinytufty2040.
package main

import (
	"flag"
	"image"
	"image/color"
	"image/gif"
	"log"
	"math/rand/v2"
	"os"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/widget"
)

const (
	w, h  = 64, 48 // fire cells
	pitch = 5      // pixels per cell: 320×240
	hot   = 36     // hottest palette index
)

// palette is the classic Doom fire ramp; index 0 is an unlit LED.
var palette = func() []color.RGBA {
	hex := [hot + 1]uint32{
		0x120e0e, 0x1f0707, 0x2f0f07, 0x470f07, 0x571707, 0x671f07, 0x771f07, 0x8f2707, 0x9f2f07,
		0xaf3f07, 0xbf4707, 0xc74707, 0xdf4f07, 0xdf5707, 0xdf5707, 0xd75f07, 0xd75f07, 0xd7670f,
		0xcf6f0f, 0xcf770f, 0xcf7f0f, 0xcf8717, 0xc78717, 0xc78f17, 0xc7971f, 0xbf9f1f, 0xbf9f1f,
		0xbfa727, 0xbfa727, 0xbfaf2f, 0xb7af2f, 0xb7b72f, 0xb7b737, 0xcfcf6f, 0xdfdf9f, 0xefefc7,
		0xffffff,
	}
	p := make([]color.RGBA, len(hex))
	for i, c := range hex {
		p[i] = color.RGBA{uint8(c >> 16), uint8(c >> 8), uint8(c), 255}
	}
	return p
}()

// step spreads heat up from the burning bottom row: each cell takes the
// heat of the cell below, drifted sideways and cooled by 0-2 at random
// (on average 1 per row: flames about 36 of the 48 rows tall).
func step(cells []uint8) {
	for x := range w {
		cells[(h-1)*w+x] = hot
	}
	for y := 1; y < h; y++ {
		for x := range w {
			v := cells[y*w+x]
			nx := min(max(x+rand.IntN(3)-1, 0), w-1)
			if cool := uint8(rand.IntN(3)); v > cool {
				v -= cool
			} else {
				v = 0
			}
			cells[(y-1)*w+nx] = v
		}
	}
}

func main() {
	out := flag.String("out", "fire.gif", "animated GIF to write")
	frames := flag.Int("frames", 60, "frames to render")
	flag.Parse()

	cells := make([]uint8, w*h)
	grid := widget.NewCellGrid(w, h, pitch, 1, widget.CellRound, cells, palette, color.RGBA{8, 6, 6, 255})
	fb := framebuffer.New(w*pitch, h*pitch, make([]uint16, w*pitch*h*pitch), nil)
	ctx := ui.NewContext(fb, w*pitch, h*pitch, 0, 0)

	gifPal := make(color.Palette, 0, len(palette)+1)
	for _, c := range palette {
		gifPal = append(gifPal, c)
	}
	gifPal = append(gifPal, color.RGBA{8, 6, 6, 255})

	anim := &gif.GIF{}
	for f := 0; f < *frames+40; f++ { // 40 frames to let the fire rise first
		step(cells)
		if f < 40 {
			continue
		}
		grid.Draw(&ctx)
		img := image.NewPaletted(image.Rect(0, 0, w*pitch, h*pitch), gifPal)
		src := fb.Image()
		for y := range h * pitch {
			for x := range w * pitch {
				img.Set(x, y, src.At(x, y))
			}
		}
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, 4) // 25 fps
	}
	file, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if err := gif.EncodeAll(file, anim); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d frames", *out, len(anim.Image))
}
