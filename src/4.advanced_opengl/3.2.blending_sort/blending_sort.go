// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/2.lighting/1.colors/colors.cpp

package main

import(
	"runtime"
	"log"
	"sort"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	loadTexture "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
	loadModel "github.com/nicholasblaskey/go-learn-opengl/includes/model"
)

// Settings
const windowWidth  = 1280
const windowHeight = 720

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

//https://golang.org/pkg/sort/ using this example (Example (sortkeys))
type By func(w1, w2 *mgl32.Vec3) bool
func (by By) Sort(windows []mgl32.Vec3) {
	ws := &windowSorter{
		windows: []mgl32.Vec3,
		by:      by,
	}
	sort.Sort(ws)
}
type windowSorter struct {
	planets []mgl32.Vec3
	by      func(w1, w1 *mlg32.Vec3) bool
}
func (w *windowSorter) Len() int {
	return len(w.windows)
}
func (s *windowSorter) Swap(i, j int) {
	s.windows[i], s.windows[j] = s.windows[j], s.windows[i]
}
func (s *windowSorter) Less(i, j int) bool {
	return s.by(&s.windows[i], &s.windows[j])
}

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
	gl.Enable(gl.BLEND)

	return window
}

func makeCubeBuffers() (uint32, uint32, uint32, uint32, uint32, uint32) {
	cubeVertices := []float32{
        // positions       // texture Coords
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
	planeVertices := []float32{
		// Positions      // texture coords
		5.0, -0.5,  5.0,  2.0, 0.0,
        -5.0, -0.5,  5.0,  0.0, 0.0,
        -5.0, -0.5, -5.0,  0.0, 2.0,
		
         5.0, -0.5,  5.0,  2.0, 0.0,
        -5.0, -0.5, -5.0,  0.0, 2.0,
		5.0, -0.5, -5.0,  2.0, 2.0,
	}
	transparentVertices := []float32{
		// positions         // texture Coords (swap y to flip texture) 
        0.0,  0.5,  0.0,  0.0,  0.0,
        0.0, -0.5,  0.0,  0.0,  1.0,
        1.0, -0.5,  0.0,  1.0,  1.0,

        0.0,  0.5,  0.0,  0.0,  0.0,
        1.0, -0.5,  0.0,  1.0,  1.0,
        1.0,  0.5,  0.0,  1.0,  0.0,
	}
	// cube VAO
	var cubeVBO, cubeVAO uint32		
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVertices) * 4,
		gl.Ptr(cubeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)	
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5 * 4,
		gl.PtrOffset(3 * 4))
	gl.BindVertexArray(0)
	// plane VAO
	var planeVAO, planeVBO uint32
	gl.GenVertexArrays(1, &planeVAO)
	gl.GenBuffers(1, &planeVBO)
	gl.BindVertexArray(planeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, planeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices) * 4,
		gl.Ptr(planeVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)	
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5 * 4,
		gl.PtrOffset(3 * 4))
	gl.BindVertexArray(0)
	// transparent VAO
	var transparentVAO, transparentVBO uint32
	gl.GenVertexArrays(1, &transparentVAO)
	gl.GenBuffers(1, &transparentVBO)
	gl.BindVertexArray(transparentVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, transparentVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(transparentVertices) * 4,
		gl.Ptr(transparentVertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)	
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5 * 4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5 * 4,
		gl.PtrOffset(3 * 4))
	gl.BindVertexArray(0)
	
	return cubeVBO, cubeVAO, planeVAO, planeVBO, transparentVAO, transparentVBO
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()
	
	ourShader := shader.MakeShaders("3.1.blending.vs", "3.1.blending.fs")
	cubeVAO, cubeVBO, planeVAO, planeVBO, transparentVAO, transparentVBO := makeCubeBuffers()

	defer gl.DeleteVertexArrays(1, &cubeVAO)
	defer gl.DeleteVertexArrays(1, &cubeVBO)
	defer gl.DeleteVertexArrays(1, &planeVAO)
	defer gl.DeleteVertexArrays(1, &planeVBO)
	defer gl.DeleteVertexArrays(1, &transparentVAO)
	defer gl.DeleteVertexArrays(1, &transparentVBO)


	
	dir := "../../../resources/textures"
	cubeTexture := loadModel.TextureFromFile("marble.jpg", dir, false)
	floorTexture := loadModel.TextureFromFile("metal.png", dir, false)
	// Use a local function instead of the load model to using clamp
	// instead of wrapping. This function should likely take an argument
	// for clamping or repeating.
	transparentTexture := textureFromFile("window.png", dir, false) 

	windows := []mgl32.Vec3{
		mgl32.Vec3{-1.5, 0.0, -0.48},
		mgl32.Vec3{1.5, 0.0, 0.51},
		mgl32.Vec3{0.0, 0.0, 0.7},
		mgl32.Vec3{-0.3, 0.0, -2.3},
		mgl32.Vec3{0.5, 0.0, -0.6},
	}

	decreasingDistance := func(w1, w2 *mgl32.Vec3) bool {
		return w1.Sub(ourCamera.Position).Len() < w2.Sub(ourCamera.Position).Len()
	}

	By(decreasingDistance).Sort(windows)
	fmt.Println(windows)
	
	// shader config
	ourShader.Use()
	ourShader.SetInt("texture1", 0)
	
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
		
		ourShader.Use()	
		projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
			float32(windowHeight) / windowWidth, 0.1, 100.0)
		view := ourCamera.GetViewMatrix()
		ourShader.SetMat4("projection", projection)
		ourShader.SetMat4("view", view)
		// Cubes
		gl.BindVertexArray(cubeVAO)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, cubeTexture)		
		model := mgl32.Translate3D(-1.0, 0.0, -1.0)
		ourShader.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		model = mgl32.Translate3D(2.0, 0.0, 0.0)
		ourShader.SetMat4("model", model)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		// Floor
		gl.BindVertexArray(planeVAO)
		gl.BindTexture(gl.TEXTURE_2D, floorTexture)
		ourShader.SetMat4("model", mgl32.Ident4())
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		gl.BindVertexArray(0)
		// Vegetation
		gl.BindVertexArray(transparentVAO)
		gl.BindTexture(gl.TEXTURE_2D, transparentTexture)
		for i := 0; i < len(windows); i++ {
			model = mgl32.Translate3D(
				windows[i][0], windows[i][1], windows[i][2])
			ourShader.SetMat4("model", model)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
		}
		
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

func textureFromFile(path string, directory string, gamma bool) uint32 {
	filePath := directory + "/" + path

	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	data := loadTexture.ImageLoad(filePath)
	
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
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER,
        gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	
	return textureID
}

