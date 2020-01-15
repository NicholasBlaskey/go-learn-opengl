// translated from

package mesh

import(
	"log"
	"unsafe"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
)


type Vertex struct {
	Position  mgl32.Vec3
	Normal    mgl32.Vec3
	TexCoords mgl32.Vec2
	Tangent   mgl32.Vec3
	Bitangent mgl32.Vec3
}

type Texture struct {
	Id          uint32
	TextureType string
	Path        string
}
	
type Mesh struct {
	vertices []Vertex
	indices  []uint32
	textures []Texture
	VAO      uint32
	VBO      uint32
	EBO      uint32
}

func NewMesh(vertices []Vertex, indices []uint32, textures []Texture) *Mesh {
	log.Println("start new mesh")

	// give buffers value of 0 to avoid complaing
	mesh := Mesh{vertices, indices, textures, 0, 0, 0}
	mesh.setUpMesh()

	log.Println("end new mesh")
	
	return &mesh
}

func (m *Mesh) Draw(shader shader.Shader) {
	log.Println("start draw")
	
	// Bind appropriate textures
	var diffuseNr  uint32 = 1
	var specularNr uint32 = 1
	var normalNr   uint32 = 1
	var heightNr   uint32 = 1

	for i := 0; i < len(m.textures); i++ {
		// Active proper texture unit before binding it
		gl.ActiveTexture(uint32(gl.TEXTURE0 + int32(i)))

		// retrieve textre number (the n in diffuse_textureN)
		name := m.textures[i].TextureType
		var number string
		if name == "texture_diffuse" {
			number = strconv.Itoa(int(diffuseNr))
			diffuseNr++
		} else if name == "texture_specular" {
			number = strconv.Itoa(int(specularNr))
			specularNr++
		} else if name == "texture_normal" {
			number = strconv.Itoa(int(normalNr))
			normalNr++
		} else if name == "texture_height" {
			number = strconv.Itoa(int(heightNr))
			heightNr++
		}

		gl.Uniform1i(gl.GetUniformLocation(
			shader.ID, gl.Str(name + number + "\x00")), int32(i))
		gl.BindTexture(gl.TEXTURE_2D, m.textures[i].Id)
	}

	// Draw the mesh
	gl.BindVertexArray(m.VAO)
	gl.DrawElements(gl.TRIANGLES, int32(len(m.indices)), gl.UNSIGNED_INT,
		unsafe.Pointer(nil))
	gl.BindVertexArray(0)

	// Set back to defaults once configed as a good practice
	gl.ActiveTexture(gl.TEXTURE0)

	log.Println("end draw")
}

func (m *Mesh) setUpMesh() {
	log.Println("start setUpMesh")
	
	vertexSize := int(unsafe.Sizeof(m.vertices[0]))
	
	// Create buffers / arrays
	gl.GenVertexArrays(1, &m.VAO)
	gl.GenBuffers(1, &m.VBO)
	gl.GenBuffers(1, &m.EBO)

	gl.BindVertexArray(m.VAO)
	// Load data into vertex buffers
	gl.BindBuffer(gl.ARRAY_BUFFER, m.VBO)
	// Take advantage of sequential struct layout
	gl.BufferData(gl.ARRAY_BUFFER, len(m.vertices) * vertexSize, 
		gl.Ptr(m.vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.EBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.indices) * 4,
		gl.Ptr(m.indices), gl.STATIC_DRAW)

	// Set the vertex attrib pointers
	// Vertex positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(vertexSize),
		gl.PtrOffset(0))
	// Vertex normals
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(vertexSize),
		gl.PtrOffset(int(unsafe.Offsetof(m.vertices[0].Normal))))
	// Vertex texture coords
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(vertexSize),
		gl.PtrOffset(int(unsafe.Offsetof(m.vertices[0].TexCoords))))
	// Vertex tangent
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, int32(vertexSize),
		gl.PtrOffset(int(unsafe.Offsetof(m.vertices[0].Tangent))))
	// Vertex bitangent
	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 3, gl.FLOAT, false, int32(vertexSize),
		gl.PtrOffset(int(unsafe.Offsetof(m.vertices[0].Bitangent))))

	gl.BindVertexArray(0)

	log.Println("setUpMesh end")
}

