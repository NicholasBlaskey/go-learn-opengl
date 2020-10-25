package game

import (
	"fmt"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/nicholasblaskey/glfont"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/audio"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/ballObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameLevel"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/particle"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/postProcessor"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/powerUp"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/resourceManager"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/spriteRenderer"
)

const (
	GameActive int = iota
	GameMenu
	GameWin
)

const (
	Up int = iota
	Right
	Down
	Left
)

type Game struct {
	State         int
	Keys          []bool
	KeysProcessed []bool
	Width         int
	Height        int
	Levels        []*gameLevel.GameLevel
	Level         int
	PowerUps      []*powerUp.PowerUp
	Lives         int
}

var (
	Player         *gameObject.GameObject
	PlayerSize     mgl32.Vec2 = mgl32.Vec2{100.0, 20.0}
	PlayerVelocity float32    = 500.0

	Ball                *ballObject.BallObject
	InitialBallVelocity mgl32.Vec2 = mgl32.Vec2{100.0, -350.0}
	BallRadius          float32    = 12.5

	ShakeTime float32 = 0.0
)

var textureDir string = "../../../../resources/textures/"
var levelDir string = "../../../../resources/levels/"
var audioDir string = "../../../../resources/audio/"
var fontDir string = "../../../../resources/fonts/"
var renderer *spriteRenderer.SpriteRenderer
var Particles *particle.Generator
var Effects *postProcessor.PostProcessor
var AudioPlayer *audio.Player
var Text *glfont.Font

func New(width, height int) *Game {
	return &Game{GameMenu, make([]bool, 1024), make([]bool, 1024),
		width, height, nil, 0, nil, 3}
}

func (g *Game) Init() {
	// Load shaders
	resourceManager.LoadShader("shaders/sprite.vs", "shaders/sprite.fs", "sprite")
	resourceManager.LoadShader("shaders/particle.vs", "shaders/particle.fs", "particle")
	resourceManager.LoadShader("shaders/postProcessing.vs",
		"shaders/postProcessing.fs", "postprocessing")

	// Configure shaders
	projection := mgl32.Ortho(0.0, float32(g.Width), float32(g.Height), 0.0,
		-1.0, 1.0)
	resourceManager.Shaders["sprite"].SetInteger("image", 0, true)
	resourceManager.Shaders["sprite"].SetMatrix4("projection", projection, false)
	resourceManager.Shaders["particle"].SetInteger("sprite", 0, true)
	resourceManager.Shaders["particle"].SetMatrix4("projection", projection, false)

	// Load textures
	resourceManager.LoadTexture(textureDir+"background.jpg", "background")
	resourceManager.LoadTexture(textureDir+"awesomeface.png", "face")
	resourceManager.LoadTexture(textureDir+"block.png", "block")
	resourceManager.LoadTexture(textureDir+"block_solid.png", "block_solid")
	resourceManager.LoadTexture(textureDir+"paddle.png", "paddle")
	resourceManager.LoadTexture(textureDir+"particle.png", "particle")

	resourceManager.LoadTexture(textureDir+"powerup_speed.png", "powerup_speed")
	resourceManager.LoadTexture(textureDir+"powerup_sticky.png", "powerup_sticky")
	resourceManager.LoadTexture(textureDir+"powerup_increase.png", "powerup_increase")
	resourceManager.LoadTexture(textureDir+"powerup_confuse.png", "powerup_confuse")
	resourceManager.LoadTexture(textureDir+"powerup_chaos.png", "powerup_chaos")
	resourceManager.LoadTexture(textureDir+"powerup_passthrough.png", "powerup_passthrough")

	// Set render-specific controls
	renderer = spriteRenderer.New(resourceManager.Shaders["sprite"])
	Particles = particle.NewGenerator(resourceManager.Shaders["particle"],
		resourceManager.Textures["particle"], 500)
	Effects = postProcessor.New(resourceManager.Shaders["postprocessing"],
		int32(g.Width), int32(g.Height))
	var err error
	Text, err = glfont.LoadFont(fontDir+"OCRAEXT.TTF", int32(24), g.Width, g.Height)
	if err != nil {
		panic(err)
	}
	Text.SetColor(1.0, 1.0, 1.0, 1.0)

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

	AudioPlayer = audio.New()
	go AudioPlayer.Play(audioDir+"breakout.mp3", true)
	//go AudioPlayer.Play(audioDir+"breakout.mp3", false)
	//go AudioPlayer.Play(audioDir+"bleep.wav", true)
}

