# Kinetix — Go port

[![Latest release](https://img.shields.io/github/v/release/chrplr/kinetix-go)](https://github.com/chrplr/kinetix-go/releases/latest)

A Go re-implementation of the Pygame Zero game **Kinetix** from *Code the Classics
Volume 2* (Raspberry Pi Press), built on
[go-sdl3](https://github.com/Zyko0/go-sdl3) and the
[pgzgo](https://github.com/chrplr/pgzgo) harness.

All images, sounds and music are embedded, so `go build` produces a single
self-contained binary that needs no asset files at run time.

**▶ Play it in your browser: <https://chrplr.github.io/kinetix-go/>**

*The in-browser version runs at the correct speed on any monitor, including high-refresh (120/144 Hz) displays, since [pgzgo](https://github.com/chrplr/pgzgo) v0.4.0.*

## Controls

| Action | Keyboard | Gamepad |
|--------|----------|---------|
| Move bat | Left / Right | D-pad or left stick |
| Launch / fire | Space or Enter | A |
| Start  | Space / Enter | A |
| Quit   | Esc | Start |

Gamepad support is native-only; the in-browser build is keyboard-only.

## Download

Prebuilt, self-contained binaries — no install, no dependencies, assets embedded.
Grab the latest for your platform:

- **Linux** (amd64) — [kinetix-linux-amd64.tar.gz](https://github.com/chrplr/kinetix-go/releases/latest/download/kinetix-linux-amd64.tar.gz)
- **macOS** (Apple Silicon) — [kinetix-macos-arm64.tar.gz](https://github.com/chrplr/kinetix-go/releases/latest/download/kinetix-macos-arm64.tar.gz)
- **Windows** (amd64) — [kinetix-windows-amd64.zip](https://github.com/chrplr/kinetix-go/releases/latest/download/kinetix-windows-amd64.zip)

All versions are on the [releases page](https://github.com/chrplr/kinetix-go/releases).

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

## WebAssembly (play in a browser)

The game compiles to WebAssembly and runs in a browser using `web/index.html` as
the page shell. Building it needs a patched SDL3 (compiled for the browser, plus a
few js bindings not yet in upstream go-sdl3); these live in the
[go-sdl3-wasm](https://github.com/chrplr/go-sdl3-wasm) fork, which also ships the
`wasmsdl` tool that compiles the game and serves it alongside its `sdl.js` +
`sdl.wasm` runtime.

To build and run it locally:

1. **Get the fork**, as a sibling of this repo (one directory up):

   ```sh
   git clone -b wasm-render-fixes https://github.com/chrplr/go-sdl3-wasm ../go-sdl3-wasm
   ```

2. **Point the build at the fork.** From this repo's directory:

   ```sh
   go mod edit -replace github.com/Zyko0/go-sdl3=../go-sdl3-wasm
   ```

   (This is a local-only change — do not commit it. Step 4 undoes it.)

3. **Build and serve** the browser bundle, using `web/index.html` as the page:

   ```sh
   (cd ../go-sdl3-wasm && go run ./cmd/wasmsdl serve -html "$OLDPWD/web/index.html" "$OLDPWD")
   ```

   Then open <http://localhost:8080>.

4. **Restore `go.mod`** when you're done, so native builds keep using upstream
   go-sdl3:

   ```sh
   go mod edit -dropreplace github.com/Zyko0/go-sdl3
   ```

In the browser the game is **keyboard-only** (no gamepad on js yet). Audio works,
but SDL's Web Audio backend only resumes on your first keypress, so the title music
stays silent until the game starts.

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
design and original assets are © their respective authors / Raspberry Pi Press.

The original Python code and assets are in Raspberry Pi Press's [Code the Classics — Volume 2](https://github.com/raspberrypipress/Code-the-Classics-Vol2) repository.

The Go source code of this port is distributed under the MIT License.

See `Python_and_Go_implementation_comparison.md` for a walkthrough of how the port
maps onto the original.
