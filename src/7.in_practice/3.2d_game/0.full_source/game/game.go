package game

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/ballObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameLevel"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
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

var (
	Player         *gameObject.GameObject
	PlayerSize     mgl32.Vec2 = mgl32.Vec2{100.0, 20.0}
	PlayerVelocity float32    = 500.0

	Ball                *ballObject.BallObject
	InitialBallVelocity mgl32.Vec2 = mgl32.Vec2{100.0, -350.0}
	BallRadius          float32    = 12.5
)

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
	resourceManager.LoadTexture(textureDir+"background.jpg", "background")
	resourceManager.LoadTexture(textureDir+"awesomeface.png", "face")
	resourceManager.LoadTexture(textureDir+"block.png", "block")
	resourceManager.LoadTexture(textureDir+"block_solid.png", "block_solid")
	resourceManager.LoadTexture(textureDir+"paddle.png", "paddle")

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
	playerPos := mgl32.Vec2{float32(g.Width)/2.0 - PlayerSize[0]/2.0,
		float32(g.Height) - PlayerSize[1]}
	Player = gameObject.New(playerPos, PlayerSize,
		resourceManager.Textures["paddle"], mgl32.Vec3{1.0, 1.0, 1.0},
		mgl32.Vec2{0.0, 0.0})

	ballPos := playerPos.Add(mgl32.Vec2{
		PlayerSize[0]/2.0 - BallRadius, -BallRadius * 2.0})
	Ball = ballObject.New(ballPos, BallRadius, InitialBallVelocity,
		resourceManager.Textures["face"])
}

func (g *Game) ProcessInput(dt float64) {
	if g.State == GameActive {
		velocity := PlayerVelocity * float32(dt)

		if g.Keys[glfw.KeyA] {
			if Player.Position[0] >= 0.0 {
				Player.Position[0] -= velocity
				if Ball.Stuck {
					Ball.Object.Position[0] -= velocity
				}
			}
		}
		if g.Keys[glfw.KeyD] {
			if Player.Position[0] <= (float32(g.Width) - Player.Size[0]) {
				Player.Position[0] += velocity
				if Ball.Stuck {
					Ball.Object.Position[0] += velocity
				}
			}
		}

		if g.Keys[glfw.KeySpace] {
			Ball.Stuck = false
		}
	}
}

func (g *Game) Update(dt float64) {
	Ball.Move(float32(dt), uint32(g.Width))
	g.DoCollisions()
}

func (g *Game) Render() {
	if g.State == GameActive {
		renderer.DrawSprite(resourceManager.Textures["background"],
			mgl32.Vec2{0.0, 0.0}, mgl32.Vec2{float32(g.Width), float32(g.Height)},
			0.0, mgl32.Vec3{1.0, 1.0, 1.0})

		g.Levels[g.Level].Draw(renderer)
		Player.Draw(renderer)

		Ball.Object.Draw(renderer)
	}
}

func (g *Game) DoCollisions() {
	for _, box := range g.Levels[g.Level].Bricks {
		if !box.Destroyed {
			if CheckCollisionBall(Ball, box) {
				//if CheckCollision(Ball.Object, box) {
				if !box.IsSolid {
					box.Destroyed = true
				}
			}
		}
	}
}

func CheckCollision(one, two *gameObject.GameObject) bool {
	// Collision x-axis?
	collisionX := (one.Position[0]+one.Size[0]) >= two.Position[0] &&
		(two.Position[0]+two.Size[0]) >= one.Position[0]
	// Collision y-axis?
	collisionY := (one.Position[1]+one.Size[1]) >= two.Position[1] &&
		(two.Position[1]+two.Size[1]) >= one.Position[1]
	// Collision only if both sides
	return collisionX && collisionY
}

func CheckCollisionBall(one *ballObject.BallObject, two *gameObject.GameObject) bool {
	center := one.Object.Position.Add(mgl32.Vec2{one.Radius, one.Radius})
	// Calculate AABB info (center, half-extents)
	aabbHalfExtents := mgl32.Vec2{two.Size[0] / 2.0, two.Size[1] / 2.0}
	aabbCenter := aabbHalfExtents.Add(two.Position)

	// Get difference vector between both centers
	difference := center.Sub(aabbCenter)
	clamped := mgl32.Vec2{
		mgl32.Clamp(difference[0], -aabbHalfExtents[0], aabbHalfExtents[0]),
		mgl32.Clamp(difference[1], -aabbHalfExtents[1], aabbHalfExtents[1]),
	}

	// Add clamped value to AABB_center and we get the value of box closet to circle
	closest := aabbCenter.Add(clamped)
	difference = closest.Sub(center)

	return difference.Len() < one.Radius
}
