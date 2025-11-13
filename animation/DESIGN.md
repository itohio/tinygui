# Animation Package Design

## Motivation
Interactive widgets like choice selectors benefit from smooth animated transitions. A dedicated animation package lets components opt into reusable motion behaviours (instant jumps, linear interpolation, easing) without embedding animation logic per widget. Because this project targets embedded systems, the design must avoid heap allocations and keep the API minimal.

## Goals
- Provide an `Animator` interface that operates on multiple scalar channels simultaneously (e.g., position X/Y, opacity, size).
- Store start/end values externally to avoid allocations; the caller supplies slices that remain valid for the animation lifetime.
- Constructors define the duration; there is no separate setter and no reset call—starting a new animation overwrites the previous configuration.
- `Update` accepts the destination slice, mutates it in place, and returns `true` once the animation reaches its end values.
- Timestamps use microseconds (`int64`, `time.Time.UnixMicro()`) for deterministic fixed-point style progress without floating-point drift.

## Core Interface

```go
package animation

type Animator interface {
    // Start initializes a new animation between start and end channel values.
    // The provided slices must have identical length; ownership stays with the caller and must live for the animation lifetime.
    Start(start, end []float32, startUnixMicro int64)
    // Update advances the animation based on `nowUnixMicro`, writing interpolated channel values back into `dst`.
    // Returns true when the animation has reached (or surpassed) its end state. The caller typically calls Update each frame until it returns true.
    Update(dst []float32, nowUnixMicro int64) bool
}
```

Key points:
- `Start` never allocates. Implementations copy channel data into internal fixed-size arrays or reuse the caller slices by reference (documented requirement: the caller keeps slices alive until another `Start`).
- `dst` passed to `Update` must have the same length as the slices originally provided to `Start`. Widgets can reuse the same buffer they render from.
- When duration is zero or negative, `Update` immediately copies the end values and returns `true`.

## Internal Structure
To avoid duplication, concrete implementations embed a shared `base` struct that tracks progress metadata while letting each easing curve define how normalized progress (`0…1`) is mapped.

```go
type base struct {
    start      []float32
    end        []float32
    channels   int
    startedUS  int64
    durationUS int64
    active     bool
}

func (b *base) begin(start, end []float32, startUS int64) {
    b.channels = len(start)
    b.start = start
    b.end = end
    b.startedUS = startUS
    b.active = true
}

func (b *base) fraction(now int64) (float32, bool) {
    if !b.active || b.durationUS <= 0 {
        return 1, true
    }
    elapsed := now - b.startedUS
    if elapsed <= 0 {
        return 0, false
    }
    if elapsed >= b.durationUS {
        return 1, true
    }
    return float32(elapsed) / float32(b.durationUS), false
}
```

Each concrete animator implements:

```go
type easingFn func(t float32) float32

type animator struct {
    base
    ease easingFn
}

func (a *animator) Update(now int64, dst []float32) bool {
    t, done := a.fraction(now)
    eased := a.ease(t)
    for i := 0; i < a.channels; i++ {
        dst[i] = a.start[i] + (a.end[i]-a.start[i])*eased
    }
    if done {
        copy(dst, a.end)
        a.active = false
    }
    return done
}
```

## Concrete Animators

| Constructor        | Behaviour description                                               | Easing                                                     |
|--------------------|---------------------------------------------------------------------|------------------------------------------------------------|
| `NewJump(us)`      | Immediate jump on the first update                                   | Returns `1` for any `t > 0`                                |
| `NewLinear(us)`    | Linear interpolation over duration                                   | Returns `t`                                                |
| `NewEaseIn(us)`    | Accelerating curve (ease-in quadratic)                               | Returns `t * t`                                            |
| `NewEaseOut(us)`   | Decelerating curve (ease-out quadratic)                              | Returns `1 - (1 - t)*(1 - t)`                              |
| `NewEaseInOut(us)` | Slow start/finish with faster middle (smoothstep)                    | Returns `t * t * (3 - 2*t)`                                |

Durations are specified via constructor parameter `us`; this value is stored and returned by `Duration()`. Passing `us <= 0` collapses the animation into a single frame.

## Usage Pattern
1. Widget constructs the chosen animator (e.g., `anim := animation.NewEaseOut(150_000)` for a 150 ms animation).
2. On state transition, widget calls `anim.Start(startValues, endValues, nowUnixMicro)` using slices it controls.
3. During each frame, widget allocates or reuses a buffer `dst` (often the same as the widget’s live state slice) and calls `done := anim.Update(nowUnixMicro, dst)`.
4. Once `done` is true, widget can switch to static rendering until the next `Start` call.

## Future Extensions
- Vector animators that operate on `[]int16` for pixel-perfect movement without float math.
- Looping or ping-pong behaviour wrappers (`Repeat`, `YoYo`).
- Composition helpers that sequence multiple animators or delay start times.
- Additional easing curves (cubic Bezier, elastic, bounce) implemented by plugging alternative `easingFn` functions.
