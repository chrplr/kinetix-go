package main

// Anchor y-kinds. X is always centred in this game.
const (
	akCenter = 0
	akAbs    = 1
)

// Anchor mirrors Pygame Zero's anchor tuples: x is centred, y is either centred
// or a fixed pixel offset from the top (e.g. the bat's ("center", 15)).
type Anchor struct {
	yKind int
	yVal  float64
}

var (
	AnchorCentre = Anchor{akCenter, 0}
	AnchorBat    = Anchor{akAbs, 15}
)

func (an Anchor) offset(w, h float64) (float64, float64) {
	ax := w / 2
	ay := h / 2
	if an.yKind == akAbs {
		ay = an.yVal
	}
	return ax, ay
}

// Actor is a positioned sprite with an anchor.
type Actor struct {
	X, Y   float64
	Image  string
	anchor Anchor
}

func newActor(image string, x, y float64, anchor Anchor) Actor {
	return Actor{X: x, Y: y, Image: image, anchor: anchor}
}

func (a *Actor) width(as *Assets) float64  { w, _ := as.Size(a.Image); return w }
func (a *Actor) height(as *Assets) float64 { _, h := as.Size(a.Image); return h }

// draw blits the sprite at its anchored screen position.
func (a *Actor) draw(as *Assets) {
	w, h := as.Size(a.Image)
	ax, ay := a.anchor.offset(w, h)
	as.Blit(a.Image, a.X-ax, a.Y-ay)
}
