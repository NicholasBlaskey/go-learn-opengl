// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/6.1.coordinate_systems/coordinate_systems.cpp

package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	"github.com/nicholasblaskey/go-learn-opengl/includes/texture"

	"github.com/disintegration/imaging"
)

const windowWidth = 800
const windowHeight = 600

func createTriangleObjects() (uint32, uint32, uint32) {
	vertices := []float32{
		//Positions      // Texture coords
		0.5, 0.5, 0.0, 1.0, 1.0, // Top right
		0.5, -0.5, 0.0, 1.0, 0.0, // Bottom right
		-0.5, -0.5, 0.0, 0.0, 0.0, // Bottom left
		-0.5, 0.5, 0.0, 0.0, 1.0, // Top left
	}
	indices := []uint32{
		0, 1, 3, // First triangle
		1, 2, 3, // Second triangle
	}

	var VAO, VBO, EBO uint32

	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &EBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4,
		gl.Ptr(indices), gl.STATIC_DRAW)

	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coord attributes
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4,
		gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	return VAO, VBO, EBO
}

func init() {
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to init glfw:", err)
	}
	defer glfw.Terminate()

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

	if err := gl.Init(); err != nil {
		panic(err)
	}

	window.SetKeyCallback(keyCallback)

	ourShader := shader.MakeShaders("6.1.coordinate_systems.vs", "6.1.coordinate_systems.fs")

	VBO, VAO, EBO := createTriangleObjects()

	// Optional to delete all of our objects
	defer gl.DeleteVertexArrays(1, &VBO)
	defer gl.DeleteVertexArrays(1, &VAO)
	defer gl.DeleteVertexArrays(1, &EBO)

	// Load and create our textures
	var texture1ID, texture2ID uint32
	gl.GenTextures(1, &texture1ID)
	gl.BindTexture(gl.TEXTURE_2D, texture1ID)
	// Set texture parameters for wrapping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// Set texture filtering parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	data := texture.ImageLoad("../../../resources/textures/container.jpg")

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(data.Rect.Size().X),
		int32(data.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(data.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	gl.GenTextures(1, &texture2ID)
	gl.BindTexture(gl.TEXTURE_2D, texture2ID)
	// Set texture parameters for wrapping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	// Set texture filtering parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	data = texture.ImageLoad("../../../resources/textures/awesomeface.png")
	flippedData := imaging.FlipV(data)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(data.Rect.Size().X),
		int32(data.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(flippedData.Pix))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	ourShader.Use()
	ourShader.SetInt("texture1", 0)
	ourShader.SetInt("texture2", 1)

	// Program loop
	for !window.ShouldClose() {
		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Bind textures
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture1ID)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, texture2ID)

		// Activate shader
		ourShader.Use()

		// Create our matrixes to transform with
		model := mgl32.HomogRotate3D(mgl32.DegToRad(-55),
			mgl32.Vec3{1.0, 0.0, 0.0})
		view := mgl32.Translate3D(0.0, 0.0, -3.0)
		projection := mgl32.Perspective(mgl32.DegToRad(45.0),
			float32(windowHeight)/windowWidth, 0.1, 100.0)

		// Get the matrix location and set the matrix in shader program
		modelLoc := gl.GetUniformLocation(ourShader.ID,
			gl.Str("model"+"\x00"))
		viewLoc := gl.GetUniformLocation(ourShader.ID,
			gl.Str("view"+"\x00"))
		projLoc := gl.GetUniformLocation(ourShader.ID,
			gl.Str("projection"+"\x00"))

		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])
		gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
		gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

		// Drawing loop
		gl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT,
			unsafe.Pointer(nil))
		gl.BindVertexArray(0)

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
