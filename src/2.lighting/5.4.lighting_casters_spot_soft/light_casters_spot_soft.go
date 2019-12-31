// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/2.lighting/1.colors/colors.cpp

package main

import(
	"runtime"
	"log"
	"math"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	"github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

// Settings
const windowWidth  = 800
const windowHeight = 600

// Camera
var ourCamera camera.Camera = camera.NewCamera(
	0.0, 0.0, 3.0, // pos xyz
	0.0, 1.0, 0.0, // up xyz
	-90.0, 0.0,    // Yaw and pitch
	80.0, 45.0, 0.1)   // Speed, zoom, and mouse sensitivity 
var firstMouse bool = true
var lastX float32   = windowWidth / 2
var lastY float32   = windowHeight / 2

// Timing
var deltaTime float32 = 0.0
var lastFrame float32 = 0.0

// Lighting
var lightPos mgl32.Vec3 = mgl32.Vec3{1.2, 1.0, 2.0}

// Controls
var heldW bool = false;
var heldA bool = false;
var heldS bool = false;
var heldD bool = false;

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
         
    if err := gl.Init(); err != nil {
        panic(err)
    }
    window.SetKeyCallback(keyCallback)

    // Config gl global state
    gl.Enable(gl.DEPTH_TEST)

	return window
}


func createBuffers() (uint32, uint32, uint32) {
	vertices := []float32{
		// positions          // normals           // texture coords
        -0.5, -0.5, -0.5,  0.0,  0.0, -1.0,  0.0,  0.0,
         0.5, -0.5, -0.5,  0.0,  0.0, -1.0,  1.0,  0.0,
         0.5,  0.5, -0.5,  0.0,  0.0, -1.0,  1.0,  1.0,
         0.5,  0.5, -0.5,  0.0,  0.0, -1.0,  1.0,  1.0,
        -0.5,  0.5, -0.5,  0.0,  0.0, -1.0,  0.0,  1.0,
        -0.5, -0.5, -0.5,  0.0,  0.0, -1.0,  0.0,  0.0,

        -0.5, -0.5,  0.5,  0.0,  0.0,  1.0,  0.0,  0.0,
         0.5, -0.5,  0.5,  0.0,  0.0,  1.0,  1.0,  0.0,
         0.5,  0.5,  0.5,  0.0,  0.0,  1.0,  1.0,  1.0,
         0.5,  0.5,  0.5,  0.0,  0.0,  1.0,  1.0,  1.0,
        -0.5,  0.5,  0.5,  0.0,  0.0,  1.0,  0.0,  1.0,
        -0.5, -0.5,  0.5,  0.0,  0.0,  1.0,  0.0,  0.0,

        -0.5,  0.5,  0.5, -1.0,  0.0,  0.0,  1.0,  0.0,
        -0.5,  0.5, -0.5, -1.0,  0.0,  0.0,  1.0,  1.0,
        -0.5, -0.5, -0.5, -1.0,  0.0,  0.0,  0.0,  1.0,
        -0.5, -0.5, -0.5, -1.0,  0.0,  0.0,  0.0,  1.0,
        -0.5, -0.5,  0.5, -1.0,  0.0,  0.0,  0.0,  0.0,
        -0.5,  0.5,  0.5, -1.0,  0.0,  0.0,  1.0,  0.0,

         0.5,  0.5,  0.5,  1.0,  0.0,  0.0,  1.0,  0.0,
         0.5,  0.5, -0.5,  1.0,  0.0,  0.0,  1.0,  1.0,
         0.5, -0.5, -0.5,  1.0,  0.0,  0.0,  0.0,  1.0,
         0.5, -0.5, -0.5,  1.0,  0.0,  0.0,  0.0,  1.0,
         0.5, -0.5,  0.5,  1.0,  0.0,  0.0,  0.0,  0.0,
         0.5,  0.5,  0.5,  1.0,  0.0,  0.0,  1.0,  0.0,

        -0.5, -0.5, -0.5,  0.0, -1.0,  0.0,  0.0,  1.0,
         0.5, -0.5, -0.5,  0.0, -1.0,  0.0,  1.0,  1.0,
         0.5, -0.5,  0.5,  0.0, -1.0,  0.0,  1.0,  0.0,
         0.5, -0.5,  0.5,  0.0, -1.0,  0.0,  1.0,  0.0,
        -0.5, -0.5,  0.5,  0.0, -1.0,  0.0,  0.0,  0.0,
        -0.5, -0.5, -0.5,  0.0, -1.0,  0.0,  0.0,  1.0,

        -0.5,  0.5, -0.5,  0.0,  1.0,  0.0,  0.0,  1.0,
         0.5,  0.5, -0.5,  0.0,  1.0,  0.0,  1.0,  1.0,
         0.5,  0.5,  0.5,  0.0,  1.0,  0.0,  1.0,  0.0,
         0.5,  0.5,  0.5,  0.0,  1.0,  0.0,  1.0,  0.0,
        -0.5,  0.5,  0.5,  0.0,  1.0,  0.0,  0.0,  0.0,
        -0.5,  0.5, -0.5,  0.0,  1.0,  0.0,  0.0,  1.0,
	}
			
	var VBO, cubeVAO uint32		
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	gl.BindVertexArray(cubeVAO)
	
	// Specify our position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Now specify our normal attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8 * 4,
		gl.PtrOffset(3 * 4))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8 * 4,
		gl.PtrOffset(6 * 4))
	gl.EnableVertexAttribArray(2)
	
	var lightVAO uint32
	gl.GenVertexArrays(1, &lightVAO)
	gl.BindVertexArray(lightVAO)
	
	// Now configure the light VBO (we already have bound it with the previous one)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	
	return VBO, cubeVAO, lightVAO
}

