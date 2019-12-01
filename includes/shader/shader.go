// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/includes/learnopengl/shader_s.h

package shader

import(
	"log"
	"github.com/go-gl/gl/v4.1-core/gl"
	"unsafe"
	"io/ioutil"
)

type shader struct {
	ID uint32 
}

func MakeShaders(vertexPath string, fragmentPath string) shader {
	// Read the source code into strings
	vertexCodeBytes, err := ioutil.ReadFile(vertexPath)
	if err != nil {
		panic(err)
	}
	vertexCode := string(vertexCodeBytes)

	fragmentCodeBytes, err := ioutil.ReadFile(fragmentPath)
	if err != nil {
		panic(err)
	}
	fragmentCode := string(fragmentCodeBytes)

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

	return shader{ID: ID}
}

func (s shader) Use() {
	gl.UseProgram(s.ID)
}

func (s shader) SetBool(name string, value bool) {
	var intValue int32 = 0
	if value {
		intValue = 1
	}
	
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name)), intValue)
}

func (s shader) SetInt(name string, value int32) {
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name)), value)
}

func (s shader) SetFloat(name string, value float32) {
	gl.Uniform1f(gl.GetUniformLocation(s.ID, gl.Str(name)), value)
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
		errorFunc(shader, 1024, nil, (*uint8) (unsafe.Pointer(&infoLog)))
		log.Fatalln(stageMessage + shaderType + string(infoLog[:1024]))
	}
}
