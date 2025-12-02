//go:build m5stickc

package main

import (
	"fmt"
	"image/color"
	"time"

	"machine"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	m5stick "github.com/itohio/tinygui/examples/m5stick"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
	"tinygo.org/x/tinyfont"
)

const (
	displayWidth  = 240
	displayHeight = 135

	shortPressThreshold = 150 * time.Millisecond
	longPressThreshold  = 600 * time.Millisecond
)

type buttonEvent uint8

const (
	eventNone buttonEvent = iota
	eventShort
	eventLong
)

type buttonTracker struct {
	pin       machine.Pin
	activeLow bool
	pressed   bool
	pressedAt time.Time
}

func newButton(pin machine.Pin, activeLow bool) *buttonTracker {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	return &buttonTracker{pin: pin, activeLow: activeLow}
}

func (b *buttonTracker) Update(now time.Time) buttonEvent {
	level := b.pin.Get()
	pressed := level
	if b.activeLow {
		pressed = !level
	}
	if pressed && !b.pressed {
		b.pressed = true
		b.pressedAt = now
		return eventNone
	}
	if !pressed && b.pressed {
		duration := now.Sub(b.pressedAt)
		b.pressed = false
		if duration >= longPressThreshold {
			return eventLong
		}
		if duration >= shortPressThreshold {
			return eventShort
		}
		return eventShort
	}
	return eventNone
}

func configureDisplay() ui.Displayer {
	m5stick.InitPower()
	display := m5stick.NewDisplay()
	display.FillScreen(color.RGBA{0, 0, 0, 255})
	return display
}

type zoneModel struct {
	name      string
	moisture  float32
	low       float32
	high      float32
	pump      float32
	soak      float32
	readout   *widget.Gauge[float32]
	lowGauge  *widget.InteractiveGauge[float32]
	highGauge *widget.InteractiveGauge[float32]
	pumpGauge *widget.InteractiveGauge[float32]
	soakGauge *widget.InteractiveGauge[float32]
	control   *container.Base[ui.Widget]
}

func buildZone(name string, moisture, low, high, pump, soak float32) *zoneModel {
	zone := &zoneModel{name: name, moisture: moisture, low: low, high: high, pump: pump, soak: soak}

	zone.readout = widget.NewGauge[float32](displayWidth-20, 12, &zone.moisture, 0, 100,
		color.RGBA{0, 180, 255, 255}, color.RGBA{25, 25, 25, 255},
	)

	zone.lowGauge = widget.NewInteractiveGauge[float32](displayWidth-40, 18,
		widget.WithValue(&zone.low),
		widget.WithRange[float32](0, 100),
		widget.WithSteps[float32](1, 5),
		widget.WithForeground[float32](color.RGBA{80, 200, 120, 255}),
		widget.WithBackground[float32](color.RGBA{30, 30, 30, 255}),
	)

	zone.highGauge = widget.NewInteractiveGauge[float32](displayWidth-40, 18,
		widget.WithValue(&zone.high),
		widget.WithRange[float32](0, 100),
		widget.WithSteps[float32](1, 5),
		widget.WithForeground[float32](color.RGBA{200, 120, 60, 255}),
		widget.WithBackground[float32](color.RGBA{30, 30, 30, 255}),
	)

	zone.pumpGauge = widget.NewInteractiveGauge[float32](displayWidth-40, 18,
		widget.WithValue(&zone.pump),
		widget.WithRange[float32](0, 15),
		widget.WithSteps[float32](0.5, 2),
		widget.WithForeground[float32](color.RGBA{220, 180, 30, 255}),
		widget.WithBackground[float32](color.RGBA{30, 30, 30, 255}),
	)

	zone.soakGauge = widget.NewInteractiveGauge[float32](displayWidth-40, 18,
		widget.WithValue(&zone.soak),
		widget.WithRange[float32](0, 10),
		widget.WithSteps[float32](0.5, 1.5),
		widget.WithForeground[float32](color.RGBA{200, 80, 200, 255}),
		widget.WithBackground[float32](color.RGBA{30, 30, 30, 255}),
	)

	zone.control = container.New[ui.Widget](displayWidth-20, 100,
		container.WithLayout[ui.Widget](layout.VList(6)),
		container.WithChildren[ui.Widget](zone.lowGauge, zone.highGauge, zone.pumpGauge, zone.soakGauge),
	)
	zone.control.SetIndex(0)

	return zone
}

