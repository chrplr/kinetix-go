# Kinetix — Python vs. Go implementation comparison

This document analyses how the Go port in this folder relates to the original
`kinetix.py`. It covers the structural mapping, the language‑paradigm differences
that shaped the port, the framework substitutions, and a set of subtle
numeric/semantic details that had to be reproduced exactly for the game to
behave the same way.

The goal throughout the port was **behavioural fidelity**: the Go code is a
faithful translation of the game logic, deviating only where a language or
library difference forces a different expression of the same idea, or where a
platform feature (game controllers) is out of scope for the port.

Kinetix is an *Arkanoid*/*Breakout* game: a bat, one or more balls with real
vector physics, destructible/indestructible brick grids, falling powerups,
bullets, and an exit portal. The Python original is ~1,200 lines in one file; the
Go port is ~1,500 lines across 17 focused files. The extra volume is mostly
explicit struct/interface declarations, per-type slice filters, and a small vector
type. The asset/sound/loop plumbing itself now lives in the pgzgo harness, not per game.

---

## 1. High‑level architecture

Both versions share the same conceptual design:

- A **bat‑and‑ball brick breaker**. The bat is driven left/right; balls bounce
  off walls, bricks, and the bat, with the bounce angle depending on where the
  ball strikes the bat. Destroying bricks scores points and sometimes drops a
  **barrel** (powerup). Clearing all destructible bricks — or collecting a portal
  powerup — opens an exit the bat walks through to the next level.
- **Flat actors**: unlike the platformer ports in this repo, there is no deep
  actor hierarchy. `Ball`, `Bat`, `Barrel`, `Bullet`, and `Impact` each extend
  Pygame Zero's `Actor` directly; collision is done functionally, not through an
  actor base class.
- A **title → play → game‑over** state machine, where the title screen runs a
  full **AI demo** game in the background.
- **Cycling levels**: six hand‑designed half‑levels, each mirrored horizontally
  at load, cycling forever.

The two largest pieces of logic in both — `Ball.update`/`(*Ball).Update` (the
per‑pixel physics and bat‑bounce handling) and `Game.collide`/`(*Game).collide`
(wall/brick collision with brick damage) — are ported statement‑for‑statement.

### File layout

| Concern | Python | Go |
|---|---|---|
| Constants / levels / powerup tables | top of `kinetix.py` | `constants.go` |
| Vector maths | `pygame.math.Vector2` | `vec.go` |
| Hex/util/rng helpers | inline / `random.*` | `util.go`, `rng.go` |
| Actor + anchors | Pygame Zero `Actor` | `actor.go` |
| Controls (keyboard + AI) | `Controls` ABC + subclasses | `controls.go` |
| Ball / Bat / Barrel / Bullet / Impact | those classes | `ball.go`, `bat.go`, `barrel.go`, `bullet.go`, `impact.go` |
| Brick collision geometry | `brick_collide` | `brickcollide.go` |
| Game (levels, collide, update, draw) | `Game` | `game.go` |
| Assets | Pygame Zero `images`/`screen` | pgzgo `Screen` |
| Audio | Pygame Zero `sounds`/`music` | pgzgo `Audio` |
| Input | `keyboard.*` | pgzgo `Keyboard` |
| State machine / entry point | `update`/`draw`/module code | `main.go` |

---

## 2. Language paradigm: ABC/inheritance → interface + embedding

The one real polymorphic hierarchy in Kinetix is the input abstraction. Python
uses an **abstract base class**:

```python
class Controls(ABC):
    def update(self):
        fire_down = self.fire_down()             # calls subclass method
        self.is_fire_pressed = fire_down and not self.fire_previously_down
        self.fire_previously_down = fire_down
    @abstractmethod
    def get_x(self): ...
    @abstractmethod
    def fire_down(self): ...

class KeyboardControls(Controls): ...
class AIControls(Controls): ...
```

The Go port expresses this with an **interface** for the polymorphism plus a
small embedded struct for the shared edge‑detection state:

```go
type Controls interface {
    update()
    getX() float64
    fireDown() bool
    firePressed() bool
    isAI() bool
}

type controlBase struct {           // shared fire-edge state
    firePreviouslyDown bool
    isFirePressed      bool
}
func (c *controlBase) updateFire(fireDown bool) { ... }

type KeyboardControls struct{ controlBase }
type AIControls struct{ controlBase; offset float64 }
```

Because Go embedding is not inheritance, the base `update()` can't call a
subclass `fire_down()` automatically. So each concrete type has its own tiny
`update()` that forwards its own `fireDown()` into the shared helper:
`func (c *KeyboardControls) update() { c.updateFire(c.fireDown()) }`.

The `isAI()` method replaces Python's `isinstance(self.controls, AIControls)`
check used by `in_demo_mode()` — Go favours an explicit method over a runtime
type test.

The game objects themselves need no interface: `Game` holds separate typed
slices (`[]*Ball`, `[]*Barrel`, …), and each is iterated in its own loop rather
than through a common base type.

---

## 3. Vector2 → a small `Vec2` value type

Ball direction is a unit `pygame.math.Vector2`, and the code leans on several of
its methods. The port provides just those in `vec.go`:

```go
type Vec2 struct{ X, Y float64 }
func (v Vec2) length() float64
func (v Vec2) lengthSquared() float64
func (v Vec2) normalize() Vec2
func (v Vec2) rotate(deg float64) Vec2   // used by multiball
```

A subtle but important behaviour: Python comments explicitly that
`self.dir = Vector2(dir)` makes a **copy**, because `Vector2` is a reference type
and sharing it would link two balls' directions. In Go this is free — `Vec2` is a
**value type**, so every assignment and struct field is already a copy. The port
notes this where `NewBall` stores `dir`.

`rotate` is implemented with the standard rotation matrix matching pygame's
convention; since multiball only ever rotates by 120° and 240°, the result is
three directions evenly spread regardless of orientation.

---

## 4. Framework: Pygame Zero → pgzgo (on go-sdl3)

| Pygame Zero feature | pgzgo equivalent (over go-sdl3) |
|---|---|
| `Actor("name", …)` auto-loads a PNG | `Screen.Texture` — pgzgo's lazily-cached texture |
| `screen.blit(name, (x,y))` | `Screen.Blit` / `BlitCentred` |
| `actor.width` (current sprite width) | `Actor.width(assets)` → live texture size |
| anchor tuples resolved internally | `Anchor` struct + `offset(w,h)` (§7) |
| `screen.surface.set_clip(rect)` | `Screen.SetClip` (pgzgo) |
| `keyboard.left`, `keyboard.space` | `app.Keyboard.Held(sc)` snapshot |
| `sounds.foo.play()` via `getattr` | `Audio.PlaySound(name, count)` |
| `music.play`/`music.stop` | `Audio.PlayMusic`/`StopMusic` |
| the `update()`/`draw()` loop | `app.Loop(update, draw)` — pgzgo's fixed-step, FPS-capped loop |

`actor.width` deserves a note: in Pygame Zero it is a **property** returning the
current sprite's width, and Kinetix reads it constantly (bat boundaries, ball/bat
overlap tests, barrel collection width) *after* the bat's image has been chosen
for the frame. The Go `width(assets)` likewise queries the current texture's
size, so it tracks the bat growing/shrinking between types exactly as the
original does.

---

## 5. Surface pre‑rendering → direct per‑frame drawing

This is the most significant rendering difference. For performance, Python
pre‑renders the bricks and their shadows onto two persistent
`pygame.Surface` buffers, updating them incrementally via `redraw_brick` whenever
a brick changes, then blits the whole buffer each frame:

```python
self.brick_surface  = surface.Surface((WIDTH, HEIGHT), flags=pygame.SRCALPHA)
self.shadow_surface = surface.Surface((WIDTH, HEIGHT), flags=pygame.SRCALPHA)
...
def redraw_brick(self, x, y):
    if self.bricks[y][x] is not None:
        self.brick_surface.blit(brick_image, (screen_x, screen_y))
        self.shadow_surface.blit(images.bricks, (screen_x+SHADOW_OFFSET, ...))
    else:
        self.brick_surface.fill((0,0,0,0), (screen_x, screen_y, ...))   # erase
```

The **authoritative state** is always the `self.bricks` grid; the surfaces are
just a cached render of it. The Go port drops the caching entirely and **draws
each brick (and each brick shadow) directly from the grid every frame**:

```go
for y := range g.bricks {
    for x := range g.bricks[y] {
        if g.bricks[y][x] != -1 {
            g.assets.Blit("brick"+hexDigit(g.bricks[y][x]), screenX, screenY)
        }
    }
}
```

`redraw_brick` therefore has no Go counterpart — mutating the grid *is* the
update. This is behaviourally identical (at most ~300 brick blits per frame, and
`SetClipRect` reproduces the shadow clipping that keeps shadows off the walls),
and it sidesteps needing SDL render‑to‑texture targets. The layered draw order —
arena, portal, brick shadows, object shadows, bricks, objects, impacts, HUD — is
preserved exactly.

---

## 6. Hexadecimal brick IDs and level mirroring

Brick types and impact/animation types are encoded as **hex digits**. Python
leans on built‑in `hex()` and `int(s, 16)`:

```python
self.bricks[y][x] = int(level[y][x], 16)        # 'a' -> 10
brick_image = getattr(images, "brick" + hex(v)[2:])   # 10 -> "bricka"
self.image  = "impact" + hex(self.type)[2:] + str(...) # 12 -> "impactc..."
```

The port provides two small helpers in `util.go`: `parseHexDigit(byte) int` for
loading, and `hexDigit(int) string` for building sprite names (returning
`"0".."f"`). So `brick10 → "bricka"`, `impact type 12 → "impactc"`, matching the
originals.

**Level mirroring.** Each `LEVELS` entry is only the left half; Python mirrors it
with a slice trick:

```python
return [row + row[-2::-1] for row in level]   # row + reversed(row without last char)
```

The Go `getMirroredLevel` builds the same string — the row plus its reverse
excluding the last character — turning an 8‑wide half into a 15‑wide symmetric
level.

---

## 7. The anchor system and shadows

Kinetix uses only two anchors: default centre, and the bat's `("center", 15)`
(centre‑x, 15px down from the top). `actor.go` models this minimally:

```go
type Anchor struct{ yKind int; yVal float64 }
var AnchorCentre = Anchor{akCenter, 0}
var AnchorBat    = Anchor{akAbs, 15}
```

Every moving object also owns a **shadow** — a second `Actor` drawn at a `+16`
(or `+SHADOW_OFFSET`) pixel offset with its own sprite (`balls`, `barrels`,
`bats{type}{frame}`). The port keeps a `shadow Actor` field on each object and
updates its position/sprite in the same places the Python code does, then draws
all shadows in a dedicated pass before the objects.

---

## 8. Numeric and semantic details reproduced exactly

Ball feel is sensitive to the exact arithmetic, so several details were
transliterated with care:

- **Integer speed, float direction.** `speed` is an integer number of 1‑pixel
  sub‑steps per frame; `dir` is a float unit vector. The port keeps `speed int`
  and the `for i := 0; i < b.speed; i++` sub‑step loop, matching Python's
  `range(self.speed)`.
- **X‑then‑Y separable collision.** Each sub‑step moves and resolves the X axis,
  then the Y axis, independently — including reverting the axis and inverting that
  component of `dir` on a hit. This ordering (and the `oy` "previous Y" capture
  used for the bat‑top test) is copied verbatim; it is what makes corner bounces
  behave the way they do.
- **Bat bounce vector.** `Vector2(dx / w, -0.5).normalize()` where
  `w = (bat.width // 2) + BALL_RADIUS` — the `//2` integer floor is preserved
  (`float64(int(width)/2 + BallRadius)`), so the bounce angle matches pixel‑for‑pixel.
- **Speed‑up cadence.** The dual‑threshold timer
  (`speed_up_timer > interval` **or** `speed_up_timer > interval2 and
  time_since_touched_bat > interval2`, with `interval2 = interval * 0.75`) and the
  fast‑ball threshold switching `interval` from 600 to 900 frames are ported as‑is,
  mixing int and float comparisons exactly as Python does.
- **Impacts live 16 updates.** `time` starts at 0, the sprite frame is `time // 4`,
  and impacts are culled once `time >= 16` — so only frames 0–3 ever show, even for
  the 5‑frame impact sets. The port keeps this (it is faithful, if slightly wasteful
  of art).
- **`None` brick sentinel.** Python stores `None` for an empty cell; Go can't put
  that in an `int`, so the port uses **`-1`**, matching every `is not None` /
  `is None` test, and counts destructible bricks as "not `-1` and not `13`".
- **Snapshot iteration.** Python updates `impacts + barrels + bullets` over a
  concatenated snapshot, so objects appended mid‑update aren't visited until next
  frame. Go's `range` captures the slice header up front, giving the same
  semantics without extra work.

---

## 9. The `game` global

Python reaches a module‑level `game` global from inside object methods
(`game.bat`, `game.collide(...)`, `game.impacts.append(...)`,
`game.play_sound(...)`). The Go port threads **`g *Game`** through the object
`Update`/`Draw` methods instead, making the data flow explicit.

The **one deliberate exception** is `AIControls`, which — like the Python
original — reads the live game state (`game.portal_active`, `game.balls[0].x`,
`game.bat.x`) to steer the demo bat. Since a `Controls` value is created before
its `Game` and outlives individual `Game` objects, the port keeps a single
package‑level `game *Game` that `AIControls.getX` consults, mirroring the Python
structure rather than inventing a back‑reference.

---

## 10. Intentional differences (out of scope for the port)

- **Game controllers.** Python has `JoystickControls` with dpad/analogue/dead‑zone
  handling and hot‑plug detection. The Go port ships keyboard only (arrows to
  move, **SPACE** to launch/fire). `AIControls` is fully ported since it drives the
  title‑screen demo.
- **Brick/shadow surfaces** are replaced by direct drawing (§5).
- **Version checks** for Python / Pygame Zero at startup have no Go analogue.
- **A `-selftest` flag** is *added* in Go: it runs a demo phase (the AI actually
  clears bricks and travels through the exit portal to the next level), then loads
  every level and applies one of each powerup, printing per‑level grid/brick/score
  counts. It exists only to verify the port without a display and has no Python
  counterpart.
- The **unused enemy‑door art** (`portal_meanie…`) is drawn exactly where the
  Python `draw` puts it — the original leaves these in as a hook for adding
  enemies, and the port preserves that.

---

## 11. Summary

The port is a close, behaviour‑preserving translation. The substantive rewrites
are all forced by the language or framework:

- the `Controls` ABC → an **interface + embedded base** with per‑type `update`
  forwarding;
- `Vector2` (a reference type) → a **`Vec2` value type** (copy‑by‑default);
- pre‑rendered brick/shadow **surfaces → direct per‑frame drawing** from the grid;
- `hex()`/`int(s,16)` → small **hex digit helpers**;
- the `game` global → an **explicit `*Game` parameter** everywhere except the
  demo AI, which keeps a package‑level reference as the original does;
- Pygame Zero's implicit asset/sound/clip/loop machinery → the **pgzgo harness**.

Everything that affects how the game *plays* — the per‑pixel X‑then‑Y ball
physics, the integer‑floored bat bounce vector, the speed‑up cadence, the brick
hex IDs and horizontal mirroring, the powerup weightings and effects, the exit
portal, and the stuck‑ball softening — is reproduced as‑is. The verification path
is `go build` + `-selftest` (the demo AI progresses through a portal, all six
levels load and every powerup applies, with no panics); on‑screen visuals and
audio require a real display to confirm.
