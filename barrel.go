package main

// Barrel is a falling collectable powerup dropped by a destroyed brick.
type Barrel struct {
	Actor
	btype  int
	time   int
	shadow Actor
}

func NewBarrel(g *Game, x, y float64) *Barrel {
	// Weight table: higher weight = more likely. PORTAL can't drop unless only a
	// few bricks remain (and no portal is already open), when it becomes likely.
	portalWeight := 20
	if g.bricksRemaining > 20 || g.portalActive {
		portalWeight = 0
	}
	weights := []struct {
		t, w int
	}{
		{PowerupExtendBat, 6}, {PowerupGun, 6}, {PowerupSmallBat, 6},
		{PowerupMagnet, 6}, {PowerupMultiBall, 6}, {PowerupFastBalls, 6},
		{PowerupSlowBalls, 6}, {PowerupExtraLife, 2}, {PowerupPortal, portalWeight},
	}
	var types []int
	for _, e := range weights {
		for i := 0; i < e.w; i++ {
			types = append(types, e.t)
		}
	}

	b := &Barrel{btype: choiceInt(types)}
	b.Actor = newActor("blank", x, y, AnchorCentre)
	b.shadow = newActor("barrels", x+ShadowOffset, y+ShadowOffset, AnchorCentre)
	return b
}

func (b *Barrel) Update(g *Game) {
	b.time++
	b.Y++

	w := float64(int(g.bat.width(g.assets))/2 + BallRadius)

	// Collected by the bat?
	if b.Y >= BatTopEdge-10 && b.Y <= BatTopEdge+30 && absf(b.X-g.bat.X) < w {
		g.impacts = append(g.impacts, NewImpact(b.X, b.Y-11, 14)) // 14 == 'e'

		if snd, ok := powerupSound[b.btype]; ok {
			g.playSound(snd, 1)
		}

		// Move off the bottom of the screen so it is culled.
		b.Y = Height + 100

		if bt, ok := powerupBatType[b.btype]; ok {
			g.bat.changeType(bt)
		} else {
			switch b.btype {
			case PowerupMultiBall:
				var nb []*Ball
				for _, ball := range g.balls {
					nb = append(nb, ball.generateMultiballs()...)
				}
				g.balls = nb
			case PowerupFastBalls:
				g.changeAllBallSpeeds(3)
			case PowerupSlowBalls:
				g.changeAllBallSpeeds(-3)
			case PowerupPortal:
				g.activatePortal()
			case PowerupExtraLife:
				g.lives++
			}
		}
	}

	b.Image = "barrel" + itoa(b.btype) + itoa(b.time/10%10)
	b.shadow.X = b.X + ShadowOffset
	b.shadow.Y = b.Y + ShadowOffset
}

func (b *Barrel) Draw(g *Game) { b.draw(g.assets) }
