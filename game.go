package main

// Collision describes a ball/bullet collision result.
type Collision struct {
	posX, posY float64
	showImpact bool
	ctype      CollisionType
}

// Game holds all level and object state.
type Game struct {
	controls Controls
	lives    int
	score    int

	numRows, numCols int
	bricks           [][]int // -1 == no brick, else 0..15
	bricksRemaining  int

	balls   []*Ball
	bat     *Bat
	bullets []*Bullet
	barrels []*Barrel
	impacts []*Impact

	levelNum     int
	portalActive bool
	portalFrame  int
	portalTimer  int

	assets *Assets
	audio  *Audio
}

func NewGame(controls Controls, lives int, assets *Assets, audio *Audio) *Game {
	g := &Game{
		controls: controls,
		lives:    lives,
		assets:   assets,
		audio:    audio,
	}
	g.newLevel(0)
	return g
}

// getMirroredLevel mirrors each row: the row plus its reverse minus the last char.
func getMirroredLevel(level []string) []string {
	out := make([]string, len(level))
	for i, row := range level {
		rev := make([]byte, 0, len(row))
		for j := len(row) - 2; j >= 0; j-- {
			rev = append(rev, row[j])
		}
		out[i] = row + string(rev)
	}
	return out
}

func (g *Game) newLevel(levelNum int) {
	g.playSound("start_game", 1)

	if levelNum >= len(LEVELS) {
		levelNum = 0
	}

	level := getMirroredLevel(LEVELS[levelNum])
	g.numRows = len(level)
	g.numCols = len(level[0])

	// Parse hex brick IDs (-1 for empty).
	g.bricks = make([][]int, g.numRows)
	g.bricksRemaining = 0
	for y := 0; y < g.numRows; y++ {
		g.bricks[y] = make([]int, g.numCols)
		for x := 0; x < g.numCols; x++ {
			c := level[y][x]
			if c == ' ' {
				g.bricks[y][x] = -1
			} else {
				g.bricks[y][x] = parseHexDigit(c)
			}
			// Count destructible bricks (everything except empty and ID 13).
			if g.bricks[y][x] != -1 && g.bricks[y][x] != 13 {
				g.bricksRemaining++
			}
		}
	}

	g.balls = []*Ball{NewDefaultBall()}
	g.bat = NewBat(g.controls)
	g.bullets = nil
	g.barrels = nil
	g.impacts = nil

	g.levelNum = levelNum
	g.portalActive = false
	g.portalFrame = 0
	g.portalTimer = 0
}

// collide checks whether a ball/bullet at (x, y) heading in dir would hit a wall
// or brick, damaging destructible bricks as a side effect.
func (g *Game) collide(x, y float64, dir Vec2, r float64) (Collision, bool) {
	dx, dy := dir.X, dir.Y

	if dx < 0 && x < LeftEdge+r {
		return Collision{LeftEdge, y, true, CollWall}, true
	}
	if dx > 0 && x > RightEdge-r {
		return Collision{RightEdge, y, true, CollWall}, true
	}
	if dy < 0 && y < TopEdge+r {
		return Collision{x, TopEdge, true, CollWall}, true
	}

	// Restrict the brick check to the cells the ball could overlap.
	x0 := maxInt(0, floorToInt((x-BricksXStart-r)/BrickWidth))
	y0 := maxInt(0, floorToInt((y-BricksYStart-r)/BrickHeight))
	x1 := minInt(g.numCols-1, floorToInt((x-BricksXStart+r)/BrickWidth))
	y1 := minInt(g.numRows-1, floorToInt((y-BricksYStart+r)/BrickHeight))

	for yb := y0; yb <= y1; yb++ {
		for xb := x0; xb <= x1; xb++ {
			if g.bricks[yb][xb] == -1 {
				continue
			}
			px, py, ok := brickCollide(x, y, xb, yb, r)
			if !ok {
				continue
			}
			centreX := float64(xb*BrickWidth + BricksXStart + BrickWidth/2)
			centreY := float64(yb*BrickHeight + BricksYStart + BrickHeight/2)
			ctype := CollBrick

			if g.bricks[yb][xb] >= 12 {
				// Brick 13 is indestructible; brick 12 takes one hit to become 11.
				if g.bricks[yb][xb] == 13 {
					ctype = CollIndestructibleBrick
				}
				g.impacts = append(g.impacts, NewImpact(centreX, centreY, 13))
				if g.bricks[yb][xb] == 12 {
					g.bricks[yb][xb] = 11
				}
			} else {
				g.impacts = append(g.impacts, NewImpact(centreX, centreY, g.bricks[yb][xb]))
				if randFloat() < PowerupChance {
					g.barrels = append(g.barrels, NewBarrel(g, centreX, centreY))
				}
				g.bricks[yb][xb] = -1
				g.bricksRemaining--
				if g.bricksRemaining == 0 {
					g.activatePortal()
				}
				g.score += 10
			}
			return Collision{px, py, false, ctype}, true
		}
	}

	return Collision{}, false
}

func (g *Game) activatePortal() {
	g.portalActive = true
	g.playSound("portal_exit", 1)
}

