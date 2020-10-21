package shader

import (
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Shader struct {
	ID uint32
}

func MakeShaders(vertexCode string, fragmentCode string) Shader {
	// Compile the shaders
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	shaderSource, freeVertex := gl.Strs(vertexCode + "\x00")
	defer freeVertex()
	gl.ShaderSource(vertexShader, 1, shaderSource, nil)
	gl.CompileShader(vertexShader)
	checkCompileErrors(vertexShader, "VERTEX")

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	shaderSource, freeFragment := gl.Strs(fragmentCode + "\x00")
	defer freeFragment()
	gl.ShaderSource(fragmentShader, 1, shaderSource, nil)
	gl.CompileShader(fragmentShader)
	checkCompileErrors(fragmentShader, "FRAGMENT")

	// Create a shader program
	ID := gl.CreateProgram()
	gl.AttachShader(ID, vertexShader)
	gl.AttachShader(ID, fragmentShader)
	gl.LinkProgram(ID)

	checkCompileErrors(ID, "PROGRAM")

	// Delete shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return Shader{ID: ID}
}

func MakeGeomShaders(vertexCode, fragmentCode, geoCode string) Shader {
	// Compile the shaders
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	shaderSource, freeVertex := gl.Strs(vertexCode + "\x00")
	defer freeVertex()
	gl.ShaderSource(vertexShader, 1, shaderSource, nil)
	gl.CompileShader(vertexShader)
	checkCompileErrors(vertexShader, "VERTEX")

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	shaderSource, freeFragment := gl.Strs(fragmentCode + "\x00")
	defer freeFragment()
	gl.ShaderSource(fragmentShader, 1, shaderSource, nil)
	gl.CompileShader(fragmentShader)
	checkCompileErrors(fragmentShader, "FRAGMENT")

	geoShader := gl.CreateShader(gl.GEOMETRY_SHADER)
	shaderSource, freeGeo := gl.Strs(geoCode + "\x00")
	defer freeGeo()
	gl.ShaderSource(geoShader, 1, shaderSource, nil)
	gl.CompileShader(geoShader)
	checkCompileErrors(geoShader, "GEOMETRY")

	// Create a shader program
	ID := gl.CreateProgram()
	gl.AttachShader(ID, vertexShader)
	gl.AttachShader(ID, fragmentShader)
	gl.AttachShader(ID, geoShader)
	gl.LinkProgram(ID)

	checkCompileErrors(ID, "PROGRAM")

	// Delete shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	gl.DeleteShader(geoShader)

	return Shader{ID: ID}
}

func (s *Shader) Use() *Shader {
	gl.UseProgram(s.ID)
	return s
}

func (s *Shader) SetInteger(name string, value int32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value)
}

func (s *Shader) SetFloat(name string, value float32, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform1f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value)
}

func (s *Shader) SetVector2f(name string, value mgl32.Vec2, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform2fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")),
		1, &value[0])
}

func (s *Shader) SetVector3(name string, value mgl32.Vec3, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform3fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")),
		1, &value[0])
}

func (s *Shader) SetVector4(name string, value mgl32.Vec3, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.Uniform4fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")),
		1, &value[0])
}

func (s *Shader) SetMatrix4(name string, value mgl32.Mat4, useShader bool) {
	if useShader {
		s.Use()
	}
	gl.UniformMatrix4fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")),
		1, false, &value[0])
}

func checkCompileErrors(shader uint32, shaderType string) {
	var success int32
	var infoLog [1024]byte

	var status uint32 = gl.COMPILE_STATUS
	stageMessage := "Shader_Compilation_error"
	errorFunc := gl.GetShaderInfoLog
	getIV := gl.GetShaderiv
	if shaderType == "PROGRAM" {
		status = gl.LINK_STATUS
		stageMessage = "Program_link_error"
		errorFunc = gl.GetProgramInfoLog
		getIV = gl.GetProgramiv
	}

	getIV(shader, status, &success)
	if success != 1 {
		test := &success
		errorFunc(shader, 1024, test, (*uint8)(unsafe.Pointer(&infoLog)))
		log.Fatalln(stageMessage + shaderType + "|" + string(infoLog[:1024]) + "|")
	}
}
