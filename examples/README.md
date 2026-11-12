# TinyGUI Examples

TinyGUI examples illustrate how to assemble embedded user interfaces by composing containers, layouts, and interactive widgets without managing drawing contexts manually. Each sample focuses on small, purpose-built screens driven entirely by `UserCommand` inputs and the shared navigator.

- `choice/`: Shows how a complete navigator-driven menu is built for the M5StickC by wiring `ScrollChoice`, `InteractiveLabelChoice`, and `InteractiveWidgetChoice` into a single container. Icons are generated under `choice/icons`.
- `fire/`: The classic "Doom fire" drawn by `widget.CellGrid` as an LED matrix of round dots into the `framebuffer` helper; runs on the host and writes an animated GIF (`go run ./examples/fire -out fire.gif`). No board needed.
- `gauge/`: Mimics the TinyWater dashboard UI using `Gauge` and `InteractiveGauge` widgets across multiple zones. Selection and editing flow mirror the original device but remain UI-only for quick prototyping.
