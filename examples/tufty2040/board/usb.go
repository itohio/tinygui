//go:build tufty2040

package board

import (
	"machine"
	"machine/usb"
)

// SetUSBSerial makes the USB serial number the flash chip's unique ID in hex,
// the same value MicroPython reported (df6750955f2a3536 on the owner's badge),
// so host tools find the badge by serial number instead of by port name. Call
// it first in main, before the host asks for the string descriptor; in an
// init() the flash command stops TinyGo from evaluating package initialisers
// at compile time.
func SetUSBSerial() {
	const hex = "0123456789abcdef"
	id := machine.DeviceID()
	s := make([]byte, 0, 2*len(id))
	for _, b := range id {
		s = append(s, hex[b>>4], hex[b&15])
	}
	usb.Serial = string(s)
}
