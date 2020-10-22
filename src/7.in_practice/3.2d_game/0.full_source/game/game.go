package game

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameLevel"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/resourceManager"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/spriteRenderer"
)

const (
	GameActive int = iota
	GameMenu
	GameWin
)

type Game struct {
	State  int
	Keys   []bool
	Width  int
	Height int
	Levels []*gameLevel.GameLevel
	Level  int
}

var textureDir string = "../../../../resources/textures/"
var levelDir string = "../../../../resources/levels/"
var renderer *spriteRenderer.SpriteRenderer

func New(width, height int) *Game {
	return &Game{GameActive, make([]bool, 1024), width, height, nil, 0}
}

func (g *Game) Init() {
	// Load shaders
	resourceManager.LoadShader("shaders/sprite.vs", "shaders/sprite.fs", "sprite")

	// Configure shaders
	projection := mgl32.Ortho(0.0, float32(g.Width), float32(g.Height), 0.0,
		-1.0, 1.0)
	resourceManager.Shaders["sprite"].SetInteger("image", 0, true)
	resourceManager.Shaders["sprite"].SetMatrix4("projection", projection, false)

	// Set render-specific controls
	renderer = spriteRenderer.New(resourceManager.Shaders["sprite"])

	// Load textures
	//resourceManager.LoadTexture(textureDir+"ba.png", "face")
	resourceManager.LoadTexture(textureDir+"awesomeface.png", "face")
	resourceManager.LoadTexture(textureDir+"block.png", "block")
	resourceManager.LoadTexture(textureDir+"block_solid.png", "block_solid")

	// Load levels
	w := uint32(g.Width)
	h := uint32(g.Height / 2)
	g.Levels = make([]*gameLevel.GameLevel, 4)
	for i, lName := range []string{"one", "two", "three", "four"} {
		g.Levels[i] = &gameLevel.GameLevel{}
		g.Levels[i].Load(levelDir+lName+".lvl", w, h)
	}
	g.Level = 0

	// Configure game objects
}

func (g *Game) ProcessInput(dt float64) {

}

func (g *Game) Update(dt float64) {

}

func (g *Game) Render() {

	//renderer.DrawSprite(resourceManager.Textures["face"], mgl32.Vec2{200.0, 200.0},
	//	mgl32.Vec2{300.0, 400.0}, 45.0, mgl32.Vec3{0.0, 1.0, 0.0})

	if g.State == GameActive {
		//renderer.DrawSprite(resourceManager.Textures["background"], mgl32.Vec2
		g.Levels[g.Level].Draw(renderer)
	}
}
