package gameObject

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/spriteRenderer"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

type GameObject struct {
	Position  mgl32.Vec2
	Size      mgl32.Vec2
	Velocity  mgl32.Vec2
	Color     mgl32.Vec3
	Rotation  float32
	Sprite    *texture.Texture
	IsSolid   bool
	Destroyed bool
}

func New(pos, size mgl32.Vec2, sprite *texture.Texture,
	color mgl32.Vec3, velocity mgl32.Vec2) *GameObject {

	return &GameObject{pos, size, velocity, color, 0.0, sprite, false, false}
}

func NewDefault() *GameObject {
	return &GameObject{Size: mgl32.Vec2{1, 1}, Color: mgl32.Vec3{1.0, 1.0, 1.0}}
}

func (g *GameObject) Draw(sr *spriteRenderer.SpriteRenderer) {
	sr.DrawSprite(g.Sprite, g.Position, g.Size, g.Rotation, g.Color)
}
