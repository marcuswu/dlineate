package core

import "math"

// Vector represents a 2D vector
type Vector struct {
	x float64
	y float64
}

// GetX return the x value of the vector
func (v *Vector) GetX() float64 {
	return v.x
}

// GetY return the y value of the vector
func (v *Vector) GetY() float64 {
	return v.y
}

// Dot product with another vector
func (v *Vector) Dot(u Vector) float64 {
	return v.x*u.x + v.y*u.y
}

// SquareMagnitude returns the squared magnitude of the vector
func (v *Vector) SquareMagnitude() float64 {
	return v.x*v.x + v.y*v.y
}

// Magnitude returns the magnitude of the vector
func (v *Vector) Magnitude() float64 {
	return math.Sqrt(v.SquareMagnitude())
}

// AngleTo returns the angle to another vector in radians
func (v *Vector) AngleTo(u Vector) float64 {
	return math.Acos(v.Dot(u) / (u.Magnitude() * v.Magnitude()))
}

// Rotated returns a vector representing this vector rotated around the origin by angle radians
func (v *Vector) Rotated(angle float64) Vector {
	sinAngle := math.Sin(angle)
	cosAngle := math.Cos(angle)

	newX := v.x*cosAngle - v.y*sinAngle
	newY := v.x*sinAngle + v.y*cosAngle

	return Vector{newX, newY}
}

// Rotate rotates the vector around the origin by angle radians
func (v *Vector) Rotate(angle float64) {
	rotated := v.Rotated(angle)

	v.x = rotated.x
	v.y = rotated.y
}

// Translated returns a vector representing this vector by an x and y distance
func (v *Vector) Translated(dx float64, dy float64) Vector {
	return Vector{v.x + dx, v.y + dy}
}

// Translate translates the vectory by an x and y distance
func (v *Vector) Translate(dx float64, dy float64) {
	translated := v.Translated(dx, dy)

	v.x = translated.x
	v.y = translated.y
}

// UnitVector returns a unit vector with the same direction
func (v *Vector) UnitVector() Vector, bool {
	mag := v.Magnitude()
	if mag == 0 {
		return nil, false
	}
	return Vector{v.x / mag, v.y / mag}, true
}

// Scaled multiplies this vector by a magnitude
func (v *Vector) Scaled(scale float64) {
	v.x *= scale
	v.y *= scale
}
