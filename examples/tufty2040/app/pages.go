package app

import (
	"image/color"

	ui "github.com/itohio/tinygui"
	"github.com/itohio/tinygui/container"
	"github.com/itohio/tinygui/examples/choice/icons"
	"github.com/itohio/tinygui/layout"
	"github.com/itohio/tinygui/widget"
)

const (
	pageW    = ScreenW
	innerW   = pageW - 16 // page padding 8 left/right
	captionW = 84
	valueW   = innerW - captionW - 12
)

// newPage builds a vertically stacked page with its own navigator, so every
// page keeps its focus when you switch away and back.
func (a *App) newPage(name string, children ...ui.Widget) *page {
	return a.newPageWith(name, layout.VList(4), children...)
}

func (a *App) newPageWith(name string, l layout.Strategy, children ...ui.Widget) *page {
	return a.newPageOpt(name, container.WithLayout[ui.Widget](l), children...)
}

// newPageOpt makes a page whose children are placed by lay (a cursor
// strategy or an arranger).
func (a *App) newPageOpt(name string, lay container.Option[ui.Widget], children ...ui.Widget) *page {
	root := container.New[ui.Widget](pageW, bodyH,
		container.WithPadding[ui.Widget](8, 6),
		lay,
		container.WithTimeout[ui.Widget](a.idleTimeout),
		container.WithChildren[ui.Widget](children...),
	)
	p := &page{name: name, root: root, nav: ui.NewNavigator(root)}
	p.nav.AddObserver(navLogger{a: a, page: p})
	p.nav.Focus(0)
	return p
}

func (a *App) text(w uint16, fn func() string, c color.RGBA) *widget.Label {
	return widget.NewLabel(w, 11, fontBody, fn, c)
}

func (a *App) caption(w uint16, s string) ui.Widget {
	return still(widget.NewLabel(w, 11, fontBody, func() string { return s }, colMuted))
}

func box(w, h uint16, l layout.Strategy, children ...ui.Widget) *container.Base[ui.Widget] {
	return container.New[ui.Widget](w, h,
		container.WithLayout[ui.Widget](l),
		container.WithChildren[ui.Widget](children...),
	)
}

// Dashboard: live sensors, graphs, frame statistics. Nothing is focusable.
func (a *App) pageDashboard() *page {
	uptime := newHeading(120, 22, func() string {
		s := int(a.now.Sub(a.start).Seconds())
		return a.txtUptime.reset().pad2(s / 3600).str(":").pad2(s / 60 % 60).str(":").pad2(s % 60).String()
	}, &colAccent)
	readouts := box(170, 35, layout.VList(1),
		a.text(170, func() string {
			return a.txtLight.reset().str("Light ").int(int(a.light)).str(" %  raw ").int(int(a.lightRaw)).String()
		}, colFG),
		a.text(170, func() string {
			t := a.txtBatt.reset().str("Battery ").fixed(a.vbat, 2).str(" V")
			if a.usb {
				t.str("  USB")
			}
			return t.String()
		}, colFG),
		a.text(170, func() string {
			return a.txtFrame.reset().str("draw ").fixed(float32(a.drawUS)/1000, 1).
				str(" push ").fixed(float32(a.pushUS)/1000, 1).str(" ms").String()
		}, colMuted),
	)
	a.spark = newSparkline(innerW, 52, 0, 100)
	a.battGauge = widget.NewVolumeGauge[float32](innerW/2-6, 12, &a.battPct, 0, 100, 10, colGood, colTrack)
	a.lightGauge = widget.NewSolidGauge[float32](innerW/2-6, 12, &a.light, 0, 100, colAccent, colTrack)
	return a.newPage("Dash",
		still(box(innerW, 36, layout.HList(10), uptime, readouts)),
		a.caption(innerW, "Light sensor history (sparkline)"),
		still(a.spark),
		still(box(innerW, 11, layout.HList(12),
			still(widget.NewLabel(innerW/2-6, 11, fontBody, func() string { return "Battery" }, colMuted)),
			still(widget.NewLabel(innerW/2-6, 11, fontBody, func() string { return "Light" }, colMuted)))),
		still(box(innerW, 12, layout.HList(12), a.battGauge, a.lightGauge)),
		a.caption(innerW, "Button presses (bar chart)"),
		still(newBars(innerW, 40, &a.presses, [6]string{"U", "D", "A", "B", "C", "*"})),
	)
}

