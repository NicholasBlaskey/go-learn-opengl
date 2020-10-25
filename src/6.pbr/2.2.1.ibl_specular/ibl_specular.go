// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/6.pbr/2.2.1.ibl_specular/ibl_specular.cpp

package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/stbi"

	"github.com/nicholasblaskey/go-learn-opengl/includes/camera"
	//loadModel "github.com/nicholasblaskey/go-learn-opengl/includes/model"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	//"github.com/disintegration/imaging"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
)

// Camera
var ourCamera camera.Camera = camera.NewCamera(
	0.0, 0.0, 3.0, // pos xyz
	0.0, 1.0, 0.0, // up xyz
	-90.0, 0.0, // Yaw and pitch
	80.0, 45.0, 0.1) // Speed, zoom, and mouse sensitivity
var firstMouse bool = true
var lastX float32 = windowWidth / 2
var lastY float32 = windowHeight / 2

// Timing
var deltaTime float32 = 0.0
var lastFrame float32 = 0.0

// Lighting
var lightPos mgl32.Vec3 = mgl32.Vec3{1.2, 1.0, 2.0}

// Controls
var heldW bool = false
var heldA bool = false
var heldS bool = false
var heldD bool = false

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
	gl.DepthFunc(gl.LEQUAL) // Set the depth function to less than AND equal for skybox depth trick.
	gl.Enable(gl.TEXTURE_CUBE_MAP_SEAMLESS)

	return window
}

