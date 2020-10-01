// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

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

func makeCubeBuffers() (uint32, uint32) {
	planeVertices := []float32{
		// positions            // normals         // texcoords
		25.0, -0.5, 25.0, 0.0, 1.0, 0.0, 25.0, 0.0,
		-25.0, -0.5, 25.0, 0.0, 1.0, 0.0, 0.0, 0.0,
		-25.0, -0.5, -25.0, 0.0, 1.0, 0.0, 0.0, 25.0,

		25.0, -0.5, 25.0, 0.0, 1.0, 0.0, 25.0, 0.0,
		-25.0, -0.5, -25.0, 0.0, 1.0, 0.0, 0.0, 25.0,
		25.0, -0.5, -25.0, 0.0, 1.0, 0.0, 25.0, 25.0,
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

	ourShader := shader.MakeShaders("5.1.parallax_mapping.vs", "5.1.parallax_mapping.fs")

	dir := "../../../resources/textures"
	diffuseMap := loadModel.TextureFromFile("bricks2.jpg", dir, false)
	normalMap := loadModel.TextureFromFile("bricks2_normal.jpg", dir, false)
	heightMap := loadModel.TextureFromFile("bricks2_disp.jpg", dir, false)

	ourShader.Use()
	ourShader.SetInt("diffuseMap", 0)
	ourShader.SetInt("normalMap", 1)
	ourShader.SetInt("depthMap", 2)

	lightPos := mgl32.Vec3{0.5, 1.0, 0.3}

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

		// configure view / projection matrices
		ourShader.Use()
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowWidth)/windowHeight, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		ourShader.SetMat4("projection", projection)
		ourShader.SetMat4("view", view)
		// render normal-mapped quad
		model := mgl32.HomogRotate3D(
			mgl32.DegToRad(float32(glfw.GetTime())*-10.0),
			mgl32.Vec3{1.0, 0.0, 1.0}.Normalize())
		ourShader.SetMat4("model", model)
		ourShader.SetVec3("viewPos", ourCamera.Position)
		ourShader.SetVec3("lightPos", lightPos)
		ourShader.SetFloat("heightScale", heightScale)
		fmt.Println(heightScale)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, diffuseMap)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, normalMap)
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, heightMap)
		renderQuad()

		// Render light source (simply re-renders a smaller plane at the light pos for debug)
		ourShader.SetMat4("model",
			mgl32.Translate3D(lightPos[0], lightPos[1], lightPos[2]).Mul4(
				mgl32.Scale3D(0.1, 0.1, 0.1)))
		renderQuad()

		window.SwapBuffers()
	}
}

var (
	quadVAO uint32
	quadVBO uint32
)

