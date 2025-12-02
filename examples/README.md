# TinyGUI Examples

TinyGUI examples illustrate how to assemble embedded user interfaces by composing containers, layouts, and interactive widgets without managing drawing contexts manually. Each sample focuses on small, purpose-built screens driven entirely by `UserCommand` inputs and the shared navigator.

- `choice/`: Shows how a complete navigator-driven menu is built for the M5StickC by wiring `ScrollChoice`, `InteractiveLabelChoice`, and `InteractiveWidgetChoice` into a single container. Icons are generated under `choice/icons`.
- `gauge/`: Mimics the TinyWater dashboard UI using `Gauge` and `InteractiveGauge` widgets across multiple zones. Selection and editing flow mirror the original device but remain UI-only for quick prototyping.
