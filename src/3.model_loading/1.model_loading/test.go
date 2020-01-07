package main

import(
	"log"
	"fmt"
	
	"github.com/tbogdala/assimp-go"
)

func main() {
	srcMeshes, err := assimp.ParseFile(
		"../../../resources/objects/nanosuit/nanosuit.obj")
	log.Println(err)

	fmt.Printf("%+v\n", srcMeshes[0].UVChannels)
}
