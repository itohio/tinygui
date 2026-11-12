---
aliases:
  - "Demo and benchmark plan"
  - "Badge demos"
tags:
  - tinygui
  - analysis
  - tufty2040
type: reference
created: 2026-11-12
updated: 2026-11-12
revision: f12a9d0+wip
---

# Demo and benchmark plan (Tufty 2040)

What the Tufty 2040 can draw with TinyGo, measured, and the demos the owner asked for, in his order: **Game of Life → water (64×64, LED-matrix style) → fire → ping-pong → boids**. Part of [[analysis]]. Board: [[example-tufty2040]].

## How every step is verified (owner's standing rule)

Flash the badge → drive it over USB serial (`badgectl`, [[tufty2040-commands]]) → capture with the badge cam (esp32cam's camclient snapshot, or `grab`, both with the owner's `tmp/badgecam/settings.json` once he has saved it) → compare with the host simulator and the framebuffer dump → improve → evidence per round in `docs/evidence/`. Each demo lands as small commits that build, pass their tests and were seen on the badge.

## Measured (2026-11-12, `tinytufty2040/bench`)

RP2040 at 200 MHz (`machine.CPUFrequency`), ST7789 on the bit-banged 8-bit parallel bus (83 ns per byte), TinyGo 0.43. Each number is the mean over 20 steps (simulations) or one call (pushes); console log in `docs/evidence/TG-1-bench.txt`.

| Operation | Size | Time | Note |
|---|---|---|---|
| push RGB565 | 320×240 | **16.2 ms** | 153.6 KB buffer; the bus is the floor: ≈ 60 fps for pushing alone |
| push 8-bit indexed (LUT → RGB565 on the way out) | 320×240 | 17.3 ms | 76.8 KB buffer (half), +7 % time |
| push 4-bit indexed (16-colour LUT) | 320×240 | 17.7 ms | 38.4 KB buffer (quarter), +9 % |
| push RGB565 window | 64×64 | 0.88 ms | partial updates cost ≈ 0.21 µs per pixel: a dirty rectangle pays only for its area |
| Game of Life step, naive (`%` for wrap) | 64×64 | 15.7 ms | the M0+ has no divide instruction |
| Game of Life step, border cells, no `%` | 64×64 | **1.8 ms** | 8.7× faster |
| Game of Life step, border cells | 160×120 | 8.3 ms | |
| ×3 upscale to pixels, naive (`/` per pixel) | 192×192 | 54.5 ms | |
| ×3 upscale, copy cells and rows | 192×192 | **2.5 ms** | 22× faster |
| ×2 upscale, copy cells and rows | 320×240 | 6.3 ms | |
| fire ("Doom fire") step | 160×120 | **5.5 ms** | byte per cell, no division |
| fire step | 320×240 | 22.0 ms | full resolution, one byte per pixel |
| LED-matrix render, round dots, naive per-pixel test | 64×48 cells at 5 px | 23.7 ms | 320×240 output; a precomputed dot mask per row is the planned fix (estimate ≤ 5 ms) |
| LED-matrix render | 64×64 cells at 3 px | 10.8 ms | 192×192 output; 2 px dots are too small to look round |

Lessons that shape every demo: keep division and modulo out of inner loops; push only what changed; 8-bit indexed frames halve RAM for a 7 % push cost.

## Resolution × colour depth × update strategy

| Strategy | RAM for the frame | Frame time floor (push) | Fits with the showcase UI? | Best for |
|---|---|---|---|---|
| full RGB565 frame | 153.6 KB | 16.2 ms | no (the UI needs ≈ 110 KB of heap) | standalone effects with smooth gradients |
| RGB565 in 3 bands (current showcase) | 51.2 KB | 19.4 ms + 3 draw passes | yes | widget UIs |
| full 8-bit indexed frame + 256-colour LUT | 76.8 KB | 17.3 ms | tight | fire, plasma, Life, water: palette effects |
| full 4-bit indexed frame + 16-colour LUT | 38.4 KB | 17.7 ms | yes | Life, LED matrices, retro games |
| no frame: cell grid rendered row by row into one strip and pushed | ≈ 3.2 KB per 5-px LED row | 16.2 ms full screen | yes | LED-matrix demos (water, Life, fire) |
| dirty rectangles | whatever changed | ≈ 0.21 µs × changed pixels | yes | ping-pong, games, UIs where little moves |
| second core (`-scheduler=cores`, TinyGo 0.43 has `numCPU = 2` on RP2040) | — | simulation on one core overlaps the push on the other: up to ≈ 2× | — | water, boids; **untested**, a spike before relying on it |