func main() {
	display := configureDisplay()

	selectButton := newButton(m5stick.BtnSide, true)
	adjustButton := newButton(m5stick.BtnMain, true)

	zones := []*zoneModel{
		buildZone("Herbs", 42, 30, 60, 5, 2),
		buildZone("Tomatoes", 58, 35, 65, 6, 3),
		buildZone("Peppers", 36, 28, 55, 4, 1.5),
	}

	readoutChildren := make([]ui.Widget, 0, len(zones)*2)
	for _, z := range zones {
		nameLabel := widget.NewLabel(displayWidth-20, 12, &tinyfont.TomThumb, func(zone *zoneModel) func() string {
			return func() string { return fmt.Sprintf("%s – %.0f%%", zone.name, zone.moisture) }
		}(z), color.RGBA{180, 180, 180, 255})
		readoutChildren = append(readoutChildren, nameLabel, z.readout)
	}

	readouts := container.New[ui.Widget](displayWidth-20, 70,
		container.WithLayout[ui.Widget](layout.VList(4)),
		container.WithChildren(readoutChildren...),
	)

	zoneWidgets := make([]ui.Widget, len(zones))
	for i, z := range zones {
		zoneWidgets[i] = z.control
	}

	interactives := container.NewScrollChoice(displayWidth-20, 110, layout.VList(8), zoneWidgets)
	interactives.SetIndex(0)

	statusText := "Press Side to select a zone"
	summaryLabel := widget.NewLabel(displayWidth-20, 14, &tinyfont.TomThumb, func() string {
		idx := interactives.Index()
		if idx < 0 {
			idx = 0
		}
		if idx >= len(zones) {
			idx = len(zones) - 1
		}
		z := zones[idx]
		return fmt.Sprintf("%s | Moist %.0f%% | Low %.0f%% | High %.0f%% | Pump %.1fs | Soak %.1fs",
			z.name, z.moisture, z.low, z.high, z.pump, z.soak)
	}, color.RGBA{220, 220, 220, 255})

	statusLabel := widget.NewLabel(displayWidth-20, 14, &tinyfont.TomThumb, func() string { return statusText }, color.RGBA{200, 200, 0, 255})

	root := container.New[ui.Widget](displayWidth, displayHeight,
		container.WithPadding[ui.Widget](10, 10),
		container.WithLayout[ui.Widget](layout.VList(10)),
		container.WithChildren[ui.Widget](readouts, interactives, statusLabel, summaryLabel),
	)

	navigator := ui.NewNavigator(root)
	navigator.Focus(1)

	updateStatus := func() {
		idx := interactives.Index()
		if idx < 0 {
			idx = 0
		}
		if idx >= len(zones) {
			idx = len(zones) - 1
		}
		z := zones[idx]
		if navigator.Depth() > 1 {
			child := z.control.Index()
			switch child {
			case 0:
				statusText = fmt.Sprintf("%s – Low threshold", z.name)
			case 1:
				statusText = fmt.Sprintf("%s – High threshold", z.name)
			case 2:
				statusText = fmt.Sprintf("%s – Pump duration", z.name)
			case 3:
				statusText = fmt.Sprintf("%s – Soak time", z.name)
			default:
				statusText = z.name
			}
		} else {
			statusText = fmt.Sprintf("%s – press and hold to edit", z.name)
		}
	}
	updateStatus()

	ctx := ui.NewContext(display, displayWidth, displayHeight, 0, 0)
	draw := func() {
		display.FillScreen(color.RGBA{0, 0, 0, 255})
		root.Draw(&ctx)
		_ = display.Display()
	}

	for {
		now := time.Now()

		if ev := selectButton.Update(now); ev != eventNone {
			switch ev {
			case eventShort:
				if navigator.Depth() > 1 {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.ENTER)
					}
					navigator.Next()
				} else {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.NEXT)
					}
				}
				updateStatus()
			case eventLong:
				if navigator.Depth() > 1 {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.ENTER)
					}
					navigator.Back()
					if navigator.Depth() > 1 {
						navigator.Back()
					}
				} else {
					navigator.Enter()
				}
				updateStatus()
			}
		}

		if ev := adjustButton.Update(now); ev != eventNone {
			if navigator.Depth() > 1 {
				switch ev {
				case eventShort:
					if current := navigator.Current(); current != nil {
						current.Interact(ui.UP)
					}
				case eventLong:
					if current := navigator.Current(); current != nil {
						current.Interact(ui.DOWN)
					}
				}
			}
		}

		draw()
		time.Sleep(33 * time.Millisecond)
	}
}
