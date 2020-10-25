// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/7.in_practice/2.text_rendering/text_rendering.cpp

package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/nicholasblaskey/glfont"
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
		windowWidth, windowHeight, "Hello!", nil, nil)

	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Add in auto resizing
	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))

	// Tell glfw to capture the mouse
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Config gl global state
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	//pathToFont := "../../../resources/fonts/huge_agb_v5.ttf"
	pathToFont := "../../../resources/fonts/Antonio-Bold.ttf"
	font, err := glfont.LoadFont(pathToFont, int32(48), windowWidth, windowHeight)
	if err != nil {
		log.Panicf("LoadFont: %v", err)
	}

	for !window.ShouldClose() {
		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		font.SetColor(0.5, 0.8, 0.2, 1.0)
		font.Printf(25.0, float32(windowHeight)-25.0, 1.0, "This is sample text")
		font.SetColor(0.3, 0.7, 0.9, 1.0)
		font.Printf(540.0, float32(windowHeight)-570.0, 0.5, "(C) LearnOpenGL.com")

		window.SwapBuffers()
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
