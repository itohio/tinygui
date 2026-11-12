//go:build tufty2040

package main

import (
	"machine"
	"time"

	"github.com/itohio/tinygui/examples/tufty2040/app"
)

// badge implements app.Hardware on the Tufty 2040.
type badge struct {
	light, vbat, vref machine.ADC
	led, backlight    pwmOut
}

type pwmOut struct {
	pwm interface {
		Configure(machine.PWMConfig) error
		Channel(machine.Pin) (uint8, error)
		Set(channel uint8, value uint32)
		Top() uint32
	}
	ch uint8
}

func newPWM(p interface {
	Configure(machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
	Set(channel uint8, value uint32)
	Top() uint32
}, pin machine.Pin) pwmOut {
	_ = p.Configure(machine.PWMConfig{Period: 50_000}) // 20 kHz
	ch, _ := p.Channel(pin)
	return pwmOut{pwm: p, ch: ch}
}

// set applies a perceptual (squared) duty cycle.
func (o pwmOut) set(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	top := o.pwm.Top()
	o.pwm.Set(o.ch, uint32(uint64(top)*uint64(percent*percent)/10000))
}

func newBadge() *badge {
	machine.InitADC()
	machine.SENSOR_POWER.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SENSOR_POWER.High() // powers the light sensor and the 1.24 V reference
	machine.VBUS_DETECT.Configure(machine.PinConfig{Mode: machine.PinInput})
	b := &badge{
		light:     machine.ADC{Pin: machine.LIGHT_SENSE},
		vbat:      machine.ADC{Pin: machine.BATTERY},
		vref:      machine.ADC{Pin: machine.ADC2},
		backlight: newPWM(machine.PWM1, machine.LCD_BACKLIGHT),
		led:       newPWM(machine.PWM4, machine.LED),
	}
	for _, a := range []machine.ADC{b.light, b.vbat, b.vref} {
		_ = a.Configure(machine.ADCConfig{})
	}
	return b
}

func (b *badge) Light() uint16 { return b.light.Get() }

// Battery follows Pimoroni's battery.py: vdd from the 1.24 V reference, ×3 divider.
func (b *badge) Battery() (float32, bool) {
	ref := b.vref.Get()
	if ref == 0 {
		return 0, machine.VBUS_DETECT.Get()
	}
	vdd := 1.24 * 65535 / float32(ref)
	return float32(b.vbat.Get()) / 65535 * 3 * vdd, machine.VBUS_DETECT.Get()
}

func (b *badge) SetLED(p int)       { b.led.set(p) }
func (b *badge) SetBacklight(p int) { b.backlight.set(p) }

// button turns a GPIO into app.Press events. Navigation keys fire on press
// and auto-repeat as long presses; A/B/C fire on release (held ≥ 500 ms = long).
type button struct {
	pin        machine.Pin
	id         app.Button
	repeat     bool
	down       bool
	since, rep time.Time
}

const (
	holdTime   = 500 * time.Millisecond
	repeatTime = 120 * time.Millisecond
)

func newButton(pin machine.Pin, id app.Button, repeat bool) *button {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // active high
	return &button{pin: pin, id: id, repeat: repeat}
}

func (b *button) poll(a *app.App, now time.Time) {
	pressed := b.pin.Get()
	switch {
	case pressed && !b.down:
		b.down, b.since, b.rep = true, now, now
		if b.repeat {
			a.Press(b.id, false)
		}
	case pressed && b.down && b.repeat:
		if now.Sub(b.since) >= holdTime && now.Sub(b.rep) >= repeatTime {
			b.rep = now
			a.Press(b.id, true)
		}
	case !pressed && b.down:
		b.down = false
		if !b.repeat {
			a.Press(b.id, now.Sub(b.since) >= holdTime)
		}
	}
}
