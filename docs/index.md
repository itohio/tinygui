---
aliases:
  - "TinyGUI docs"
  - "Documentation index"
tags:
  - tinygui
  - home
type: hub
updated: 2026-11-10
---

# TinyGUI

A retained widget tree for TinyGo microcontrollers, drawn by full redraw into any `tinygo.org/x/drivers` display, driven by abstract `UserCommand` inputs through a `Navigator` (as built at `b140156`). Details: [[architecture]]. Scope set by the owner: [[0001-library-scope-and-layers]].

> [!note] Owner's description #verify
> The owner's one-sentence description of the project goes here.

> [!tip] Open this folder as an Obsidian vault
> `make -C ../obsidian up OPEN=tinygui/docs` from the project root (no project Makefile yet): the shared Obsidian starts at https://localhost:3011 (self-signed certificate). Links are wikilinks, metadata lives in properties (frontmatter), new notes come from `templates/`, pasted files land in `attachments/`. Open questions carry the #verify tag; find them in the tag pane.

## Start here

- [[guides]]: [[guide-first-screen]], [[guide-bind-data]], [[guide-bind-input]], [[guide-port-to-board]]
- [[example-tufty2040]]: everything TinyGUI has, on the owner's Tufty 2040 badge and in a host simulator

## Library

- [[architecture]]: packages, core types, design properties
  - [[topology]]: dependency graph, build tags, modules, display contract
  - [[render-path]]: from button press to pixels
  - [[navigation-and-input]]: commands, navigator, input sources
  - [[widget-catalog]]: every widget and container
  - [[memory-and-buffers]]: buffer ownership and allocations
  - [[concurrency]]: goroutines, scheduler, watchdog
  - [[design-intent]]: the author's goals and roadmap (former `DESIGN.md`), with drift
  - [[animation-design]]: the `animation` package (former `animation/DESIGN.md`)
- [[examples]]: the demos and tools, how to build them, current status
- [[analysis]]: baseline measurements, findings, improvement analysis

## Project work

- [[workflow]]: how work flows from idea to done; rules for humans and agents
- [[backlog]]: tickets (`t-NNNN`); every implementation starts from one
- [[inbox]]: new ideas
- [[discovery]]: what we know and what we don't (KK / KU / UK / UU)
- [[plans]]: plans of accepted ideas
- [[decisions]]: decision records (ADRs)
- [[research]]: research with sources

## The funnel

| Stage | Where | Who |
|---|---|---|
| Idea | [[inbox]] | the user writes it |
| Dossier | [[in-progress]] | an agent researches it and maps what is known |
| Decision | [[parked]] or [[accepted]] | the user decides |
| Plan | `plans/<date>-<slug>/` ([[plans]]) | an agent, after acceptance |
| Tickets | [[backlog]] | implementation units, then done |

## Conventions

- Unique kebab-case note names, wikilinks only, properties on every note: [[workflow#Properties]].
- Code claims cite `file:line` at the revision in the note's `revision` property; re-check when HEAD moves.
- #verify marks a claim that still needs evidence.
