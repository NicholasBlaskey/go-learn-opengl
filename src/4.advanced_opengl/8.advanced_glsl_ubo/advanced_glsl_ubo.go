// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/4.advanced_opengl/8.advanced_glsl_ubo/advanced_glsl_ubo.cpp

package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	loadTexture "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
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
	//window.SetScrollCallback(glfw.ScrollCallback(scroll_callback))

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

func makeCubeBuffers() (uint32, uint32) {
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

	return cubeVBO, cubeVAO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Create shaders
	shaderRed := shader.MakeShaders("8.advanced_glsl.vs", "8.red.fs")
	shaderGreen := shader.MakeShaders("8.advanced_glsl.vs", "8.green.fs")
	shaderBlue := shader.MakeShaders("8.advanced_glsl.vs", "8.blue.fs")
	shaderYellow := shader.MakeShaders("8.advanced_glsl.vs", "8.yellow.fs")

	cubeVAO, cubeVBO := makeCubeBuffers()
	defer gl.DeleteVertexArrays(1, &cubeVAO)
	defer gl.DeleteVertexArrays(1, &cubeVBO)

	// Configure uniform buffer object
	// First get relevant block indices
	uniformBlockRed := gl.GetUniformBlockIndex(shaderRed.ID,
		gl.Str("Matrices"+"\x00"))
	uniformBlockGreen := gl.GetUniformBlockIndex(shaderGreen.ID,
		gl.Str("Matrices"+"\x00"))
	uniformBlockBlue := gl.GetUniformBlockIndex(shaderBlue.ID,
		gl.Str("Matrices"+"\x00"))
	uniformBlockYellow := gl.GetUniformBlockIndex(shaderYellow.ID,
		gl.Str("Matrices"+"\x00"))
	// Then link each shader's uniform block to this uniform binding point
	gl.UniformBlockBinding(shaderRed.ID, uniformBlockRed, 0)
	gl.UniformBlockBinding(shaderGreen.ID, uniformBlockGreen, 0)
	gl.UniformBlockBinding(shaderBlue.ID, uniformBlockBlue, 0)
	gl.UniformBlockBinding(shaderYellow.ID, uniformBlockYellow, 0)
	// Now actually make the buffer
	var uboMatrices uint32
	sizeOfMat4 := 16 * 4
	gl.GenBuffers(1, &uboMatrices)
	gl.BindBuffer(gl.UNIFORM_BUFFER, uboMatrices)
	gl.BufferData(gl.UNIFORM_BUFFER, 2*sizeOfMat4, nil, gl.STATIC_DRAW)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	// Define the range of the buffer that links to a uniform binding point
	gl.BindBufferRange(gl.UNIFORM_BUFFER, 0, uboMatrices, 0, 2*sizeOfMat4)

	// Store the project matrix (not using zoom so only need to do it once)
	projection := mgl32.Perspective(mgl32.DegToRad(45.0),
		float32(windowHeight)/windowWidth, 0.1, 100.0)
	gl.BindBuffer(gl.UNIFORM_BUFFER, uboMatrices)
	gl.BufferSubData(gl.UNIFORM_BUFFER, 0, sizeOfMat4, unsafe.Pointer(&projection[0]))
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)

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

		view := ourCamera.GetViewMatrix()
		gl.BindBuffer(gl.UNIFORM_BUFFER, uboMatrices)
		gl.BufferSubData(gl.UNIFORM_BUFFER, sizeOfMat4,
			sizeOfMat4, unsafe.Pointer(&view[0]))
		gl.BindBuffer(gl.UNIFORM_BUFFER, 0)

		// Draw 4 cubes
		// Red
		gl.BindVertexArray(cubeVAO)
		shaderRed.Use()
		model := mgl32.Translate3D(-0.75, 0.75, 0.0)
		shaderRed.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		// Green
		gl.BindVertexArray(cubeVAO)
		shaderGreen.Use()
		model = mgl32.Translate3D(0.75, 0.75, 0.0)
		shaderGreen.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		// Yellow
		gl.BindVertexArray(cubeVAO)
		shaderYellow.Use()
		model = mgl32.Translate3D(-0.75, -0.75, 0.0)
		shaderYellow.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		// Blue
		gl.BindVertexArray(cubeVAO)
		shaderBlue.Use()
		model = mgl32.Translate3D(0.75, -0.75, 0.0)
		shaderBlue.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

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

/*
func scroll_callback(w *glfw.Window, xOffset float64, yOffset float64) {
	ourCamera.ProcessMouseScroll(float32(yOffset))
}
*/

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func loadCubemap(faces []string) uint32 {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, textureID)

	for i := uint32(0); i < uint32(len(faces)); i++ {
		data := loadTexture.ImageLoad(faces[i])
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i, 0, gl.RGBA,
			int32(data.Rect.Size().X), int32(data.Rect.Size().Y),
			0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.Pix))
	}
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S,
		gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T,
		gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R,
		gl.CLAMP_TO_EDGE)

	return textureID
}
