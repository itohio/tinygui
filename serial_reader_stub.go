//go:build !(stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid)

package ui

type serialer interface {
	Buffered() int
	ReadByte() (byte, error)
}

type SerialReader struct{}

func NewSerialReader(_ serialer) *SerialReader {
	return &SerialReader{}
}

func (s *SerialReader) Read(_ []byte) (int, error) {
	return 0, nil
}