// Gauges: every gauge type, four animation curves, editable values.
func (a *App) pageGauges() *page {
	names := [4]string{"lin", "in", "out", "i/o"}
	var cols [4]ui.Widget
	for i := range cols {
		cols[i] = box(30, 84, layout.VList(2),
			widget.NewGauge[float32](22, 70, &a.curves[i], 0, 100, accents[i], colTrack),
			widget.NewLabel(30, 11, fontBody, func() string { return names[i] }, colMuted),
		)
	}
	a.hGauge = widget.NewGauge[float32](150, 10, &a.curves[3], 0, 100, colAccent, colTrack)
	a.multi = widget.NewMultiGauge[float32](150, 10, 0, 100, &a.mix, []color.RGBA{colBad, colGood, accents[0]}, colTrack, colFG)
	right := box(160, 84, layout.VList(3),
		a.caption(160, "Gauge (ease i/o)"), a.hGauge,
		a.caption(160, "MultiGauge (mix)"), a.multi,
		a.caption(160, "VolumeGauge (level)"),
		widget.NewVolumeGauge[float32](150, 12, &a.level, 0, 100, 12, colGood, colTrack),
	)
	curves := box(136, 84, layout.HList(4), cols[:]...)

	level := widget.NewInteractiveGauge[float32](valueW, 10,
		widget.WithValue(&a.level), widget.WithRange[float32](0, 100), widget.WithSteps[float32](2, 10),
		widget.WithForeground[float32](colGood), widget.WithBackground[float32](colTrack),
		widget.WithCommit(func(v float32) { a.logf("level", int(v)) }))
	mix := widget.NewInteractiveMultiGauge[float32](valueW, 10,
		widget.WithValues(&a.mix), widget.WithRange[float32](0, 100), widget.WithSteps[float32](2, 10),
		widget.WithSegmentColors[float32]([]color.RGBA{colBad, colGood, accents[0]}),
		widget.WithBackground[float32](colTrack),
		widget.WithMultiCommit(func([]float32) { a.log("mix committed") }))
	speed := widget.NewInteractiveLabel[int](valueW, 11,
		widget.WithValue(&a.speedMS), widget.WithRange(200, 5000), widget.WithSteps(100, 1000),
		widget.WithFont[int](fontBody), widget.WithTextColor[int](colFG),
		widget.WithFormatter(func(v int) string { return a.txtSpeed.reset().int(v).str(" ms").String() }),
		widget.WithCommit(func(v int) { a.restartCurves() }))
	return a.newPage("Gauge",
		still(box(innerW, 84, layout.HList(8), curves, right)),
		newField("Level", captionW, innerW, level),
		newField("Mix (A=next)", captionW, innerW, mix),
		newField("Curve time", captionW, innerW, speed),
		a.caption(innerW, "A edit, UP/DOWN change, A commit, B revert"),
	)
}

