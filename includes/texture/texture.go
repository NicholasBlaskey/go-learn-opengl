package texture

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// https://github.com/raedatoui/glutils/blob/master/texture.go
// taken from
func ImageLoad(path string) *image.RGBA {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("Unsupported stride")
	}

	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba
}
