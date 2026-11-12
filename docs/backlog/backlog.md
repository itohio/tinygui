---
aliases:
  - "Backlog"
  - "Tickets"
tags:
  - tinygui
  - backlog
type: hub
updated: 2026-11-10
---

# Backlog

The local ticket tracker. A **ticket** is one implementable, verifiable change (about 1–2 days at most); humans and agents start every implementation from one ([[workflow#Tickets]]). Table views: `backlog.base`. Template: [[ticket]]; copy it to `backlog/t-NNNN-<slug>.md` with the next free number.

| Status | Meaning |
|---|---|
| `backlog` | captured, not ready |
| `todo` | ready: goal, acceptance criteria and verification written |
| `in-progress` | someone works on it (`assignee`) |
| `review` | done, evidence in the log, awaiting the user |
| `done` | accepted; `closed:` set |
| `blocked` | `blocked-by` lists tickets or discovery items |
| `cancelled` | will not be done; reason in the log |

Kinds: `feature`, `bug`, `chore`, `spike` (time-boxed investigation that answers a [[discovery]] item), `docs`.

## Tickets

Plain fallback for `backlog.base`; whoever creates a ticket adds it here.

| Ticket | Status | Priority | Kind | Area |
|---|---|---|---|---|
| [[t-0002-tufty2040-board-package]] | review (built; waiting for the owner's go to flash) | p0 | feature | examples |
| [[t-0001-m5stickc-examples-build]] | backlog (proposed, owner picks option B) | p1 | bug | build |
