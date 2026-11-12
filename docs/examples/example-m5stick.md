---
aliases:
  - "M5StickC board support"
  - "examples/m5stick"
tags:
  - tinygui
  - examples
  - m5stickc
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# examples/m5stick — board support

Shared board package for the demos, not a runnable example. Part of [[examples]].

## What it provides

| Item | Source | Notes |
|---|---|---|
| Pin map | `examples/m5stick/cfg_m5stickc.go:15-34` | LED IO10, buttons Main IO37 / Side IO39, display SPI SCK IO13 / SDO IO15 / RST IO18 / DC IO23 / CS IO5, Grove I2C IO33/IO32, internal I2C IO22/IO21 |
| `InitPower()` | `:37-43` | soft-I2C on the internal bus, AXP192 LDO2 on (panel backlight) |
| `NewDisplay()` | `:46-65` | `machine.SPI3` at 32 MHz, ST7735 `MINI80x160`, 80×160, column offset 26, row offset 1, `ROTATION_270`; sleeps 1 s before init; returns `st7735.Device` **by value** |
| `i2c/` | `examples/m5stick/i2c/` | vendored soft-I2C with an ESP32 `nop` delay loop; unused |
| `st7735/` | `examples/m5stick/st7735/` | vendored, non-generic ST7735 driver that implements the full `ui.Displayer` (incl. `DrawRGBBitmap([]uint16)`, `:316`); unused (import commented out at `cfg_m5stickc.go:9`) |
| `mlx90614/` | `examples/m5stick/mlx90614/` | IR thermometer driver; unused |

Build constraint `tinygo || m5stickc` (`cfg_m5stickc.go:1`): the file is included for every TinyGo target, but it compiles only for ESP32 targets because it names ESP32 pins (`tinygo build -target=pico` of an importer: `cfg_m5stickc.go:16:23: undefined: machine.IO10`).

## Hardware facts to confirm on the device

- Panel: the config is for the original M5StickC (ST7735S, 80×160). The demos draw for 240×135, the M5StickC Plus panel (ST7789). Which board the owner hands over decides the fix ([[findings]] F-02).
- Buttons are active-low with pull-ups configured by the demos (`examples/choice/main.go:46-48`); IO37/IO39 are input-only on ESP32 and have no internal pull-up #verify against the ESP32 datasheet and the StickC schematic (external pull-ups expected).

## Build status

The package itself builds only as part of a demo; status and errors: [[t-0001-m5stickc-examples-build]] and [[baseline]].