func (g *Game) ProcessInput(dt float64) {
	if g.State == GameMenu {
		if g.Keys[glfw.KeyEnter] && !g.KeysProcessed[glfw.KeyEnter] {
			g.State = GameActive
			g.KeysProcessed[glfw.KeyEnter] = true
		}
		if g.Keys[glfw.KeyW] && !g.KeysProcessed[glfw.KeyW] {
			g.Level = (g.Level + 1) % 4
			g.KeysProcessed[glfw.KeyW] = true
		}
		if g.Keys[glfw.KeyS] && !g.KeysProcessed[glfw.KeyS] {
			g.Level -= 1
			if g.Level < 0 {
				g.Level = 3
			}
			g.KeysProcessed[glfw.KeyS] = true
		}
	}

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
	} else if g.State == GameWin {
		if g.Keys[glfw.KeyEnter] {
			g.KeysProcessed[glfw.KeyEnter] = true
			Effects.Chaos = false
			g.State = GameMenu
		}
	}
}

func (g *Game) Update(dt float64) {
	Ball.Move(float32(dt), uint32(g.Width))
	g.DoCollisions()

	Particles.Update(float32(dt), Ball.Object, 2,
		mgl32.Vec2{Ball.Radius / 2.0, Ball.Radius / 2.0})

	g.UpdatePowerUps(float32(dt))

	// Reduce shake time
	if ShakeTime > 0.0 {
		ShakeTime -= float32(dt)
		if ShakeTime <= 0 {
			Effects.Shake = false
		}
	}

	// Check loss condition
	if Ball.Object.Position[1] >= float32(g.Height) {
		g.Lives -= 1
		if g.Lives == 0 {
			g.ResetLevel()
			g.State = GameMenu
		}
		g.ResetPlayer()
	}

	// Check win condition
	if g.State == GameActive && g.Levels[g.Level].IsCompleted() {
		g.ResetLevel()
		g.ResetPlayer()
		Effects.Chaos = true
		g.State = GameWin
	}
}

func (g *Game) Render() {
	if g.State == GameActive || g.State == GameMenu || g.State == GameWin {
		Effects.BeginRender()

		renderer.DrawSprite(resourceManager.Textures["background"],
			mgl32.Vec2{0.0, 0.0}, mgl32.Vec2{float32(g.Width), float32(g.Height)},
			0.0, mgl32.Vec3{1.0, 1.0, 1.0})

		g.Levels[g.Level].Draw(renderer)
		Player.Draw(renderer)
		for _, p := range g.PowerUps {
			if !p.Object.Destroyed {
				p.Object.Draw(renderer)
			}
		}
		Particles.Draw()
		Ball.Object.Draw(renderer)

		Effects.EndRender()
		Effects.Render(float32(glfw.GetTime()))
		Text.Printf(5.0, 20.0, 1.0, fmt.Sprintf("Lives: %d", g.Lives))
	}
	if g.State == GameMenu {
		Text.Printf(250.0, float32(g.Height)/2.0, 1.0, "Press enter to start")
		Text.Printf(245.0, float32(g.Height)/2.0+20.0, 0.75, "Press W or S to select a level")
	} else if g.State == GameWin {
		Text.SetColor(0.0, 1.0, 0.0, 1.0)
		Text.Printf(320.0, float32(g.Height)/2.0-20, 1.0, "You Won!!")
		Text.SetColor(1.0, 1.0, 1.0, 1.0)
		Text.Printf(130.0, float32(g.Height)/2.0, 1.0, "Press ENTER to retry or ESC to quit")
	}
}

func (g *Game) ResetLevel() {
	lName := []string{"one", "two", "three", "four"}[g.Level]
	g.Levels[g.Level].Load(levelDir+lName+".lvl",
		uint32(g.Width), uint32(g.Height/2))
	g.Lives = 3
}

