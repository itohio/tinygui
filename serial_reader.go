//go:build stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid

package ui

import "machine"

// SerialReader adapts machine.Serialer to io.Reader so code can consume serial
// input using standard Go APIs.
type SerialReader struct {
	s machine.Serialer
}

// NewSerialReader wraps a machine.Serialer with io.Reader semantics.
func NewSerialReader(s machine.Serialer) *SerialReader {
	return &SerialReader{s: s}
}

func (s *SerialReader) Read(buf []byte) (int, error) {
	n := 0
	for i := range buf {
		if s.s.Buffered() == 0 {
			break
		}
		b, err := s.s.ReadByte()
		if err != nil {
			return n, err
		}
		buf[i] = b
		n++
	}
	return n, nil
}