func main() {
	window := initGLFW()
	defer glfw.Terminate()

	// Build and compile shaders
	pbrShader := shader.MakeShaders("2.2.1.pbr.vs", "2.2.1.pbr.fs")
	equaiRectToCubeMapShader := shader.MakeShaders(
		"2.2.1.cubemap.vs", "2.2.1.equirectangular_to_cubemap.fs")
	irradianceShader := shader.MakeShaders("2.2.1.cubemap.vs",
		"2.2.1.irradiance_convolution.fs")
	prefilterShader := shader.MakeShaders("2.2.1.cubemap.vs", "2.2.1.prefilter.fs")
	brdfShader := shader.MakeShaders("2.2.1.brdf.vs", "2.2.1.brdf.fs")
	backgroundShader := shader.MakeShaders("2.2.1.background.vs", "2.2.1.background.fs")

	pbrShader.Use()
	pbrShader.SetInt("irradianceMap", 0)
	pbrShader.SetInt("prefilterMap", 1)
	pbrShader.SetInt("brdfLUT", 2)
	pbrShader.SetVec3("albedo", mgl32.Vec3{0.5, 0.0, 0.0})
	pbrShader.SetFloat("ao", 1.0)

	backgroundShader.Use()
	backgroundShader.SetInt("environmentMap", 0)

	lightPositions := []mgl32.Vec3{
		mgl32.Vec3{-10.0, +10.0, 10.0},
		mgl32.Vec3{+10.0, +10.0, 10.0},
		mgl32.Vec3{-10.0, -10.0, 10.0},
		mgl32.Vec3{+10.0, -10.0, 10.0},
	}
	lightColors := []mgl32.Vec3{
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
		mgl32.Vec3{300.0, 300.0, 300.0},
	}
	nrRows := 7
	nrCols := 7
	spacing := float32(2.5)

	// Pbr: set up the framebuffer
	var captureFBO, captureRBO uint32
	gl.GenFramebuffers(1, &captureFBO)
	gl.GenRenderbuffers(1, &captureRBO)

	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	gl.BindFramebuffer(gl.RENDERBUFFER, captureRBO)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, 512, 512)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, captureRBO)

	// Pbr load the HDR environment map
	path := "../../../resources/textures/hdr/newport_loft.hdr"
	data, width, height, _, cleanup, err := stbi.Loadf(path, true, 0)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	var hdrTexture uint32
	gl.GenTextures(1, &hdrTexture)
	gl.BindTexture(gl.TEXTURE_2D, hdrTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB16F, int32(width), int32(height), 0,
		gl.RGB, gl.FLOAT, data)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Pbr: set up the cubemap to render to and attach to framebuffer
	var envCubemap uint32
	gl.GenTextures(1, &envCubemap)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)
	for i := 0; i < 6; i++ {
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0, gl.RGB16F,
			512, 512, 0, gl.RGB, gl.FLOAT, nil)
	}
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER,
		gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Pbr: set up projection and view matrices for capturing data onto the 6
	// cubemap face directions
	captureProjection := mgl32.Perspective(mgl32.DegToRad(90.0), 1.0, 0.1, 10.0)
	captureViews := []mgl32.Mat4{
		mgl32.LookAt(0.0, 0.0, 0.0, +1.0, +0.0, +0.0, +0.0, -1.0, +0.0),
		mgl32.LookAt(0.0, 0.0, 0.0, -1.0, +0.0, +0.0, +0.0, -1.0, +0.0),
		mgl32.LookAt(0.0, 0.0, 0.0, +0.0, +1.0, +0.0, +0.0, +0.0, +1.0),
		mgl32.LookAt(0.0, 0.0, 0.0, +0.0, -1.0, +0.0, +0.0, +0.0, -1.0),
		mgl32.LookAt(0.0, 0.0, 0.0, +0.0, +0.0, +1.0, +0.0, -1.0, +0.0),
		mgl32.LookAt(0.0, 0.0, 0.0, +0.0, +0.0, -1.0, +0.0, -1.0, +0.0),
	}

	// Pbr: convert HDR equirectangular environment map to cubemap equivalent
	equaiRectToCubeMapShader.Use()
	equaiRectToCubeMapShader.SetInt("equirectangularMap", 0)
	equaiRectToCubeMapShader.SetMat4("projection", captureProjection)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, hdrTexture)

	gl.Viewport(0, 0, 512, 512)
	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	for i, captureView := range captureViews {
		equaiRectToCubeMapShader.SetMat4("view", captureView)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
			gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), envCubemap, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		renderCube()
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Then let OpenGL generate mipmaps from the first mip face (combatting visible dots artifact)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)
	gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)

	// Pbr: create an irradiance cubemap, and re-scale capture FBO to irradiance scale.
	var irradianceMap uint32
	gl.GenTextures(1, &irradianceMap)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, irradianceMap)
	for i := 0; i < 6; i++ {
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0,
			gl.RGB16F, 32, 32, 0, gl.RGB, gl.FLOAT, nil)
	}
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	gl.BindRenderbuffer(gl.RENDERBUFFER, captureRBO)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, 32, 32)

	// Pbr: solve diffuse integral by convolution to create an irradiance cubemap
	irradianceShader.Use()
	irradianceShader.SetInt("environmentMap", 0)
	irradianceShader.SetMat4("projection", captureProjection)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)

	gl.Viewport(0, 0, 32, 32)
	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	for i := 0; i < 6; i++ {
		irradianceShader.SetMat4("view", captureViews[i])
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
			gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), irradianceMap, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		renderCube()
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Pbr: create a pre-filter cubemap, and re-scale capture FBO to pre-filter scale
	var prefilterMap uint32
	gl.GenTextures(1, &prefilterMap)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, prefilterMap)
	for i := 0; i < 6; i++ {
		gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), 0,
			gl.RGB16F, 128, 128, 0, gl.RGB, gl.FLOAT, nil)
	}
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	// Generate mipmaps for the cubemap so OpenGL automatically allocates the needed memory
	gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)

	// Pbr: run a quasi monte-carlo simulation ont he environment lighting to create
	// a prefilter cubemap
	prefilterShader.Use()
	prefilterShader.SetInt("environmentMap", 0)
	prefilterShader.SetMat4("projection", captureProjection)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)

	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	maxMipLevels := 5
	for mip := 0; mip < maxMipLevels; mip++ {
		mipWidth := int32(128.0 * float32(math.Pow(0.5, float64(mip))))
		mipHeight := int32(128.0 * float32(math.Pow(0.5, float64(mip))))
		gl.BindRenderbuffer(gl.RENDERBUFFER, captureRBO)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24,
			mipWidth, mipHeight)
		gl.Viewport(0, 0, mipWidth, mipHeight)

		roughness := float32(mip) / float32(maxMipLevels-1)
		prefilterShader.SetFloat("roughness", roughness)
		for i := 0; i < 6; i++ {
			prefilterShader.SetMat4("view", captureViews[i])
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
				gl.TEXTURE_CUBE_MAP_POSITIVE_X+uint32(i), prefilterMap, int32(mip))

			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			renderCube()
		}
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Pbr: generate a 2D LUT from the BRDF equations used
	var brdfLUTTexture uint32
	gl.GenTextures(1, &brdfLUTTexture)

	// Pre-allocate enough memory for the LUT texture
	gl.BindTexture(gl.TEXTURE_2D, brdfLUTTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RG16F,
		512, 512, 0, gl.RG, gl.FLOAT, nil)
	// Be sure to set wrapping mode to GL_CLAMP_TO_EDGE
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Then re-configure capture framebuffer object and render screen-space quad with BRDF shader
	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	gl.BindRenderbuffer(gl.RENDERBUFFER, captureRBO)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, 512, 512)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D,
		brdfLUTTexture, 0)

	gl.Viewport(0, 0, 512, 512)
	brdfShader.Use()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	renderQuad()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Init static shader uniform before rendering
	projection := mgl32.Perspective(mgl32.DegToRad(ourCamera.Zoom),
		float32(windowWidth)/windowHeight, 0.1, 100.0)
	pbrShader.Use()
	pbrShader.SetMat4("projection", projection)
	backgroundShader.Use()
	backgroundShader.SetMat4("projection", projection)

	// Then before rendering, configure the viewport to the original framebuffer's
	// screen dimensions
	gl.Viewport(0, 0, windowWidth, windowHeight)

	// Program loop
	for !window.ShouldClose() {
		// Pre frame logic
		currentFrame := float32(glfw.GetTime())
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Input
		glfw.PollEvents()

		// Render
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// render scene, supplying the convoluted irradiance map to the final shader
		pbrShader.Use()
		view := ourCamera.GetViewMatrix()
		pbrShader.SetMat4("view", view)
		pbrShader.SetVec3("camPos", ourCamera.Position)

		// Bind pre-computed IBL data
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, irradianceMap)
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, prefilterMap)
		gl.ActiveTexture(gl.TEXTURE2)
		gl.BindTexture(gl.TEXTURE_2D, brdfLUTTexture)

		// Render rows * cols number of spheres
		// with varying material properties
		for row := 0; row < nrRows; row++ {
			pbrShader.SetFloat("metallic", float32(row)/float32(nrRows))
			for col := 0; col < nrCols; col++ {
				// We clamp the roughness to 0.025 - 1.0 as perfectly smooth surfaces
				// (roughness of 0.0) tend to look a bit off on direct lighting.
				pbrShader.SetFloat("roughness",
					mgl32.Clamp(float32(col)/float32(nrCols), 0.05, 1.0))
				model := mgl32.Translate3D(
					(float32(col)-(float32(nrCols)/2.0))*spacing,
					(float32(row)-(float32(nrRows)/2.0))*spacing, 0.0)
				pbrShader.SetMat4("model", model)
				renderSphere()
			}
		}

		for i := 0; i < len(lightPositions); i++ {
			newPos := lightPositions[i].Add(
				mgl32.Vec3{float32(math.Sin(glfw.GetTime()*5.0)) * 5.0, 0.0, 0.0})
			newPos = lightPositions[i] // Confused on why overwrite previous assignment
			pbrShader.SetVec3(fmt.Sprintf("lightPositions[%d]", i), newPos)
			pbrShader.SetVec3(fmt.Sprintf("lightColors[%d]", i), lightColors[i])

			model := mgl32.Translate3D(newPos[0], newPos[1], newPos[2]).Mul4(
				mgl32.Scale3D(0.5, 0.5, 0.5))
			pbrShader.SetMat4("model", model)
			renderSphere()
		}

		// Render skybox (render as last to prevent overdraw)
		backgroundShader.Use()
		backgroundShader.SetMat4("view", view)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)
		//gl.BindTexture(gl.TEXTURE_CUBE_MAP, irradianceMap) // display irradiance map
		//gl.BindTexture(gl.TEXTURE_CUBE_MAP, prefilterMap) // display prefilter map
		renderCube()

		// Render BRDF map to screen
		// brdfShader.Use()
		// renderQuad()

		window.SwapBuffers()
	}
}

