// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/1.advanced_lighting/advanced_lighting.cpp

package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	loadModel "github.com/nicholasblaskey/go-learn-opengl/includes/model"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
)

// Settings
const windowWidth = 1280
const windowHeight = 720

// Camera
var ourCamera camera.Camera = camera.NewCamera(
	0.0, 0.0, 3.0, // pos xyz
	0.0, 1.0, 0.0, // up xyz
	-90.0, 0.0, // Yaw and pitch
	80.0, 45.0, 0.1) // Speed, zoom, and mouse sensitivity
var firstMouse bool = true
var lastX float32 = windowWidth / 2
var lastY float32 = windowHeight / 2

// Timing
var deltaTime float32 = 0.0
var lastFrame float32 = 0.0

// Lighting
var lightPos mgl32.Vec3 = mgl32.Vec3{1.2, 1.0, 2.0}

// Controls
var heldW bool = false
var heldA bool = false
var heldS bool = false
var heldD bool = false

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

	// Add in auto resizing
	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))
	window.SetCursorPosCallback(glfw.CursorPosCallback(mouse_callback))
	window.SetScrollCallback(glfw.ScrollCallback(scroll_callback))

	// Tell glfw to capture the mouse
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Config gl global state
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return window
}

func makeCubeBuffers() (uint32, uint32) {
	planeVertices := []float32{
		// positions            // normals         // texcoords
		10.0, -0.5, 10.0, 0.0, 1.0, 0.0, 10.0, 0.0,
		-10.0, -0.5, 10.0, 0.0, 1.0, 0.0, 0.0, 0.0,
		-10.0, -0.5, -10.0, 0.0, 1.0, 0.0, 0.0, 10.0,

		10.0, -0.5, 10.0, 0.0, 1.0, 0.0, 10.0, 0.0,
		-10.0, -0.5, -10.0, 0.0, 1.0, 0.0, 0.0, 10.0,
		10.0, -0.5, -10.0, 0.0, 1.0, 0.0, 10.0, 10.0,
	}
	// planeVAO
	var planeVAO, planeVBO uint32
	gl.GenVertexArrays(1, &planeVAO)
	gl.GenBuffers(1, &planeVBO)
	gl.BindVertexArray(planeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, planeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*4,
		gl.Ptr(planeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.BindVertexArray(0)

	return planeVAO, planeVBO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	ourShader := shader.MakeShaders("1.advanced_lighting.vs",
		"1.advanced_lighting.fs")
	planeVAO, planeVBO := makeCubeBuffers()
	defer gl.DeleteVertexArrays(1, &planeVAO)
	defer gl.DeleteVertexArrays(1, &planeVBO)

	dir := "../../../resources/textures"
	floorTexture := loadModel.TextureFromFile("wood.png", dir, false)

	// shader config
	ourShader.Use()
	ourShader.SetInt("texture1", 0)

	lightPos := mgl32.Vec3{0.0, 0.0, 0.0}

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.05, 0.05, 0.05, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Draw objects
		ourShader.Use()
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowWidth)/windowHeight, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		ourShader.SetMat4("projection", projection)
		ourShader.SetMat4("view", view)
		// Set light uniforms
		ourShader.SetVec3("viewPos", ourCamera.Position)
		ourShader.SetVec3("lightPos", lightPos)
		ourShader.SetBool("blinn", blinn)
		// Floor
		gl.BindVertexArray(planeVAO)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, floorTexture)
		//ourShader.SetMat4("model", mgl32.Ident4())
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		window.SwapBuffers()
	}
}

var (
	blinn           = true
	blinnKeyPressed = false
)

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {

	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}

	if key == glfw.KeyW && action == glfw.Press || heldW {
		ourCamera.ProcessKeyboard(camera.FORWARD, deltaTime)
		heldW = true
	}
	if key == glfw.KeyS && action == glfw.Press || heldS {
		ourCamera.ProcessKeyboard(camera.BACKWARD, deltaTime)
		heldS = true
	}
	if key == glfw.KeyA && action == glfw.Press || heldA {
		ourCamera.ProcessKeyboard(camera.LEFT, deltaTime)
		heldA = true
	}
	if key == glfw.KeyD && action == glfw.Press || heldD {
		ourCamera.ProcessKeyboard(camera.RIGHT, deltaTime)
		heldD = true
	}

	if key == glfw.KeyW && action == glfw.Release {
		heldW = false
	}
	if key == glfw.KeyS && action == glfw.Release {
		heldS = false
	}
	if key == glfw.KeyA && action == glfw.Release {
		heldA = false
	}
	if key == glfw.KeyD && action == glfw.Release {
		heldD = false
	}

	if key == glfw.KeyB && action == glfw.Press && !blinnKeyPressed {
		blinn = !blinn
		blinnKeyPressed = true
	}
	if key == glfw.KeyB && action == glfw.Release {
		blinnKeyPressed = false
	}
}

func mouse_callback(w *glfw.Window, xPos float64, yPos float64) {
	if firstMouse {
		lastX = float32(xPos)
		lastY = float32(yPos)
		firstMouse = false
	}

	xOffset := float32(xPos) - lastX
	// Reversed due to y coords go from bot up
	yOffset := lastY - float32(yPos)

	lastX = float32(xPos)
	lastY = float32(yPos)

	ourCamera.ProcessMouseMovement(xOffset, yOffset, true)
}

func scroll_callback(w *glfw.Window, xOffset float64, yOffset float64) {
	ourCamera.ProcessMouseScroll(float32(yOffset))
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
