// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/3.1.1.shadow_mapping_depth/shadow_mapping_depth.cpp

package main

import (
	"fmt"
	"log"
	"runtime"
	//"unsafe"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
	"io/ioutil"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	//tLoad "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

var (
	VAO        uint32
	VBO        uint32
	characters []*Character
)

type Character struct {
	TextureID uint32
	Size      [2]int32
	Bearing   [2]int32
	Advance   uint32
}

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
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Build and compile shaders
	ourShader := shader.MakeShaders("text.vs", "text.fs")
	projection := mgl32.Ortho(0.0, float32(windowWidth),
		0.0, float32(windowHeight), 0.0, 0.0)
	ourShader.Use()
	ourShader.SetMat4("projection", projection)

	// FreeType
	// Some code taken from
	// https://github.com/nullboundary/glfont/blob/master/truetype.go
	fd, err := os.Open("../../../resources/fonts/Antonio-Bold.ttf")
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}

	ttf, err := truetype.Parse(data)
	if err != nil {
		panic(err)
	}

	characters = make([]*Character, 128)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	scale := float32(48.0)
	for c := 0; c < 128; c++ {
		ttfFace := truetype.NewFace(ttf, &truetype.Options{
			Size:    float64(scale),
			DPI:     72.0,
			Hinting: font.HintingFull,
		})

		gBnd, gAdv, ok := ttfFace.GlyphBounds(rune(c))
		if !ok {
			panic("ttfFace gylph bounds had an error")
		}
		gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)
		gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		if gw == 0 || gh == 0 {
			gBnd = ttf.Bounds(fixed.Int26_6(scale))
			gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
			gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)

			if gw == 0 || gh == 0 {
				gw = 1
				gh = 1
			}
		}

		char := new(Character)
		char.Size = [2]int32{gw, gh}
		char.Bearing = [2]int32{int32(gBnd.Min.X) >> 6, int32(gBnd.Max.Y) >> 6}
		char.Advance = uint32(gAdv)

		fg, bg := image.White, image.Black
		rect := image.Rect(0, 0, int(gw), int(gh))
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

		freeContext := freetype.NewContext()
		freeContext.SetDPI(72)
		freeContext.SetFont(ttf)
		freeContext.SetFontSize(float64(scale))
		freeContext.SetClip(rgba.Bounds())
		freeContext.SetDst(rgba)
		freeContext.SetSrc(fg)
		freeContext.SetHinting(font.HintingFull)

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED,
			int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), 0, gl.RGBA,
			gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		char.TextureID = texture
		characters[c] = char
		//fmt.Printf("%+v\n", char)
	}

	fmt.Println(characters)
	// End free type

	// Configure VAO/VBO for texture quads
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, 4*6*4, nil, gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	// Program loop
	// Draw in polygon mode
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	for !window.ShouldClose() {
		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		renderText(ourShader, "This is sample text", 25.0, 25.0, 1.0,
			mgl32.Vec3{0.5, 0.8, 0.2})

		window.SwapBuffers()
	}
}

func renderText(ourShader shader.Shader, text string,
	x, y, scale float32, col mgl32.Vec3) {

	ourShader.Use()
	ourShader.SetVec3("textColor", col)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindVertexArray(VAO)

	for i := 0; i < len(text); i++ {
		ch := characters[text[i]]

		xPos := float32(25.0)
		yPos := float32(25.0)
		//xPos := x + float32(ch.Bearing[0])*scale // bearingH
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