## Owner's demos

| Demo | Grid / resolution | RAM (state + frame) | Per-frame budget (measured / estimated) | Expected fps |
|---|---|---|---|---|
| **Game of Life** (must) | 64×64 LED style, or 160×120 ×2 | 2 × 4.2 KB grids (+ border) + strip or 4-bit frame | step 1.8 ms + render ≤ 5 ms (est.) + push 16 ms | ≈ 40 fps capped by the push; 30 fps target |
| **water** (64×64 LED style) | 64×64 cells; see the fit question below | sand/particle model ("PixelDust"): occupancy 4 KB + ≤ 2 000 grains × 8 B = 16 KB; FLIP in fixed point: ≈ 40–60 KB; frame: strip (3 KB) or 4-bit frame (38.4 KB) | sand step ≈ 2–5 ms (est.), FLIP 15–30 ms (est., M0+ has no FPU) | sand 30+ fps; FLIP 15–25 fps unless the second core helps |
| **fire** | 160×120 ×2 (or 320×240) | 19.2 KB heat grid + 76.8 KB 8-bit frame (or the heat grid *is* the frame at 320×240) | step 5.5 ms + upscale ≈ 3 ms (fast copy) + push 17.3 ms | ≈ 35 fps; full-res 320×240: 22 + 17.3 ms ≈ 25 fps |
| **ping-pong** | full 320×240, dirty rectangles (two paddles, ball, score) | < 10 KB | physics < 0.1 ms; push of 3–4 small rectangles ≈ 1–2 ms | 60 fps possible; vsync-free tearing to check on the glass |
| **boids** | 200–400 boids on 320×240, 4-bit frame or dirty rects | boids × 12 B (fixed-point pos/vel) ≈ 5 KB + frame | naive O(n²) 400 boids ≈ 160 000 pair checks ≈ 15–25 ms (est.); a spatial grid brings it to ≈ 3–5 ms | 30 fps with a grid |

