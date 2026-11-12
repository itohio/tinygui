---
aliases:
  - "Tufty 2040 board package and showcase"
tags:
  - tinygui
  - ticket
type: ticket
status: review
priority: p0
kind: feature
area: examples
plan:
idea:
discovery: []
blocked-by: []
estimate: 2d
assignee: agent
created: 2026-11-12
closed:
updated: 2026-11-12
---

# T-0002: Tufty 2040 board package and showcase demo

Part of [[backlog]]. Ranked ahead of [[t-0001-m5stickc-examples-build]]: the owner's screen device is a Pimoroni Tufty 2040 (RP2040; USB VID 0x2E8A, serial df6750955f2a3536), so the first thing he runs should run on it (UX-28 D2).

## Goal

A board package and a demo for the Tufty 2040 that show every TinyGUI feature, with the UI in a board-independent package that also runs in a host simulator.

## Context and links

Owner request 2026-11-10 (direct): "a demo for this board that showcases all features of tinygui as of right now". Scope rules: [[0001-library-scope-and-layers]]. Implementation and facts: [[example-tufty2040]].

## Acceptance criteria

- [x] `tinytufty2040` (demo repo) builds for `-target=tufty2040` (223 064 B flash / 219 424 B RAM, 2026-11-12).
- [x] UI package has no `machine` import; board files carry `//go:build tufty2040`.
- [x] Host simulator renders every page; 13 screenshots inspected (`docs/evidence/tufty-sim/`).
- [x] App code allocates only at construction (`-print-allocs`).
- [x] On the device: runs stably; `stats` reports fps 15–21, push 12.8 ms, heap; `pins` shows idle button levels (2026-11-12).
- [ ] Owner confirms on the glass: picture upright, colours right, all six buttons act as labelled.
- [ ] Review (code) and verify accept.

## Verification

`go vet ./examples/framebuffer ./tinytufty2040/...`; `go run ./tinytufty2040/sim -out <dir>`; `tinygo build -target=tufty2040 -size short`; on the device: `tinygo flash -target=tufty2040 ./tinytufty2040`, then the serial `stats`, `pins` and `dump` commands.

## Log

- 2026-11-12: built from the owner's direct request; simulator and TinyGo build pass; library fix F-22 (Scroll) with regression test. Waiting for the owner's go to flash.
- 2026-11-12: flashed (owner approval relayed by Leader). First build ran out of memory (full framebuffer, F-30) → two-band rendering; Icon decoded PNG per draw (F-08) → decode once; edit marker stuck on rows (F-28) → drawn only on the focused row. Stable on the badge.
