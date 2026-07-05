package main

import "math"

// Vec2 is a 2D vector, mirroring the subset of pygame.math.Vector2 the game uses.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) length() float64        { return math.Hypot(v.X, v.Y) }
func (v Vec2) lengthSquared() float64 { return v.X*v.X + v.Y*v.Y }

func (v Vec2) sub(o Vec2) Vec2 { return Vec2{v.X - o.X, v.Y - o.Y} }

// normalize returns the unit vector. A zero vector is returned unchanged
// (pygame raises in that case; the game avoids calling it on a zero vector).
func (v Vec2) normalize() Vec2 {
	l := v.length()
	if l == 0 {
		return v
	}
	return Vec2{v.X / l, v.Y / l}
}

// rotate rotates the vector by the given angle in degrees, matching
// pygame.math.Vector2.rotate (counterclockwise in vector space).
func (v Vec2) rotate(deg float64) Vec2 {
	rad := deg * math.Pi / 180
	c, s := math.Cos(rad), math.Sin(rad)
	return Vec2{c*v.X - s*v.Y, s*v.X + c*v.Y}
}