// Controls: values that drive the badge hardware and the theme.
func (a *App) pageControls() *page {
	led := widget.NewToggle(60, 12, fontBody, colFG, "ON", "OFF", colGood, colTrack,
		func() bool { return a.ledOn }, func(v bool) { a.ledOn = v; a.applyLED(); a.logf("led on", btoi(v)) })
	ledPct := widget.NewInteractiveLabel[int](valueW, 11,
		widget.WithValue(&a.ledPct), widget.WithRange(0, 100), widget.WithSteps(5, 25),
		widget.WithFont[int](fontBody), widget.WithTextColor[int](colFG),
		widget.WithFormatter(func(v int) string { return a.txtLED.reset().int(v).str(" %").String() }),
		widget.WithCommit(func(int) { a.applyLED() }))
	bl := widget.NewInteractiveLabel[int](valueW, 11,
		widget.WithValue(&a.backlight), widget.WithRange(10, 100), widget.WithSteps(5, 25),
		widget.WithFont[int](fontBody), widget.WithTextColor[int](colFG),
		widget.WithFormatter(func(v int) string { return a.txtBL.reset().int(v).str(" %").String() }),
		widget.WithCommit(func(v int) { a.hw.SetBacklight(v); a.logf("backlight", v) }))
	gain := widget.NewInteractiveLabel[float32](valueW, 11,
		widget.WithValue(&a.gain), widget.WithRange[float32](-12, 12), widget.WithSteps[float32](0.5, 3),
		widget.WithFont[float32](fontBody), widget.WithTextColor[float32](colFG),
		widget.WithFormatter(func(v float32) string { return a.txtGain.reset().fixed(v, 1).str(" dB").String() }))
	accent := widget.NewInteractiveLabelChoice(valueW, 11, accentNames,
		widget.WithLabelChoiceIndex(&a.accent), widget.WithLabelChoiceFont(fontBody),
		widget.WithLabelChoiceColor(colFG),
		widget.WithLabelChoiceChange(func(i int, _ string) { a.setAccent(i) }))
	sample := newDurationLabel(valueW, &a.samplePeriod, a.cfg.DurationAdd, a.cfg.DurationSub)
	help := widget.NewMultilineLabel(innerW, 11, 3, widget.WithMultilineFont(fontBody), widget.WithMultilineColor(colMuted))
	help.SetLines([]string{
		"Toggle: A flips it. Numbers: A edit, UP/DOWN,",
		"A commit, B revert. Accent recolours the UI.",
		"Sample period uses ui.DurationAdd steps.",
	})
	return a.newPage("Ctrl",
		newField("User LED", captionW, innerW, led),
		newField("LED level", captionW, innerW, ledPct),
		newField("Backlight", captionW, innerW, bl),
		newField("Sample", captionW, innerW, sample),
		newField("Accent", captionW, innerW, accent),
		newField("Gain", captionW, innerW, gain),
		still(help),
	)
}

// Pictures: PNG icons, RGB565 and RGB332 bitmaps, an animated bitmap.
func (a *App) pagePictures() *page {
	icon := widget.NewInteractiveIcon(pngSize, pngSize, pngIcons, widget.WithIconIndex(&a.pngIdx),
		widget.WithIconChange(func(i int) { a.log("icon " + pngNames[i]) }))
	iconChoice := widget.NewInteractiveIconChoice(pngSize, pngSize, pngIcons, widget.WithIconChoiceIndex(&a.pngChoiceIdx))
	bitmaps := []ui.Widget{
		widget.NewBitmap16(icons.AquariumWidth, icons.AquariumHeight, icons.AquariumPng),
		widget.NewBitmap16(icons.FilterWidth, icons.FilterHeight, icons.FilterPng),
		widget.NewBitmap16(icons.FoodWidth, icons.FoodHeight, icons.FoodPng),
		widget.NewBitmap16(icons.ThermometerWidth, icons.ThermometerHeight, icons.ThermometerPng),
	}
	pick := widget.NewInteractiveWidgetChoice[ui.Widget](40, 40, bitmaps, widget.WithWidgetChoiceIndex[ui.Widget](&a.bitmapIdx))
	renderGradient(a.gradient[:], plasmaW, plasmaH)
	a.plasmaBmp = widget.NewBitmap16(plasmaW, plasmaH, a.plasma[:])
	strip := box(innerW, plasmaH+12, layout.HList(8),
		box(plasmaW, plasmaH+12, layout.VList(1), a.caption(plasmaW, "Bitmap16"), a.plasmaBmp),
		widget.NewSeparator(3, plasmaH+12, colTrack),
		box(plasmaW, plasmaH+12, layout.VList(1), a.caption(plasmaW, "Bitmap8"), widget.NewBitmap8(plasmaW, plasmaH, a.gradient[:])),
		widget.NewSeparator(3, plasmaH+12, colTrack),
		box(100, plasmaH+12, layout.VList(1), a.caption(100, "Separator"), widget.NewSeparator(100, 9, colAccent),
			widget.NewSeparator(100, 9, colMuted), widget.NewSeparator(100, 9, colGood)),
	)
	return a.newPage("Image",
		newField("InteractiveIcon", 100, innerW, icon),
		newField("IconChoice", 100, innerW, iconChoice),
		newField("WidgetChoice", 100, innerW, pick),
		still(strip),
	)
}

