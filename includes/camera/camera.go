// Translated from
// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/1.getting_started/7.4.camera_class/camera_class.cpp

package camera

import(
	"math"
	
	"github.com/go-gl/mathgl/mgl32"
)

const FORWARD  uint32 = 0
const BACKWARD uint32 = 1
const LEFT     uint32 = 2
const RIGHT    uint32 = 3

// Default camera values
const YAW         float32 = -90.0
const PITCH       float32 = 0.0
const SPEED       float32 = 2.5
const SENSITIVITY float32 = 0.1
const ZOOM        float32 = 45.0

type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	// Euler Angles
	Yaw      float32
	Pitch    float32

	// Camera options
	MovementSpeed    float32
	MouseSensitivity float32
	Zoom             float32
}

// Construct camera with vectors
func NewCameraV(position mgl32.Vec3, up mgl32.Vec3, yaw float32,
	pitch, movementSpeed, zoom, mouseSen float32) Camera {

	c := Camera{Position: position, WorldUp: up, Yaw: yaw,
		Pitch: pitch, MovementSpeed: movementSpeed, Zoom: zoom,
		MouseSensitivity: mouseSen}
	c.updateCameraVectors()

	return c
}

func NewCamera(posX, posY, posZ, upX, upY, upZ, yaw, pitch,
	movementSpeed, zoom, mouseSen float32) Camera {
	
	position := mgl32.Vec3{posX, posY, posZ}
	up  := mgl32.Vec3{upX, upY, upZ}
	
	c := Camera{Position: position, WorldUp: up, Yaw: yaw,
		Pitch: pitch, MovementSpeed: movementSpeed, Zoom: zoom,
		MouseSensitivity: mouseSen, Front: mgl32.Vec3{0.0, 0.0, -1.0}}
	c.updateCameraVectors()

	return c
}

func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

func (c *Camera) ProcessKeyboard(direction uint32, deltaTime float32) {
	velocity := c.MovementSpeed * deltaTime
	if direction == FORWARD {
		c.Position = c.Position.Add(c.Front.Mul(velocity)) 
	}
	if direction == BACKWARD {
		c.Position = c.Position.Sub(c.Front.Mul(velocity))
	}
	if direction == LEFT {
		c.Position = c.Position.Sub(c.Right.Mul(velocity))
	}
	if direction == RIGHT {
		c.Position = c.Position.Add(c.Right.Mul(velocity))
	}
}

func (c *Camera) ProcessMouseMovement(xOffset, yOffset float32, constrainPitch bool) {
	xOffset *= c.MouseSensitivity
	yOffset *= c.MouseSensitivity

	c.Yaw += xOffset
	c.Pitch += yOffset

	if constrainPitch {
		if c.Pitch > 89.0 {
			c.Pitch = 89.0
		}
		if c.Pitch < -89.0 {
			c.Pitch = -89.0
		}
	}

	c.updateCameraVectors()
}

func (c *Camera) ProcessMouseScroll(yOffset float32) {
	if c.Zoom >= 1.0 && c.Zoom <= 45.0 {
		c.Zoom -= yOffset
	}
	if c.Zoom <= 1.0 {
		c.Zoom = 1.0
	}
	if c.Zoom >= 45.0 {
		c.Zoom = 45.0
	}
}

func (c *Camera) updateCameraVectors() {
	c.Front = mgl32.Vec3{0, 0, 0}

	
	c.Front[0] = float32(math.Cos(float64(mgl32.DegToRad(c.Yaw))) *
		math.Cos(float64(mgl32.DegToRad(c.Pitch))))
	c.Front[1] = float32(math.Sin(float64(mgl32.DegToRad(c.Pitch))))
    c.Front[2] = float32(math.Sin(float64(mgl32.DegToRad(c.Yaw))) *
		math.Cos(float64(mgl32.DegToRad(c.Pitch))))
	c.Front = c.Front.Normalize()
	
	c.Right = c.Front.Cross(c.WorldUp).Normalize()

	c.Up    = c.Right.Cross(c.Front).Normalize()
}
