---
aliases:
  - "GUI prior art"
tags:
  - tinygui
  - research
type: research
status: draft
idea:
created: 2026-11-12
updated: 2026-11-12
---

# GUI prior art

Part of [[research]]. Seed list for TG-1b's prior-art survey (the full comparison and the gap matrix vs LVGL and Fyne are TG-1b). Evidence, not decisions; unconfirmed claims carry #verify.

## Question

Which embedded and small-footprint GUI libraries should TinyGUI learn from, and which is closest to it?

## Findings

| Project | What it is | Why it matters here | Source (read) |
|---|---|---|---|
| **aykevl/tinygl** (Go, BSD) | "Tiny graphics library for SPI connected displays", by TinyGo's lead developer; marked work in progress. Inspired by LVGL and Fyne. Colours stored as target pixel data (usually RGB565 big-endian); `int` coordinates in the API, `int16` in storage; "no memory allocations should be needed while writing data to the display"; supports most TinyGo drivers displays through `DrawRGBBitmap8`; styles kept out of the core (`style/basic`, `style/mono`); has a `gfx` canvas, `listbox`, flex layout and image (QOI, mono) packages | **closest prior art**: same language, same targets, same LVGL/Fyne inspiration, and its display contract (`DrawRGBBitmap8` + raw pixel buffers) is the alternative to TinyGUI's `BitmapDisplayer` discussed as option B3 in [[t-0001-m5stickc-examples-build]] | https://github.com/aykevl/tinygl README (2026-11-12) |
| **aykevl/board** (Go, BSD) | experimental board abstraction for TinyGo: common board API, a simulated board, board definitions generated from a devicetree-like format; boards include badger2040, gopher-badge, pybadge, pyportal, thumby, pinetime, MCH2022, GBA | prior art for TinyGUI's host simulator and for keeping board code out of the library ([[0001-library-scope-and-layers]]); no Tufty 2040 definition | https://github.com/aykevl/board README (2026-11-12) |
| LVGL, Fyne, Gio, Slint, emWin, TouchGFX, Embedded Wizard, uGFX, GUIslice, u8g2, Adafruit GFX, TFT_eSPI, LovyanGFX, embedded-graphics + Kolibri, egui, Nuklear, Dear ImGui, TinyGo drivers + tinydraw/tinyfont/tinyterm | the TG-1 brief's list | to be surveyed in TG-1b | `docs/tasks/TG-1.md` |
| mitxela fluid pendant, Ten Minute Physics FLIP, Adafruit PixelDust | liquid/sand simulations on small displays | prior art for the owner's "digital liquid" demo | — |

## Comparable work

tinygl and TinyGUI overlap almost completely in goal; tinygl's README lists as unfinished what TinyGUI also lacks (partial redraw for smooth animations, theming separated from layout, more widgets and containers). #verify how far tinygl's code is past its README (not read).

## Prior art in this vault

[[design-intent]] (the author's goals), [[findings]] (what TinyGUI does not do yet).

## Open questions

- Does tinygl's display contract avoid the per-call allocations TinyGUI hits in tinyfont and the drivers ([[findings]] F-21, F-29)? TG-1b reads its code.
