package main

import (
	"flag"
	"fmt"

	"github.com/chrplr/pgzgo"
)

type State int

const (
	StateTitle State = iota
	StatePlay
	StateGameOver
)

var (
	state       State
	game        *Game
	totalFrames int

	assets           *Assets
	audio            *Audio
	keyboardControls *KeyboardControls
	aiControls       *AIControls
	joystickControls *JoystickControls
)

// startControls returns the human controller that just pressed fire (keyboard
// takes priority over the gamepad), or nil if neither did.
func startControls() Controls {
	if keyboardControls.firePressed() {
		return keyboardControls
	}
	if joystickControls.firePressed() {
		return joystickControls
	}
	return nil
}

func update() {
	totalFrames++

	keyboardControls.update()
	joystickControls.update()

	switch state {
	case StateTitle:
		aiControls.update()
		game.Update()
		if c := startControls(); c != nil {
			game = NewGame(c, 3, assets, audio)
			state = StatePlay
			audio.StopMusic()
		}

	case StatePlay:
		if game.lives > 0 {
			game.Update()
		} else {
			game.playSound("game_over", 1)
			state = StateGameOver
		}

	case StateGameOver:
		if startControls() != nil {
			game = NewGame(aiControls, 3, assets, audio)
			state = StateTitle
			audio.PlayMusic("title_theme", 0.3)
		}
	}
}

func draw() {
	game.Draw()

	switch state {
	case StateTitle:
		assets.Blit("title", 0, 0)
		assets.Blit("startgame", 20, 80)
		assets.Blit("start"+itoa((totalFrames/4)%13), Width/2-125, 530)

	case StateGameOver:
		assets.Blit("gameover"+itoa((totalFrames/4)%15), Width/2-225, 450)
	}
}

func main() {
	selftest := flag.Bool("selftest", false, "run headlessly across levels, then exit")
	flag.Parse()

	a, err := pgzgo.New(pgzgo.Config{
		Title:  "Kinetix",
		Width:  Width,
		Height: Height,
		Images: imagesFS,
		Audio:  audioFS,
	})
	if err != nil {
		panic(err)
	}
	defer a.Close()

	// Publish the harness and its sub-systems to the game globals.
	app = a
	assets = a.Screen
	audio = a.Audio

	keyboardControls = &KeyboardControls{}
	aiControls = &AIControls{}
	joystickControls = &JoystickControls{}

	if *selftest {
		runSelftest()
		return
	}

	audio.PlayMusic("title_theme", 0.3)

	state = StateTitle
	game = NewGame(aiControls, 3, assets, audio)

	// The harness runs the loop: event polling, keyboard/gamepad snapshots,
	// Escape/Start/window-close quit, clear and present all happen inside.
	a.Loop(
		func(*pgzgo.App) { update() },
		func(*pgzgo.App) { draw() },
	)
}

// runSelftest drives the game headlessly to exercise the logic without a display.
func runSelftest() {
	// A keyboard-controlled game so sounds/scoring paths run (audio is a no-op
	// under the dummy driver). Demo AI drives the bat.
	game = NewGame(aiControls, 3, assets, audio)

	// Free-running demo phase: ball physics, bat AI, brick collisions, barrels.
	for i := 0; i < 4000; i++ {
		aiControls.update()
		game.Update()
	}
	fmt.Printf("after demo: level %d, score %d, %d balls, %d barrels, %d impacts, bricks left %d\n",
		game.levelNum, game.score, len(game.balls), len(game.barrels), len(game.impacts), game.bricksRemaining)

	// Exercise every powerup and a real (non-demo) game with each level.
	pg := NewGame(keyboardControls, 3, assets, audio)
	for lvl := 0; lvl < len(LEVELS); lvl++ {
		pg.newLevel(lvl)
		// Apply one of each powerup to hit bat/ball/portal code paths.
		for _, pu := range []int{PowerupExtendBat, PowerupGun, PowerupMagnet, PowerupSmallBat,
			PowerupMultiBall, PowerupFastBalls, PowerupSlowBalls, PowerupExtraLife} {
			b := NewBarrel(pg, pg.bat.X, BatTopEdge-5)
			b.btype = pu
			pg.barrels = append(pg.barrels, b)
		}
		for i := 0; i < 300; i++ {
			pg.Update()
		}
		fmt.Printf("level %d: %dx%d grid, bricks left %d, %d balls, score %d, lives %d\n",
			lvl, pg.numRows, pg.numCols, pg.bricksRemaining, len(pg.balls), pg.score, pg.lives)
	}

	fmt.Println("SELFTEST OK")
}
