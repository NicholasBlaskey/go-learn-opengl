// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
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

func lerp(a, b, f float32) float32 {
	return a + f*(b-a)
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Build and compile shaders
	shaderGeometryPass := shader.MakeShaders("9.ssao_geometry.vs", "9.ssao_geometry.fs")
	shaderLightingPass := shader.MakeShaders("9.ssao.vs", "9.ssao_lighting.fs")
	shaderSSAO := shader.MakeShaders("9.ssao.vs", "9.ssao.fs")
	shaderSSAOBlur := shader.MakeShaders("9.ssao.vs", "9.ssao_blur.fs")

	// Load model
	backpack := loadModel.NewModel(
		"../../../resources/objects/backpack/backpack.obj", false)

	// Configure g-buffer framebuffer
	var gBuffer uint32
	gl.GenFramebuffers(1, &gBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, gBuffer)
	var gPosition, gNormal, gAlbedoSpec uint32
	// Position color buffer
	gl.GenTextures(1, &gPosition)
	gl.BindTexture(gl.TEXTURE_2D, gPosition)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, windowWidth,
		windowHeight, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, gPosition, 0)
	// Normal color buffer
	gl.GenTextures(1, &gNormal)
	gl.BindTexture(gl.TEXTURE_2D, gNormal)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, windowWidth,
		windowHeight, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1,
		gl.TEXTURE_2D, gNormal, 0)
	// Color + specular color buffer
	gl.GenTextures(1, &gAlbedoSpec)
	gl.BindTexture(gl.TEXTURE_2D, gAlbedoSpec)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, windowWidth,
		windowHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2,
		gl.TEXTURE_2D, gAlbedoSpec, 0)
	// Tell OpenGL which color attachments we will use (of this framebuffer) for rendering
	attachments := []uint32{gl.COLOR_ATTACHMENT0,
		gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2}
	gl.DrawBuffers(3, &attachments[0])

	// Create and attach depth buffer (renderbuffer)
	var rboDepth uint32
	gl.GenRenderbuffers(1, &rboDepth)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rboDepth)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, windowWidth, windowHeight)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rboDepth)
	// Check if frame buffer is complete
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Framebuffer not complete")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Also create a framebuffer to hold SSAO processing stage
	var ssaoFBO, ssaoBlurFBO uint32
	gl.GenFramebuffers(1, &ssaoFBO)
	gl.GenFramebuffers(1, &ssaoBlurFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, ssaoFBO)
	var ssaoColorBuffer, ssaoColorBufferBlur uint32
	// SSAO color buffer
	gl.GenTextures(1, &ssaoColorBuffer)
	gl.BindTexture(gl.TEXTURE_2D, ssaoColorBuffer)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, windowWidth,
		windowHeight, 0, gl.RED, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, ssaoColorBuffer, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("SSAO Framebuffer not complete")
	}
	// Blur stage too
	gl.BindFramebuffer(gl.FRAMEBUFFER, ssaoBlurFBO)
	gl.GenTextures(1, &ssaoColorBufferBlur)
	gl.BindTexture(gl.TEXTURE_2D, ssaoColorBufferBlur)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, windowWidth,
		windowHeight, 0, gl.RED, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, ssaoColorBufferBlur, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("SSAO Blur Framebuffer not complete")
	}

	// Generate sample kernel
	ssaoKernel := []mgl32.Vec3{}
	for i := 0; i < 64; i++ {
		sample := mgl32.Vec3{rand.Float32()*2.0 - 1.0,
			rand.Float32()*2.0 - 1.0, rand.Float32()}.Normalize()
		scale := float32(i) / 64.0

		ssaoKernel = append(ssaoKernel,
			sample.Mul(rand.Float32()*lerp(0.1, 1.0, scale*scale)))
	}

	// Generate noise texture
	ssaoNoise := []mgl32.Vec3{}
	for i := 0; i < 16; i++ {
		// Rotate around z-axis (in tangent space)
		ssaoNoise = append(ssaoNoise,
			mgl32.Vec3{rand.Float32()*2.0 - 1.0, rand.Float32()*2.0 - 1.0, 0.0})
	}
	var noiseTexture uint32
	gl.GenTextures(1, &noiseTexture)
	gl.BindTexture(gl.TEXTURE_2D, 0, gl.RGBA32F, 4, 4,
		0, gl.RGB, gl.FLOAT, &ssaoNoise[0])
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	// Lighting info
	lightPos := mgl32.Vec3{2.0, 4.0, -2.0}
	lightColor := mgl32.Vec3{0.2, 0.2, 0.7}

	// Shader config
	shaderLightingPass.Use()
	shaderLightingPass.SetInt("gPosition", 0)
	shaderLightingPass.SetInt("gNormal", 1)
	shaderLightingPass.SetInt("gAlbedo", 2)
	shaderLightingPass.SetInt("ssao", 3)
	shaderSSAO.Use()
	shaderSSAO.SetInt("gPosition", 0)
	shaderSSAO.SetInt("gNormal", 1)
	shaderSSAO.SetInt("texNoise", 2)
	shaderSSAOBlur.Use()
	shaderSSAOBlur.SetInt("ssaoInput", 0)

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// 1. Geometry pass: render scene's geometry / color data into gbuffer
		gl.BindFramebuffer(gl.FRAMEBUFFER, gBuffer)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowWidth)/windowHeight, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		shaderGeometryPass.Use()
		shaderGeometryPass.SetMat4("projection", projection)
		shaderGeometryPass.SetMat4("view", view)
		for _, pos := range objectPositions {
			model := mgl32.Translate3D(pos[0], pos[1], pos[2]).Mul4(
				mgl32.Scale3D(0.5, 0.5, 0.5))
			shaderGeometryPass.SetMat4("model", model)
			backpack.Draw(shaderGeometryPass)
		}
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// 2. lighting pass: calculate lighting by iterating over a screen filled quad
		// pixel by pixel using the gbuffer's content.
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		shaderLightingPass.Use()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, gPosition)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, gNormal)
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, gAlbedoSpec)
		// Set light relevant uniforms
		for i := 0; i < len(lightPositions); i++ {
			shaderLightingPass.SetVec3(
				fmt.Sprintf("lights[%d].Position", i), lightPositions[i])
			shaderLightingPass.SetVec3(
				fmt.Sprintf("lights[%d].Color", i), lightColors[i])
			// Update attenuation parameters and calculate radius
			constant := float32(1.0) // Don't send to shader we assume it is 1.0
			linear := float32(0.7)
			quadratic := float32(1.8)
			shaderLightingPass.SetFloat(fmt.Sprintf("lights[%d].Linear", i), linear)
			shaderLightingPass.SetFloat(
				fmt.Sprintf("lights[%d].Quadratic", i), quadratic)
			maxBrightness := float32(math.Max(
				math.Max(float64(lightColors[i][0]), float64(lightColors[i][1])),
				float64(lightColors[i][2])))
			radius := -linear + float32(math.Sqrt(float64(
				linear*linear-4*quadratic*
					(constant-(256.0/5.0)*maxBrightness))))/(2.0*quadratic)
			shaderLightingPass.SetFloat(fmt.Sprintf("lights[%d].Radius", i), radius)
		}
		shaderLightingPass.SetVec3("viewPos", ourCamera.Position)
		renderQuad()

		// 2.5. Copy content of geometry's depth buffer to default framebuffer's depth buffer
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, gBuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0) // Write to default framebuffer
		gl.BlitFramebuffer(0, 0, windowWidth, windowHeight, 0, 0,
			windowWidth, windowWidth, gl.DEPTH_BUFFER_BIT, gl.NEAREST)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// 3. Render lights on top of scene
		shaderLightBox.Use()
		shaderLightBox.SetMat4("projection", projection)
		shaderLightBox.SetMat4("view", view)
		for i := 0; i < len(lightPositions); i++ {
			model := mgl32.Translate3D(lightPositions[i][0],
				lightPositions[i][1], lightPositions[i][2]).Mul4(
				mgl32.Scale3D(0.125, 0.125, 0.125))
			shaderLightBox.SetMat4("model", model)
			shaderLightBox.SetVec3("lightColor", lightColors[i])
			renderCube()
		}

		window.SwapBuffers()
	}
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
