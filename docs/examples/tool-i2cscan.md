---
aliases:
  - "i2cscan"
  - "cmd/i2cscan"
tags:
  - tinygui
  - examples
  - tools
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# cmd/i2cscan

Firmware that probes I2C addresses 1–126 on `machine.I2C0` and prints the ones that answer. Part of [[examples]].

## Behaviour

- Build constraint: `stm32 || rp2040 || esp32 || esp32c3 || nrf || sam || teensy || fe310 || hid` (`cmd/i2cscan/main.go:1`).
- Configures `machine.I2C0` with default `I2CConfig{}` (default pins and frequency of the target), waits 500 ms, then `bus.Tx(addr, nil, nil)` per address with 1 ms between probes (`:13-23`, `:25-35`).
- Prints `Found device at address: 0x` + two hex digits via `println` (serial console) (`:33`, `:38-41`).
- Unrelated to the GUI: a bring-up helper for sensors and PMICs (on the M5StickC the internal bus holds the AXP192, but that bus is on IO21/IO22 driven by soft-I2C in [[example-m5stick]], not necessarily `I2C0`'s default pins #verify against the target's `I2C0` defaults).

## Build

```sh
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -o /dev/null ./cmd/i2cscan   # builds at b140156
```

Host `go build` reports "build constraints exclude all Go files" (expected). Not run: running it needs the board ([[baseline]]).

## Notes

- `println("…0x", hex(addr))` puts a space between `0x` and the digits (`println` separates arguments with spaces).
- No result when nothing answers: the program exits silently.
