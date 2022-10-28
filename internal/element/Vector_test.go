package element

import (
	"fmt"
	"math"
	"testing"

	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

// func (v *Vector) Dot(u *Vector) float64 {
func TestDotProduct(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Vector
		v2       *Vector
		expected float64
	}{
		{"DotProduct1", &Vector{1.5, 2.3}, &Vector{-0.2, 1.45}, 3.035},
		{"DotProduct2", &Vector{-1.1, 0.3}, &Vector{0.25, 1.1}, 0.055},
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.expected, tt.v1.Dot(tt.v2), utils.StandardCompare)
	}
}

func TestVectorAngleTo(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Vector
		v2       *Vector
		expected float64
	}{
		{"AngleTo1", &Vector{1.5, 2.3}, &Vector{-0.2, 1.45}, 0.7149681112724342},
		{"AngleTo2", &Vector{-1.1, 0.3}, &Vector{0.25, 1.1}, 4.755164428394983 - (math.Pi * 2)},
		{"AngleTo3", &Vector{-1.06, 0.06}, &Vector{-1.06, -0.06}, -6.170098433289952 + (math.Pi * 2)},
		{"AngleTo4", &Vector{-1.06, -0.06}, &Vector{-1.06, 0.06}, 6.170098433289952 - (math.Pi * 2)},
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.expected, tt.v1.AngleTo(tt.v2), utils.StandardCompare, tt.name)
	}
}

func TestUnitVectorAndScale(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Vector
		scale    float64
		expected bool
	}{
		{"Unit Vector 1", &Vector{1.5, 2.3}, 2.5, true},
		{"Zero Magnitude", &Vector{0.0, 0.0}, 1.45, false},
	}
	for _, tt := range tests {
		v, ok := tt.v1.UnitVector()
		assert.Equal(t, tt.expected, ok, tt.name)
		if !tt.expected {
			assert.Nil(t, v, tt.name)
		} else {
			assert.InDelta(t, 1, v.Magnitude(), utils.StandardCompare, tt.name)
			v.Scaled(tt.scale)
			assert.InDelta(t, tt.scale, v.Magnitude(), utils.StandardCompare, tt.name)
		}
	}
}

func TestVectorString(t *testing.T) {
	//Vector((0,0),(%f,%f))
	tests := []struct {
		name string
		v1   *Vector
	}{
		{"Unit Vector 1", &Vector{1.5, 2.3}},
		{"Zero Magnitude", &Vector{0.0, 0.0}},
	}
	for _, tt := range tests {
		str := tt.v1.String()
		assert.Contains(t, str, fmt.Sprintf("(%f,%f)", tt.v1.X, tt.v1.Y))
	}
}
