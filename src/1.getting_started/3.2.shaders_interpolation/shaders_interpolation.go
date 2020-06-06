// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/3.2.shaders_interpolation/shaders_interpolation.cpp

package main

import (
	"log"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const windowWidth = 800
const windowHeight = 600

var vertexShaderSource = `
#version 410 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec3 aColor;
out vec3 ourColor;
void main()
{
    gl_Position = vec4(aPos, 1.0);
    ourColor = aColor;  
}
`

var fragmentShaderSource = `
#version 410 core
out vec4 FragColor;
in vec3 ourColor;
void main()
{
    FragColor = vec4(ourColor, 1.0f);
}
` + "\x00"

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

func createTriangleVAO() uint32 {
	vertices := []float32{
		// Positions        // Colors
		-0.5, -0.5, 0.0, 1.0, 0.0, 0.0, // Bottom right
		0.5, -0.5, 0.0, 0.0, 1.0, 0.0, // Bottom left
		0.0, 0.5, 0.0, 0.0, 0.0, 1.0, // Top
	}

	var VAO uint32
	gl.GenVertexArrays(1, &VAO)

	var VBO uint32
	gl.GenBuffers(1, &VBO)

	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Specify our color attributes
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	// Optional to delete VBO
	gl.DeleteBuffers(1, &VBO)

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

	shaders := compileShaders()
	shaderProgram := linkShaders(shaders)

	VAO := createTriangleVAO()

	// Program loop
	for !window.ShouldClose() {
		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.5, 0.5, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Drawing loop
		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

		window.SwapBuffers()
	}

	// Optional to delete VAO
	gl.DeleteVertexArrays(1, &VAO)
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
