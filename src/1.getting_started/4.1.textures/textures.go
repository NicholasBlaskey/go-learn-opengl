// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/4.1.textures/textures.cpp

package main

import(
	"runtime"
	"log"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-learn-opengl/includes/shader"
	"github.com/go-learn-opengl/includes/texture"
	"unsafe"
)

const windowWidth  = 800
const windowHeight = 600

func createTriangleObjects() (uint32, uint32, uint32) {
	vertices := []float32{
		//Positions      // Colors       // Texture coords
		0.5, 0.5, 0.0,   1.0, 0.0, 0.0,  1.0, 1.0, // Top right
		0.5, -0.5, 0.0,  0.0, 1.0, 0.0,  1.0, 0.0, // Bottom right
		-0.5, -0.5, 0.0, 0.0, 0.0, 1.0,  0.0, 0.0, // Bottom left
		-0.5, 0.5, 0.0,  1.0, 1.0, 0.0,  0.0, 1.0, // Top left 
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
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices) * 4,
		gl.Ptr(indices), gl.STATIC_DRAW)

	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Specify our color attributes
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8 * 4,
		gl.PtrOffset(3 * 4))
	gl.EnableVertexAttribArray(1)
	// Texture coord attributes
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8 * 4,
		gl.PtrOffset(6 * 4))
	gl.EnableVertexAttribArray(2)
	
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

	ourShader := shader.MakeShaders("4.1.texture.vs", "4.1.texture.fs")
	
	VBO, VAO, EBO := createTriangleObjects()
	
	// Optional to delete all of our objects
	defer gl.DeleteVertexArrays(1, &VBO);
	defer gl.DeleteVertexArrays(1, &VAO);
	defer gl.DeleteVertexArrays(1, &EBO);

	// Load and create our textures
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)
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
	
	
	// Program loop
	for !window.ShouldClose() {
		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Bind texture
		gl.BindTexture(gl.TEXTURE_2D, textureID)
		
		// Drawing loop
		ourShader.Use()
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
