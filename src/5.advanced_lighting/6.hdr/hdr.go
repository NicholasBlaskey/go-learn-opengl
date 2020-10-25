// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/6.hdr/hdr.cpp

package main

import (
	"fmt"
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
const windowWidth = 800
const windowHeight = 600

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

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	ourShader := shader.MakeShaders("6.lighting.vs", "6.lighting.fs")
	hdrShader := shader.MakeShaders("6.hdr.vs", "6.hdr.fs")

	dir := "../../../resources/textures"
	woodTexture := loadModel.TextureFromFile("wood.png", dir, false)

	// Configure floating point framebuffer
	var hdrFBO, colorBuffer uint32
	gl.GenFramebuffers(1, &hdrFBO)
	// Create depth texture
	gl.GenTextures(1, &colorBuffer)
	gl.BindTexture(gl.TEXTURE_2D, colorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, windowWidth, windowHeight,
		0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// Create depth buffer (renderbuffer)
	var rboDepth uint32
	gl.GenRenderbuffers(1, &rboDepth)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rboDepth)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, windowWidth, windowHeight)
	// Attach buffers
	gl.BindFramebuffer(gl.FRAMEBUFFER, hdrFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, colorBuffer, 0)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rboDepth)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Framebuffer not complete")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Lighting info
	lightPositions := []mgl32.Vec3{
		mgl32.Vec3{0.0, 0.0, 49.5},
		mgl32.Vec3{-1.4, -1.9, 9.0},
		mgl32.Vec3{0.0, -1.8, 4.0},
		mgl32.Vec3{0.8, -1.7, 6.0},
	}
	lightColors := []mgl32.Vec3{
		mgl32.Vec3{200.0, 200.0, 200.0},
		mgl32.Vec3{0.1, 0.0, 0.0},
		mgl32.Vec3{0.0, 0.0, 0.2},
		mgl32.Vec3{0.0, 0.1, 0.0},
	}

	// shader config
	ourShader.Use()
	ourShader.SetInt("diffuseTexture", 0)
	hdrShader.Use()
	ourShader.SetInt("hdrBuffer", 0)

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// 1. Render the scene into the floating point framebuffer
		gl.BindFramebuffer(gl.FRAMEBUFFER, hdrFBO)
		{
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			ourShader.Use()
			projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
				float32(windowWidth)/windowHeight, 0.1, 100.0)
			view := ourCamera.GetViewMatrix()
			ourShader.SetMat4("projection", projection)
			ourShader.SetMat4("view", view)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, woodTexture)
			// Set the lighting uniforms
			for i := 0; i < len(lightPositions); i++ {
				ourShader.SetVec3(fmt.Sprintf("lights[%d].Position", i), lightPositions[i])
				ourShader.SetVec3(fmt.Sprintf("lights[%d].Color", i), lightColors[i])
			}
			ourShader.SetVec3("viewPos", ourCamera.Position)
			// Render the tunnel
			model := mgl32.Translate3D(0.0, 0.0, 25.0).Mul4(
				mgl32.Scale3D(2.5, 2.5, 27.5))
			ourShader.SetMat4("model", model)
			ourShader.SetBool("inverse_normals", true)
			renderCube()
		}
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// 2. Now render the floating point color buffer to a 2d quad
		// and tonemap HDR colors to default framebuffer's (clamed) color range
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		hdrShader.Use()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, colorBuffer)

		hdrShader.SetBool("hdr", hdr)
		if hdr {
			fmt.Println("Hdr is on and exposure is", exposure)
		} else {
			fmt.Println("Hdr is off and exposure is", exposure)
		}
		hdrShader.SetFloat("exposure", exposure)
		renderQuad()

		window.SwapBuffers()
	}
}

