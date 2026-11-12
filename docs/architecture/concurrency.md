---
aliases:
  - "Concurrency"
  - "Threading model"
tags:
  - tinygui
  - architecture
  - concurrency
type: reference
created: 2026-11-10
updated: 2026-11-10
revision: b140156
---

# Concurrency

Part of [[architecture]]. Sources at `b140156`.

## Model as built

- **Single goroutine, by convention.** The library starts no goroutines and uses no channels, mutexes or atomics (no `go `, `chan`, `sync` in non-test library code). All widget, container and navigator state is plain fields mutated from whichever goroutine calls `Interact`, `Navigator.*` or `Draw`.
- **App-owned loop.** Both examples run input polling, state changes and drawing on the main goroutine at ~30 Hz with `time.Sleep(33 * time.Millisecond)` (`examples/choice/main.go:191-248`).
- **Cooperative yields.** `CommandStreamMux.Run` calls `runtime.Gosched()` before every 1-byte read (`mux.go:30-33`) so it can run in its own goroutine next to the UI loop. With `SerialReader` returning 0 bytes when idle (`serial_reader.go:18-31`), that goroutine never blocks: it yields and polls again, i.e. busy-polls the scheduler.
- **Blocking helper.** `PeekButton` sleeps in 1 ms steps until release (`ui.go:12-20`); the caller's goroutine is blocked for the whole press. `time.Sleep` is a scheduler yield under `scheduler.tasks`, which all five targets checked use (`tinygo info` for `m5stick-c`, `esp32c3-supermini`, `xiao-esp32s3`, `pico`, `xiao`: tags include `scheduler.tasks … tinygo.unicore`), so other goroutines keep running.
- **Watchdog.** `updateWatchdog` is called only inside `PeekButton` (`ui.go:17`). It feeds `machine.Watchdog` on RP2040 (`watchdog_rp2040.go:7-9`) and is a no-op on ESP32/ESP32-C3 (`watchdog_esp32.go:5`) and elsewhere (`watchdog_stub.go:6`).
- **Data binding races.** Widgets read bound pointers and closures during `Draw` (`widget/gauge.go:67`, `widget/label.go:36`). If a sensor goroutine writes `*T` while the UI goroutine draws, the access is unsynchronised but cannot tear between goroutines: the checked targets are `tinygo.unicore` with the cooperative `scheduler.tasks`, so a switch happens only at a yield point (sleep, channel op, `Gosched`, blocking I/O), never inside an assignment. Interrupt handlers do preempt: an ISR that writes a multi-word value (`string`, slice, `float64` on 32-bit) the UI is reading can be observed half-updated. TinyGo 0.43 also ships `-scheduler=cores` with two CPUs on RP2040/RP2350 (`src/runtime/runtime_rp2_flashsafe_cores.go:1`, `runtime_rp2.go:21`); there goroutines run in parallel and unsynchronised bindings become real data races.
- **Interrupts.** No interrupt handlers in the library. Pin interrupts (`machine.Pin.SetInterrupt`) are not used; an app that pushes commands from an ISR has no safe queue in the library to push into.

## Implications to carry into TG-1b

- A lock-free hand-off from producer goroutines/ISRs to the UI loop (single-producer ring of `UserCommand`, or a dirty flag set by value setters) is missing.
- Frame pacing is a fixed sleep; there is no "redraw only when something changed" signal.
- The serial path needs a blocking or interrupt-driven read to stop busy-polling.

Evidence for the "no goroutines/sync" claim: [[baseline]] (grep command and output).
