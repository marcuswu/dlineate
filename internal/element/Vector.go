package element

import (
	"fmt"
	"math"
	"math/big"

	"github.com/marcuswu/dlineate/utils"
)

// Vector represents a 2D vector
type Vector struct {
	X big.Float
	Y big.Float
}

// GetX return the x value of the vector
func (v *Vector) GetX() *big.Float {
	var ret big.Float
	return ret.Copy(&v.X)
}

// GetY return the y value of the vector
func (v *Vector) GetY() *big.Float {
	var ret big.Float
	return ret.Copy(&v.Y)
}

// Dot product with another vector
func (v *Vector) Dot(u *Vector) (ret *big.Float) {
	var x, y, res, zero big.Float
	zero.SetPrec(utils.FloatPrecision).SetFloat64(0)
	defer func() {
		if r := recover(); r != nil {
			ret = &zero
		}
	}()
	x.Mul(&v.X, &u.X)
	y.Mul(&v.Y, &u.Y)
	res.Add(&x, &y)
	return &res
}

// SquareMagnitude returns the squared magnitude of the vector
func (v *Vector) SquareMagnitude() (ret *big.Float) {
	var x, y, res, zero big.Float
	zero.SetPrec(utils.FloatPrecision).SetFloat64(0)
	x.Mul(&v.X, &v.X)
	y.Mul(&v.Y, &v.Y)
	res.Add(&x, &y)
	ret = &res
	return
}

// Magnitude returns the magnitude of the vector
func (v *Vector) Magnitude() *big.Float {
	res := v.SquareMagnitude()
	return res.Sqrt(res)
}

// AngleTo returns the angle to another vector in radians
// https://stackoverflow.com/a/21484228
// With this math, counter clockwise is positive
func (v *Vector) AngleTo(u *Vector) *big.Float {
	// Probably should think about how to handle accuracy values from the conversion
	vX, _ := v.X.Float64()
	vY, _ := v.Y.Float64()
	uX, _ := u.X.Float64()
	uY, _ := u.Y.Float64()

	angle := math.Atan2(uY, uX) - math.Atan2(vY, vX)
	if angle > math.Pi {
		angle -= 2 * math.Pi
	} else if angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	return new(big.Float).SetPrec(utils.FloatPrecision).SetFloat64(angle)
}

// Rotated returns a vector representing this vector rotated around the origin by angle radians
func (v *Vector) Rotated(angle *big.Float) (ret Vector) {
	defer func() {
		if r := recover(); r != nil {
			ret = Vector{*v.GetX(), *v.GetY()}
		}
	}()
	a, _ := angle.Float64()
	var sinAngle, cosAngle big.Float
	sinAngle.SetPrec(utils.FloatPrecision).SetFloat64(math.Sin(a))
	cosAngle.SetPrec(utils.FloatPrecision).SetFloat64(math.Cos(a))

	var xSin, xCos, ySin, yCos big.Float
	xCos.Mul(&v.X, &cosAngle)
	yCos.Mul(&v.Y, &cosAngle)
	xSin.Mul(&v.X, &sinAngle)
	ySin.Mul(&v.Y, &sinAngle)

	return Vector{*xCos.Sub(&xCos, &ySin), *xSin.Add(&xSin, &yCos)}
}

// Rotate rotates the vector around the origin by angle radians
func (v *Vector) Rotate(angle *big.Float) {
	rotated := v.Rotated(angle)

	v.X = rotated.X
	v.Y = rotated.Y
}

// Translated returns a vector representing this vector by an x and y distance
func (v *Vector) Translated(dx *big.Float, dy *big.Float) (ret Vector) {
	defer func() {
		if r := recover(); r != nil {
			ret = Vector{*v.GetX(), *v.GetY()}
		}
	}()
	var x, y big.Float
	x.SetPrec(utils.FloatPrecision)
	y.SetPrec(utils.FloatPrecision)
	return Vector{*x.Add(&v.X, dx), *y.Add(&v.Y, dy)}
}

// Translate translates the vectory by an x and y distance
func (v *Vector) Translate(dx *big.Float, dy *big.Float) {
	translated := v.Translated(dx, dy)

	v.X = translated.X
	v.Y = translated.Y
}

// UnitVector returns a unit vector with the same direction
func (v *Vector) UnitVector() (*Vector, bool) {
	mag := v.Magnitude()
	var x, y, zero big.Float
	x.SetPrec(utils.FloatPrecision)
	y.SetPrec(utils.FloatPrecision)
	zero.SetPrec(utils.FloatPrecision).SetInt64(0)
	if mag.Cmp(&zero) == 0 {
		return nil, false
	}
	return &Vector{*x.Quo(&v.X, mag), *y.Quo(&v.Y, mag)}, true
}

// Scaled multiplies this vector by a magnitude
func (v *Vector) Scaled(scale *big.Float) {
	defer func() {
		if r := recover(); r != nil {
			v.X.SetPrec(utils.FloatPrecision).SetFloat64(0)
			v.Y.SetPrec(utils.FloatPrecision).SetFloat64(0)
		}
	}()
	v.X.Mul(&v.X, scale)
	v.Y.Mul(&v.Y, scale)
}

func (v *Vector) String() string {
	return fmt.Sprintf("Vector(%s,%s)", v.X.String(), v.Y.String())
}
