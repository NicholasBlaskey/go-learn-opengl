package powerUp

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

var (
	PowerUpSize mgl32.Vec2 = mgl32.Vec2{60.0, 20.0}
	Velocity    mgl32.Vec2 = mgl32.Vec2{0.0, 150.0}
)

type PowerUp struct {
	Type      string
	Duration  float32
	Activated bool
	Object    *gameObject.GameObject
}

func New(t string, col mgl32.Vec3, duration float32,
	pos mgl32.Vec2, text *texture.Texture) *PowerUp {

	return &PowerUp{t, duration, false,
		gameObject.New(pos, PowerUpSize, text, col, Velocity)}
}
