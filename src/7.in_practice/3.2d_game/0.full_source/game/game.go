package game

import ()

const (
	GameActive int = iota
	GameMenu
	GameWin
)

type Game struct {
	GameState int
	Keys      []bool
	Width     int
	Height    int
}

func New(width, height int) *Game {
	return &Game{GameActive, make([]bool, 1024), width, height}
}

func (g *Game) Init() {

}

func (g *Game) ProcessInput(dt float64) {

}

func (g *Game) Update(dt float64) {

}

func (g *Game) Render() {

}
