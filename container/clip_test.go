package container_test

import (
	"testing"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/framebuffer"
	"github.com/itohio/tinygui/layout"
)

// counted records which bands drew it and its visibility events.
type counted struct {
	ui.WidgetBase
	bands  []int16 // the band's first row at each Draw
	events []bool
}

func (c *counted) Draw(ctx ui.Context) {
	c.bands = append(c.bands, ctx.D().(*framebuffer.RGB565).Y0)
}
func (c *counted) Interact(ui.UserCommand) bool { return false }
func (c *counted) OnVisible(v bool)             { c.events = append(c.events, v) }

// On a display drawn band by band (ui.Clipper), a container draws each
// child only in the bands it reaches, so a frame costs each child one draw,
// not one per band (TG-5 5d: on the badge's three 80-row bands up to 25 %
// of a page's draw went to children outside the band). Visibility is the
// container's box as before: culling by band sends no visibility events.
func TestChildrenDrawOnlyInTheBandsTheyReach(t *testing.T) {
	const screenW, screenH, bandH = 100, 240, 80
	for _, tc := range []struct {
		name string
		opt  func(ws ...ui.Widget) *container.Base[ui.Widget]
	}{
		{"layouter (VList)", func(ws ...ui.Widget) *container.Base[ui.Widget] {
			return container.New[ui.Widget](screenW, screenH, container.WithLayout[ui.Widget](layout.VList(40)), container.WithChildren[ui.Widget](ws...))
		}},
		{"arranger (Column)", func(ws ...ui.Widget) *container.Base[ui.Widget] {
			return container.New[ui.Widget](screenW, screenH, container.WithArranger[ui.Widget](layout.Column(40, layout.Start, layout.Start)), container.WithChildren[ui.Widget](ws...))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// the container at display row 10, children 40 rows tall and 40
			// apart: rows 10–50, 90–130, 170–210, each in one 80-row band
			// clear of its edges by more than the slack
			a := &counted{WidgetBase: ui.NewWidgetBase(screenW, 40)}
			b := &counted{WidgetBase: ui.NewWidgetBase(screenW, 40)}
			c := &counted{WidgetBase: ui.NewWidgetBase(screenW, 40)}
			cont := tc.opt(a, b, c)
			pix := make([]uint16, screenW*bandH)
			fb := framebuffer.NewBand(screenW, screenH, bandH, pix, nil)
			ctx := ui.NewContext(fb, screenW, screenH, 0, 10)
			for frame := range 2 {
				for y0 := int16(0); y0 < screenH; y0 += bandH {
					fb.SetY0(y0)
					c := ctx
					cont.Draw(&c)
				}
				_ = frame
			}
			for _, k := range []struct {
				name string
				w    *counted
				band int16
			}{{"a", a, 0}, {"b", b, 80}, {"c", c, 160}} {
				if len(k.w.bands) != 2 || k.w.bands[0] != k.band || k.w.bands[1] != k.band {
					t.Errorf("%s drawn in bands %v over two frames, want only band %d once a frame", k.name, k.w.bands, k.band)
				}
				if len(k.w.events) != 1 || !k.w.events[0] {
					t.Errorf("%s visibility events %v, want one true (the box, not the band)", k.name, k.w.events)
				}
			}
		})
	}
}

// Without a Clipper (a full-screen display) every visible child draws, as
// before.
func TestChildrenDrawWithoutAClipper(t *testing.T) {
	a := &counted{WidgetBase: ui.NewWidgetBase(100, 50)}
	b := &counted{WidgetBase: ui.NewWidgetBase(100, 50)}
	cont := container.New[ui.Widget](100, 240, container.WithLayout[ui.Widget](layout.VList(30)), container.WithChildren[ui.Widget](a, b))
	pix := make([]uint16, 100*240)
	fb := framebuffer.New(100, 240, pix, nil) // one band: the whole screen
	ctx := ui.NewContext(fb, 100, 240, 0, 0)
	cont.Draw(&ctx)
	if len(a.bands) != 1 || len(b.bands) != 1 {
		t.Fatalf("full screen: a drawn %d times, b %d, want once each", len(a.bands), len(b.bands))
	}
}
