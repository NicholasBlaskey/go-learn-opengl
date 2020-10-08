// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	//	loadModel "github.com/nicholasblaskey/go-learn-opengl/includes/model"
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

	// Build and compile shaders
	ourShader := shader.MakeShaders("1.1.pbr.vs", "1.1.pbr.fs")
	ourShader.Use()
	ourShader.SetVec3("albedo", mgl32.Vec3{0.5, 0.0, 0.0})
	ourShader.SetFloat("ao", 1.0)

	lightPositions := []mgl32.Vec3{
		mgl32.Vec3{-10.0, +10.0, 10.0},
		mgl32.Vec3{+10.0, +10.0, 10.0},
		mgl32.Vec3{-10.0, -10.0, 10.0},
		mgl32.Vec3{+10.0, -10.0, 10.0},
	}
	lightColors := []mgl32.Vec3{
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
	}
	nrRows := 7
	nrCols := 7
	spacing := float32(2.5)

	// Init static shader uniform before rendering
	projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
		float32(windowWidth)/windowHeight, 0.1, 100.0)
	ourShader.SetMat4("projection", projection)

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

		ourShader.Use()
		view := ourCamera.GetViewMatrix()
		ourShader.SetMat4("view", view)
		ourShader.SetVec3("camPos", ourCamera.Position)

		// Render rows * cols number of spehers with varying metallic / roughness
		// values scaled by rows and columns respectively
		for row := 0; row < nrRows; row++ {
			ourShader.SetFloat("metallic", float32(row)/float32(nrRows))
			for col := 0; col < nrCols; col++ {
				// Clamp to 0.025 -1 to avoid perfectly smooth surfaces (roughness = 0.0)
				// which can look off on direct lighting
				ourShader.SetFloat("roughness", mgl32.Clamp(
					float32(col)/float32(nrCols), 0.05, 1.0))
				model := mgl32.Translate3D(
					(float32(col)-(float32(nrCols)/2.0))*spacing,
					(float32(row)-(float32(nrRows)/2.0))*spacing, 0.0)
				ourShader.SetMat4("model", model)
				renderSphere()
			}
		}

		for i := 0; i < len(lightPositions); i++ {
			newPos := lightPositions[i].Add(
				mgl32.Vec3{float32(math.Sin(glfw.GetTime()*5.0)) * 5.0, 0.0, 0.0})
			// newPos = lightPositions[i] // Confused on why overwrite previous assignment
			ourShader.SetVec3(fmt.Sprintf("lightPositions[%d]", i), newPos)
			ourShader.SetVec3(fmt.Sprintf("lightColors[%d]", i), lightColors[i])

			model := mgl32.Translate3D(newPos[0], newPos[1], newPos[2]).Mul4(
				mgl32.Scale3D(0.5, 0.5, 0.5))
			ourShader.SetMat4("model", model)
			renderSphere()
		}

		window.SwapBuffers()
	}
}

var (
	sphereVAO  uint32 = 0
	indexCount uint32
)

func renderSphere() {
	if sphereVAO != 0 {
		gl.BindVertexArray(sphereVAO)
		gl.DrawElements(gl.TRIANGLE_STRIP, int32(indexCount),
			gl.UNSIGNED_INT, unsafe.Pointer(nil))
		return
	}

	gl.GenVertexArrays(1, &sphereVAO)

	var vbo, ebo uint32
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	positions := []mgl32.Vec3{}
	uv := []mgl32.Vec2{}
	normals := []mgl32.Vec3{}
	indices := []uint32{}

	xSegments := 64
	ySegments := 64
	pi := float32(math.Pi)
	for y := 0; y <= ySegments; y++ {
		for x := 0; x <= xSegments; x++ {
			xSegment := float32(x) / float32(xSegments)
			ySegment := float32(y) / float32(ySegments)
			xPos := float32(math.Cos(float64(xSegment*2.0*pi)) *
				math.Sin(float64(ySegment*pi)))
			yPos := float32(math.Cos(float64(ySegment * pi)))
			zPos := float32(math.Sin(float64(xSegment*2.0*pi)) *
				math.Sin(float64(ySegment*pi)))

			positions = append(positions, mgl32.Vec3{xPos, yPos, zPos})
			uv = append(uv, mgl32.Vec2{xSegment, ySegment})
			normals = append(normals, mgl32.Vec3{xPos, yPos, zPos})
		}
	}

	oddRow := false
	for y := 0; y < ySegments; y++ {
		if oddRow {
			for x := 0; x <= xSegments; x++ {
				indices = append(indices, uint32(y*(xSegments+1)+x))
				indices = append(indices, uint32((y+1)*(xSegments+1)+x))
			}
		} else {
			for x := xSegments; x >= 0; x-- {
				indices = append(indices, uint32((y+1)*(xSegments+1)+x))
				indices = append(indices, uint32(y*(xSegments+1)+x))
			}
		}
		oddRow = !oddRow
	}
	indexCount = uint32(len(indices))

	data := []float32{}
	for i := 0; i < len(positions); i++ {
		data = append(data, positions[i][0], positions[i][1], positions[i][2])
		if len(uv) > 0 {
			data = append(data, uv[i][0], uv[i][1])
		}
		if len(normals) > 0 {
			data = append(data, normals[i][0], normals[i][1], normals[i][2])
		}
	}

	gl.BindVertexArray(sphereVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4,
		gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32((3 + 2 + 3) * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))

	renderSphere()
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
