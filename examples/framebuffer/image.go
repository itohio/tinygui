//go:build !tinygo

package framebuffer

import "image"

// Image converts the frame to an image (host tools: simulator, screenshots).
func (f *RGB565) Image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(f.W), int(f.H)))
	for y := range int(f.H) {
		for x := range int(f.W) {
			img.SetRGBA(x, y, From565(f.Pix[y*int(f.W)+x]))
		}
	}
	return img
}
