# Kinetix — Go port

A Go re-implementation of the Pygame Zero game **Kinetix** from *Code the Classics
Volume 2* (Raspberry Pi Press), built on
[go-sdl3](https://github.com/Zyko0/go-sdl3) and the
[pgzgo](https://github.com/chrplr/pgzgo) harness.

All images, sounds and music are embedded, so `go build` produces a single
self-contained binary that needs no asset files at run time. Keyboard and gamepad
are both supported.

## Run

```sh
go run .
```

go-sdl3 bundles the SDL3, SDL3_image and SDL3_mixer libraries and extracts them at
startup, so no system SDL install is needed.

## Headless self-test

```sh
go run . -selftest   # steps the game logic without a window, then exits
```

## Ebitengine vs. go-sdl3 + pgzgo

This game exists as two repositories that share their gameplay code verbatim and
differ only in the backend: **[kinetix-go](https://github.com/chrplr/kinetix-go)**
uses go-sdl3 with the [pgzgo](https://github.com/chrplr/pgzgo) harness, and
**[kinetix-go-ebitengine](https://github.com/chrplr/kinetix-go-ebitengine)** uses
[Ebitengine](https://ebitengine.org) (and is playable
[in your browser](https://chrplr.github.io/kinetix-go-ebitengine/)). Comparing them
is really a comparison of the two Go game stacks. Where they differ:

| Dimension | Comes out ahead | Why |
|-----------|-----------------|-----|
| Web / mobile reach | **Ebitengine** | Compiles to WebAssembly (see the live demo) plus iOS/Android; SDL3-via-purego has no real wasm story. |
| Build & dependencies | **Ebitengine** | Pure Go — `go build` just works and cross-compiles cleanly. go-sdl3 bundles SDL's C libraries and extracts them at runtime. |
| Maturity & ecosystem | **Ebitengine** | Years old, widely used and documented; go-sdl3 is a young (v0.1.x) binding over battle-tested SDL. |
| Built-in audio | **go-sdl3 + pgzgo** | SDL3_mixer offers looping tracks, gain and fades out of the box; Ebitengine's audio is lower-level. |
| Low-level control | **go-sdl3 + pgzgo** | Direct renderer, blend-mode and clip-rect access, with a small harness you fully own. |
| Headless testing | **go-sdl3 + pgzgo** | SDL's dummy driver runs the game with no display, so CI can `-selftest` the real loop. |

**Bottom line:** for something you want to ship or share, Ebitengine is the more
pragmatic default — its maturity and one-command web/mobile builds are hard to beat.
Reach for go-sdl3 + pgzgo when you want low-level SDL control, richer built-in audio,
or a minimal, transparent stack you own end to end.

Two things this pair shows beyond the scorecard:

1. **The engine choice barely touches the game.** Only `main.go` and the backend
   glue differ — the 13 gameplay files are identical. Keeping game logic
   engine-agnostic (plain structs behind an assets/audio/input seam) lets you swap
   backends later.
2. **The two APIs converged.** The Ebitengine adapter is thin because both stacks
   descend from the same `update()` / `draw()` / `screen` loop (Pygame Zero → pgzgo;
   Ebitengine's `ebiten.Game`). Good 2D game APIs keep rediscovering the same shape.

## Provenance & license

Ported to Go from the Python original in *Code the Classics Volume 2*. The game
design and original assets are © their respective authors / Raspberry Pi Press —
add the appropriate license before redistributing.

The Go source code of this port is distributed under the MIT License.

See `Python_and_Go_implementation_comparison.md` for a walkthrough of how the port
maps onto the original.
