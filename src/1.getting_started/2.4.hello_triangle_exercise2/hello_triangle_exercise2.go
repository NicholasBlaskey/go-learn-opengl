// Basically just translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/2.4.hello_triangle_exercise2/hello_triangle_exercise2.cpp

package main

import(
	"runtime"
	"log"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"unsafe"
)

const windowWidth  = 800
const windowHeight = 600

var vertexShaderSource = `
#version 410 core
layout (location = 0) in vec3 position;
void main()
{
    gl_Position = vec4(position.x, position.y, position.z, 1.0);
}
`

var fragmentShaderSource = `
#version 410 core
out vec4 color;
void main()
{
    color = vec4(1.0f, 0.5f, 0.2f, 1.0f);
}
`

func compileShaders() []uint32 {
	// Create the vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	shaderSourceChars, freeVertexShaderFunc := gl.Strs(
		vertexShaderSource)
	gl.ShaderSource(vertexShader, 1, shaderSourceChars, nil)
	gl.CompileShader(vertexShader)

	// Create the fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	shaderSourceChars, freeFragmentShaderFunc := gl.Strs(
		fragmentShaderSource)
	gl.ShaderSource(fragmentShader, 1, shaderSourceChars, nil)
	gl.CompileShader(fragmentShader)

	defer freeFragmentShaderFunc()
	defer freeVertexShaderFunc()

	// Handles error checking
	var success int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &success)
	if success != 1 {
		var infoLog [512]byte
		gl.GetShaderInfoLog(vertexShader, 512, nil,
			(*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln("Vertex shader failed", "\n", string(infoLog[:512]))
	}
	
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &success)
	if success != 1 {
		var infoLog [512]byte
		gl.GetShaderInfoLog(fragmentShader, 512, nil,
			(*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln("Fragment shader failed", "\n", string(infoLog[:512]))
	}

	return []uint32{vertexShader, fragmentShader}
}

func linkShaders(shaders []uint32) uint32 {
	program := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(program, shader)
	}
	gl.LinkProgram(program)

	// Check program link errors
	var success int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success != 1 {
		var infoLog [512]byte
		gl.GetProgramInfoLog(program, 512, nil,
			(*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln("Program link failed", "\n", string(infoLog[:512]))
	}

	// Delete the shaders because we are done with them
	for _, shader := range shaders {
		gl.DeleteShader(shader)
	}


	return program
}

func createTriangleVAO() [2]uint32 {
	firstTriangle := []float32{
		-0.9, -0.5, 0.0, // Left 
		-0.0, -0.5, 0.0, // Right
		-0.45, 0.5, 0.0, // Top
	}

	secondTriangle := []float32{
		0.0, -0.5, 0.0, // Left
		0.9, -0.5, 0.0, // Right
		0.45, 0.5, 0.0, // Top
	}

	var VAOs, VBOs [2]uint32
	
	gl.GenVertexArrays(2, &VAOs[0])
	gl.GenBuffers(2, &VBOs[0])

	// First triangle setup
	gl.BindVertexArray(VAOs[0])
	gl.BindBuffer(gl.ARRAY_BUFFER, VBOs[0])
	gl.BufferData(gl.ARRAY_BUFFER, len(firstTriangle) * 4,
		gl.Ptr(firstTriangle), gl.STATIC_DRAW)

	// Specifies the format of the vertex input
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	// Second triangle setup
	gl.BindVertexArray(VAOs[1])
	gl.BindBuffer(gl.ARRAY_BUFFER, VBOs[1])
	gl.BufferData(gl.ARRAY_BUFFER, len(secondTriangle) * 4,
		gl.Ptr(secondTriangle), gl.STATIC_DRAW)

	// Specifies the format of the vertex input
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)
	
	// Optional to delete VBO
	gl.DeleteBuffers(2, &VBOs[0]);

	return VAOs
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

	shaders := compileShaders()
	shaderProgram := linkShaders(shaders)

	VAOs := createTriangleVAO()
	
	// Program loop
	for !window.ShouldClose() {
		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.5, 0.5, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Drawing loop
		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(VAOs[0])
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

		gl.BindVertexArray(VAOs[1])
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

		window.SwapBuffers()
	}

	// Optional to delete VAO
	gl.DeleteVertexArrays(2, &VAOs[0]);
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