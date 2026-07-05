package main

// Controls abstracts player input (keyboard) and the AI/demo driver.
type Controls interface {
	update()
	getX() float64
	fireDown() bool
	firePressed() bool
	isAI() bool
}

// controlBase provides the shared fire-edge detection.
type controlBase struct {
	firePreviouslyDown bool
	isFirePressed      bool
}

func (c *controlBase) updateFire(fireDown bool) {
	c.isFirePressed = fireDown && !c.firePreviouslyDown
	c.firePreviouslyDown = fireDown
}

func (c *controlBase) firePressed() bool { return c.isFirePressed }

// KeyboardControls reads the arrow keys and space bar.
type KeyboardControls struct {
	controlBase
}

func (c *KeyboardControls) update()    { c.updateFire(c.fireDown()) }
func (c *KeyboardControls) isAI() bool { return false }

func (c *KeyboardControls) getX() float64 {
	if keyLeft() {
		return -BatSpeed
	} else if keyRight() {
		return BatSpeed
	}
	return 0
}

func (c *KeyboardControls) fireDown() bool { return keySpace() || keyEnter() }

// AIControls drives the bat automatically for the title-screen demo.
type AIControls struct {
	controlBase
	offset float64
}

func (c *AIControls) update()    { c.updateFire(c.fireDown()) }
func (c *AIControls) isAI() bool { return true }

func (c *AIControls) getX() float64 {
	if game.portalActive {
		// Portal is open: head right through it.
		return BatSpeed
	}
	// Drift the aim offset so the AI doesn't hit dead-centre every time.
	c.offset += float64(randInt(-1, 1))
	c.offset = minf(maxf(-40, c.offset), 40)
	if len(game.balls) == 0 {
		return 0
	}
	return minf(BatSpeed, maxf(-BatSpeed, game.balls[0].X-(game.bat.X+c.offset)))
}

func (c *AIControls) fireDown() bool { return randInt(0, 5) == 0 }

// JoystickControls reads a connected gamepad: the d-pad or left analogue stick
// moves the bat, the South (A) button launches/fires. Mirrors the Python
// JoystickControls (d-pad first, else analogue with a 0.2 dead-zone).
type JoystickControls struct {
	controlBase
}

func (c *JoystickControls) update()    { c.updateFire(c.fireDown()) }
func (c *JoystickControls) isAI() bool { return false }

func (c *JoystickControls) getX() float64 {
	if padLeft() {
		return -BatSpeed
	}
	if padRight() {
		return BatSpeed
	}
	ax := padAxisX()
	if absf(ax) < 0.2 {
		return 0
	}
	return ax * BatSpeed
}

func (c *JoystickControls) fireDown() bool { return padButton0() }