func (g *Game) ResetPlayer() {
	// Reset player / ball states
	Player.Size = PlayerSize
	Player.Position = mgl32.Vec2{
		float32(g.Width)/2.0 - PlayerSize[0]/2.0,
		float32(g.Height) - PlayerSize[1],
	}
	Effects.Chaos = false
	Effects.Confuse = false
	Ball.Passthrough = false
	Ball.Sticky = false
	Ball.Reset(Player.Position.Add(
		mgl32.Vec2{PlayerSize[0]/2.0 - BallRadius, -(BallRadius * 2.0)}),
		InitialBallVelocity)
}

func (g *Game) DoCollisions() {
	for _, box := range g.Levels[g.Level].Bricks {
		if !box.Destroyed {
			hit, dir, diffVector := CheckCollisionBall(Ball, box)
			if hit {
				if !box.IsSolid {
					box.Destroyed = true
					g.SpawnPowerUps(box)
					go AudioPlayer.Play(audioDir+"bleep.mp3", false)
				} else {
					ShakeTime = 0.05
					Effects.Shake = true
					go AudioPlayer.Play(audioDir+"solid.wav", false)
				}

				// Collision resolution
				if !(Ball.Passthrough && !box.IsSolid) {
					if dir == Left || dir == Right { // Horizontal collison
						// Reverse horizontal velocity
						Ball.Object.Velocity[0] = -Ball.Object.Velocity[0]
						// Relocate
						penetration := Ball.Radius - mgl32.Abs(diffVector[0])
						if dir == Left {
							Ball.Object.Position[0] += penetration
						} else {
							Ball.Object.Position[0] -= penetration
						}
					} else { // Vertical collision
						// Reverse vertical velocity
						Ball.Object.Velocity[1] = -Ball.Object.Velocity[1]
						// Relocate
						penetration := Ball.Radius - mgl32.Abs(diffVector[1])
						if dir == Up {
							Ball.Object.Position[1] -= penetration
						} else {
							Ball.Object.Position[1] += penetration
						}
					}
				}
			}
		}
	}

	for _, powerUp := range g.PowerUps {
		if !powerUp.Object.Destroyed {
			if powerUp.Object.Position[1] >= float32(g.Height) {
				powerUp.Object.Destroyed = true
			}
			if CheckCollision(Player, powerUp.Object) {
				ActivatePowerUp(powerUp)
				powerUp.Object.Destroyed = true
				powerUp.Activated = true
				go AudioPlayer.Play(audioDir+"powerup.wav", false)
			}
		}
	}

	hit, _, _ := CheckCollisionBall(Ball, Player)
	if !Ball.Stuck && hit {
		// Check where it hit the board
		centerBoard := Player.Position[0] + Player.Size[0]/2.0
		distance := (Ball.Object.Position[0] + Ball.Radius) - centerBoard
		percentage := distance / (Player.Size[0] / 2.0)
		// Chage velocity according to where the ball was hit
		strength := float32(2.0)
		oldVelocity := Ball.Object.Velocity
		lenOldVelocity := oldVelocity.Len()
		Ball.Object.Velocity[0] = InitialBallVelocity[0] * percentage * strength

		Ball.Object.Velocity[1] = -1 * mgl32.Abs(Ball.Object.Velocity[1])
		Ball.Object.Velocity = Ball.Object.Velocity.Normalize().Mul(lenOldVelocity)

		Ball.Stuck = Ball.Sticky
		go AudioPlayer.Play(audioDir+"bleep.wav", false)
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

func CheckCollisionBall(one *ballObject.BallObject,
	two *gameObject.GameObject) (bool, int, mgl32.Vec2) {

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

	if difference.Len() < one.Radius {
		return true, VectorDirection(difference), difference
	}
	return false, Up, mgl32.Vec2{0, 0}
}

func VectorDirection(target mgl32.Vec2) int {
	compass := []mgl32.Vec2{
		mgl32.Vec2{0.0, 1.0},  // Up
		mgl32.Vec2{1.0, 0.0},  // Right
		mgl32.Vec2{0.0, -1.0}, // Down
		mgl32.Vec2{-1.0, 0.0}, // Left
	}
	max := float32(0.0)
	bestMatch := -1
	for i, direction := range compass {
		dotProd := target.Normalize().Dot(direction)
		if dotProd > max {
			max = dotProd
			bestMatch = i
		}
	}
	return bestMatch
}

func shouldSpawn(chance int32) bool {
	return rand.Int31n(chance) == 0
}

const (
	goodChance = 50
	badChance  = 15
)

func (g *Game) SpawnPowerUps(block *gameObject.GameObject) {
	if shouldSpawn(goodChance) { // 1 in goodChance chance
		g.PowerUps = append(g.PowerUps, powerUp.New("speed",
			mgl32.Vec3{0.5, 0.5, 1.0}, 0.0, block.Position,
			resourceManager.Textures["powerup_speed"]))
	}
	if shouldSpawn(goodChance) {
		g.PowerUps = append(g.PowerUps, powerUp.New("sticky",
			mgl32.Vec3{1.0, 0.5, 1.0}, 20.0, block.Position,
			resourceManager.Textures["powerup_sticky"]))
	}
	if shouldSpawn(goodChance) {
		g.PowerUps = append(g.PowerUps, powerUp.New("pass-through",
			mgl32.Vec3{0.5, 1.0, 0.5}, 10.0, block.Position,
			resourceManager.Textures["powerup_passthrough"]))
	}
	if shouldSpawn(goodChance) {
		g.PowerUps = append(g.PowerUps, powerUp.New("pad-size-increase",
			mgl32.Vec3{1.0, 0.6, 0.4}, 0.0, block.Position,
			resourceManager.Textures["powerup_increase"]))
	}
	if shouldSpawn(badChance) {
		g.PowerUps = append(g.PowerUps, powerUp.New("confuse",
			mgl32.Vec3{1.0, 0.3, 0.3}, 15.0, block.Position,
			resourceManager.Textures["powerup_confuse"]))
	}
	if shouldSpawn(badChance) {
		g.PowerUps = append(g.PowerUps, powerUp.New("chaos",
			mgl32.Vec3{0.9, 0.25, 0.25}, 15.0, block.Position,
			resourceManager.Textures["powerup_chaos"]))
	}
}

func ActivatePowerUp(p *powerUp.PowerUp) {
	if p.Type == "speed" {
		Ball.Object.Velocity = Ball.Object.Velocity.Mul(1.2)
	} else if p.Type == "sticky" {
		Ball.Sticky = true
		Player.Color = mgl32.Vec3{1.0, 0.5, 1.0}
	} else if p.Type == "pass-through" {
		Ball.Passthrough = true
		Ball.Object.Color = mgl32.Vec3{1.0, 0.5, 0.5}
	} else if p.Type == "pad-size-increase" {
		Player.Size[0] += 50
	} else if p.Type == "confuse" {
		if !Effects.Chaos {
			Effects.Confuse = true
		}
	} else if p.Type == "chaos" {
		if !Effects.Confuse {
			Effects.Chaos = true
		}
	}
}

func (g *Game) UpdatePowerUps(dt float32) {
	for i := 0; i < len(g.PowerUps); i++ {
		p := g.PowerUps[i]
		p.Object.Position = p.Object.Position.Add(powerUp.Velocity.Mul(dt))
		if p.Activated {
			p.Duration -= dt
			if p.Duration <= 0.0 {
				p.Activated = false
				// Deativate effects
				if p.Type == "sticky" {
					if !g.isPowerActive("sticky") {
						Ball.Sticky = false
						Player.Color = mgl32.Vec3{1.0, 1.0, 1.0}
					}
				} else if p.Type == "pass-through" {
					if !g.isPowerActive("pass-through") {
						Ball.Passthrough = false
						Ball.Object.Color = mgl32.Vec3{1.0, 1.0, 1.0}
					}
				} else if p.Type == "confuse" {
					if !g.isPowerActive("confuse") {
						Effects.Confuse = false
					}
				} else if p.Type == "chaos" {
					if !g.isPowerActive("chaos") {
						Effects.Chaos = false
					}
				}
			}
		}

		if !p.Activated && p.Object.Destroyed {
			g.PowerUps = append(g.PowerUps[:i], g.PowerUps[i+1:]...)
		}
	}
}

func (g *Game) isPowerActive(power string) bool {
	for _, p := range g.PowerUps {
		if p.Activated && p.Type == power {
			return true
		}
	}
	return false
}