func (g *Game) Update() {
	// Update bat then balls.
	g.bat.Update(g)
	for _, b := range g.balls {
		b.Update(g)
	}

	// Remove balls that have fallen off the bottom.
	g.balls = filterBalls(g.balls, func(b *Ball) bool { return b.Y < Height })

	// Lose a life if no balls remain.
	if len(g.balls) == 0 {
		if g.lives > 0 || g.inDemoMode() {
			g.lives--
			g.balls = []*Ball{NewDefaultBall()}
			g.bat.changeType(BatNormal)
		}
		g.playSound("lose_life", 1)
	}

	// Update impacts, barrels and bullets (snapshot semantics via range).
	for _, o := range g.impacts {
		o.Update(g)
	}
	for _, o := range g.barrels {
		o.Update(g)
	}
	for _, o := range g.bullets {
		o.Update(g)
	}

	g.impacts = filterImpacts(g.impacts, func(o *Impact) bool { return o.time < 16 })
	g.barrels = filterBarrels(g.barrels, func(o *Barrel) bool { return o.Y < Height })
	g.bullets = filterBullets(g.bullets, func(o *Bullet) bool { return o.alive })

	// Advance the exit portal.
	if g.portalActive {
		if g.portalFrame < 3 {
			g.portalTimer--
			if g.portalTimer <= 0 {
				g.portalTimer = PortalAnimationSpeed
				g.portalFrame++
			}
		} else if g.bat.isPortalTransitionComplete(g) {
			g.newLevel(g.levelNum + 1)
		}
	}

	// If balls are stuck bouncing between indestructible bricks, soften them.
	if g.detectStuckBalls() {
		changedAny := false
		for row := 0; row < g.numRows; row++ {
			for col := 0; col < g.numCols; col++ {
				if g.bricks[row][col] == 13 {
					g.bricks[row][col] = 12
					changedAny = true
				}
			}
		}
		if changedAny {
			g.playSound("bat_small", 1)
		}
		if len(g.balls) > 0 {
			g.balls[0].timeSinceTouchedBat = 0
		}
	}
}

func (g *Game) detectStuckBalls() bool {
	if len(g.balls) == 0 {
		return false
	}
	for _, ball := range g.balls {
		if ball.timeSinceDamagedBrick < 30*60 || ball.timeSinceTouchedBat < 30*60 {
			return false
		}
	}
	return true
}

func (g *Game) Draw() {
	g.assets.Blit("arena"+itoa(g.levelNum%len(LEVELS)), 0, 0)

	// Exit portal and (unused) enemy doors.
	g.assets.Blit("portal_exit"+itoa(g.portalFrame), Width-70-20, Height-70)
	g.assets.Blit("portal_meanie00", 110, 40)
	g.assets.Blit("portal_meanie10", 440, 40)

	// Clip so shadows don't spill onto the walls.
	g.assets.SetClip(20, 42, 600, 598)

	// Brick shadows.
	for y := 0; y < g.numRows; y++ {
		for x := 0; x < g.numCols; x++ {
			if g.bricks[y][x] != -1 {
				g.assets.Blit("bricks",
					float64(x*BrickWidth+BricksXStart+ShadowOffset),
					float64(y*BrickHeight+BricksYStart+ShadowOffset))
			}
		}
	}

	// Object shadows: barrels, balls, bat.
	for _, o := range g.barrels {
		o.shadow.draw(g.assets)
	}
	for _, o := range g.balls {
		o.shadow.draw(g.assets)
	}
	g.bat.shadow.draw(g.assets)

	// Bricks.
	for y := 0; y < g.numRows; y++ {
		for x := 0; x < g.numCols; x++ {
			if g.bricks[y][x] != -1 {
				g.assets.Blit("brick"+hexDigit(g.bricks[y][x]),
					float64(x*BrickWidth+BricksXStart),
					float64(y*BrickHeight+BricksYStart))
			}
		}
	}

	// Balls, bat, barrels, bullets.
	for _, o := range g.balls {
		o.Draw(g)
	}
	g.bat.draw(g.assets)
	for _, o := range g.barrels {
		o.Draw(g)
	}
	for _, o := range g.bullets {
		o.Draw(g)
	}

	g.assets.ClearClip()

	// Impacts draw on top, outside the clip.
	for _, o := range g.impacts {
		o.Draw(g)
	}

	if !g.inDemoMode() {
		g.drawScore()
		g.drawLives()
	}
}

func (g *Game) drawScore() {
	x := 15.0
	for _, digit := range itoa(g.score) {
		g.assets.Blit("digit"+string(digit), x, 50)
		x += 55
	}
}

func (g *Game) drawLives() {
	x := 0.0
	for i := 0; i < g.lives; i++ {
		g.assets.Blit("life", x, Height-20)
		x += 50
	}
}

func (g *Game) changeAllBallSpeeds(change int) {
	for _, b := range g.balls {
		b.speed = minInt(maxInt(b.speed+change, BallMinSpeed), BallMaxSpeed)
	}
}

func (g *Game) inDemoMode() bool { return g.controls.isAI() }

// playSound plays a sound, but never in demo/AI mode (the title screen).
func (g *Game) playSound(name string, count int) {
	if g.inDemoMode() {
		return
	}
	g.audio.PlaySound(name, count)
}

// Typed slice filters.
func filterBalls(s []*Ball, keep func(*Ball) bool) []*Ball {
	var out []*Ball
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterImpacts(s []*Impact, keep func(*Impact) bool) []*Impact {
	var out []*Impact
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterBarrels(s []*Barrel, keep func(*Barrel) bool) []*Barrel {
	var out []*Barrel
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
func filterBullets(s []*Bullet, keep func(*Bullet) bool) []*Bullet {
	var out []*Bullet
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}
