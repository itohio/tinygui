---
aliases:
  - "Analysis"
tags:
  - tinygui
  - analysis
type: hub
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Analysis

Measurements and evaluations of TinyGUI. Home: [[index]].

| Note | Content | Status |
|---|---|---|
| [[baseline]] | host tests, TinyGo builds, size, allocations, host probes at `b140156` | done (TG-1a) |
| [[findings]] | 27 defects and risks with evidence (F-22…F-27 from the Tufty demo) | done (TG-1a + demo) |
| prior art, gap matrix vs LVGL and Fyne, performance, concurrency, usability, ranked proposals | | TG-1b, not started; UX-28 A1 (non-generic container entry points), A2 (button adapter + `nav.Handle`), A5 (`tinygui_debug` tag for silent failures) and its minors feed its proposals |

## Answered

| Question | Answer | Source |
|---|---|---|
| Which device with a screen does the owner hand over? | **Pimoroni Tufty 2040**: RP2040, ST7789 320×240 on an 8-bit parallel bus, buttons A/B/C/UP/DOWN/USER, user LED, light sensor, battery sense. Enumerates as VID 0x2E8A, serial df6750955f2a3536 (currently MicroPython) | owner, 2026-11-10; [[example-tufty2040]] |
| How do we see a screen without hardware? | `examples/framebuffer` + the Tufty simulator write PNG screenshots of the real app code; making this host preview a library-level tool is the first TG-1c deliverable (UX-28 D3) | [[guide-port-to-board#4. See it before flashing]] |

## Known unknowns

| Question | Why it matters | How it gets resolved |
|---|---|---|
| Tufty 2040 on the device: panel orientation (MADCTL `0xB0`), colour inversion, parallel write timing, button polarity, ADC scaling | the demo's first run | owner-present flash; serial `madctl`, `pins`, `stats` ([[t-0002-tufty2040-board-package]]) |
| Frame time and SPI bytes per frame for the demos on the device | sizes the payoff of dirty rectangles, line buffers and faster text (F-05, F-06) | TG-1c: timestamp around `Draw` + `Display`, count bytes in a counting displayer on the host first |
| GC pause length with per-frame allocations under `gc.conservative` | decides how urgent zero-allocation rendering is (F-07, F-21) | TG-1c: `runtime.ReadMemStats` before/after N frames on the device |
| Which data forms TinyGo 0.43 keeps in flash (`.rodata`) vs RAM | 16 KB of the choice demo's 22 KB RAM is fonts and icons (F-09) | probe builds with `string`, array and slice tables, `-size full` |
| Is the clear-then-draw flicker visible on an unbuffered panel? | ranks dirty-region work; the Tufty demo sidesteps it with a RAM framebuffer, which costs 153.6 KB | M5StickC run (T-0001), camera or owner's eyes |
| Do IO37/IO39 need external pull-ups on the StickC (input-only pins)? | button reliability in the M5StickC demos (now p1) | schematic / datasheet |

## Known knowns

Listed with evidence in [[baseline]] and [[findings]]. Headline facts: the library is ≈ 9.6 KB of code with no static RAM; both demos fail to build at `b140156` ([[t-0001-m5stickc-examples-build]]); every frame repaints everything and allocates on the heap; the stateful grid/flow layouts drift between frames.
