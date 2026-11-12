---
aliases:
  - "Discovery"
  - "Knowledge map catalog"
tags:
  - tinygui
  - discovery
type: hub
updated: 2026-11-10
---

# Discovery

What we know and what we don't. Dossiers and plans list their items here per quadrant ([[workflow#Knowledge map]]). Table views: `discovery.base`. Template: [[discovery-item]].

| Quadrant | Meaning | Techniques |
|---|---|---|
| Known knowns (KK) | facts with evidence | link the evidence; re-check when it gets stale |
| Known unknowns (KU) | open questions we must answer | desk research, interviews, spikes, prototypes, #verify scan |
| Unknown knowns (UK) | tacit assumptions, unwritten knowledge | assumption audit, "what would have to be true?", ask experts |
| Unknown unknowns (UU) | surprises we cannot name yet | pre-mortem, red-team prompts, failures of comparable work, early real tests |

## Lifecycle

- **Raise:** `d-NNNN-<slug>.md` from [[discovery-item]] (next free number, never reused); set `quadrant`, `impact`, `owner`; list it below.
- **Research:** `status: researching`; dated findings with evidence links.
- **Answer:** write the answer, `status: answered`, `answered:`, `confidence`; note what changed (notes, ADRs, tickets).
- **Quadrant moves:** change `quadrant`, append the old one to `quadrant-history`, never rename the note.
- **Obsolete:** no longer relevant, with a one-line reason.

## Index

Plain fallback for `discovery.base`.

_No `d-NNNN` items yet._ The open questions from the TG-1a analysis are listed in [[analysis#Known unknowns]]; each becomes a `d-` item when a plan depends on it.
