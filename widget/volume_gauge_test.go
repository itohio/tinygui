package widget

import (
	"image/color"
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
	"tinygo.org/x/drivers"
)

type mockDisplay struct{}

func (mockDisplay) SetPixel(int16, int16, color.RGBA) {}
func (mockDisplay) Display() error                    { return nil }
func (mockDisplay) Size() (int16, int16)              { return 128, 64 }
func (mockDisplay) DrawBuffer(int16, int16, int16, int16, []uint8) error {
	return nil
}
func (mockDisplay) DrawRGBBitmap(int16, int16, []uint16, int16, int16) error {
	return nil
}
func (mockDisplay) DrawRGBBitmap8(int16, int16, []uint8, int16, int16) error {
	return nil
}
func (mockDisplay) FillRectangle(int16, int16, int16, int16, color.RGBA) error {
	return nil
}
func (mockDisplay) FillRectangleWithBuffer(int16, int16, int16, int16, []color.RGBA) error {
	return nil
}
func (mockDisplay) FillScreen(color.RGBA) {}

var _ drivers.Displayer = mockDisplay{}
var _ ui.RectangleDisplayer = mockDisplay{}
var _ ui.BitmapDisplayer = mockDisplay{}

func TestVolumeGaugeDefaultsAndDraw(t *testing.T) {
	value := uint16(50)
	g := NewVolumeGauge[uint16](60, 10, &value, 0, 100, 0, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255})
	require.NotNil(t, g)
	require.Equal(t, uint8(defaultVolumeBars), g.bars)

	ctx := ui.NewContext(mockDisplay{}, 60, 10, 0, 0)
	require.NotPanics(t, func() { g.Draw(&ctx) })
}

func TestSolidGaugeDraw(t *testing.T) {
	value := float32(0.5)
	g := NewSolidGauge[float32](80, 12, &value, 0, 1, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 0, 255})
	require.NotNil(t, g)

	ctx := ui.NewContext(mockDisplay{}, 80, 12, 0, 0)
	require.NotPanics(t, func() { g.Draw(&ctx) })
}
