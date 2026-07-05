package main

// Ball is the ball in play. It carries a unit direction vector and an integer
// speed, moving one pixel at a time and reflecting off walls, bricks and the bat.
type Ball struct {
	Actor
	dir        Vec2
	stuckToBat bool
	batOffset  float64
	speed      int

	speedUpTimer          int
	timeSinceTouchedBat   int
	timeSinceDamagedBrick int

	shadow Actor
}

func NewBall(x, y float64, dir Vec2, stuckToBat bool, speed int) *Ball {
	b := &Ball{
		dir:        dir, // Vec2 is a value type, so this is a copy.
		stuckToBat: stuckToBat,
		batOffset:  BallInitialOffset,
		speed:      speed,
	}
	b.Actor = newActor("ball0", x, y, AnchorCentre)
	b.shadow = newActor("balls", x+16, y+16, AnchorCentre)
	return b
}

// NewDefaultBall creates a fresh ball stuck to the bat.
func NewDefaultBall() *Ball {
	return NewBall(0, 0, Vec2{0, 0}, true, BallStartSpeed)
}

func (b *Ball) Update(g *Game) {
	b.timeSinceDamagedBrick++

	if b.stuckToBat {
		b.X = g.bat.X + b.batOffset
		b.Y = g.bat.Y - BallRadius

		if g.controls.firePressed() {
			b.stuckToBat = false
			_, b.dir = b.getBatBounceVector(g)
		}
	} else {
		b.timeSinceTouchedBat++

		// Speed up periodically; faster cadence if the bat hasn't been touched.
		b.speedUpTimer++
		if b.timeSinceTouchedBat > 5*60 {
			b.speedUpTimer++
		}
		interval := BallSpeedUpInterval
		if b.speed >= BallFastSpeedThreshold {
			interval = BallSpeedUpIntervalFast
		}
		interval2 := float64(interval) * 0.75
		if float64(b.speedUpTimer) > float64(interval) ||
			(float64(b.speedUpTimer) > interval2 && float64(b.timeSinceTouchedBat) > interval2) {
			b.incrementSpeed()
			b.speedUpTimer = 0
		}

		// Move one pixel at a time, speed times.
		for i := 0; i < b.speed; i++ {
			// X axis.
			b.X += b.dir.X
			if c, ok := g.collide(b.X, b.Y, b.dir, BallRadius); ok {
				b.dir.X = -b.dir.X
				b.X += b.dir.X
				if c.showImpact {
					g.impacts = append(g.impacts, NewImpact(c.posX, c.posY, 0xc))
				}
				if c.ctype == CollBrick {
					b.timeSinceDamagedBrick = 0
				}
				ballCollisionSound(g, c.ctype)
			}

			oy := b.Y

			// Y axis.
			b.Y += b.dir.Y
			if c, ok := g.collide(b.X, b.Y, b.dir, BallRadius); ok {
				b.dir.Y = -b.dir.Y
				b.Y += b.dir.Y
				if c.showImpact {
					g.impacts = append(g.impacts, NewImpact(c.posX, c.posY, 0xc))
				}
				if c.ctype == CollBrick {
					b.timeSinceDamagedBrick = 0
				}
				ballCollisionSound(g, c.ctype)
			} else if b.dir.Y > 0 {
				// Moving down - check for a bat collision.
				if oy+BallRadius <= BatTopEdge && b.Y+BallRadius > BatTopEdge {
					// Ball's bottom just crossed the top edge of the bat.
					collidedX, newDir := b.getBatBounceVector(g)
					if collidedX {
						if g.bat.currentType == BatMagnet {
							b.stuckToBat = true
							b.batOffset = b.X - g.bat.X
							b.dir = Vec2{0, 0}
						} else {
							b.dir = newDir
						}
						b.timeSinceTouchedBat = 0
						g.impacts = append(g.impacts, NewImpact(b.X, b.Y, 0xc))
						ballCollisionSound(g, CollBat)
						if b.stuckToBat {
							break
						}
					}
				} else if b.Y+BallRadius > BatTopEdge && b.Y < BatTopEdge+15 {
					// Hit the side of the bat - fling off at an extreme angle.
					collidedX, _ := b.getBatBounceVector(g)
					if collidedX {
						dx := -1.0
						if b.X > g.bat.X {
							dx = 1
						}
						b.dir = Vec2{dx, uniform(-0.3, -0.1)}.normalize()
						b.timeSinceTouchedBat = 0
						g.impacts = append(g.impacts, NewImpact(b.X, BatTopEdge, 0xc))
						b.speed = minInt(b.speed+4, BallMaxSpeed)
						ballCollisionSound(g, CollBatEdge)
					}
				}
			}
		}
	}

	b.shadow.X = b.X + 16
	b.shadow.Y = b.Y + 16
}

func (b *Ball) incrementSpeed() {
	b.speed = minInt(b.speed+1, BallMaxSpeed)
}

// getBatBounceVector returns whether the ball overlaps the bat on the X axis and,
// if so, the bounce direction for a top-of-bat hit.
func (b *Ball) getBatBounceVector(g *Game) (bool, Vec2) {
	dx := b.X - g.bat.X
	w := float64(int(g.bat.width(g.assets))/2 + BallRadius)
	if absf(dx) < w {
		return true, Vec2{dx / w, -0.5}.normalize()
	}
	return false, Vec2{0, -1}
}

// generateMultiballs returns three new balls spread 120 degrees apart.
func (b *Ball) generateMultiballs() []*Ball {
	var balls []*Ball
	for i := 0; i < 3; i++ {
		vec := b.dir.rotate(float64(i) * 120)
		if absf(vec.Y) < 0.15 {
			// Avoid a near-horizontal direction (e.g. if stuck to the bat).
			vec = Vec2{uniform(-1, 1), -1}.normalize()
		}
		balls = append(balls, NewBall(b.X, b.Y, vec, false, b.speed))
	}
	return balls
}

func (b *Ball) Draw(g *Game) { b.draw(g.assets) }

// ballCollisionSound plays the sound for a given collision type.
func ballCollisionSound(g *Game, ct CollisionType) {
	switch ct {
	case CollBrick, CollIndestructibleBrick:
		g.playSound("hit_brick", 1)
	case CollWall:
		g.playSound("hit_wall", 1)
	case CollBat:
		if g.bat.currentType == BatMagnet {
			g.playSound("ball_stick", 1)
		} else {
			g.playSound("hit_fast", 1)
		}
	case CollBatEdge:
		if g.bat.currentType == BatMagnet {
			g.playSound("ball_stick", 1)
		} else {
			g.playSound("hit_veryfast", 1)
		}
	}
}
