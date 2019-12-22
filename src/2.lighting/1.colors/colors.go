// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/2.lighting/1.colors/colors.cpp

package main

import(
	"runtime"
	"log"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
)

// Settings
const windowWidth  = 800
const windowHeight = 600

// Camera
var ourCamera camera.Camera = camera.NewCamera(
	0.0, 0.0, 3.0, // pos xyz
	0.0, 1.0, 0.0, // up xyz
	-90.0, 0.0,    // Yaw and pitch
	25.0, 45.0, 0.1)   // Speed, zoom, and mouse sensitivity 
var firstMouse bool = true
var lastX float32   = windowWidth / 2
var lastY float32   = windowHeight / 2

// Timing
var deltaTime float32 = 0.0
var lastFrame float32 = 0.0

// Lighting
var lightPos mgl32.Vec3 = mgl32.Vec3{1.2, 1.0, 2.0}

func init() {
	runtime.LockOSThread()
}

func createBuffers() (uint32, uint32, uint32) {
	vertices := []float32{
		-0.5, -0.5, -0.5, 
         0.5, -0.5, -0.5,  
         0.5,  0.5, -0.5,  
         0.5,  0.5, -0.5,  
        -0.5,  0.5, -0.5, 
        -0.5, -0.5, -0.5, 

        -0.5, -0.5,  0.5, 
         0.5, -0.5,  0.5,  
         0.5,  0.5,  0.5,  
         0.5,  0.5,  0.5,  
        -0.5,  0.5,  0.5, 
        -0.5, -0.5,  0.5, 

        -0.5,  0.5,  0.5, 
        -0.5,  0.5, -0.5, 
        -0.5, -0.5, -0.5, 
        -0.5, -0.5, -0.5, 
        -0.5, -0.5,  0.5, 
        -0.5,  0.5,  0.5, 

         0.5,  0.5,  0.5,  
         0.5,  0.5, -0.5,  
         0.5, -0.5, -0.5,  
         0.5, -0.5, -0.5,  
         0.5, -0.5,  0.5,  
         0.5,  0.5,  0.5,  

        -0.5, -0.5, -0.5, 
         0.5, -0.5, -0.5,  
         0.5, -0.5,  0.5,  
         0.5, -0.5,  0.5,  
        -0.5, -0.5,  0.5, 
        -0.5, -0.5, -0.5, 

        -0.5,  0.5, -0.5, 
         0.5,  0.5, -0.5,  
         0.5,  0.5,  0.5,  
         0.5,  0.5,  0.5,  
        -0.5,  0.5,  0.5, 
        -0.5,  0.5, -0.5,
	}
			
	var VBO, cubeVAO uint32		
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindVertexArray(cubeVAO)
	
	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	var lightVAO uint32
	gl.GenVertexArrays(1, &lightVAO)
	gl.BindVertexArray(lightVAO)
	
	// Now configure the light VBO (we already have bound it with the previous one)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	
	return VBO, cubeVAO, lightVAO
}

func configGLFW() *glfw.Window { 
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
         
    if err := gl.Init(); err != nil {
        panic(err)
    }
    window.SetKeyCallback(keyCallback)

    // Config gl global state
    gl.Enable(gl.DEPTH_TEST)

	return window
}

func main() {
	window := configGLFW()
	defer glfw.Terminate()
	
	lightingShader := shader.MakeShaders("1.colors.vs", "1.colors.fs")
	lampShader := shader.MakeShaders("1.lamp.vs", "1.lamp.fs")

	
	VBO, cubeVAO, lightVAO := createBuffers()

	// Optional to delete all of our objects
	defer gl.DeleteVertexArrays(1, &VBO)
	defer gl.DeleteVertexArrays(1, &cubeVAO)
	defer gl.DeleteVertexArrays(1, &lightVAO)
	
	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Poll events and call their registered callbacks
		glfw.PollEvents()

		
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
	
		// Set lighting uniforms
		lightingShader.Use()
		lightingShader.SetVec3("objectColor", mgl32.Vec3{1.0, 0.5, 0.31})
		lightingShader.SetVec3("lightColor", mgl32.Vec3{1.0, 1.0, 1.0})

		// View / projection transformations
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowHeight) / windowWidth, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		lightingShader.SetMat4("projection", projection)
		lightingShader.SetMat4("view", view)
		
		// World transformation
		model := mgl32.Ident4()
		lightingShader.SetMat4("model", model)
		
		// Render the cube
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)

		// Draw the lamp object
		lampShader.Use()
		lampShader.SetMat4("projection", projection)
		lampShader.SetMat4("view", view)
		model = mgl32.Translate3D(lightPos[0], lightPos[1], lightPos[2])
		model = model.Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
		lampShader.SetMat4("model", model)

		gl.BindVertexArray(lightVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		
		window.SwapBuffers()
	}
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	// Escape closes window
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
 
	if key == glfw.KeyW && action == glfw.Press {
		ourCamera.ProcessKeyboard(camera.FORWARD, deltaTime)
	}
	if key == glfw.KeyS && action == glfw.Press {
		ourCamera.ProcessKeyboard(camera.BACKWARD, deltaTime)
	}
	if key == glfw.KeyA && action == glfw.Press {
		ourCamera.ProcessKeyboard(camera.LEFT, deltaTime)
	}
	if key == glfw.KeyD && action == glfw.Press {
		ourCamera.ProcessKeyboard(camera.RIGHT, deltaTime)
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