func loadTexture(filePath string) uint32 {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	data := texture.ImageLoad(filePath)

	gl.BindTexture(gl.TEXTURE_2D, textureID)
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

    // Set texture parameters for wrapping
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER,
        gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()
	
	lightingShader := shader.MakeShaders("5.4.light_casters.vs",
		"5.4.light_casters.fs")
	lampShader := shader.MakeShaders("5.4.lamp.vs", "5.4.lamp.fs")
	
	VBO, cubeVAO, lightVAO := createBuffers()

	// Optional to delete all of our objects
	defer gl.DeleteVertexArrays(1, &VBO)
	defer gl.DeleteVertexArrays(1, &cubeVAO)
	defer gl.DeleteVertexArrays(1, &lightVAO)
	
	cubePosition := []mgl32.Vec3{
		mgl32.Vec3{0.0, 0.0, 0.0},
		mgl32.Vec3{ 2.0,  5.0, -15.0},
        mgl32.Vec3{-1.5, -2.2, -2.5},
        mgl32.Vec3{-3.8, -2.0, -12.3},
        mgl32.Vec3{ 2.4, -0.4, -3.5},
        mgl32.Vec3{-1.7,  3.0, -7.5},
        mgl32.Vec3{ 1.3, -2.0, -2.5},
        mgl32.Vec3{ 1.5,  2.0, -2.5},
        mgl32.Vec3{ 1.5,  0.2, -1.5},
        mgl32.Vec3{-1.3,  1.0, -1.5},
	}

	diffuseMap := loadTexture("../../../resources/textures/container2.png")
	specularMap := loadTexture(
		"../../../resources/textures/container2_specular.png")

	lightingShader.Use()
	lightingShader.SetInt("material.diffuse", 0)
	lightingShader.SetInt("material.specular", 1)

	lampShader.Use() // not needed but lets do it anyway
	
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
		lightingShader.SetVec3("light.position", ourCamera.Position)
		lightingShader.SetVec3("light.direction", ourCamera.Front)
		lightingShader.SetFloat("light.cutOff",
			float32(math.Cos(float64(mgl32.DegToRad(12.5)))))
		lightingShader.SetFloat("light.outerCutOff",
			float32(math.Cos(float64(mgl32.DegToRad(17.5)))))
		lightingShader.SetVec3("viewPos", ourCamera.Position)

		// Light properties
		lightingShader.SetVec3("light.ambient", mgl32.Vec3{0.1, 0.1, 0.1})
		// We set diffuse to higher intensity to get a better scene. This
		// requires tweaking to your needs.
		lightingShader.SetVec3("light.diffuse", mgl32.Vec3{0.8, 0.8, 0.8})
		lightingShader.SetVec3("light.specular", mgl32.Vec3{1.0, 1.0, 1.0})
		lightingShader.SetFloat("light.constant", 1.0)
		lightingShader.SetFloat("light.linear", 0.09)
		lightingShader.SetFloat("light.quadratic", 0.032)
		
		// Material properties
		lightingShader.SetFloat("material.shininess", 64.0)
			
		// View / projection transformations
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowHeight) / windowWidth, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		lightingShader.SetMat4("projection", projection)
		lightingShader.SetMat4("view", view)
		
		
		// Activate textures
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, diffuseMap)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, specularMap)

		// World transformation
		model := mgl32.Ident4()
		lightingShader.SetMat4("model", model)
		
		// Render the cubes
		gl.BindVertexArray(cubeVAO)
		for i := 0; i < 10; i++ {
			model := mgl32.Translate3D(
				cubePosition[i][0], cubePosition[i][1], cubePosition[i][2])
			model = model.Mul4(mgl32.HomogRotate3D(mgl32.DegToRad(
				float32(20.0 * i)), mgl32.Vec3{1.0, 0.3, 0.5}))
			lightingShader.SetMat4("model", model)

			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}

		// Draw the lamp object
		/* No need to draw when we have the spotlight
        lampShader.Use()
		lampShader.SetMat4("projection", projection)
		lampShader.SetMat4("view", view)
		model = mgl32.Translate3D(lightPos[0], lightPos[1], lightPos[2])
		model = model.Mul4(mgl32.Scale3D(0.2, 0.2, 0.2))
		lampShader.SetMat4("model", model)
		gl.BindVertexArray(lightVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		
*/
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

