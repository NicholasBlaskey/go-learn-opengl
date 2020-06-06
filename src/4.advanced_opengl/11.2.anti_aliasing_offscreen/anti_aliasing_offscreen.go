// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/tree/master/src/4.advanced_opengl/11.2.anti_aliasing_offscreen

package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
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

	glfw.WindowHint(glfw.Samples, 4) // For anti-aliasing
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
	gl.Enable(gl.MULTISAMPLE)

	return window
}

func makeCubeBuffers() (uint32, uint32, uint32, uint32) {
	cubeVertices := []float32{
		// positions
		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5, -0.5,
		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,

		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,

		-0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, -0.5,
		-0.5, -0.5, 0.5,
		-0.5, 0.5, 0.5,

		0.5, 0.5, 0.5,
		0.5, 0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,

		-0.5, -0.5, -0.5,
		0.5, -0.5, -0.5,
		0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
		-0.5, -0.5, 0.5,
		-0.5, -0.5, -0.5,

		-0.5, 0.5, -0.5,
		0.5, 0.5, -0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		-0.5, 0.5, -0.5,
	}
	quadVertices := []float32{
		// Positions      // texture coords
		-1.0, 1.0, 0.0, 1.0,
		-1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 0.0,

		-1.0, 1.0, 0.0, 1.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
	}

	// cube VAO
	var cubeVBO, cubeVAO uint32
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices)*4,
		gl.Ptr(cubeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.BindVertexArray(0)
	// Quad VBO
	var quadVAO, quadVBO uint32
	gl.GenVertexArrays(1, &quadVAO)
	gl.GenBuffers(1, &quadVBO)
	gl.BindVertexArray(quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4,
		gl.Ptr(quadVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4,
		gl.PtrOffset(2*4))
	gl.BindVertexArray(0)

	return cubeVAO, cubeVBO, quadVAO, quadVBO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	ourShader := shader.MakeShaders("11.2.anti_aliasing.vs",
		"11.2.anti_aliasing.fs")
	screenShader := shader.MakeShaders("11.2.aa_post.vs",
		"11.2.aa_post.fs")

	cubeVAO, cubeVBO, quadVAO, quadVBO := makeCubeBuffers()

	defer gl.DeleteVertexArrays(1, &cubeVAO)
	defer gl.DeleteVertexArrays(1, &cubeVBO)
	defer gl.DeleteVertexArrays(1, &quadVAO)
	defer gl.DeleteVertexArrays(1, &quadVBO)

	// Configure framebuffer
	var framebuffer uint32
	gl.GenFramebuffers(1, &framebuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
	// Create a multisampled color attachment texture
	var textureColorBufferMulti uint32
	gl.GenTextures(1, &textureColorBufferMulti)
	gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, textureColorBufferMulti)
	gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, 4, gl.RGB,
		windowWidth, windowHeight, true)
	gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, 0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D_MULTISAMPLE, textureColorBufferMulti, 0)
	// Create a renderbuffer object for depth and stencil atachment (multi)
	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, 4,
		gl.DEPTH24_STENCIL8, windowWidth, windowHeight)
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT,
		gl.RENDERBUFFER, rbo)
	// Ensure framebuffer is complete
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Framebuffer is not complete")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Configure second post-processing framebuffer
	var intermediateFBO uint32
	gl.GenFramebuffers(1, &intermediateFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, intermediateFBO)
	// Create a color attachment texture
	var screenTexture uint32
	gl.GenTextures(1, &screenTexture)
	gl.BindTexture(gl.TEXTURE_2D, screenTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, windowWidth, windowHeight,
		0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, screenTexture, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Intermediate framebuffer is not complete")
	}

	screenShader.Use()
	screenShader.SetInt("screenTexture", 0)

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		// Draw scene as normal in multisampled buffers
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		gl.Enable(gl.DEPTH_TEST)

		// Set transformation matrices
		ourShader.Use()
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowHeight)/windowWidth, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		ourShader.SetMat4("projection", projection)
		ourShader.SetMat4("view", view)
		ourShader.SetMat4("model", mgl32.Ident4())

		// Render the images into buffer
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

		// Now blit multisampled buffer to normal colorbuffer
		// of intermediate FBO. Image stored in screenTexture
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, intermediateFBO)
		gl.BlitFramebuffer(0, 0, windowWidth, windowHeight, 0, 0,
			windowWidth, windowHeight, gl.COLOR_BUFFER_BIT, gl.NEAREST)

		// Now render quad with scene's visuals as its texture image
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.ClearColor(1.0, 1.0, 1.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Disable(gl.DEPTH_TEST)

		// Draw Screen quad
		screenShader.Use()
		gl.BindVertexArray(quadVAO)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, screenTexture)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		window.SwapBuffers()
	}
}

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
