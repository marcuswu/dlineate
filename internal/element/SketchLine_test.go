package element

import (
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestLineBasics(t *testing.T) {
	var zero, one, two big.Float
	zero.SetFloat64(0)
	one.SetFloat64(1)
	two.SetFloat64(2)
	l1 := NewSketchLine(0, &zero, &one, &zero)

	assert.Equal(t, uint(0), l1.GetID())
	l1.SetID(1)
	assert.Equal(t, uint(1), l1.GetID())
	l1.SetA(&one)
	assert.Equal(t, one.Cmp(l1.GetA()), 0)
	l1.SetB(&two)
	assert.Equal(t, two.Cmp(l1.GetB()), 0)
	l1.SetC(&one)
	assert.Equal(t, one.Cmp(l1.GetC()), 0)

	l2 := NewSketchLine(1, &one, &zero, &zero)
	assert.True(t, l1.Is(l2))
	assert.False(t, l1.IsEquivalent(l2))
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		line *SketchLine
	}{
		{NewSketchLine(0, big.NewFloat(5), big.NewFloat(10), big.NewFloat(3))},
		{NewSketchLine(1, big.NewFloat(3.5), big.NewFloat(1.04), big.NewFloat(3.7))},
		{NewSketchLine(2, big.NewFloat(8.2), big.NewFloat(2.8), big.NewFloat(9.1))},
	}
	for _, tt := range tests {
		slope := tt.line.GetSlope()
		point := tt.line.PointNearestOrigin()
		tt.line.Normalize()
		assert.Equal(t, 0, utils.StandardBigFloatCompare(slope, tt.line.GetSlope()))
		assert.True(t, point.Is(tt.line.PointNearestOrigin()))
	}
}

func TestAngleToLine(t *testing.T) {
	// 72 * math.Pi / 180
	angle := big.NewFloat(math.Pi)
	angle.Quo(angle, big.NewFloat(180))
	angle.Mul(angle, big.NewFloat(72))
	l1 := NewSketchLine(0, big.NewFloat(0.5877852523), big.NewFloat(0.8090169944), big.NewFloat(-6.155367074))
	l2 := NewSketchLine(1, big.NewFloat(-0.5877852523), big.NewFloat(0.8090169944), big.NewFloat(-3.804226065))

	a := l1.AngleToLine(l2)
	b := l2.AngleToLine(l1)

	var negAngle big.Float
	negAngle.Neg(angle)
	if utils.StandardBigFloatCompare(a, angle) != 0 {
		t.Errorf("Expected angle to be 108º (%s), got %s\n", negAngle.String(), a.String())
	}
	if utils.StandardBigFloatCompare(b, &negAngle) != 0 {
		t.Errorf("Expected angle to be -108º (%s), got %s\n", angle.String(), b.String())
	}
}

func TestNearestPoint(t *testing.T) {
	tests := []struct {
		line    *SketchLine
		point   *SketchPoint
		nearest *SketchPoint
	}{
		{
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0)),
		},
		{
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(0)),
		},
		{
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			NewSketchPoint(0, big.NewFloat(1), big.NewFloat(0)),
		},
	}
	for _, tt := range tests {
		assert.True(t, tt.nearest.Is(tt.line.NearestPoint(&tt.point.X, &tt.point.Y)))
	}

}

func TestTranslateDistance(t *testing.T) {
	tests := []struct {
		line       *SketchLine
		distance   *big.Float
		translated *SketchLine
	}{
		{
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			big.NewFloat(4),
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(-4)),
		},
		{
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			big.NewFloat(-4),
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(4))},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.translated, tt.line.TranslatedDistance(tt.distance))
	}
	for _, tt := range tests {
		tt.line.TranslateDistance(tt.distance)
		assert.Equal(t, tt.translated, tt.line)
	}
}

func TestRotate(t *testing.T) {
	pi := big.NewFloat(math.Pi)
	tests := []struct {
		line  *SketchLine
		angle *big.Float
	}{
		{NewSketchLine(0, big.NewFloat(1), big.NewFloat(4), big.NewFloat(7)), big.NewFloat(0).Quo(pi, big.NewFloat(6))},
		{NewSketchLine(0, big.NewFloat(0.67), big.NewFloat(1.455), big.NewFloat(2.34)), big.NewFloat(0).Quo(pi, big.NewFloat(2))},
	}
	for _, tt := range tests {
		original := CopySketchElement(tt.line).AsLine()
		tt.line.Rotate(tt.angle)

		flipRotAngle := big.NewFloat(math.Pi)
		flipRotAngle.Sub(flipRotAngle, tt.angle)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.angle, original.AngleToLine(tt.line)))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(original.GetOriginDistance(), tt.line.GetOriginDistance()))
	}
}
