---
aliases:
  - "Library scope and layers"
tags:
  - tinygui
  - adr
type: adr
status: accepted
decided: 2026-11-10
owner: user
created: 2026-11-12
updated: 2026-11-12
---

# 0001: TinyGUI is a board-, font- and renderer-agnostic GUI layer

Part of [[decisions]].

## Decision (owner, verbatim, 2026-11-10)

> "tinygui itself should not be specific to a font or font library or drawing primitives. it is just gui library that can be used to create interactive animated guis. i was inspired by fyne where drawing, data, layout, control, and logic layers were clearly separated. but of course i made some compromises. so, the actual rpi2040 code and actual badge code should be only in examples, not in the library itself. library should ease the development of actual drawing stuff with as little boilerplate as possible. ofc heavy on build flags since embedded don't need runtime detection and all that stuff - build flags would strip all that is unnecessary for really tight spaces(just like tinygo proclaims)."

## Consequences

- Board code (pins, panel drivers, buttons, ADC, PWM) lives under `examples/`. The Tufty 2040 demo follows this: `tinytufty2040/app` uses only TinyGUI; board files carry `//go:build tufty2040` ([[example-tufty2040]]).
- Gaps against this decision at `b140156` (input to TG-1b):
  - `widget` imports `tinygo.org/x/tinyfont` directly (`widget/label.go:7`, `toggle.go`, `multiline.go`, `interactive_choice.go`, `interactive_options.go`): text widgets are tied to one font library.
  - `ui` imports `machine` in `ui.go` (`PeekButton`), `serial_reader.go`, `watchdog_rp2040.go`: board code inside the library, gated by hand-written chip lists ([[topology#Build tags and per-target files]]).
  - `drawing.go` imports the drivers' PNG decoder; every app that links `widget.Icon` pays for it (≈ 13 KB flash + 34 KB RAM for the inflate window in the Tufty build, [[example-tufty2040#Size]]).
  - Widgets draw with `drivers.Displayer` primitives directly; there is no renderer/canvas layer between widgets and pixels (Fyne's split of widget → renderer → canvas objects).
- Build tags, not runtime detection, select optional features.