**Water at 64×64 LED style (owner's choice: a 64×64 simulation grid).** The panel is 320×240, so 64 rows fit only at a pitch of ≤ 3.75 px. Rendered with the real `CellGrid` (`docs/evidence/water-fit/fit-options.png`):

| Display of the 64×64 grid | Screen area | Dot | Rows shown | Verdict |
|---|---|---|---|---|
| pitch 3 | 192×192 | 2 px square-ish dots, 1 px gap | all 64 | everything visible, but the dots don't read as round LEDs |
| pitch 4 | 256×240 | 3 px rounded dots | 60 (rows 2–61) | LED look, loses 2 rows top and bottom (the bottom rows are where water settles) |
| pitch 5 | 320×240 | 4 px round dots | 48 (a 64×48 window) | the LED-badge look; the window can follow the water surface |

Plan: simulate 64×64; show it at pitch 3 inside the showcase page (fits the page body exactly), and offer a full-screen mode with pitch 4 or 5 where the owner picks the look on the badge itself.

**RAM, "in the same framebuffer".** As a page of the showcase the water draws into the showcase's banded framebuffer (51.2 KB, already allocated), so it adds only its own state:

| Item | Bytes |
|---|---|
| water state, 64×64 cells, 1 byte each (water bit + moved-this-step bit) | 4 096 |
| shade buffer the `CellGrid` reads (palette index per cell: depth/foam) | 4 096 |
| total for a cellular-automaton water (falling-liquid rules on the grid) | **8 192** |
| alternative: particle water ("PixelDust"), up to 2 000 particles × 6 B (int16 x, y + int8 vx, vy) + 4 096 B occupancy | 16 096 |
| alternative: full-screen RGB565 frame just for the demo, if it were its own firmware | + 153 600 |

The showcase heap is 142 KB with peaks of 116 KB, so the 8 KB automaton fits inside the showcase; the particle version (16 KB) also fits but leaves ≈ 10 KB of headroom. A standalone water firmware with a full RGB565 frame would use 153.6 + 8 ≈ 162 KB of 264 KB.

## Classic badge and game benchmarks

| Benchmark | Feasible in TinyGo on the Tufty? | Main cost | TinyGUI features it exercises |
|---|---|---|---|
| Game of Life | yes, measured above | step + push | cell-grid / LED widget, palette, fixed-step timing |
| Pong | yes | push (dirty rects) | sprites/blits, input latency, dirty rectangles, score labels |
| Snake | yes, trivial | push | cell grid, input repeat |
| Breakout | yes | collision + dirty rects | sprites, dirty rectangles, fixed-point physics |
| Tetris | yes | trivial CPU | cell grid, input repeat/long press, labels |
| Flappy-style scroller | yes | scrolling a band | hardware scroll (ST7789 VSCRSADD) or band redraw, sprites |
| Asteroids / Space Invaders | yes | vector lines / sprites | line drawing, rotation (fixed-point sin LUT), dirty rects |
| Starfield / plasma / tunnel / fire (demo effects) | yes; fire and plasma measured (plasma runs in the showcase) | per-pixel math | indexed frames, LUTs, fixed-point |
| Mandelbrot | yes, slowly: 320×240 × ~30 iterations of fixed-point math ≈ 1–3 s per frame (est.) | integer multiply | progressive rendering (partial updates) |
| Rotating 3D cube | yes | 8 vertices, 12 lines | line drawing, fixed-point matrices, double buffering |
| Wolfenstein-style raycaster | yes: 320 rays × DDA per frame ≈ 10–20 ms (est.); textures 64×64 × 8 bit = 4 KB each in flash | per-column texture fetch | column rendering into an indexed frame, flash-resident assets |
| "Can it run Doom" | not realistic now: rp2040-doom (C) needs heavily optimised C, ~2 MB of flash for code + compressed WAD, and its own RAM tricks; TinyGo would need the whole engine ported to Go, the `tufty2040` target's 1 020 KB flash limit raised (board has 8 MB), and a far smaller GC footprint | engine port | — |

## Flash budget

| Item | Size |
|---|---|
| board flash | 8 MB (Pimoroni) |
| what TinyGo's `tufty2040` target links | **1 020 KB** (`--defsym=__flash_size=1020K` in `targets/tufty2040.json`; raisable with a custom target) |
| showcase firmware today (12 pages incl. Life, water, fire, pong, boids) | 242 KB (`tinygo build -size short`: 242 132 B flash, 95 160 B static RAM) |
| bench / bring-up | 18 KB / 15 KB |
| per demo, **measured** | Life page + `CellGrid` + `Clipper` 3.7 KB (225 104 → 228 812 B); water page 3.3 KB (→ 232 068 B), its full-screen view + scanline path 1.9 KB (→ 233 924 B); fire 2.3 KB (→ 236 212 B); pong + tab scrolling 2.9 KB (→ 239 124 B); boids 3.0 KB (→ 242 132 B) |
| per demo, estimated | raycaster 10 KB + 4 KB per texture |

**Recommendation: one firmware with a demo menu** — the TinyGUI navigator is the menu (dogfooding), all demos together stay under ≈ 300 KB of the 1 020 KB, and the owner never re-flashes to switch. Demos share one static work buffer (largest demo wins) instead of allocating per demo, so switching can't fragment the heap (findings F-31). The raycaster is un-parked by this budget: it fits.

## Library features this adds (to land as tinygui widgets, not demo code)

- `widget.CellGrid` / LED-matrix renderer: W×H cells of `uint8`, palette (LUT), cell pitch, gap, round/square dots; renders row strips (no frame buffer needed) — used by Life, water and fire.
- indexed-colour frames (8-bit / 4-bit + LUT) next to `examples/framebuffer.RGB565`.
- dirty-rectangle tracking in the frame loop (ping-pong, boids, UIs).
- fixed-point helpers (Q16.16 mul/div, sin/cos LUT) for physics without an FPU.
- frame pacing and timing (`ui.Run` from UX-28 A3) with the `stats` numbers built in.

## Steps (each a small commit, verified on the badge)

1. **done** `7d135c2` `widget.CellGrid` with round/square cells, pitch, gap, palette; host tests. `ad8809f` `ui.Clipper` so banded displays let it skip rows outside the band.
2. **done** `bacafec` Game of Life page (60×32 cells at 5 px, ages fade, Run/Pause, Random / glider gun / Acorn / R-pentomino); badge: **17 fps** (draw 34 ms, push 18.6 ms; 14 fps before `Clipper`), framebuffer dump `docs/evidence/life-badge-dump.png`; badge-cam photo pending (camera client being rebuilt).
3. **done** `0aef293` Water page: falling-liquid automaton on a 64×64 grid, `CellGrid` at 3 px (192×192), gravity in 8 directions on UP/DOWN, Pour, Fill, 3 steps per frame. Badge: step 7.8 ms → **5.1 ms** after one-pass depth shading; page **14 fps** (draw 31 ms, push 18.5 ms, steps 15 ms); dump `docs/evidence/water-badge-dump.png`. `1dfe41c` CellGrid copies RGB565 scanlines on bitmap displays (draw: full-screen 31.9 → 18.1 ms, Life 34 → 25 ms → **20 fps**); `765fe21` full-screen views: LEDs 5 px (64×48, round 4 px dots) or 4 px (64×60), **18 fps** on the badge, rows follow gravity; dumps `water-full5-badge-dump.png`, `water-full4-badge-dump.png`.
4. **done** `f1ec9ee` demos share one tick / full-screen hook; `a2c9ecf` Fire page: Doom fire on 64×48, `CellGrid` 3 px in the page, 5 px full screen, wind and height; badge: step **0.87 ms**, page **21 fps**, full screen **25 fps**; dumps `fire-badge-dump.png`, `fire-full-badge-dump.png`. The 8-bit indexed 160×120 variant is not built: the LED look is the owner's choice.
5. **done** `086a7e4` scrolling tab bar (11 tabs); `3f4a0bc` Pong on `CellGrid` 64×48: CPU vs CPU in the page, you vs CPU full screen; host tests (fast-ball return, scoring and serve, ball stays on court); badge: **160 µs** per frame for the game, page **22 fps**, full screen **25 fps**; dumps `pong-badge-dump.png`, `pong-full-badge-dump.png`. Dirty rectangles not needed for the LED look.
6. **done** Boids (Reynolds, 32/64/128) on `CellGrid` 64×48; a plain neighbour scan with early reject is enough: 64 boids **2.4 ms**/step, page **21 fps**, full **23 fps**, 128 boids **21 fps**; no spatial grid or second core needed. Two bias bugs found on the way (minimum speed vs soft walls; `>>4` rounding drove the flock into a corner). Dumps `boids-badge-dump.png`, `boids-full-badge-dump.png`.
7. **done** Demo menu (owner 2026-11-12): `815b6f4` layout arrangers (Row/Column with weights and alignment, Wrap, Table, Anchor; inventory in [[layouts-and-overlays]]); `a0b917e` Life 64×48 + full screen; `2284816` Demos page with a live Life tile, `516d3ab` the other four tiles; `3698e5d` + `49f58e1` allocation fixes (findings F-34, F-35: ~70 → 7 GCs in the same window, water's tick 22.6 → 12.2 ms); `90a2821` + `262e609` + `ab01c77` overlays over the live tiles (dropdown, choice question, modal confirm, popup). Badge: Demos page **14–15 fps** at ×1; `tiles` budget (µs, tick / draw): Life 2 290 / 4 300 (focused, 2 px), water 12 230 / 2 450, fire 800 / 2 400, pong 170 / 2 400, boids 2 060 / 2 440. Water's three steps per frame are the largest share.
8. Frame cost shared by every page: the UI redraws whole in 3 bands (draw 31–34 ms on every page). Redrawing only changed widgets is the next large fps win (library: dirty tracking).
