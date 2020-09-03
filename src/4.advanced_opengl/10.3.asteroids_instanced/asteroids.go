// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/4.advanced_opengl/10.3.asteroids_instanced/asteroids_instanced.cpp

package main

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"unsafe"

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

	asteroidShader := shader.MakeShaders("10.3.asteriods.vs",
		"10.3.asteriods.fs")
	planetShader := shader.MakeShaders("10.3.planet.vs",
		"10.3.planet.fs")

	rock := loadModel.NewModel(
		"../../../resources/objects/rock/rock.obj", false)
	planet := loadModel.NewModel(
		"../../../resources/objects/planet/planet.obj", false)

	// Generate large list of semi random model transform matrices
	amount := 10000
	modelMatrices := []mgl32.Mat4{}
	rand.Seed(int64(glfw.GetTime()))
	radius := 150.0
	offset := 25.0
	for i := 0; i < amount; i++ {
		angle := float32(i) / float32(amount) * 360.0
		displacement := float64(rand.Int31()%
			int32(2*offset*100))/100.0 - offset
		x := float32(math.Sin(float64(mgl32.DegToRad(angle)))*
			radius + displacement)
		displacement = float64(rand.Int31()%
			int32(2*offset*100))/100.0 - offset
		y := float32(displacement * 0.4)
		displacement = float64(rand.Int31()%
			int32(2*offset*100))/100.0 - offset
		z := float32(math.Cos(float64(mgl32.DegToRad(angle)))*
			radius + displacement)
		model := mgl32.Translate3D(x, y, z)

		scale := float32(rand.Int31()%20)/100.0 + 0.05
		model = model.Mul4(mgl32.Scale3D(scale, scale, scale))

		rotAngle := float32(mgl32.DegToRad(float32(rand.Int31() % 360)))
		model = model.Mul4(
			mgl32.HomogRotate3D(rotAngle, mgl32.Vec3{0.4, 0.6, 0.8}))

		modelMatrices = append(modelMatrices, model)
	}

	// Config instanced array
	var buffer uint32
	sizeOfMat4 := int32(16 * 4)
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, amount*int(sizeOfMat4),
		unsafe.Pointer(&modelMatrices[0]), gl.STATIC_DRAW)

	// Set transformation matrices as an instance vertex attrib
	// Note we are using a hack by adding new
	// vertexAttribPointers to the model meshes
	for i := 0; i < len(rock.Meshes); i++ {
		VAO := rock.Meshes[i].VAO
		gl.BindVertexArray(VAO)
		// Set attrib points for matrix (4 times vec4)
		gl.EnableVertexAttribArray(3)
		gl.VertexAttribPointer(3, 4, gl.FLOAT, false, sizeOfMat4,
			gl.PtrOffset(0))
		gl.EnableVertexAttribArray(4)
		gl.VertexAttribPointer(4, 4, gl.FLOAT, false, sizeOfMat4,
			gl.PtrOffset(int(sizeOfMat4/4)))
		gl.EnableVertexAttribArray(5)
		gl.VertexAttribPointer(5, 4, gl.FLOAT, false, sizeOfMat4,
			gl.PtrOffset(int(2*sizeOfMat4/4)))
		gl.EnableVertexAttribArray(6)
		gl.VertexAttribPointer(6, 4, gl.FLOAT, false, sizeOfMat4,
			gl.PtrOffset(int(3*sizeOfMat4/4)))

		gl.VertexAttribDivisor(3, 1)
		gl.VertexAttribDivisor(4, 1)
		gl.VertexAttribDivisor(5, 1)
		gl.VertexAttribDivisor(6, 1)

		gl.BindVertexArray(0)
	}

	// Draw in polygon mode
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.05, 0.05, 0.05, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		// Configure transformation matrices
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowHeight)/windowWidth, 0.1, 1000.0)
		view := ourCamera.GetViewMatrix()
		asteroidShader.Use()
		asteroidShader.SetMat4("projection", projection)
		asteroidShader.SetMat4("view", view)
		planetShader.Use()
		planetShader.SetMat4("projection", projection)
		planetShader.SetMat4("view", view)

		// Render the planet
		model := mgl32.Translate3D(0.0, -3.0, 0)
		model = model.Mul4(mgl32.Scale3D(4.0, 4.0, 4.0))
		planetShader.SetMat4("model", model)
		planet.Draw(planetShader)

		// Draw meteorites
		asteroidShader.Use()
		asteroidShader.SetInt("texture_diffuse1", 0)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, rock.TexturesLoaded[0].Id)
		for i := 0; i < len(rock.Meshes); i++ {
			gl.BindVertexArray(rock.Meshes[i].VAO)
			gl.DrawElementsInstanced(gl.TRIANGLES,
				int32(len(rock.Meshes[i].Indices)),
				gl.UNSIGNED_INT, gl.PtrOffset(0), int32(amount))
			gl.BindVertexArray(0)
		}

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