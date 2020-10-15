// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

package main

import (
	//"fmt"
	"log"
	"runtime"
	//"unsafe"

	"github.com/raedatoui/glfont"

	"github.com/go-gl/gl/v4.1-core/gl"
	//"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/glfw/v3.2/glfw"
	//"github.com/nullboundary/glfont"
	//"github.com/go-gl/mathgl/mgl32"
	//"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	//tLoad "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

func init() {
	runtime.LockOSThread()
}

func initGLFW() *glfw.Window {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to init glfw:", err)
	}

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

	// Tell glfw to capture the mouse
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Config gl global state
	//gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	//pathToFont := "../../../resources/fonts/Antonio-Bold.ttf"
	pathToFont := "../../../resources/fonts/huge_agb_v5.ttf"
	font, err := glfont.LoadFont(pathToFont, int32(52), windowWidth, windowHeight)
	if err != nil {
		log.Panicf("LoadFont: %v", err)
	}
	font.SetColor(1.0, 1.0, 1.0, 1.0)

	for !window.ShouldClose() {
		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		font.Printf(30.0, 30.0, 0.5, "This is sample text")
		//renderText(ourShader, "This is sample text", 25.0, 25.0, 1.0,
		//	mgl32.Vec3{0.5, 0.8, 0.2})
		//renderText(ourShader, "(C) LearnOpenGL.com", 540.0, 570.0, 0.5,
		//	mgl32.Vec3{0.3, 0.7, 0.9})

		window.SwapBuffers()
	}
}

/*
func renderText(ourShader shader.Shader, text string,
	x, y, scale float32, col mgl32.Vec3) {

	ourShader.Use()
	ourShader.SetVec3("textColor", col)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(VAO)

	for i := 0; i < len(text); i++ {
		ch := characters[text[i]]

		//xPos := float32(25.0)
		yPos := float32(25.0)
		xPos := x + float32(ch.Bearing[0])*scale // bearingH
		//yPos := y - float32(ch.Size[1]-ch.Bearing[1])*scale // bearingV

		w := float32(ch.Size[0]) * scale
		h := float32(ch.Size[0]) * scale
		fmt.Printf("\n\n %f - (%d - %d)*%f\n", y, ch.Size[1], ch.Bearing[1], scale)
		fmt.Printf("(xPos, yPos) = (%0.2f, %0.2f)\n", xPos, yPos)
		fmt.Printf("(w, h) = (%0.2f, %0.2f) (bearing) = (%d, %d) size = (%d, %d) \n",
			w, h, ch.Bearing[0], ch.Bearing[1], ch.Size[0], ch.Size[1])
		//		fmt.Printf("(x = %0.2f, y = %0.2f) \n", x, y)
		// Update VBO for each character
		vertices := []float32{
			xPos, yPos + h, 0.0, 0.0,
			xPos, yPos, 0.0, 1.0,
			xPos + w, yPos, 1.0, 1.0,

			xPos, yPos + h, 0.0, 0.0,
			xPos + w, yPos, 1.0, 1.0,
			xPos + w, yPos + h, 1.0, 0.0,
		}

		// Render glyph texture over quad
		gl.BindTexture(gl.TEXTURE_2D, ch.TextureID)
		// Update content of VBO memory
		gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, 4*len(vertices), gl.Ptr(vertices))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		x += float32(ch.Advance>>6) * scale
	}
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}
*/

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
