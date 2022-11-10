package dlineation

import (
	"testing"

	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

func TestDotProduct(t *testing.T) {
	v1 := NewVector(1, 2, 3)
	v2 := NewVector(4, 5, 6)
	dot := v1.Dot(v2)
	assert.Equal(t, 32.0, dot, "dot product")
}

func TestSquareMagnitude(t *testing.T) {
	v1 := NewVector(1, 2, 3)
	sqMag := v1.SquareMagnitude()
	assert.Equal(t, 14.0, sqMag, "square magnitude")
}

func TestMagnitude(t *testing.T) {
	v1 := NewVector(1, 2, 3)
	mag := v1.Magnitude()
	assert.InDelta(t, 3.741657, mag, utils.StandardCompare, "magnitude")
}

func TestUnitVector(t *testing.T) {
	v1 := NewVector(1, 2, 3)
	vec, ok := v1.UnitVector()
	assert.True(t, ok, "unit vector success")
	assert.InDelta(t, 0.267261, vec.X, utils.StandardCompare, "unit vector X")
	assert.InDelta(t, 0.534522, vec.Y, utils.StandardCompare, "unit vector Y")
	assert.InDelta(t, 0.801784, vec.Z, utils.StandardCompare, "unit vector Z")

	v2 := NewVector(0, 0, 0)
	vec, ok = v2.UnitVector()
	assert.False(t, ok, "unit vector fail")
	assert.Nil(t, vec, "unit vector fail")
}
