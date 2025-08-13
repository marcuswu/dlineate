package element

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

// func (v *Vector) Dot(u *Vector) float64 {
func TestDotProduct(t *testing.T) {
	var posInf, negInf big.Float
	posInf.SetInf(false)
	negInf.SetInf(true)
	tests := []struct {
		name     string
		v1       *Vector
		v2       *Vector
		expected *big.Float
	}{
		{"DotProduct1", &Vector{*big.NewFloat(1.5), *big.NewFloat(2.3)}, &Vector{*big.NewFloat(-0.2), *big.NewFloat(1.45)}, big.NewFloat(3.035)},
		{"DotProduct2", &Vector{*big.NewFloat(-1.1), *big.NewFloat(0.3)}, &Vector{*big.NewFloat(0.25), *big.NewFloat(1.1)}, big.NewFloat(0.055)},
		{"Invalid Values", &Vector{posInf, *big.NewFloat(0.3)}, &Vector{*big.NewFloat(0), *big.NewFloat(1.1)}, big.NewFloat(0)},
	}
	for _, tt := range tests {
		result := tt.v1.Dot(tt.v2)
		assert.Equal(t, tt.expected.String(), result.String())
	}
}

func TestSquareMagnitude(t *testing.T) {
	var posInf, negInf big.Float
	posInf.SetInf(false)
	negInf.SetInf(true)
	tests := []struct {
		name     string
		v1       *Vector
		expected *big.Float
	}{
		{"SquareMagnitude", &Vector{*big.NewFloat(3), *big.NewFloat(4)}, big.NewFloat(25)},
		{"Invalid Values", &Vector{posInf, negInf}, &posInf},
	}
	for _, tt := range tests {
		result := tt.v1.SquareMagnitude()
		assert.Equal(t, tt.expected.String(), result.String(), tt.name)
	}
}

func TestVectorAngleTo(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Vector
		v2       *Vector
		expected *big.Float
	}{
		{
			"AngleTo1",
			&Vector{*big.NewFloat(1.5), *big.NewFloat(2.3)},
			&Vector{*big.NewFloat(-0.2), *big.NewFloat(1.45)},
			big.NewFloat(0.7149681112724342),
		},
		{
			"AngleTo2",
			&Vector{*big.NewFloat(-1.1), *big.NewFloat(0.3)},
			&Vector{*big.NewFloat(0.25), *big.NewFloat(1.1)},
			new(big.Float).Sub(big.NewFloat(4.755164428394983), new(big.Float).Mul(big.NewFloat(math.Pi), big.NewFloat(2))),
		},
		{
			"AngleTo3",
			&Vector{*big.NewFloat(-1.06), *big.NewFloat(0.06)},
			&Vector{*big.NewFloat(-1.06), *big.NewFloat(-0.06)},
			new(big.Float).Add(big.NewFloat(-6.170098433289952), new(big.Float).Mul(big.NewFloat(math.Pi), big.NewFloat(2))),
		},
		{
			"AngleTo4",
			&Vector{*big.NewFloat(-1.06), *big.NewFloat(-0.06)},
			&Vector{*big.NewFloat(-1.06), *big.NewFloat(0.06)},
			new(big.Float).Sub(big.NewFloat(6.170098433289952), new(big.Float).Mul(big.NewFloat(math.Pi), big.NewFloat(2))),
		},
	}
	for _, tt := range tests {
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.expected, tt.v1.AngleTo(tt.v2)))
	}
}

func TestUnitVectorAndScale(t *testing.T) {
	tests := []struct {
		name     string
		v1       *Vector
		scale    *big.Float
		expected bool
	}{
		{"Unit Vector 1", &Vector{*big.NewFloat(1.5), *big.NewFloat(2.3)}, big.NewFloat(2.5), true},
		{"Zero Magnitude", &Vector{*big.NewFloat(0.0), *big.NewFloat(0.0)}, big.NewFloat(1.45), false},
	}
	for _, tt := range tests {
		v, ok := tt.v1.UnitVector()
		assert.Equal(t, tt.expected, ok, tt.name)
		if !tt.expected {
			assert.Nil(t, v, tt.name)
		} else {
			assert.Equal(t, 0, utils.StandardBigFloatCompare(v.Magnitude(), new(big.Float).SetPrec(utils.FloatPrecision).SetFloat64(1)), "check unit magnitude")
			v.Scaled(tt.scale)
			assert.Equal(t, 0, utils.StandardBigFloatCompare(v.Magnitude(), tt.scale), "check scaled magnitude")
		}
	}
}

func TestVectorString(t *testing.T) {
	//Vector((0,0),(%f,%f))
	tests := []struct {
		name string
		v1   *Vector
	}{
		{"Unit Vector 1", &Vector{*big.NewFloat(1.5), *big.NewFloat(2.3)}},
		{"Zero Magnitude", &Vector{*big.NewFloat(0.0), *big.NewFloat(0.0)}},
	}
	for _, tt := range tests {
		str := tt.v1.String()
		assert.Contains(t, str, fmt.Sprintf("(%s,%s)", tt.v1.X.String(), tt.v1.Y.String()))
	}
}
