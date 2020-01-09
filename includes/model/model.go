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

// TODO add in texture

struct aiVector3D* mesh_tangent_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mTangents[index]);
}

struct aiVector3D* mesh_bitangent_at(struct aiMesh* m, unsigned int index) 
{
	return &(m->mBitangents[index]);
}
*/
import "C"

import(
	"log"
	"unsafe"
	//"strconv"
	"strings"
	
	//"github.com/go-gl/mathgl/mgl32"
	//"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/includes/mesh"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
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
	aiScene *C.struct_aiScene)  {

	// Data to fill
	//var vertices []mesh.Vertex 
	//var indices  []uint32
	//var textures []mesh.Texture

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

		// Texture coords
		
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

		log.Println(vertex)
	}


		/*
        // Walk through each of the mesh's vertices
        for(unsigned int i = 0; i < mesh->mNumVertices; i++)
        {
            Vertex vertex;
            glm::vec3 vector; // we declare a placeholder vector since assimp uses its own vector class that doesn't directly convert to glm's vec3 class so we transfer the data to this placeholder glm::vec3 first.
            // positions
            vector.x = mesh->mVertices[i].x;
            vector.y = mesh->mVertices[i].y;
            vector.z = mesh->mVertices[i].z;
            vertex.Position = vector;
            // normals
            vector.x = mesh->mNormals[i].x;
            vector.y = mesh->mNormals[i].y;
            vector.z = mesh->mNormals[i].z;
            vertex.Normal = vector;
            // texture coordinates
            if(mesh->mTextureCoords[0]) // does the mesh contain texture coordinates?
            {
                glm::vec2 vec;
                // a vertex can contain up to 8 different texture coordinates. We thus make the assumption that we won't 
                // use models where a vertex can have multiple texture coordinates so we always take the first set (0).
                vec.x = mesh->mTextureCoords[0][i].x; 
                vec.y = mesh->mTextureCoords[0][i].y;
                vertex.TexCoords = vec;
            }
            else
                vertex.TexCoords = glm::vec2(0.0f, 0.0f);
            // tangent
            vector.x = mesh->mTangents[i].x;
            vector.y = mesh->mTangents[i].y;
            vector.z = mesh->mTangents[i].z;
            vertex.Tangent = vector;
            // bitangent
            vector.x = mesh->mBitangents[i].x;
            vector.y = mesh->mBitangents[i].y;
            vector.z = mesh->mBitangents[i].z;
            vertex.Bitangent = vector;
            vertices.push_back(vertex);
        }
        // now wak through each of the mesh's faces (a face is a mesh its triangle) and retrieve the corresponding vertex indices.
        for(unsigned int i = 0; i < mesh->mNumFaces; i++)
        {
            aiFace face = mesh->mFaces[i];
            // retrieve all indices of the face and store them in the indices vector
            for(unsigned int j = 0; j < face.mNumIndices; j++)
                indices.push_back(face.mIndices[j]);
        }
        // process materials
        aiMaterial* material = scene->mMaterials[mesh->mMaterialIndex];    
        // we assume a convention for sampler names in the shaders. Each diffuse texture should be named
        // as 'texture_diffuseN' where N is a sequential number ranging from 1 to MAX_SAMPLER_NUMBER. 
        // Same applies to other texture as the following list summarizes:
        // diffuse: texture_diffuseN
        // specular: texture_specularN
        // normal: texture_normalN

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

// loadMaterialTextures

// Not part of class
// TextureFromFile
