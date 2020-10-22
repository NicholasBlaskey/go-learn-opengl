package ballObject

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

type BallObject struct {
	Radius float32
	Stuck  bool
	Object *gameObject.GameObject
}

func NewDefault() *BallObject {
	return &BallObject{12.5, true, gameObject.NewDefault()}
}

func New(pos mgl32.Vec2, radius float32,
	velocity mgl32.Vec2, sprite *texture.Texture) *BallObject {

	return &BallObject{radius, true,
		gameObject.New(pos, mgl32.Vec2{radius * 2.0, radius * 2.0},
			sprite, mgl32.Vec3{1.0, 1.0, 1.0}, velocity,
		)}
}

func (b *BallObject) Move(dt float32, windowWidth uint32) mgl32.Vec2 {
	if !b.Stuck {
		// Move the ball
		b.Object.Position = b.Object.Position.Add(b.Object.Velocity.Mul(dt))
		// Check if it falls outside window bounds if so reverse velocity
		// and then restore correct position
		if b.Object.Position[0] <= 0.0 {
			b.Object.Velocity[0] = -b.Object.Velocity[0]
			b.Object.Position[0] = 0.0
		} else if (b.Object.Position[0] + b.Object.Size[0]) >= float32(windowWidth) {
			b.Object.Velocity[0] = -b.Object.Velocity[0]
			b.Object.Position[0] = float32(windowWidth) - b.Object.Size[0]
		}

		if b.Object.Position[1] <= 0.0 {
			b.Object.Velocity[1] = -b.Object.Velocity[1]
			b.Object.Position[1] = 0.0
		}
	}
	return b.Object.Position
}

func (b *BallObject) Reset(position mgl32.Vec2, velocity mgl32.Vec2) {
	b.Object.Position = position
	b.Object.Velocity = velocity
	b.Stuck = true
}
