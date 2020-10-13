// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

package main

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	tLoad "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

func glCheckError() {
	for {
		errorCode := gl.GetError()
		if errorCode == gl.NO_ERROR {
			break
		}

		error := ""
		switch errorCode {
		case gl.INVALID_ENUM:
			error = "INVALID_ENUM"
		case gl.INVALID_VALUE:
			error = "INVALID_VALUE"
		case gl.INVALID_OPERATION:
			error = "INVALID_OPERATION"
		case gl.STACK_OVERFLOW:
			error = "STACK_OVERFLOW"
		case gl.STACK_UNDERFLOW:
			error = "STACK_UNDERFLOW"
		case gl.OUT_OF_MEMORY:
			error = "OUT_OF_MEMORY"
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			error = "INVALID_FRAMEBUFFER_OPERATION"
		}
		fmt.Println(error)
	}
}

func glDebugOutput(source, gltype, id, severity uint32, length int32,
	message string, userParam unsafe.Pointer) {

	if id == 131169 || id == 131185 || id == 131218 || id == 13204 {
		return // Ignore these non significant error codes
	}

	fmt.Println("----------------")
	fmt.Printf("Debug message (%d) %s", id, message)
	switch source {
	case gl.DEBUG_SOURCE_API:
		fmt.Println("Source: API")
	case gl.DEBUG_SOURCE_WINDOW_SYSTEM:
		fmt.Println("Source: Window System")
	case gl.DEBUG_SOURCE_SHADER_COMPILER:
		fmt.Println("Source: Shader Compiler")
	case gl.DEBUG_SOURCE_THIRD_PARTY:
		fmt.Println("Source: Third Party")
	case gl.DEBUG_SOURCE_APPLICATION:
		fmt.Println("Source: Application")
	case gl.DEBUG_SOURCE_OTHER:
		fmt.Println("Source: Other")
	}

	switch gltype {
	case gl.DEBUG_TYPE_ERROR:
		fmt.Println("Type: Error")
	case gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR:
		fmt.Println("Type: Deprecated behavior")
	case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
		fmt.Println("Type: Undefined Behaviour")
	case gl.DEBUG_TYPE_PORTABILITY:
		fmt.Println("Type: Portability")
	case gl.DEBUG_TYPE_PERFORMANCE:
		fmt.Println("Type: Performance")
	case gl.DEBUG_TYPE_MARKER:
		fmt.Println("Type: Marker")
	case gl.DEBUG_TYPE_PUSH_GROUP:
		fmt.Println("Type: Push Group")
	case gl.DEBUG_TYPE_POP_GROUP:
		fmt.Println("Type: Pop Group")
	case gl.DEBUG_TYPE_OTHER:
		fmt.Println("Type: Other")
	}

	switch severity {
	case gl.DEBUG_SEVERITY_HIGH:
		fmt.Println("Severity: high")
	case gl.DEBUG_SEVERITY_MEDIUM:
		fmt.Println("Severity: medium")
	case gl.DEBUG_SEVERITY_LOW:
		fmt.Println("Severity: low")
	case gl.DEBUG_SEVERITY_NOTIFICATION:
		fmt.Println("Severity: notification")
	}
}

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
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True) // comment this line in a release build!

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

	var flags int32
	gl.GetIntegerv(gl.CONTEXT_FLAGS, &flags)
	if flags&gl.CONTEXT_FLAG_DEBUG_BIT != 0 {
		fmt.Println("Enabling debug output")

		gl.Enable(gl.DEBUG_OUTPUT)
		gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
		gl.DebugMessageCallback(glDebugOutput, nil)
		gl.DebugMessageControl(gl.DONT_CARE, gl.DONT_CARE, gl.DONT_CARE,
			0, nil, true)
	}

	// Config gl global state
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Build and compile shaders
	ourShader := shader.MakeShaders("debugging.vs", "debugging.fs")

	// Configure 3D cube
	var cubeVAO, cubeVBO uint32
	vertices := []float32{
		// back face
		-0.5, -0.5, -0.5, 0.0, 0.0, // bottom-let
		0.5, 0.5, -0.5, 1.0, 1.0, // top-right
		0.5, -0.5, -0.5, 1.0, 0.0, // bottom-right
		0.5, 0.5, -0.5, 1.0, 1.0, // top-right
		-0.5, -0.5, -0.5, 0.0, 0.0, // bottom-let
		-0.5, 0.5, -0.5, 0.0, 1.0, // top-let
		// ront face
		-0.5, -0.5, 0.5, 0.0, 0.0, // bottom-let
		0.5, -0.5, 0.5, 1.0, 0.0, // bottom-right
		0.5, 0.5, 0.5, 1.0, 1.0, // top-right
		0.5, 0.5, 0.5, 1.0, 1.0, // top-right
		-0.5, 0.5, 0.5, 0.0, 1.0, // top-let
		-0.5, -0.5, 0.5, 0.0, 0.0, // bottom-let
		// let face
		-0.5, 0.5, 0.5, -1.0, 0.0, // top-right
		-0.5, 0.5, -0.5, -1.0, 1.0, // top-let
		-0.5, -0.5, -0.5, -0.0, 1.0, // bottom-let
		-0.5, -0.5, -0.5, -0.0, 1.0, // bottom-let
		-0.5, -0.5, 0.5, -0.0, 0.0, // bottom-right
		-0.5, 0.5, 0.5, -1.0, 0.0, // top-right
		// right face
		0.5, 0.5, 0.5, 1.0, 0.0, // top-let
		0.5, -0.5, -0.5, 0.0, 1.0, // bottom-right
		0.5, 0.5, -0.5, 1.0, 1.0, // top-right
		0.5, -0.5, -0.5, 0.0, 1.0, // bottom-right
		0.5, 0.5, 0.5, 1.0, 0.0, // top-let
		0.5, -0.5, 0.5, 0.0, 0.0, // bottom-let
		// bottom ace
		-0.5, -0.5, -0.5, 0.0, 1.0, // top-right
		0.5, -0.5, -0.5, 1.0, 1.0, // top-let
		0.5, -0.5, 0.5, 1.0, 0.0, // bottom-let
		0.5, -0.5, 0.5, 1.0, 0.0, // bottom-let
		-0.5, -0.5, 0.5, 0.0, 0.0, // bottom-right
		-0.5, -0.5, -0.5, 0.0, 1.0, // top-right
		// top face
		-0.5, 0.5, -0.5, 0.0, 1.0, // top-let
		0.5, 0.5, 0.5, 1.0, 0.0, // bottom-right
		0.5, 0.5, -0.5, 1.0, 1.0, // top-right
		0.5, 0.5, 0.5, 1.0, 0.0, // bottom-right
		-0.5, 0.5, -0.5, 0.0, 1.0, // top-let
		-0.5, 0.5, 0.5, 0.0, 0.0, // bottom-let
	}
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	// Fill buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4,
		gl.Ptr(vertices), gl.STATIC_DRAW)
	// Link vertex attributes
	gl.BindVertexArray(cubeVAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	// Load cube texture
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	data := tLoad.ImageLoad("../../../resources/textures/wood.png")
	gl.TexImage2D(gl.FRAMEBUFFER, 0, gl.RGBA, int32(data.Rect.Size().X),
		int32(data.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE,
		gl.Ptr(data.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Init static shader uniform before rendering
	projection := mgl32.Perspective(mgl32.DegToRad(45.0),
		float32(windowWidth)/windowHeight, 0.1, 100.0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(ourShader.ID, gl.Str("projection\x00")),
		1, false, &projection[0])
	gl.Uniform1i(gl.GetUniformLocation(ourShader.ID, gl.Str("tex\x00")), 0)

	// Program loop
	for !window.ShouldClose() {
		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		ourShader.Use()
		rotationSpeed := 10.0
		angle := float32(glfw.GetTime() * rotationSpeed)
		model := mgl32.Translate3D(0.0, 0.0, -2.5).Mul4(
			mgl32.HomogRotate3D(mgl32.DegToRad(angle),
				mgl32.Vec3{1.0, 1.0, 1.0}))
		gl.UniformMatrix4fv(gl.GetUniformLocation(ourShader.ID, gl.Str("model\x00")),
			1, false, &model[0])

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)

		window.SwapBuffers()
	}
}

var (
	sphereVAO  uint32 = 0
	quadVAO    uint32 = 0
	quadVBO    uint32 = 0
	indexCount uint32
)

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
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
