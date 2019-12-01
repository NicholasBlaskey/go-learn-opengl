// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/3.3.shaders_class/shaders_class.cpp

package main

import(
	"runtime"
	"log"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-learn-opengl/includes/shader"
)

const windowWidth  = 800
const windowHeight = 600

func createTriangleVAO() uint32 {
	vertices := []float32{
		// Positions        // Colors
		-0.5, -0.5, 0.0, 1.0, 0.0, 0.0, // Bottom right
		0.5, -0.5, 0.0,  0.0, 1.0, 0.0, // Bottom left
		0.0, 0.5, 0.0,   0.0, 0.0, 1.0, // Top
	}

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Specify our color attributes
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6 * 4, gl.PtrOffset(3 * 4))
	gl.EnableVertexAttribArray(1)


	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	// Optional to delete VBO
	gl.DeleteBuffers(1, &VBO);

	return VAO
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


	ourShader := shader.MakeShaders("3.5.shader.vs", "3.5.shader.fs")
	var offset float32 = 0.5

	// Note we need to use the program before we can set uniforms...
	ourShader.Use() 
	ourShader.SetFloat("xOffset\x00", offset) 
	
	VAO := createTriangleVAO()

	// Optional to delete VAO
	defer gl.DeleteVertexArrays(1, &VAO);
	
	// Program loop
	for !window.ShouldClose() {
		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.5, 0.5, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Drawing loop
		ourShader.Use()
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
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

