package main

const (
	Width  = 640
	Height = 640

	BatSpeed = 8

	BatMinX = 35
	BatMaxX = 605

	TopEdge   = 50
	RightEdge = 617
	LeftEdge  = 23

	BatTopEdge = 590

	BallInitialOffset = 10

	BallStartSpeed = 5
	BallMinSpeed   = 4
	BallMaxSpeed   = 11

	BallSpeedUpInterval     = 10 * 60 // 10 seconds at 60fps
	BallSpeedUpIntervalFast = 15 * 60
	BallFastSpeedThreshold  = 7

	BallRadius = 7

	BulletSpeed = 8

	BricksXStart = 20
	BricksYStart = 100

	BrickWidth   = 40
	BrickHeight  = 20
	ShadowOffset = 10

	PowerupChance = 0.2

	FireInterval = 30

	PortalAnimationSpeed = 5
)

// LEVELS holds the left half of each level; each level is mirrored horizontally
// at load time. Characters are hex brick IDs; space means no brick.
var LEVELS = [][]string{
	{
		"        ",
		"        ",
		"        ",
		"     a  ",
		"    a7a ",
		"     a  ",
		"     a55",
		"    444 ",
		"   333a ",
		"  222a  ",
		" 111a   ",
		"   11aa ",
		"    111 ",
		"    6   ",
		"     6  ",
	},
	{
		"        ",
		"        ",
		"    3   ",
		"    3   ",
		"    3   ",
		"    3000",
		"    3000",
		"   53000",
		"   53000",
		"  35a555",
		" 3 5aa55",
		"3  5aaa5",
		"  355555",
		"  333333",
		"   333  ",
		"    33  ",
		"     3  ",
	},
	{
		"   7    ",
		"  77    ",
		" 7777   ",
		" 7777   ",
		" 77777  ",
		" 77777  ",
		" 77 777 ",
		" 7  7777",
		" 7   717",
		"     777",
		"      77",
		"      7 ",
		"     c7 ",
		"      c ",
		"      c ",
	},
	{
		"   03   ",
		"   30   ",
		"    03  ",
		"    30  ",
		"     0  ",
		" 8   0  ",
		" 88 8033",
		"  883333",
		"   8333d",
		"   33733",
		"  33373d",
		" 3333333",
		" 3c 333d",
		" cc 3333",
		" c   3 3",
		"     3 3",
		"    3 3 ",
		"    c 3 ",
		"    cc3c",
		"    cccc",
		"      d ",
	},
	{
		"5   9  0",
		"0   4  3",
		"08  4  4",
		"53  47 2",
		" 39 92 1",
		" 84  2  ",
		"  47 26 ",
		"5 92 71 ",
		"08 26 1 ",
		"53971 1 ",
		" 8471c6 ",
		"  926acc",
		"   71aad",
		"039 6aac",
		"dc421ac ",
		"  dccc  ",
		"    d   ",
	},
	{
		"  dccccd",
		"  c89765",
		"  c34210",
		"  c34210",
		"  c34210",
		"  c34210",
		"  c3421d",
		"  c34210",
		"  c34210",
		"  c34210",
		"  c34210",
		"  c89765",
		"  dccccd",
	},
}

// Powerup types (values match the barrel sprite filenames).
const (
	PowerupExtendBat = 0
	PowerupGun       = 1
	PowerupSmallBat  = 2
	PowerupMagnet    = 3
	PowerupMultiBall = 4
	PowerupFastBalls = 5
	PowerupSlowBalls = 6
	PowerupPortal    = 7
	PowerupExtraLife = 8
)

// Bat types.
const (
	BatNormal   = 0
	BatMagnet   = 1
	BatGun      = 2
	BatExtended = 3
	BatSmall    = 4
)

// Collision types.
type CollisionType int

const (
	CollWall CollisionType = iota
	CollBat
	CollBatEdge
	CollBrick
	CollIndestructibleBrick
)

// powerupBatType maps a powerup to the bat type it grants (-1 = none).
var powerupBatType = map[int]int{
	PowerupExtendBat: BatExtended,
	PowerupGun:       BatGun,
	PowerupSmallBat:  BatSmall,
	PowerupMagnet:    BatMagnet,
}

// powerupSound maps a powerup to its collection sound ("" = none).
var powerupSound = map[int]string{
	PowerupExtendBat: "bat_extend",
	PowerupGun:       "bat_gun",
	PowerupMagnet:    "magnet",
	PowerupSmallBat:  "bat_small",
	PowerupExtraLife: "extra_life",
	PowerupFastBalls: "speed_up",
	PowerupSlowBalls: "powerup",
	PowerupMultiBall: "multiball",
}