// Layouts: Grid, HFlow, VFlow and HList side by side. Nothing is focusable.
func (a *App) pageLayouts() *page {
	half := uint16(innerW/2 - 4)
	var sw []ui.Widget
	for i := range 12 {
		c := accents[i%4]
		c.R, c.G, c.B = c.R/uint8(1+i/4), c.G/uint8(1+i/4), c.B/uint8(1+i/4)
		sw = append(sw, newSwatch(22, 14, c))
	}
	var hf, vf []ui.Widget
	for _, s := range []string{"tiny", "gui", "flows", "left", "to", "right", "and", "wraps"} {
		hf = append(hf, newChip(s))
	}
	for _, s := range []string{"v", "flow", "fills", "column", "first", "then", "next"} {
		vf = append(vf, newChip(s))
	}
	var hl []ui.Widget
	for i := range 6 {
		hl = append(hl, newSwatch(14, 30, accents[i%4]), widget.NewSeparator(2, 30, colMuted))
	}
	panel := func(title string, inner ui.Widget) ui.Widget {
		return box(half, 88, layout.VList(2), a.caption(half, title), inner)
	}
	return a.newPage("Flow",
		still(box(innerW, 88, layout.HList(8),
			panel("Grid(4,4)", box(half, 70, layout.Grid(4, 4), sw...)),
			panel("HFlow(4)", box(half, 70, layout.HFlow(4, 0), hf...)),
		)),
		still(box(innerW, 88, layout.HList(8),
			panel("VFlow(3)", box(half, 70, layout.VFlow(3, 0), vf...)),
			panel("HList(2) + Separator", box(half, 70, layout.HList(2), hl...)),
		)),
	)
}

// List: a ScrollChoice with a scrollbar observer and a detail panel.
func (a *App) pageList() *page {
	const rows, rowH = 20, 16 // field height: 12 px min content + 2×2 frame
	var items []ui.Widget
	for i := range rows {
		name := a.listNames[i].reset().str("Channel ").pad2(i + 1).String()
		freq := 5658 + 37*i
		items = append(items, newField(name, 90, 200, widget.NewLabel(80, 11, fontBody,
			func() string { return a.listVals[i].reset().int(freq).str(" MHz").String() }, colFG)))
	}
	viewH := uint16(bodyH - 12)
	// WithScrollChoiceChange fires only for selector-driven moves, not for
	// navigator focus, so the detail panel reads the container index.
	list := container.NewScrollChoice(200, viewH, layout.VList(2), items)
	bar := newScrollbar(viewH, rows*(rowH+2)-2, int16(viewH))
	list.AddObserver(bar)
	detail := box(innerW-200-16, viewH, layout.VList(4),
		a.caption(80, "Selected"),
		newHeading(80, 22, func() string {
			i := list.Index()
			if i < 0 {
				return "--"
			}
			return a.txtSel.reset().pad2(i + 1).String()
		}, &colAccent),
		a.caption(80, "A: enter list"),
		a.caption(80, "B: leave list"),
	)
	return a.newPageWith("List", layout.HList(4), list, still(bar), still(detail))
}

// Log: scrollable event log fed by buttons, navigator events and serial.
func (a *App) pageLog() *page {
	a.logView = widget.NewInteractiveLog(innerW-8, 11, 13, widget.WithMultilineFont(fontBody), widget.WithMultilineColor(colFG))
	a.logView.SetCapacity(logCapacity)
	info := widget.NewMultilineLabel(innerW, 11, 2, widget.WithMultilineFont(fontBody), widget.WithMultilineColor(colMuted))
	a.info = info
	return a.newPage("Log",
		newField("", 0, innerW, a.logView),
		still(info),
	)
}