func renderQuad() {
	if quadVAO != 0 {
		gl.BindVertexArray(quadVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		gl.BindVertexArray(0)
		return
	}

	// Positions
	pos1 := mgl32.Vec3{-1.0, 1.0, 0.0}
	pos2 := mgl32.Vec3{-1.0, -1.0, 0.0}
	pos3 := mgl32.Vec3{1.0, -1.0, 0.0}
	pos4 := mgl32.Vec3{1.0, 1.0, 0.0}
	// Texture coords
	uv1 := mgl32.Vec2{0.0, 1.0}
	uv2 := mgl32.Vec2{0.0, 0.0}
	uv3 := mgl32.Vec2{1.0, 0.0}
	uv4 := mgl32.Vec2{1.0, 1.0}
	// Normal vector
	nm := mgl32.Vec3{0.0, 0.0, 1.0}

	// Calculate tangent / bitangent vectors of both triangles
	var tangent1, bitangent1 mgl32.Vec3
	var tangent2, bitangent2 mgl32.Vec3
	// Triangle 1
	edge1 := pos2.Sub(pos1)
	edge2 := pos3.Sub(pos1)
	deltaUV1 := uv2.Sub(uv1)
	deltaUV2 := uv3.Sub(uv1)

	f := 1.0 / (deltaUV1[0]*deltaUV2[1] - deltaUV2[0]*deltaUV1[1])
	tangent1[0] = f * (deltaUV2[1]*edge1[0] - deltaUV1[1]*edge2[0])
	tangent1[1] = f * (deltaUV2[1]*edge1[1] - deltaUV1[1]*edge2[1])
	tangent1[2] = f * (deltaUV2[1]*edge1[2] - deltaUV1[1]*edge2[2])

	bitangent1[0] = f * (-deltaUV2[0]*edge1[0] + deltaUV1[0]*edge2[0])
	bitangent1[1] = f * (-deltaUV2[0]*edge1[1] + deltaUV1[0]*edge2[1])
	bitangent1[2] = f * (-deltaUV2[0]*edge1[2] + deltaUV1[0]*edge2[2])

	// triangle 2
	edge1 = pos3.Sub(pos1)
	edge2 = pos4.Sub(pos1)
	deltaUV1 = uv3.Sub(uv1)
	deltaUV2 = uv4.Sub(uv1)

	f = 1.0 / (deltaUV1[0]*deltaUV2[1] - deltaUV2[0]*deltaUV1[1])
	tangent2[0] = f * (deltaUV2[1]*edge1[0] - deltaUV1[1]*edge2[0])
	tangent2[1] = f * (deltaUV2[1]*edge1[1] - deltaUV1[1]*edge2[1])
	tangent2[2] = f * (deltaUV2[1]*edge1[2] - deltaUV1[1]*edge2[2])

	bitangent2[0] = f * (-deltaUV2[0]*edge1[0] + deltaUV1[0]*edge2[0])
	bitangent2[1] = f * (-deltaUV2[0]*edge1[1] + deltaUV1[0]*edge2[1])
	bitangent2[2] = f * (-deltaUV2[0]*edge1[2] + deltaUV1[0]*edge2[2])

	quadVertices := []float32{
		pos1[0], pos1[1], pos1[2], nm[0], nm[1], nm[2], uv1[0], uv1[1],
		tangent1[0], tangent1[1], tangent1[2], bitangent1[0], bitangent1[1], bitangent1[2],

		pos2[0], pos2[1], pos2[2], nm[0], nm[1], nm[2], uv2[0], uv2[1],
		tangent1[0], tangent1[1], tangent1[2], bitangent1[0], bitangent1[1], bitangent1[2],

		pos3[0], pos3[1], pos3[2], nm[0], nm[1], nm[2], uv3[0], uv3[1],
		tangent1[0], tangent1[1], tangent1[2], bitangent1[0], bitangent1[1], bitangent1[2],

		pos1[0], pos1[1], pos1[2], nm[0], nm[1], nm[2], uv1[0], uv1[1],
		tangent2[0], tangent2[1], tangent2[2], bitangent2[0], bitangent2[1], bitangent2[2],

		pos3[0], pos3[1], pos3[2], nm[0], nm[1], nm[2], uv3[0], uv3[1],
		tangent2[0], tangent2[1], tangent2[2], bitangent2[0], bitangent2[1], bitangent2[2],

		pos4[0], pos4[1], pos4[2], nm[0], nm[1], nm[2], uv4[0], uv4[1],
		tangent2[0], tangent2[1], tangent2[2], bitangent2[0], bitangent2[1], bitangent2[2],
	}

	gl.GenVertexArrays(1, &quadVAO)
	gl.GenBuffers(1, &quadVBO)
	gl.BindVertexArray(quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4,
		gl.Ptr(quadVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 14*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 14*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 14*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, 14*4, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 3, gl.FLOAT, false, 14*4, gl.PtrOffset(11*4))

	renderQuad()
}

var heightScale float32 = 0.10

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

	if key == glfw.KeyQ && action == glfw.Press {
		if heightScale > 0.0 {
			heightScale -= 0.05
		} else {
			heightScale += 0.0
		}
	} else if key == glfw.KeyE && action == glfw.Press {
		if heightScale < 1.0 {
			heightScale += 0.05
		} else {
			heightScale = 1.0
		}
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
