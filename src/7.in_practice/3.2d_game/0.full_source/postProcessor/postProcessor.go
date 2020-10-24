package postProcessor

import (
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/go-gl/mathgl/mgl32"

	// Gross import path todo fix later
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/shader"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

type PostProcessor struct {
	Shader  *shader.Shader
	Texture *texture.Texture
	Width   int32
	Height  int32
	Chaos   bool
	Shake   bool
	Confuse bool
	MSFBO   uint32
	FBO     uint32
	RBO     uint32
	VAO     uint32
}

func New(s *shader.Shader, width, height int32) *PostProcessor {
	p := &PostProcessor{Shader: s, Width: width, Height: height}
	// Initialize renderbuffer / framebuffer object
	gl.GenFramebuffers(1, &p.MSFBO)
	gl.GenFramebuffers(1, &p.FBO)
	gl.GenRenderbuffers(1, &p.RBO)

	// Initialize renderbuffer storage with a multisampled color buffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.MSFBO)
	gl.BindRenderbuffer(gl.RENDERBUFFER, p.RBO)
	// Allocate storage for RBO object
	gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 4, gl.RGB, width, height)
	// Attach MS render buffer object to framebuffer
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.RENDERBUFFER, p.RBO)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Postprocessor failed to initialize MSFBO")
	}

	// Also initialize the FBO / texture to blit multisample color-buffer to
	// be used for shader operations (for postprocessing effects)
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.FBO)
	p.Texture = texture.New()
	p.Texture.Generate(width, height, nil)
	// Attach texture to framebuffer as its color attachment
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.Texture.ID, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Postprocessor failed to initialize FBO")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Initialize render data and uniforms
	p.initRenderData()
	p.Shader.SetInteger("scene", 0, true)
	offset := float32(1.0 / 300.0)
	offsets := []mgl32.Vec2{
		mgl32.Vec2{-offset, offset},  // top-left
		mgl32.Vec2{0.0, offset},      // top-center
		mgl32.Vec2{offset, offset},   // top-right
		mgl32.Vec2{-offset, 0.0},     // center-left
		mgl32.Vec2{0.0, 0.0},         // center-center
		mgl32.Vec2{offset, 0.0},      // center-right
		mgl32.Vec2{-offset, -offset}, // bottom-left
		mgl32.Vec2{0.0, -offset},     // bottom-center
		mgl32.Vec2{offset, -offset},  // bottom right
	}
	gl.Uniform2fv(gl.GetUniformLocation(p.Shader.ID,
		gl.Str("offsets\x00")), 9, &offsets[0][0])

	edgeKernel := []int32{
		-1, -1, -1,
		-1, +8, -1,
		-1, -1, -1,
	}
	gl.Uniform1iv(gl.GetUniformLocation(p.Shader.ID,
		gl.Str("edge_kernel\x00")), 9, &edgeKernel[0])

	blurKernel := []float32{
		1.0 / 16.0, 2.0 / 16.0, 1.0 / 16.0,
		2.0 / 16.0, 4.0 / 16.0, 2.0 / 16.0,
		1.0 / 16.0, 2.0 / 16.0, 1.0 / 16.0,
	}
	gl.Uniform1fv(gl.GetUniformLocation(p.Shader.ID,
		gl.Str("blur_kernel\x00")), 9, &blurKernel[0])

	return p
}

func (p *PostProcessor) BeginRender() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.MSFBO)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (p *PostProcessor) EndRender() {
	// Now resolve multisampled color-buffer into intermidate FBO to
	// store its texture
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, p.MSFBO)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, p.FBO)
	gl.BlitFramebuffer(0, 0, p.Width, p.Height, 0, 0, p.Width, p.Height,
		gl.COLOR_BUFFER_BIT, gl.NEAREST)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func boolToInt(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func (p *PostProcessor) Render(time float32) {
	// Set unfiorms
	p.Shader.SetFloat("time", time, true)
	p.Shader.SetInteger("confuse", boolToInt(p.Confuse), false)
	p.Shader.SetInteger("chaos", boolToInt(p.Chaos), false)
	p.Shader.SetInteger("shake", boolToInt(p.Shake), false)

	// Render texture quad
	gl.ActiveTexture(gl.TEXTURE0)
	p.Texture.Bind()
	gl.BindVertexArray(p.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (p *PostProcessor) initRenderData() {
	var VBO uint32
	vertices := []float32{
		// pos    // tex
		-1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
		-1.0, 1.0, 0.0, 1.0,

		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
	}
	gl.GenVertexArrays(1, &p.VAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindVertexArray(p.VAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}
