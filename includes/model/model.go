// translated from

// used a lot of cgo stuff from
// https://github.com/tbogdala/assimp-go/blob/master/assimp.go

package model

/*
#cgo CPPFLAGS: -I/mingw64/include -std=c99
#cgo LDFLAGS: -L/mingw64/lib -lassimp -lz -lstdc++


#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <assimp/cimport.h>
#include <assimp/scene.h>
#include <assimp/mesh.h>
#include <assimp/cimport.h>
#include <assimp/matrix4x4.h>
#include <assimp/postprocess.h>

struct aiNode* get_child(struct aiNode* n, unsigned int index)
{
	return n->mChildren[index];
}

struct aiMesh* get_mesh(struct aiScene* s, struct aiNode* n, 
	unsigned int index) 
{
	return s->mMeshes[n->mMeshes[index]];
}

struct aiVector3D* mesh_vertex_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mVertices[index]);
}

struct aiVector3D* mesh_normal_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mNormals[index]);
}

_Bool has_tex_coords(struct aiMesh* m) {
	return m->mTextureCoords[0];
}

struct aiVector3D* mesh_texture_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mTextureCoords[0][index]);
}

struct aiVector3D* mesh_tangent_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mTangents[index]);
}

struct aiVector3D* mesh_bitangent_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mBitangents[index]);
}

struct aiFace* get_face(struct aiMesh* m, unsigned int index) 
{
	return &(m->mFaces[index]);
}

unsigned int get_face_indices(struct aiFace* f, unsigned int index) 
{
	return f->mIndices[index];
}

struct aiMaterial* get_material(struct aiScene* s, struct aiMesh* m) 
{
	return s->mMaterials[m->mMaterialIndex];
}

//int get_num_textures(struct aiMaterial* mat, enum aiTextureType type) {
//	return mat->GetTextureCount(type);
//}
*/
import "C"

import(
	"log"
	"unsafe"
	//"math"
	//"strconv"
	"strings"
	
	//"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/includes/mesh"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
	loadTexture "github.com/nicholasblaskey/go-learn-opengl/includes/texture"
)

type Model struct {
	texturesLoaded  []mesh.Texture
	meshes          []mesh.Mesh
	directory       string
	gammaCorrection bool
}

func NewModel(path string, gamma bool) *Model {
	model := Model{gammaCorrection: gamma}
	model.loadModel(path)

	return &model
}

func (model *Model) Draw(shader shader.Shader) {
	for i := 0; i < len(model.meshes); i++ {
		model.meshes[i].Draw(shader)
	}
}

func (model *Model) loadModel(path string) {
	cPathString := C.CString(path)
	defer C.free(unsafe.Pointer(cPathString))

	scene := C.aiImportFile(cPathString,
		C.aiProcess_Triangulate |
		C.aiProcess_FlipUVs |
		C.aiProcess_CalcTangentSpace)

	//log.Println(scene)
	//log.Printf("%T\n", scene)
	//log.Printf("%+v", scene)
	
	// Make sure we loaded meshes properly
	if uintptr(unsafe.Pointer(scene)) == 0 {
		panic("filepath: " + path + "loaded a nil scene")
	}
	if scene.mNumMeshes < 1 {
		panic("Got zero meshes when loading")
	}
	if uintptr(unsafe.Pointer(scene.mRootNode)) == 0 {
		panic("Root node of the scene was nil")
	}
	
	
	// Retrieve the directory of the filepath
	model.directory = path[0:strings.LastIndex(path, "/")]

	model.processNode(scene.mRootNode, scene)	
}

func (model *Model) processNode(aiNode *C.struct_aiNode,
	aiScene *C.struct_aiScene) {

	// Process the current node
	for i := 0; i < int(aiNode.mNumMeshes); i++ {
		// Get mesh just does scene->mMeshes[node->mMeshes[i]]
		mesh := C.get_mesh(aiScene, aiNode, C.uint(i))
		//model.meshes = append(model.meshes,
		//	model.processMesh(mesh, aiScene))
		model.processMesh(mesh, aiScene)
	}
	// Call process node on all the children nodes
	for i := 0; i < int(aiNode.mNumChildren); i++ {
		model.processNode(C.get_child(aiNode, C.uint(i)), aiScene)
	}
}

