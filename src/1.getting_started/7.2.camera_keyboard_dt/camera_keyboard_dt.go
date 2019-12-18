// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/7.2.camera_keyboard_dt/camera_keyboard_dt.cpp

package main

import(
	"runtime"
	"log"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	
	"github.com/disintegration/imaging"
	
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	"github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

const windowWidth  = 800
const windowHeight = 600

var cameraPos   mgl32.Vec3 = mgl32.Vec3{0.0, 0.0, 3.0}
var cameraFront mgl32.Vec3 = mgl32.Vec3{0.0, 0.0, -1.0}
var cameraUp    mgl32.Vec3 = mgl32.Vec3{0.0, 1.0, 0.0}

var deltaTime float32 = 0.0
var lastFrame float32 = 0.0

var heldW bool = false
var heldA bool = false
var heldS bool = false
var heldD bool = false

func createTriangleObjects() (uint32, uint32) {
	vertices := []float32{
		-0.5, -0.5, -0.5,  0.0, 0.0,
		0.5, -0.5, -0.5,  1.0, 0.0,
		0.5,  0.5, -0.5,  1.0, 1.0,
		0.5,  0.5, -0.5,  1.0, 1.0,
		-0.5,  0.5, -0.5,  0.0, 1.0,
		-0.5, -0.5, -0.5,  0.0, 0.0,

		-0.5, -0.5,  0.5,  0.0, 0.0,
		0.5, -0.5,  0.5,  1.0, 0.0,
		0.5,  0.5,  0.5,  1.0, 1.0,
		0.5,  0.5,  0.5,  1.0, 1.0,
		-0.5,  0.5,  0.5,  0.0, 1.0,
		-0.5, -0.5,  0.5,  0.0, 0.0,

		-0.5,  0.5,  0.5,  1.0, 0.0,
		-0.5,  0.5, -0.5,  1.0, 1.0,
		-0.5, -0.5, -0.5,  0.0, 1.0,
		-0.5, -0.5, -0.5,  0.0, 1.0,
		-0.5, -0.5,  0.5,  0.0, 0.0,
		-0.5,  0.5,  0.5,  1.0, 0.0,
		
		0.5,  0.5,  0.5,  1.0, 0.0,
		0.5,  0.5, -0.5,  1.0, 1.0,
		0.5, -0.5, -0.5,  0.0, 1.0,
		0.5, -0.5, -0.5,  0.0, 1.0,
		0.5, -0.5,  0.5,  0.0, 0.0,
		0.5,  0.5,  0.5,  1.0, 0.0,
		
		-0.5, -0.5, -0.5,  0.0, 1.0,
		0.5, -0.5, -0.5,  1.0, 1.0,
		0.5, -0.5,  0.5,  1.0, 0.0,
		0.5, -0.5,  0.5,  1.0, 0.0,
		-0.5, -0.5,  0.5,  0.0, 0.0,
		-0.5, -0.5, -0.5,  0.0, 1.0,
		
		-0.5,  0.5, -0.5,  0.0, 1.0,
		0.5,  0.5, -0.5,  1.0, 1.0,
		0.5,  0.5,  0.5,  1.0, 0.0,
		0.5,  0.5,  0.5,  1.0, 0.0,
		-0.5,  0.5,  0.5,  0.0, 0.0,
		-0.5,  0.5, -0.5,  0.0, 1.0,
	}
	
	var VAO, VBO uint32		
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	
	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices) * 4, gl.Ptr(vertices),
		gl.STATIC_DRAW)

	// Specify our position attributes
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// Texture coord attributes
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5 * 4,
		gl.PtrOffset(3 * 4))
	gl.EnableVertexAttribArray(1)
	
	// Unbind our vertex array so we don't mess with it later
	gl.BindVertexArray(0)

	return VAO, VBO
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

	// Config gl global state
	gl.Enable(gl.DEPTH_TEST)
	
	ourShader := shader.MakeShaders("7.2.camera.vs", "7.2.camera.fs")
	
	VBO, VAO := createTriangleObjects()

	// World space positions of our cubes
	cubePositions := []mgl32.Vec3{
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
	
	// Optional to delete all of our objects
	defer gl.DeleteVertexArrays(1, &VBO);
	defer gl.DeleteVertexArrays(1, &VAO);

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

	// Create and set our project matrix in advance since it will rarely change
	projection := mgl32.Perspective(mgl32.DegToRad(45.0),
		float32(windowHeight) / windowWidth, 0.1, 100.0)
	projLoc := gl.GetUniformLocation(ourShader.ID,
		gl.Str("projection" + "\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])
		
	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Poll events and call their registered callbacks
		glfw.PollEvents()

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		
		// Bind textures
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture1ID)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, texture2ID)

		// Activate shader
		ourShader.Use()
		
		// Camera / view transformation
		view := mgl32.LookAtV(cameraPos, cameraPos.Add(cameraFront), cameraUp)
		
		viewLoc := gl.GetUniformLocation(ourShader.ID,
			gl.Str("view" + "\x00"))		
		gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])


		// Drawing loop
		gl.BindVertexArray(VAO)
		for i := 0; i < 10; i++ {
			model := mgl32.Translate3D(cubePositions[i][0],
				cubePositions[i][1], cubePositions[i][2])

			model = model.Mul4(mgl32.HomogRotate3D(
				mgl32.DegToRad(float32(20 * i)),
				mgl32.Vec3{1.0, 0.3, 0.5}.Normalize()))
			
			modelLoc := gl.GetUniformLocation(ourShader.ID,
				gl.Str("model" + "\x00"))
			gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

			gl.DrawArrays(gl.TRIANGLES, 0, 36)
		}
		
		
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
 
	cameraSpeed := 20.0 * deltaTime
	if key == glfw.KeyW && action == glfw.Press || heldW {
		cameraPos = cameraPos.Add(cameraFront.Mul(cameraSpeed))
		heldW = true
	}
	if key == glfw.KeyS && action == glfw.Press || heldS {
		cameraPos = cameraPos.Sub(cameraFront.Mul(cameraSpeed))
		heldS = true
	}
	if key == glfw.KeyA && action == glfw.Press || heldA {
		cameraPos = cameraPos.Sub(cameraFront.Cross(
			cameraUp).Normalize().Mul(cameraSpeed))
		heldA = true
	}
	if key == glfw.KeyD && action == glfw.Press || heldD {
		cameraPos = cameraPos.Add(cameraFront.Cross(
			cameraUp).Normalize().Mul(cameraSpeed))
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

func framebuffer_size_callback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

