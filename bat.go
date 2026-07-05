package main

// Bat is the player's paddle. It animates between five types as powerups change
// it, can fire bullets in gun form, and stops balls in magnet form.
type Bat struct {
	Actor
	controls    Controls
	fireTimer   int
	currentType int
	targetType  int
	frame       int
	shadow      Actor
}

func NewBat(controls Controls) *Bat {
	b := &Bat{
		controls:    controls,
		currentType: BatNormal,
		targetType:  BatNormal,
	}
	b.Actor = newActor("blank", 320, 590, AnchorBat)
	b.shadow = newActor("blank", 320+16, 590+16, AnchorBat)
	return b
}

func (b *Bat) Update(g *Game) {
	// Animate towards a new bat type (12 frames, a new frame every 4).
	if b.targetType != BatNormal && b.targetType == b.currentType && b.frame < 12 {
		b.frame++
	}
	// Switching type from something other than normal: animate back to frame 0 first.
	if b.targetType != b.currentType && b.frame > 0 {
		b.frame--
	}
	if b.frame == 0 {
		b.currentType = b.targetType
	}

	b.Image = "bat" + itoa(b.currentType) + itoa(b.frame/4)

	b.fireTimer--

	// Fire the gun?
	if b.controls.fireDown() && b.currentType == BatGun && b.frame == 12 && b.fireTimer <= 0 {
		b.fireTimer = FireInterval
		b.Image += "f" // barely visible for its single frame
		g.bullets = append(g.bullets, NewBullet(b.X-20, b.Y, 0))
		g.bullets = append(g.bullets, NewBullet(b.X+20, b.Y, 1))
		g.playSound("laser", 1)
	}

	// Move, clamped to the screen (width comes from the current sprite).
	newX := b.X + b.controls.getX()
	minX := BatMinX + float64(int(b.width(g.assets))/2)
	newX = maxf(minX, newX)
	if !g.portalActive {
		maxX := BatMaxX - float64(int(b.width(g.assets))/2)
		newX = minf(maxX, newX)
	}
	b.X = newX

	// Shadow follows.
	b.shadow.X = b.X + 16
	b.shadow.Y = b.Y + 16
	b.shadow.Image = "bats" + itoa(b.currentType) + itoa(b.frame/4)
}

func (b *Bat) changeType(t int) { b.targetType = t }

func (b *Bat) isPortalTransitionComplete(g *Game) bool {
	return b.X-float64(int(b.width(g.assets))/2) >= Width
}