func (model *Model) processMesh(aiMesh *C.struct_aiMesh,
	aiScene *C.struct_aiScene) *mesh.Mesh {

	// Data to fill
	var vertices []mesh.Vertex 
	var indices  []uint32
	var textures []mesh.Texture

	// Loop through all of the mesh's vertices
	for i := 0; i < int(aiMesh.mNumVertices); i++ {
		var vertex mesh.Vertex

		// Position
		cVec := C.mesh_vertex_at(aiMesh, C.uint(i))
		vertex.Position[0] = float32(cVec.x)
		vertex.Position[1] = float32(cVec.y)
		vertex.Position[2] = float32(cVec.z)

		// Normals
		cVec = C.mesh_normal_at(aiMesh, C.uint(i))
		vertex.Normal[0] = float32(cVec.x)
		vertex.Normal[1] = float32(cVec.y)
		vertex.Normal[2] = float32(cVec.z)

		// Texture coords (assuming we only use the first uv channel)
		if C.has_tex_coords(aiMesh) {
			cVec = C.mesh_texture_at(aiMesh, C.uint(i))
			vertex.TexCoords[0] = float32(cVec.x)
			vertex.TexCoords[1] = float32(cVec.y)
		} // No need for else when mgl vecs are inited to 0
		
		// Tangent
		cVec = C.mesh_tangent_at(aiMesh, C.uint(i))
		vertex.Tangent[0] = float32(cVec.x)
		vertex.Tangent[1] = float32(cVec.y)
		vertex.Tangent[2] = float32(cVec.z)
		
		// Bitangent
		cVec = C.mesh_bitangent_at(aiMesh, C.uint(i))
		vertex.Bitangent[0] = float32(cVec.x)
		vertex.Bitangent[1] = float32(cVec.y)
		vertex.Bitangent[2] = float32(cVec.z)

		vertices = append(vertices, vertex)
	}

	//log.Printf("%+v", vertices)

	// Now handle all the mesh's faces abd retrieve corresponding vertex indices.
	for i := 0; i < int(aiMesh.mNumFaces); i++ {
		face := C.get_face(aiMesh, C.uint(i))

		for j := 0; j < int(face.mNumIndices); j++ {
			// TODO check the result of indices is right due to getting some very
			// large values 1312808169...2483192576? Could be an issue or could be
			// intended if something is wrong check back here.
			indices = append(indices,
				uint32(C.get_face_indices(face, C.uint(i))))
		}
	}

	//log.Printf("%+v", indices)

	// Process materias
	material := C.get_material(aiScene, aiMesh)
	// We assume a convention for sampler names in the shaders. Each diffuse
	// texture should be named as 'texture_diffuseN' where N is a sequential
	// number ranging from 1 to MAX_SAMPLER_NUMBER. 
	// Same applies to other texture as the following list summarizes:
	// diffuse: texture_diffuseN
	// specular: texture_specularN
	// normal: texture_normalN

	log.Println(material, "this is not a crash")

	
	// 1. diffuse maps
	diffuseMaps := model.loadMaterialTextures(material,
		C.aiTextureType_DIFFUSE, "texture_diffuse")
	log.Println(diffuseMaps)
	/*
	// TODO figure out return type and append it
	// 2. specular maps
	speculareMaps := model.loadMaterialTextures(material,
		C.aiTextureType_SPECULAR, "texture_specular")
	// TODO
	// 3. normal maps
	normalMaps := model.loadMaterialTextures(material,
		C.aiTextureType_HEIGHT, "texture_normal")
	// TODO
	// 4. height maps
	heightMaps := model.loadMaterialTextures(material,
		C.aiTextureType_AMBIENT, "texture_height")
	// TODO
	*/
	
	return mesh.NewMesh(vertices, indices, textures)
	
	/*
        // process materials
        aiMaterial* material = scene->mMaterials[mesh->mMaterialIndex];    
   
        // 1. diffuse maps
        vector<Texture> diffuseMaps = loadMaterialTextures(material, aiTextureType_DIFFUSE, "texture_diffuse");
        textures.insert(textures.end(), diffuseMaps.begin(), diffuseMaps.end());
        // 2. specular maps
        vector<Texture> specularMaps = loadMaterialTextures(material, aiTextureType_SPECULAR, "texture_specular");
        textures.insert(textures.end(), specularMaps.begin(), specularMaps.end());
        // 3. normal maps
        std::vector<Texture> normalMaps = loadMaterialTextures(material, aiTextureType_HEIGHT, "texture_normal");
        textures.insert(textures.end(), normalMaps.begin(), normalMaps.end());
        // 4. height maps
        std::vector<Texture> heightMaps = loadMaterialTextures(material, aiTextureType_AMBIENT, "texture_height");
        textures.insert(textures.end(), heightMaps.begin(), heightMaps.end());
        
        // return a mesh object created from the extracted mesh data
        return Mesh(vertices, indices, textures);
	*/
}

func (model *Model) loadMaterialTextures(mat *C.struct_aiMaterial,
	textType uint32/**C.enum_aiTextureType*/, typeName string) ([]mesh.Texture){

	var textures []mesh.Texture

	log.Printf("in this func brah %T", textType)
	
	textCount := C.aiGetMaterialTextureCount(mat, textType)


	//textCount := C.get_num_textures(textType)
	//textCount := C.GetTextureCount(textType)
	for i := 0; i < int(textCount); i++ {
		log.Println(textCount)
	}

	return textures
}
/*
    {
        vector<Texture> textures;
        for(unsigned int i = 0; i < mat->GetTextureCount(type); i++)
        {
            aiString str;
            mat->GetTexture(type, i, &str);
            // check if texture was loaded before and if so, continue to next iteration: skip loading a new texture
            bool skip = false;
            for(unsigned int j = 0; j < textures_loaded.size(); j++)
            {
                if(std::strcmp(textures_loaded[j].path.data(), str.C_Str()) == 0)
                {
                    textures.push_back(textures_loaded[j]);
                    skip = true; // a texture with the same filepath has already been loaded, continue to next one. (optimization)
                    break;
                }
            }
            if(!skip)
            {   // if texture hasn't been loaded already, load it
                Texture texture;
                texture.id = TextureFromFile(str.C_Str(), this->directory);
                texture.type = typeName;
                texture.path = str.C_Str();
                textures.push_back(texture);
                textures_loaded.push_back(texture);  // store it as texture loaded for entire model, to ensure we won't unnecesery load duplicate textures.
            }
        }
        return textures;
*/

// Not part of class
// TextureFromFile
func TextureFromFile(path string, directory string, gamma bool) uint32 {
	filePath := path + directory

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
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER,
        gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID
}
	
