package main

// This file is the only glue between the game and the pgzgo harness. It owns the
// embedded assets (the //go:embed directives must live in this package, since
// embed can only reach files under the importing package's directory), names the
// harness types the way the game code refers to them, and adapts the game's
// input helpers onto the harness keyboard/gamepad.

import (
	"embed"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

// Assets and Audio are the game's names for the harness drawing surface and
// mixer, so the rest of the game code reads naturally (assets.Blit, audio.Play…).
type Assets = pgzgo.Screen
type Audio = pgzgo.Audio

//go:embed images
var imagesFS embed.FS

//go:embed sounds music
var audioFS embed.FS

// app is the running harness; the input wrappers below read from its per-frame
// keyboard and gamepad snapshots.
var app *pgzgo.App

// Keyboard bindings used by KeyboardControls.
func keyLeft() bool  { return app.Keyboard.Held(sdl.SCANCODE_LEFT) }
func keyRight() bool { return app.Keyboard.Held(sdl.SCANCODE_RIGHT) }
func keySpace() bool { return app.Keyboard.Held(sdl.SCANCODE_SPACE) }

// Gamepad bindings used by JoystickControls.
func padLeft() bool     { return app.Gamepad.Left() }
func padRight() bool    { return app.Gamepad.Right() }
func padAxisX() float64 { return app.Gamepad.AxisX() }
func padButton0() bool  { return app.Gamepad.Button0() }
