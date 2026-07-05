package main

// Impact is a short-lived animation played on ball/wall/brick/barrel events.
type Impact struct {
	Actor
	itype int
	time  int
}

func NewImpact(x, y float64, itype int) *Impact {
	im := &Impact{itype: itype}
	im.Actor = newActor("blank", x, y, AnchorCentre)
	return im
}

func (im *Impact) Update(g *Game) {
	// Sprite names look like "impactc0": hex type digit + frame (time/4).
	im.Image = "impact" + hexDigit(im.itype) + itoa(im.time/4)
	im.time++
}

func (im *Impact) Draw(g *Game) { im.draw(g.assets) }
