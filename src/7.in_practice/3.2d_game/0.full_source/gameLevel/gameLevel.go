package gameLevel

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/gameObject"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/resourceManager"
	"github.com/nicholasblaskey/go-learn-opengl/src/7.in_practice/3.2d_game/0.full_source/spriteRenderer"
)

type GameLevel struct {
	Bricks []*gameObject.GameObject
}

func (gl *GameLevel) Load(filePath string, levelWidth, levelHeight uint32) {

	gl.Bricks = []*gameObject.GameObject{}

	// Load from file
	tileData := [][]uint32{}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	input := strings.Split(string(content), "\n")
	for _, line := range input {
		tileRow := []uint32{}
		for i := 0; i < len(line); i += 2 {
			val, err := strconv.Atoi(string(rune(line[i])))
			if err != nil {
				panic(err)
			}
			tileRow = append(tileRow, uint32(val))
		}
		tileData = append(tileData, tileRow)
	}

	if len(tileData) > 0 {
		gl.init(tileData, levelWidth, levelHeight)
	}
}

func (gl *GameLevel) Draw(renderer *spriteRenderer.SpriteRenderer) {
	for _, tile := range gl.Bricks {
		if !tile.Destroyed {
			tile.Draw(renderer)
		}
	}
}

func (gl *GameLevel) IsCompleted() bool {
	for _, tile := range gl.Bricks {
		if !tile.IsSolid && !tile.Destroyed {
			return false
		}
	}
	return true
}

func (gl *GameLevel) init(tileData [][]uint32, levelWidth, levelHeight uint32) {

	// Calculate dimensions
	height := len(tileData)
	width := len(tileData[0])
	unitWidth := float32(levelWidth) / float32(width)
	unitHeight := float32(levelHeight) / float32(height)

	// Init level tiles based on tileData
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			pos := mgl32.Vec2{unitWidth * float32(x), unitHeight * float32(y)}
			size := mgl32.Vec2{unitWidth, unitHeight}
			if tileData[y][x] == 1 {
				obj := gameObject.New(pos, size,
					resourceManager.Textures["block_solid"],
					mgl32.Vec3{0.8, 0.8, 0.7}, mgl32.Vec2{0, 0})
				obj.IsSolid = true
				gl.Bricks = append(gl.Bricks, obj)
			} else if tileData[y][x] > 1 {
				color := mgl32.Vec3{1.0, 1.0, 1.0}
				if tileData[y][x] == 2 {
					color = mgl32.Vec3{0.2, 0.6, 1.0}
				} else if tileData[y][x] == 3 {
					color = mgl32.Vec3{0.0, 0.7, 0.0}
				} else if tileData[y][x] == 4 {
					color = mgl32.Vec3{0.8, 0.8, 0.4}
				} else if tileData[y][x] == 5 {
					color = mgl32.Vec3{1.0, 0.5, 0.0}
				}

				obj := gameObject.New(pos, size,
					resourceManager.Textures["block"], color, mgl32.Vec2{0, 0})
				gl.Bricks = append(gl.Bricks, obj)
			}
		}
	}

}
