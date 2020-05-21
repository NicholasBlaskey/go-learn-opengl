// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/2.lighting/1.colors/colors.cpp

package main

import(
	"runtime"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
)

// Settings
const windowWidth  = 800
const windowHeight = 600

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

	window.SetFramebufferSizeCallback(
		glfw.FramebufferSizeCallback(framebuffer_size_callback))
	window.SetKeyCallback(keyCallback)

	
    if err := gl.Init(); err != nil {
        panic(err)
    }

    // Config gl global state
    gl.Enable(gl.DEPTH_TEST)

	return window
}

func makeBuffers() (uint32, uint32) {
	Vertices := []float32{
		-0.5,  0.5, 1.0, 0.0, 0.0, // top-let
         0.5,  0.5, 0.0, 1.0, 0.0, // top-right
         0.5, -0.5, 0.0, 0.0, 1.0, // bottom-right
        -0.5, -0.5, 1.0, 1.0, 0.0,  // bottom-let
	}
	//  VAO
	var VBO, VAO uint32		
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(Vertices) * 4,
		gl.Ptr(Vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)	
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)	
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5 * 4, gl.PtrOffset(2 * 4))
	gl.BindVertexArray(0)

	return VBO, VAO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Create shaders
	ourShader := shader.MakeGeomShaders("9.1.geometry_shader.vs",
		"9.1.geometry_shader.fs", "9.1.geometry_shader.gs")
	
	VAO, VBO := makeBuffers()
	defer gl.DeleteVertexArrays(1, &VAO)
	defer gl.DeleteVertexArrays(1, &VBO)
	
	// Program loop
	for !window.ShouldClose() {
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)

		ourShader.Use()
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.POINTS, 0, 4)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}
