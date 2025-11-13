//go:build m5stickc

package main

import (
	"fmt"
	"image/color"
	"time"

	"machine"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/choice/icons"
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
	ultraPressThreshold = 1500 * time.Millisecond
)

type buttonEvent uint8

const (
	eventNone buttonEvent = iota
	eventShort
	eventLong
	eventUltra
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
		switch {
		case duration >= ultraPressThreshold:
			return eventUltra
		case duration >= longPressThreshold:
			return eventLong
		case duration >= shortPressThreshold:
			return eventShort
		default:
			return eventShort
		}
	}
	return eventNone
}

func configureDisplay() ui.Displayer {
	m5stick.InitPower()
	display := m5stick.NewDisplay()
	display.FillScreen(color.RGBA{0, 0, 0, 255})
	return display
}

type navigatorStatus struct {
	choice *container.ScrollChoice
	titles []string
	status *string
}

func (s navigatorStatus) OnNavigatorEvent(ev ui.NavigatorEvent) {
	if s.choice == nil {
		return
	}
	idx := s.choice.Index()
	if idx < 0 || idx >= len(s.titles) {
		return
	}
	switch ev.Type {
	case ui.NavigatorEventFocusChanged:
		*s.status = fmt.Sprintf("Focus: %s", s.titles[idx])
	case ui.NavigatorEventActivated:
		*s.status = fmt.Sprintf("Editing: %s", s.titles[idx])
	case ui.NavigatorEventDeactivated:
		*s.status = fmt.Sprintf("Saved: %s", s.titles[idx])
	}
}

func main() {
	display := configureDisplay()

	selectButton := newButton(m5stick.BtnSide, true)
	adjustButton := newButton(m5stick.BtnMain, true)

	modes := []string{"Manual", "Schedule", "Vacation"}
	modeIndex := 0

	iconNames := []string{"Aquarium", "Filter", "Feeder", "Thermometer"}
	iconIndex := 0

	updateSummary := func() string {
		return fmt.Sprintf("Mode: %s | Icon: %s", modes[modeIndex], iconNames[iconIndex])
	}

	summaryText := updateSummary()

	labelChoice := widget.NewInteractiveLabelChoice(120, 18, modes,
		widget.WithLabelChoiceIndex(&modeIndex),
		widget.WithLabelChoiceChange(func(i int, _ string) {
			modeIndex = i
			summaryText = updateSummary()
		}),
	)

	iconWidgets := []ui.Widget{
		widget.NewBitmap16(icons.AquariumWidth, icons.AquariumHeight, icons.AquariumPng),
		widget.NewBitmap16(icons.AquariumWidth, icons.AquariumHeight, icons.FilterPng),
		widget.NewBitmap16(icons.AquariumWidth, icons.AquariumHeight, icons.FoodPng),
		widget.NewBitmap16(icons.AquariumWidth, icons.AquariumHeight, icons.ThermometerPng),
	}

	iconChoice := widget.NewInteractiveWidgetChoice[ui.Widget](icons.AquariumWidth, icons.AquariumHeight, iconWidgets,
		widget.WithWidgetChoiceIndex[ui.Widget](&iconIndex),
		widget.WithWidgetChoiceChange(func(i int, _ ui.Widget) {
			iconIndex = i
			summaryText = updateSummary()
		}),
	)

	choiceTitles := []string{"Feeder mode", "Status icon"}
	statusText := "Press Side to focus, Main to adjust"

	choices := container.NewScrollChoice(displayWidth-20, icons.AquariumHeight+60, layout.VList(12), []ui.Widget{labelChoice, iconChoice},
		container.WithScrollChoiceChange(func(i int, _ ui.Widget) {
			if i >= 0 && i < len(choiceTitles) {
				statusText = fmt.Sprintf("Focus: %s", choiceTitles[i])
			}
		}),
	)
	choices.SetIndex(0)

	statusLabel := widget.NewLabel(displayWidth-20, 14, &tinyfont.TomThumb, func() string { return statusText }, color.RGBA{200, 200, 0, 255})
	summaryLabel := widget.NewLabel(displayWidth-20, 14, &tinyfont.TomThumb, func() string { return summaryText }, color.RGBA{220, 220, 220, 255})

	updateFocusStatus := func() {
		if idx := choices.Index(); idx >= 0 && idx < len(choiceTitles) {
			statusText = fmt.Sprintf("Focus: %s", choiceTitles[idx])
		}
	}

	root := container.New[ui.Widget](displayWidth, displayHeight,
		container.WithPadding[ui.Widget](10, 10),
		container.WithLayout[ui.Widget](layout.VList(12)),
		container.WithChildren[ui.Widget](choices, statusLabel, summaryLabel),
	)

	navigator := ui.NewNavigator(root)
	navigator.AddObserver(navigatorStatus{choice: choices, titles: choiceTitles, status: &statusText})
	navigator.Focus(0)
	updateFocusStatus()

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
				if current := navigator.Current(); current != nil && !choices.Active() {
					current.Interact(ui.NEXT)
				}
			case eventLong:
				if choices.Active() {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.ENTER)
					}
					navigator.Back()
					updateFocusStatus()
				} else {
					navigator.Enter()
				}
			case eventUltra:
				if choices.Active() {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.ESC)
					}
					navigator.Back()
					updateFocusStatus()
				}
			}
		}

		if ev := adjustButton.Update(now); ev != eventNone {
			switch ev {
			case eventShort:
				if choices.Active() {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.UP)
					}
				}
			case eventLong:
				if choices.Active() {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.DOWN)
					}
				}
			case eventUltra:
				if choices.Active() {
					if current := navigator.Current(); current != nil {
						current.Interact(ui.BACK)
					}
					navigator.Back()
					updateFocusStatus()
				}
			}
		}

		draw()
		time.Sleep(33 * time.Millisecond)
	}
}
