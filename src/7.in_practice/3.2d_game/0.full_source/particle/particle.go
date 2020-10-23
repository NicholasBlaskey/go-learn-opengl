package particle

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/shader"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

type Particle struct {
	Position mgl32.Vec2
	Velocity mgl32.Vec2
	Color    mgl32.Vec4
	Life     float32
}

type Generator struct {
	Particles        []*Particle
	Amount           uint32
	Shader           *shader.Shader
	Texture          *texture.Texture
	VAO              uint32
	lastUsedParticle uint32
}

func NewGenerator(s *shader.Shader, t *texture.Texture, amount uint32) *Generator {
	g := &Generator{Shader: s, Texture: t, Amount: amount}
	g.init()

	return g
}

func (g *Generator) Update(dt float32, object *gameObject.GameObject,
	newParticles uint32, offset mgl32.Vec2) {

	// Add new particles
	for i := uint32(0); i < newParticles; i++ {
		unusedParticle := g.firstUnusedParticle()
		g.respawnParticle(g.Particles[unusedParticle], object, offset)
	}

	// Update all particles
	for i := uint32(0); i < g.Amount; i++ {
		p := g.Particles[i]
		p.Life -= dt
		if p.Life > 0.0 { // Particle is alive, thus update
			p.Position = p.Position.Sub(p.Velocity.Mul(dt))
			p.Color[3] -= dt * 2.5 // Lower alpha value
		}
	}
}

func (g *Generator) Draw() {
	// Use additive blending to give it a 'glow' effect
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	defer gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	g.Shader.Use()
	for _, p := range g.Particles {
		g.Shader.SetVector2f("offset", p.Position, false)
		g.Shader.SetVector4f("color", p.Color, false)
		g.Texture.Bind()
		gl.BindVertexArray(g.VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		gl.BindVertexArray(0)
	}
}

func (g *Generator) init() {
	var VBO uint32
	particleQuad := []float32{
		0.0, 1.0, 0.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
		0.0, 0.0, 0.0, 0.0,

		0.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, 0.0, 1.0, 0.0,
	}
	gl.GenVertexArrays(1, &g.VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(g.VAO)
	// Fill mesh buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(particleQuad)*4,
		gl.Ptr(particleQuad), gl.STATIC_DRAW)
	// Set mesh attribs
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	for i := 0; i < int(g.Amount); i++ {
		g.Particles = append(g.Particles,
			&Particle{Color: mgl32.Vec4{1.0, 1.0, 1.0, 1.0}})
	}
}

func (g *Generator) firstUnusedParticle() uint32 {
	// First search from last used particle,
	// this should almost always return instantly
	for i := g.lastUsedParticle; i < g.Amount; i++ {
		if g.Particles[i].Life <= 0.0 {
			g.lastUsedParticle = i
			return i
		}
	}

	// Otherwise, do a linear search
	for i := uint32(0); i < g.lastUsedParticle; i++ {
		if g.Particles[i].Life <= 0.0 {
			g.lastUsedParticle = i
			return i
		}
	}

	// If all particles are taken, override the first one
	// If this case keeps getting hit more particles are needed
	g.lastUsedParticle = 0
	return 0
}

func (g *Generator) respawnParticle(particle *Particle,
	object *gameObject.GameObject, offset mgl32.Vec2) {

	random := float32(rand.Int31n(100)-50) / 10.0
	rColor := 0.5 + float32(rand.Int31n(100)/100.0)
	particle.Position = object.Position.Add(
		offset).Add(mgl32.Vec2{random, random})
	particle.Color = mgl32.Vec4{rColor, rColor, rColor, 1.0}
	particle.Life = 1.0
	particle.Velocity = object.Velocity.Mul(0.1)
}
