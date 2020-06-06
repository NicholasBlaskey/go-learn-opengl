// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/4.advanced_opengl/10.1.instancing_quads/instancing_quads.cpp

package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
)

// Settings
const windowWidth = 800
const windowHeight = 600

func init() {
	runtime.LockOSThread()
}

func initGLFW() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to init glfw:", err)
	}

	//glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(
		windowWidth, windowHeight, "Hello!", nil, nil)

	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Config gl global state
	gl.Enable(gl.DEPTH_TEST)

	return window
}

func makeBuffers() (uint32, uint32, uint32) {
	// Generate list of translations
	translations := [100]mgl32.Vec2{}
	index := 0
	offset := float32(0.1)
	for y := -10; y < 10; y += 2 {
		for x := -10; x < 10; x += 2 {
			translations[index] = mgl32.Vec2{
				float32(x)/10.0 + offset, float32(y)/10.0 + offset}
			index += 1
		}
	}

	// Store instance data in an array buffer
	var instanceVBO uint32
	gl.GenBuffers(1, &instanceVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(translations)*4*2,
		unsafe.Pointer(&translations[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// Set up vertex data and buffers and config vertex attribs
	Vertices := []float32{
		// positions     // colors
		-0.05, 0.05, 1.0, 0.0, 0.0,
		0.05, -0.05, 0.0, 1.0, 0.0,
		-0.05, -0.05, 0.0, 0.0, 1.0,

		-0.05, 0.05, 1.0, 0.0, 0.0,
		0.05, -0.05, 0.0, 1.0, 0.0,
		0.05, 0.05, 0.0, 1.0, 1.0,
	}
	var quadVBO, quadVAO uint32
	gl.GenVertexArrays(1, &quadVAO)
	gl.GenBuffers(1, &quadVBO)
	gl.BindVertexArray(quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(Vertices)*4,
		gl.Ptr(Vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4))
	// Also set instance data
	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 2*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.VertexAttribDivisor(2, 1)

	return quadVAO, quadVBO, instanceVBO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Create shaders
	ourShader := shader.MakeShaders("10.1.instancing.vs",
		"10.1.instancing.fs")

	quadVAO, quadVBO, instanceVBO := makeBuffers()
	defer gl.DeleteVertexArrays(1, &quadVAO)
	defer gl.DeleteVertexArrays(1, &quadVBO)
	defer gl.DeleteVertexArrays(1, &instanceVBO)

	// Program loop
	for !window.ShouldClose() {
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		// Draw 100 instanced quads
		ourShader.Use()
		gl.BindVertexArray(quadVAO)
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, 100)
		gl.BindVertexArray(0)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}
