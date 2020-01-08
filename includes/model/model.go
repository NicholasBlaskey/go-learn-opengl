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
	log.Printf("%+v", scene)
	
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
	
	for i := 0; i < int(aiNode.mNumMeshes); i++ {
		log.Println(i)
	}
	for i := 0; i < int(aiNode.mNumChildren); i++ {
		log.Printf("%T\n", aiNode.mChildren)
		//model.processNode(aiNode.mChildren, aiScene)
	}
}
	/*
	// process each mesh located at the current node
	for(unsigned int i = 0; i < node->mNumMeshes; i++)
	{
		// the node object only contains indices to index the actual objects in the scene. 
		// the scene contains all the data, node is just to keep stuff organized (like relations between nodes).
		aiMesh* mesh = scene->mMeshes[node->mMeshes[i]];
		meshes.push_back(processMesh(mesh, scene));
	}
	// after we've processed all of the meshes (if any) we then recursively process each of the children nodes
	for(unsigned int i = 0; i < node->mNumChildren; i++)
	{
		processNode(node->mChildren[i], scene);
	}
	*/
//}

// processMesh

// loadMaterialTextures

// Not part of class
// TextureFromFile
