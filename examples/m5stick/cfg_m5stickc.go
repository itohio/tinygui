//go:build tinygo || m5stickc

package m5stick

import (
	"machine"
	"time"

	// "github.com/itohio/tinygui/examples/m5stick/st7735"
	"tinygo.org/x/drivers/axp192"
	"tinygo.org/x/drivers/i2csoft"
	"tinygo.org/x/drivers/st7735"
)

var (
	LEDPin     = machine.IO10
	ServoPin   = machine.IO26
	BtnMain    = machine.IO37
	BtnSide    = machine.IO39
	DisplaySCK = machine.IO13
	DisplaySDO = machine.IO15
	DisplayRST = machine.IO18
	DisplayDC  = machine.IO23
	DisplayCS  = machine.IO5
	DisplayBL  = machine.NoPin

	// Grove connector
	I2cCLK  = machine.IO33
	I2cDATA = machine.IO32

	// I2C bus used for internal periphery
	I2cInternalCLK  = machine.IO22
	I2cInternalDATA = machine.IO21
)

// InitPower enables the LCD LDO to light up the display backlight on M5StickC.
func InitPower() *axp192.Device {
	i2c := i2csoft.New(I2cInternalCLK, I2cInternalDATA)
	i2c.Configure(i2csoft.I2CConfig{Frequency: 100e3})
	axp := axp192.New(i2c)
	axp.SetLDOEnable(2, true)
	return axp
}

// NewDisplay sets up the ST7735 display with the offsets required by M5StickC.
func NewDisplay() st7735.Device {
	spi := machine.SPI3
	spi.Configure(machine.SPIConfig{
		Frequency: 32000000,
		SCK:       DisplaySCK,
		SDO:       DisplaySDO,
	})
	time.Sleep(time.Second)
	display := st7735.New(spi, DisplayRST, DisplayDC, DisplayCS, DisplayBL)
	display.Configure(st7735.Config{
		Model:        st7735.MINI80x160,
		Width:        80,
		Height:       160,
		ColumnOffset: 26,
		RowOffset:    1,
		Rotation:     st7735.ROTATION_270,
	})

	return display
}
