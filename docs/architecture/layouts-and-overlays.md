---
aliases:
  - "Layouts and overlays"
tags:
  - tinygui
  - architecture
  - layout
type: reference
created: 2026-11-12
updated: 2026-11-12
---

# Layouts and overlays

Owner (2026-11-12): "utilise layout algorithms properly (we should have all useful layout algorithms implemented …)" and popups/dialogs/choice questions/combo boxes drawn "over the other renderings". This note lists what the library had, what TG-1 added, and what is still missing. Widgets are fixed-size (`WidgetBase.Width/Height`), so every layout places children; none resizes them.

## Two kinds of layout

| Kind | API | How it works | Can do | Can't do |
|---|---|---|---|---|
| cursor strategy (existed) | `layout.Strategy`, `container.WithLayout` | after each child is drawn, the strategy moves the context's cursor for the next child | lists, wrapping flows, simple grids; cheap, streaming | centre or right-align a row, split space by weight, equal columns: it never sees the other children |
| arranger (added, `815b6f4`) | `layout.Arranger`, `container.WithArranger` | before drawing, computes every child's rectangle in the box from all children's sizes (measure → arrange); re-run every `Draw`, so children that change size reflow | alignment, justification, weights, tables, stacking | resize children (no stretch), spans |

## Inventory

| Algorithm | Kind | Status | What it does |
|---|---|---|---|
| `HList(gap)` / `VList(gap)` | strategy | existed | children left→right / top→bottom |
| `Grid(px, py)` | strategy | existed | wraps by context width, row height = tallest so far |
| `HFlow(gap, maxW)` / `VFlow(gap, maxH)` | strategy | existed | wrapping flow with a limit (0 = context size, which ignores padding: findings F-33) |
| `container.Scroll` + `ScrollChoice` | container | existed (fixed `811a3f5`) | viewport over content laid out by a strategy; scrollbar observer |
| `Row(gap, main, cross, weights…)` / `Column(…)` | arranger | **added** | packed with gap and placed Start / Center / End / SpaceBetween, or **weighted cells** (weight 0 keeps its size, others share the rest), child aligned in its cell; cross-axis alignment |
| `Wrap{Gap, Main, Cross, Block}` | arranger | **added** | rows that wrap by width; each row justified, children aligned in the row, the block aligned vertically (the Demos page: the focused tile grows, rows reflow) |
| `Table{Cols, GapX, GapY, H, V}` | arranger | **added** | N equal-width columns, rows as tall as their tallest child, cell alignment |
| `Anchor{H, V, DX, DY}` | arranger | **added** | every child aligned in the whole box, offset: centred dialogs, corner badges, stacks in draw order |
| `Measurer` | arranger option | **added** | a zero-sized container takes the arranger's measured size (`Line`, `Anchor`) |

## Overlays (added, `90a2821`; `262e609` DrawLayers; on the badge `ab01c77`)

| Piece | What it does |
|---|---|
| `container.Overlay` | wraps a tree; draws up to `MaxOverlays` (4) layers over it; the top layer is modal; closing a layer closes those opened after it; `DrawLayers` for apps that paint chrome first. No damage tracking: what a layer covered is redrawn by the next frame (owner: "rendering layer business", later) |
| `ui.OverlayHost`, `ui.FindOverlayHost` | widgets reach the host through their parents |
| `widget.Dialog` | title, text lines, options: none = **popup**, two = **modal dialog** (button row), more = **choice question** (list); closes itself before calling back, so callbacks chain dialogs |
| `widget.Dropdown` | **combo box**: shows the choice; ENTER opens its list under itself, above when there's no room |
| `widget.OverlayStyle` | font, row height, colours |

## Still missing (TG-1b candidates)

- **Stretch / fill**: widgets can't be resized by a layout (no `Resizer`); weights only size cells.
- **Min/max constraints, aspect ratio**.
- **Grid spans** (a child over two columns) and named areas.
- **Anchoring to another widget** in general (only `Dropdown` does it, through its last draw position).
- **Clipping** of children to their container (findings: no clipping) and **damage / dirty rectangles** under overlays.
- **Focus order independent of child order** (navigation follows `Items`).
