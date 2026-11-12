---
aliases:
  - "Guides"
  - "How-to"
tags:
  - tinygui
  - guides
type: hub
created: 2026-11-12
updated: 2026-11-12
revision: b140156+wip
---

# Guides

Task-first notes for an app author. Each one is short, gives code that was compiled and run on the host (2026-11-12, screenshots in the evidence folder), and links the reference notes for detail. Home: [[index]].

| Guide | Job |
|---|---|
| [[guide-first-screen]] | display → widgets → container → navigator → loop |
| [[guide-bind-data]] | pointers, closures, the pending-copy commit model, text without allocations |
| [[guide-bind-input]] | buttons → `UserCommand` → navigator, today's glue explained |
| [[guide-port-to-board]] | adapt a display to `ui.Displayer`, build tags, host preview, sizes |

The worked example behind all four is [[example-tufty2040]]: the same UI runs on the badge and in a host simulator that writes PNG screenshots.