var (
	sphereVAO  uint32 = 0
	cubeVAO    uint32 = 0
	cubeVBO    uint32 = 0
	quadVAO    uint32 = 0
	quadVBO    uint32 = 0
	indexCount uint32
)

func renderSphere() {
	if sphereVAO != 0 {
		gl.BindVertexArray(sphereVAO)
		gl.DrawElements(gl.TRIANGLE_STRIP, int32(indexCount),
			gl.UNSIGNED_INT, unsafe.Pointer(nil))
		return
	}

	gl.GenVertexArrays(1, &sphereVAO)

	var vbo, ebo uint32
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	positions := []mgl32.Vec3{}
	uv := []mgl32.Vec2{}
	normals := []mgl32.Vec3{}
	indices := []uint32{}

	xSegments := 64
	ySegments := 64
	pi := float32(math.Pi)
	for y := 0; y <= ySegments; y++ {
		for x := 0; x <= xSegments; x++ {
			xSegment := float32(x) / float32(xSegments)
			ySegment := float32(y) / float32(ySegments)
			xPos := float32(math.Cos(float64(xSegment*2.0*pi)) *
				math.Sin(float64(ySegment*pi)))
			yPos := float32(math.Cos(float64(ySegment * pi)))
			zPos := float32(math.Sin(float64(xSegment*2.0*pi)) *
				math.Sin(float64(ySegment*pi)))

			positions = append(positions, mgl32.Vec3{xPos, yPos, zPos})
			uv = append(uv, mgl32.Vec2{xSegment, ySegment})
			normals = append(normals, mgl32.Vec3{xPos, yPos, zPos})
		}
	}

	oddRow := false
	for y := 0; y < ySegments; y++ {
		if oddRow {
			for x := 0; x <= xSegments; x++ {
				indices = append(indices, uint32(y*(xSegments+1)+x))
				indices = append(indices, uint32((y+1)*(xSegments+1)+x))
			}
		} else {
			for x := xSegments; x >= 0; x-- {
				indices = append(indices, uint32((y+1)*(xSegments+1)+x))
				indices = append(indices, uint32(y*(xSegments+1)+x))
			}
		}
		oddRow = !oddRow
	}
	indexCount = uint32(len(indices))

	data := []float32{}
	for i := 0; i < len(positions); i++ {
		data = append(data, positions[i][0], positions[i][1], positions[i][2])
		if len(uv) > 0 {
			data = append(data, uv[i][0], uv[i][1])
		}
		if len(normals) > 0 {
			data = append(data, normals[i][0], normals[i][1], normals[i][2])
		}
	}

	gl.BindVertexArray(sphereVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4,
		gl.Ptr(indices), gl.STATIC_DRAW)

	stride := int32((3 + 2 + 3) * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))

	renderSphere()
}

