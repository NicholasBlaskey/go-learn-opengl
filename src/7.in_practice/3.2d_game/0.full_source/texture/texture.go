package texture

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Texture struct {
	ID             uint32
	Width          int32
	Height         int32
	InternalFormat int32
	ImageFormat    uint32
	WrapS          int32
	WrapT          int32
	FilterMin      int32
	FilterMax      int32
}

func New() *Texture {
	t := Texture{
		InternalFormat: gl.RGBA,
		ImageFormat:    gl.RGBA,
		WrapS:          gl.REPEAT,
		WrapT:          gl.REPEAT,
		FilterMin:      gl.LINEAR,
		FilterMax:      gl.LINEAR,
	}
	gl.GenTextures(1, &t.ID)
	return &t
}

func (t *Texture) Generate(width, height int32, data []byte) {
	t.Width, t.Height = width, height
	// Create texture
	gl.BindTexture(gl.TEXTURE_2D, t.ID)

	var dataPtr unsafe.Pointer
	if data != nil {
		dataPtr = gl.Ptr(data)
	}
	gl.TexImage2D(gl.TEXTURE_2D, 0, t.InternalFormat, width, height, 0,
		t.ImageFormat, gl.UNSIGNED_BYTE, dataPtr)

	// Set texture wrap and filter modes
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.WrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.WrapT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.FilterMin)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.FilterMax)
	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (t *Texture) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}
