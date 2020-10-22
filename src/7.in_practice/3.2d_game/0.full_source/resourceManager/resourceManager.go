package resourceManager

import (
	"io/ioutil"

	"github.com/go-gl/gl/v4.1-core/gl"

	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/shader"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/texture"
)

var (
	Textures map[string]*texture.Texture = make(map[string]*texture.Texture)
	Shaders  map[string]*shader.Shader   = make(map[string]*shader.Shader)
)

func LoadShader(vShaderFile, fShaderFile, name string) *shader.Shader {
	Shaders[name] = loadShaderFromFile(vShaderFile, fShaderFile, "")
	return Shaders[name]
}

func LoadShaderGeom(vShaderFile, fShaderFile, gShaderFile, name string) *shader.Shader {
	Shaders[name] = loadShaderFromFile(vShaderFile, fShaderFile, gShaderFile)
	return Shaders[name]
}

func LoadTexture(file string, name string) *texture.Texture {
	Textures[name] = loadTextureFromFile(file)
	return Textures[name]
}

func Clear() {
	for _, s := range Shaders {
		gl.DeleteProgram(s.ID)
	}
	for _, t := range Textures {
		gl.DeleteTextures(1, &t.ID)
	}
}

func loadShaderFromFile(vShaderFile, fShaderFile, gShaderFile string) *shader.Shader {
	vertexCodeBytes, err := ioutil.ReadFile(vShaderFile)
	if err != nil {
		panic(err)
	}
	vertexCode := string(vertexCodeBytes)

	fragmentCodeBytes, err := ioutil.ReadFile(fShaderFile)
	if err != nil {
		panic(err)
	}
	fragmentCode := string(fragmentCodeBytes)

	if gShaderFile != "" {
		geoCodeBytes, err := ioutil.ReadFile(gShaderFile)
		if err != nil {
			panic(err)
		}
		geoCode := string(geoCodeBytes)

		return shader.MakeGeomShaders(vertexCode, fragmentCode, geoCode)
	}
	return shader.MakeShaders(vertexCode, fragmentCode)
}

func loadTextureFromFile(file string) *texture.Texture {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("Unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	t := texture.New()
	t.Generate(int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), rgba.Pix)
	return t
}
