---
aliases:
  - "Examples"
  - "Demos and tools"
tags:
  - tinygui
  - examples
type: hub
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Examples

Demos and tools at `b140156`. Build status measured on 2026-11-10 with TinyGo 0.43.0 ([[baseline]]). Home: [[index]].

| Note | Path | Kind | Target | Builds at `b140156` |
|---|---|---|---|---|
| [[example-choice]] | `examples/choice` | demo | `m5stick-c`, tag `m5stickc` | **no** ([[t-0001-m5stickc-examples-build]]) |
| [[example-gauge]] | `examples/gauge` | demo | `m5stick-c`, tag `m5stickc` | **no** ([[t-0001-m5stickc-examples-build]]) |
| [[example-m5stick]] | `examples/m5stick` | board package | ESP32 targets (names ESP32 pins; fails on `pico`) | via the demos |
| [[tool-i2cscan]] | `cmd/i2cscan` | MCU tool | MCU families in its tag list | yes (`m5stick-c`) |
| [[tool-png2bin]] | `cmd/png2bin` | host tool | host | yes; not run (no PNG in the repo) |
| [[example-tufty2040]] | `examples/tufty2040` | showcase demo + host simulator | `tufty2040` (Pimoroni Tufty 2040), host sim | yes (2026-11-12, uncommitted); not yet run on the device |
| [[tufty2040-commands]] | `examples/tufty2040` | command set (serial + buttons) | `tufty2040` | — |
| — | `examples/framebuffer` | RGB565 RAM display | any | yes (used by the Tufty demo and its simulator) |

## Build commands

```sh
cd .
GOTOOLCHAIN=local GOPROXY=off go test ./container ./layout ./widget .           # host tests
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -tags=m5stickc -o /dev/null ./examples/choice
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -tags=m5stickc -size full -o /dev/null ./examples/gauge
GOTOOLCHAIN=local GOPROXY=off tinygo build -target=m5stick-c -o /dev/null ./cmd/i2cscan
GOTOOLCHAIN=local GOPROXY=off go run ./cmd/png2bin path/to/icon.png > icon.go
```

The `-tags=m5stickc` flag is required: without it every file of the demos is excluded and TinyGo prints only the package name as the error ([[baseline]]). Flashing (`tinygo flash`) is not covered here: it needs the owner and the board.

## Common shape of the demos

Both demos use the same skeleton ([[render-path]]): power up and configure the panel, build widgets → containers → `ScrollChoice` → root `container.New`, create a `Navigator`, then loop: poll two buttons, translate presses into navigator calls and `Interact` commands, clear the screen, draw the root, `Display()`, sleep 33 ms. Button handling is copied between the two `main.go` files.
