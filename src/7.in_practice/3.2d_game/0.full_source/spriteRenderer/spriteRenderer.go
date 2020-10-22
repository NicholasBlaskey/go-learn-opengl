package spriteRenderer

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/shader"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

type SpriteRenderer struct {
	SpriteShader *shader.Shader
	VAO          uint32
}

func New(s *shader.Shader) *SpriteRenderer {
	sr := &SpriteRenderer{SpriteShader: s}
	sr.initRenderData()
	return sr
}

func (sr *SpriteRenderer) Clear() {
	gl.DeleteVertexArrays(1, &sr.VAO)
}

func (sr *SpriteRenderer) DrawSprite(texture *texture.Texture,
	position mgl32.Vec2, size mgl32.Vec2, rotate float32, color mgl32.Vec3) {

	// Translate first
	model := mgl32.Translate3D(position[0], position[1], 0)

	// Move origin to center
	model = model.Mul4(mgl32.Translate3D(0.5*size[0], 0.5*size[1], 0))
	// Then rotate
	model = model.Mul4(mgl32.HomogRotate3D(
		mgl32.DegToRad(rotate), mgl32.Vec3{0, 0, 1}))
	// Move origin back
	model = model.Mul4(mgl32.Translate3D(-0.5*size[0], -0.5*size[1], 0))

	// Scale last
	model = model.Mul4(mgl32.Scale3D(size[0], size[1], 1))

	sr.SpriteShader.SetMatrix4("model", model, true)
	sr.SpriteShader.SetVector3f("spriteColor", color, false)

	gl.ActiveTexture(gl.TEXTURE0)
	texture.Bind()

	gl.BindVertexArray(sr.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (sr *SpriteRenderer) initRenderData() {
	var VBO uint32
	vertices := []float32{
		// pos    // tex
		0.0, 1.0, 0.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0,

		0.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
	}

	gl.GenVertexArrays(1, &sr.VAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindVertexArray(sr.VAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}
