package widget

import (
	"fmt"
	"image/color"
	"strings"
	"testing"
	"unicode/utf8"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/internal/check"
	"tinygo.org/x/tinyfont"
)

func TestInteractiveLabelAdjustCommitAndCancel(t *testing.T) {
	value := float32(1.5)
	var committed float32
	label := NewInteractiveLabel[float32](64, 12,
		WithValue(&value),
		WithRange[float32](0, 5),
		WithSteps[float32](0.1, 0.5),
		WithFont[float32](&tinyfont.TomThumb),
		WithTextColor[float32](color.RGBA{255, 255, 255, 255}),
		WithCommit[float32](func(v float32) { committed = v }),
	)

	label.SetSelected(true)
	check.True(t, label.Selected())
	inEpsilon(t, 1.5, label.pending, 0.001)

	check.True(t, label.Interact(ui.UP))
	inEpsilon(t, 1.6, label.pending, 0.001)

	check.True(t, label.Interact(ui.LONG_UP))
	inEpsilon(t, 2.1, label.pending, 0.001)

	// Commit and ensure setter invoked.
	check.True(t, label.Interact(ui.ENTER))
	inEpsilon(t, 2.1, committed, 0.001)
	inEpsilon(t, 2.1, value, 0.001)
	check.False(t, label.Selected())

	// Re-select and cancel.
	label.SetSelected(true)
	check.True(t, label.Interact(ui.DOWN))
	inEpsilon(t, 2.0, label.pending, 0.001)
	check.True(t, label.Interact(ui.ESC))
	inEpsilon(t, 2.1, value, 0.001) // unchanged
}

func TestInteractiveLabelDisable(t *testing.T) {
	value := float32(0)
	label := NewInteractiveLabel[float32](32, 10,
		WithValue(&value),
		WithRange[float32](0, 10),
		WithSteps[float32](1, 2),
		WithFont[float32](&tinyfont.TomThumb),
		WithTextColor[float32](color.RGBA{255, 255, 255, 255}),
		WithDisabled[float32](),
	)

	check.False(t, label.Interact(ui.UP))
	check.Equal(t, float32(0), label.pending)
}

// The label shows its bound value when that value changes elsewhere (a
// serial command, another widget); while it is being edited it shows the
// pending edit instead.
func TestInteractiveLabelFollowsItsBoundValue(t *testing.T) {
	value := 80
	label := NewInteractiveLabel[int](64, 12, WithValue(&value), WithRange[int](10, 100), WithSteps[int](5, 25),
		WithFont[int](&tinyfont.TomThumb))
	value = 60
	check.Equal(t, "60", label.displayText())
	value = 200
	check.Equal(t, "100", label.displayText(), "clamped to the range")

	value = 60
	label.SetSelected(true) // editing
	label.Interact(ui.UP)
	value = 30 // changed elsewhere mid-edit: the edit wins on screen
	if s := label.displayText(); !strings.Contains(s, "65") {
		t.Fatalf("%q does not contain %q", s, "65")
	}
}

// The label formats its number like fmt's %v, marks it while selected, and
// does so without allocating (it is drawn every frame; fmt.Sprintf and the
// marker concatenation allocated per draw).
func TestInteractiveLabelTextWithoutAllocating(t *testing.T) {
	f := float32(-2.5)
	n := 1234
	u := uint8(200)
	fl := NewInteractiveLabel[float32](80, 11, WithValue(&f), WithRange[float32](-12, 12))
	il := NewInteractiveLabel[int](80, 11, WithValue(&n), WithRange(0, 5000))
	ul := NewInteractiveLabel[uint8](80, 11, WithValue(&u), WithRange[uint8](0, 255))
	for _, c := range []struct {
		got  func() string
		want string
	}{
		{fl.displayText, fmt.Sprintf("%v", f)},
		{il.displayText, fmt.Sprintf("%v", n)},
		{ul.displayText, fmt.Sprintf("%v", u)},
	} {
		check.Equal(t, c.want, c.got())
	}
	f = 0.1
	check.Equal(t, fmt.Sprintf("%v", f), fl.displayText(), "float32 shortest form, as fmt prints it")
	fl.SetSelected(true)
	check.Equal(t, "▲ "+fmt.Sprintf("%v", f)+" ▼", fl.displayText())
	check.Zero(t, testing.AllocsPerRun(100, func() { _ = fl.displayText(); _ = il.displayText() }))
}

// The text path's corners (verify TG-1-postcommit R11): negative numbers;
// a custom formatter longer than the buffer is cut (at a rune boundary,
// keeping the closing mark) and never allocates, selected or not; a number
// type's String method isn't used (names need WithFormatter); Text is a
// view, valid until the next Text or Draw.
func TestInteractiveLabelTextCorners(t *testing.T) {
	neg := -1234
	nl := NewInteractiveLabel[int](80, 11, WithValue(&neg), WithRange(-5000, 5000))
	check.Equal(t, "-1234", nl.displayText())

	long := "abcdefghij" + "ÄÖÜabcdefg" + "abcdefghij" + "abcdefghij" + "abcdefghij" // 53 bytes
	v := 1
	fl := NewInteractiveLabel[int](80, 11, WithValue(&v), WithRange(0, 9), WithFormatter(func(int) string { return long }))
	got := fl.displayText()
	check.LessOrEqual(t, len(got), 40)
	check.True(t, utf8.ValidString(got))
	check.True(t, strings.HasPrefix(long, got))
	fl.SetSelected(true)
	got = fl.displayText()
	check.True(t, strings.HasPrefix(got, "▲ ") && strings.HasSuffix(got, " ▼") && utf8.ValidString(got), got)
	check.Zero(t, testing.AllocsPerRun(50, func() { _ = fl.displayText(); _ = nl.displayText() }))
	fl.SetSelected(false)
	check.Zero(t, testing.AllocsPerRun(50, func() { _ = fl.displayText() }))

	m := mode(1)
	ml := NewInteractiveLabel[mode](80, 11, WithValue(&m), WithRange[mode](0, 2))
	check.Equal(t, "1", ml.displayText(), "the number, not mode.String")
	named := NewInteractiveLabel[mode](80, 11, WithValue(&m), WithRange[mode](0, 2), WithFormatter(func(v mode) string { return v.String() }))
	check.Equal(t, "on", named.displayText(), "names via WithFormatter")

	x := 17
	xl := NewInteractiveLabel[int](80, 11, WithValue(&x), WithRange(0, 99))
	s := xl.Text()
	x = 42
	_ = xl.Text()
	check.Equal(t, "42", s, "Text is a view of the label's buffer: the next Text overwrites it (documented)")
}

// mode is a number type with a String method.
type mode int

func (m mode) String() string { return [...]string{"off", "on", "auto"}[m] }
