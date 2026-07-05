package main

// Bullet is fired upward by the gun bat.
type Bullet struct {
	Actor
	alive bool
}

func NewBullet(x, y float64, side int) *Bullet {
	b := &Bullet{alive: true}
	b.Actor = newActor("bullet"+itoa(side), x, y, AnchorCentre)
	return b
}

func (b *Bullet) Update(g *Game) {
	b.Y -= BulletSpeed

	if c, ok := g.collide(b.X, b.Y, Vec2{0, -1}, 2); ok {
		b.alive = false
		g.impacts = append(g.impacts, NewImpact(b.X, b.Y, 15))
		if c.ctype == CollBrick || c.ctype == CollIndestructibleBrick {
			g.playSound("bullet_hit", 4)
		}
	}
}

func (b *Bullet) Draw(g *Game) { b.draw(g.assets) }
