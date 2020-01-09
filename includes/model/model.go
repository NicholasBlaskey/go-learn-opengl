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
	unsigned int i) 
{
	return s->mMeshes[n->mMeshes[i]];
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
		model.meshes = append(model.meshes,
			model.processMesh(mesh, aiScene))
	}
	// Call process node on all the children nodes
	for i := 0; i < int(aiNode.mNumChildren); i++ {
		model.processNode(C.get_child(aiNode, C.uint(i)), aiScene)
	}
}

// processMesh

// loadMaterialTextures

// Not part of class
// TextureFromFile
