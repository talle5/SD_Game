package game

import rl "github.com/gen2brain/raylib-go/raylib"

type Player struct {
	X, Y, W, H int32
	Name       string
	Color      rl.Color
	Input      func(p *Player)
	Direction  Direction
}

type Direction int32

const (
	Stop  Direction = 0
	Up    Direction = 1
	Down  Direction = 2
	Left  Direction = 3
	Right Direction = 4
)

func (p *Player) Draw() {
	rl.DrawRectangle(p.X, p.Y, p.W, p.H, p.Color)
	if p.Name != "" {
		rl.DrawText(p.Name, p.X, p.Y-18, 14, rl.White)
	}
}

func (p *Player) Update() {
	if p.Input != nil {
		p.Input(p)
	}

	switch p.Direction {
	case Up:
		p.Y -= 5
	case Down:
		p.Y += 5
	case Left:
		p.X -= 5
	case Right:
		p.X += 5
	}

	if p.X < 0 {
		p.X = 0
		p.Direction = Stop
	}
	if p.Y < 0 {
		p.Y = 0
		p.Direction = Stop
	}
	if p.X > 750 {
		p.X = 750
		p.Direction = Stop
	}
	if p.Y > 550 {
		p.Y = 550
		p.Direction = Stop
	}
}
