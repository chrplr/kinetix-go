package main

import "math"

// brickCollide reports whether a ball/bullet of radius r at (x, y) collides with
// the brick at grid (gridX, gridY), returning the collision point if so.
func brickCollide(x, y float64, gridX, gridY int, r float64) (float64, float64, bool) {
	x0 := x - r
	y0 := y - r
	x1 := x + r
	y1 := y + r

	xb0 := float64(gridX*BrickWidth + BricksXStart)
	yb0 := float64(gridY*BrickHeight + BricksYStart)
	xb1 := xb0 + BrickWidth
	yb1 := yb0 + BrickHeight

	xbc := (xb0 + xb1) / 2
	ybc := (yb0 + yb1) / 2

	// Bounce off the left/right side of the brick.
	if x1 > xb0 && x0 < xb1 && y > yb0 && y < yb1 {
		if x < xbc {
			return xb0, y, true
		}
		return xb1, y, true
	}

	// Bounce off the top/bottom of the brick.
	if x > xb0 && x < xb1 && y1 > yb0 && y0 < yb1 {
		if y < ybc {
			return x, yb0, true
		}
		return x, yb1, true
	}

	// Otherwise, check the nearest corner.
	corners := [4][2]float64{{xb0, yb0}, {xb1, yb0}, {xb0, yb1}, {xb1, yb1}}
	best := corners[0]
	bestD := sqDist(x, y, best[0], best[1])
	for _, p := range corners[1:] {
		if d := sqDist(x, y, p[0], p[1]); d < bestD {
			best = p
			bestD = d
		}
	}
	if math.Hypot(x-best[0], y-best[1]) < r {
		return best[0], best[1], true
	}
	return 0, 0, false
}

func sqDist(x, y, px, py float64) float64 {
	dx, dy := x-px, y-py
	return dx*dx + dy*dy
}