func renderCube() {
	if cubeVAO != 0 {
		gl.BindVertexArray(cubeVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 36)
		gl.BindVertexArray(0)
		return
	}

	vertices := []float32{
		// positions            // normals         // texcoords
		// back
		-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
		1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
		1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
		-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
		-1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 1.0, // top-left
		// front
		-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
		1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
		1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
		-1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, // top-left
		-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
		// left
		-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
		-1.0, 1.0, -1.0, -1.0, 0.0, 0.0, 1.0, 1.0, // top-left
		-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
		-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
		-1.0, -1.0, 1.0, -1.0, 0.0, 0.0, 0.0, 0.0, // bottom-right
		-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
		// right
		1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
		1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
		1.0, 1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, // top-right
		1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
		1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
		1.0, -1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0, // bottom-left
		// bottom
		-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
		1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 1.0, 1.0, // top-left
		1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
		1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
		-1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 0.0, 0.0, // bottom-right
		-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
		// top
		-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
		1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
		1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 1.0, 1.0, // top-right
		1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
		-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
		-1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 0.0, // bottom-left
	}
	gl.GenVertexArrays(1, &cubeVAO)
	gl.GenBuffers(1, &cubeVBO)
	gl.BindVertexArray(cubeVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4,
		gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.BindVertexArray(0)

	renderCube()
}

func renderQuad() {
	if quadVAO != 0 {
		gl.BindVertexArray(quadVAO)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		gl.BindVertexArray(0)
		return
	}

	vertices := []float32{
		// positions        // texture Coords
		-1.0, 1.0, 0.0, 0.0, 1.0,
		-1.0, -1.0, 0.0, 0.0, 0.0,
		1.0, 1.0, 0.0, 1.0, 1.0,
		1.0, -1.0, 0.0, 1.0, 0.0,
	}
	gl.GenVertexArrays(1, &quadVAO)
	gl.GenBuffers(1, &quadVBO)
	gl.BindVertexArray(quadVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4,
		gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.BindVertexArray(0)

	renderQuad()
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
