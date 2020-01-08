// translated from

package model

import(
	//"log"
	"unsafe"
	"strconv"
	"strings"
	
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v4.1-core/gl"

	"github.com/nicholasblaskey/go-learn-opengl/includes/mesh"
	"github.com/nicholasblaskey/go-learn-opengl/includes/shader"
)

type Model struct {
	texturesLoaded  []mesh.Texture
	meshes          []mesh.Mesh
	directory       string
	gammaCorrection bool
}

func newModel(path string, gamma bool) *Model {
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
	/*
	// read file via ASSIMP
	Assimp::Importer importer;
	const aiScene* scene = importer.ReadFile(path, aiProcess_Triangulate | aiProcess_FlipUVs | aiProcess_CalcTangentSpace);
	// check for errors
	if(!scene || scene->mFlags & AI_SCENE_FLAGS_INCOMPLETE || !scene->mRootNode) // if is Not Zero
	{
		cout << "ERROR::ASSIMP:: " << importer.GetErrorString() << endl;
		return;
	}
	*/

	// Retrieve the directory of the filepath
	model.directory = path[0:strings.LastIndex(path, "/")]

	// processNode(scene->mRootNode, scene);
}

func (model *Model) processNode(aiNode *node, aiScene *scene) {
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
}

// processMesh

// loadMaterialTextures

// Not part of class
// TextureFromFile
