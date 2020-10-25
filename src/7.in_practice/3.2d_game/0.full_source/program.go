package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	// Gross import path todo fix later
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/game"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

func init() {
	runtime.LockOSThread()
}

func initGLFW() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to init glfw:", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(
		windowWidth, windowHeight, "Breakout", nil, nil)

	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Add in auto resizing
	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Config gl global state
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return window
}

var (
	breakout *game.Game
)

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Initiliaze game
	breakout = game.New(windowWidth, windowHeight)
	breakout.Init()

	// deltaTime variables
	deltaTime := 0.0
	lastFrame := 0.0
	for !window.ShouldClose() {
		// Calculate delta time
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		// Manage user input
		breakout.ProcessInput(deltaTime)

		// Update game state
		breakout.Update(deltaTime)

		// Render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		breakout.Render()

		window.SwapBuffers()
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}

	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			breakout.Keys[key] = true
		} else if action == glfw.Release {
			breakout.Keys[key] = false
			breakout.KeysProcessed[key] = false
		}
	}
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
