package widget

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/stretchr/testify/require"
)

func TestBitmap16Draw(t *testing.T) {
	pixels := make([]uint16, 4)
	bmp := NewBitmap16(2, 2, pixels)

	ctx := ui.NewContext(mockDisplay{}, 2, 2, 0, 0)
	require.NotPanics(t, func() { bmp.Draw(&ctx) })
}

func TestBitmap8Draw(t *testing.T) {
	pixels := make([]uint8, 9)
	bmp := NewBitmap8(3, 3, pixels)

	ctx := ui.NewContext(mockDisplay{}, 3, 3, 0, 0)
	require.NotPanics(t, func() { bmp.Draw(&ctx) })
}