func renderScene(s shader.Shader, planeVAO uint32) {
	/*
			// Floor
			model := mgl32.Ident4()
			s.SetMat4("model", model)
			gl.BindVertexArray(planeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	*/

	// Room cube
	s.SetMat4("model", mgl32.Scale3D(5.0, 5.0, 5.0))
	gl.Disable(gl.CULL_FACE)
	s.SetInt("reverse_normals", 1)
	renderCube()
	s.SetInt("reverse_normals", 0)
	gl.Enable(gl.CULL_FACE)
	// Cubes
	model := mgl32.Translate3D(4.0, -3.5, 0.0).Mul4(
		mgl32.Scale3D(0.5, 0.5, 0.5))
	s.SetMat4("model", model)
	renderCube()

	model = mgl32.Translate3D(2.0, 3.0, 1.0).Mul4(
		mgl32.Scale3D(0.75, 0.75, 0.75))
	s.SetMat4("model", model)
	renderCube()

	model = mgl32.Translate3D(-3.0, -1.0, 0.0).Mul4(
		mgl32.Scale3D(0.5, 0.5, 0.5))
	s.SetMat4("model", model)
	renderCube()

	model = mgl32.Translate3D(-1.5, 1.0, 1.5).Mul4(
		mgl32.Scale3D(0.5, 0.5, 0.5))
	s.SetMat4("model", model)
	renderCube()

	model = mgl32.Translate3D(-1.5, 2.0, -3.0).Mul4(
		mgl32.HomogRotate3D(mgl32.DegToRad(60),
			mgl32.Vec3{1.0, 0.0, 1.0}.Normalize())).Mul4(
		mgl32.Scale3D(0.75, 0.75, 0.75))
	s.SetMat4("model", model)
	renderCube()
}

var (
	cubeVAO uint32
	cubeVBO uint32
	quadVAO uint32
	quadVBO uint32
)

func renderCube() {
	if cubeVAO != 0 {
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)
		return
	}

	vertices := []float32{
		// positions            // normals         // texcoords
		// back
		-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
		1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
		1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
		-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
		-1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 1.0, // top-left
		// front
		-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
		1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
		1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
		-1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, // top-left
		-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
		// left
		-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
		-1.0, 1.0, -1.0, -1.0, 0.0, 0.0, 1.0, 1.0, // top-left
		-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
		-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
		-1.0, -1.0, 1.0, -1.0, 0.0, 0.0, 0.0, 0.0, // bottom-right
		-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
		// right
		1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
		1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
		1.0, 1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, // top-right
		1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
		1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
		1.0, -1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0, // bottom-left
		// bottom
		-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
		1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 1.0, 1.0, // top-left
		1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
		1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
		-1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 0.0, 0.0, // bottom-right
		-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
		// top
		-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
		1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 1.0, 1.0, // top-right
		1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
		-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
		-1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 0.0, // bottom-left
	}
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4,
		gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.BindVertexArray(0)

	renderCube()
}

func renderQuad() {
	if quadVAO != 0 {
		gl.BindVertexArray(quadVAO)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		gl.BindVertexArray(0)
		return
	}

	vertices := []float32{
		// positions        // texture Coords
		-1.0, 1.0, 0.0, 0.0, 1.0,
		-1.0, -1.0, 0.0, 0.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 1.0,
		1.0, -1.0, 0.0, 1.0, 0.0,
	}
	gl.GenVertexArrays(1, &quadVAO)
	gl.GenBuffers(1, &quadVBO)
	gl.BindVertexArray(quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4,
		gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.BindVertexArray(0)

	renderQuad()
}

var (
	hdr                   = true
	hdrKeyPressed         = false
	exposure      float32 = 5.0
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

	if key == glfw.KeySpace && action == glfw.Press && !hdrKeyPressed {
		hdr = !hdr
		hdrKeyPressed = true
	}
	if key == glfw.KeySpace && action == glfw.Release {
		hdrKeyPressed = false
	}

	if key == glfw.KeyQ && action == glfw.Press {
		if exposure > 0.0 {
			exposure -= 0.1
		} else {
			exposure = 0.0
		}
	} else if key == glfw.KeyE && action == glfw.Press {
		exposure += 0.01
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
