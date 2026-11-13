package widget

import (
	"fmt"
	"math"
	"testing"

	ui "github.com/itohio/tinygui"
)

func TestBitmap16Draw(t *testing.T) {
	pixels := make([]uint16, 4)
	bmp := NewBitmap16(2, 2, pixels)

	ctx := ui.NewContext(mockDisplay{}, 2, 2, 0, 0)
	notPanics(t, func() { bmp.Draw(&ctx) })
}

func TestBitmap8Draw(t *testing.T) {
	pixels := make([]uint8, 9)
	bmp := NewBitmap8(3, 3, pixels)

	ctx := ui.NewContext(mockDisplay{}, 3, 3, 0, 0)
	notPanics(t, func() { bmp.Draw(&ctx) })
}

// notPanics fails the test if f panics (testify's require.NotPanics).
func notPanics(t testing.TB, f func(), msg ...any) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v %v", r, fmt.Sprint(msg...))
		}
	}()
	f()
}

// inEpsilon fails unless got is within relative error eps of want
// (testify's require.InEpsilon).
func inEpsilon[W, G ~float32 | ~float64](t testing.TB, wantV W, gotV G, eps float64, msg ...any) {
	t.Helper()
	want, got := float64(wantV), float64(gotV)
	if want == 0 || math.IsNaN(want) || math.IsNaN(got) || math.Abs(want-got)/math.Abs(want) > eps {
		t.Fatalf("want %v within relative %v, got %v %v", want, eps, got, fmt.Sprint(msg...))
	}
}
