package element

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

func TestLineBasics(t *testing.T) {
	l1 := NewSketchLine(0, 0, 1, 0)

	assert.Equal(t, uint(0), l1.GetID())
	l1.SetID(1)
	assert.Equal(t, uint(1), l1.GetID())
	l1.SetA(1)
	assert.Equal(t, 1.0, l1.GetA())
	l1.SetB(2)
	assert.Equal(t, 2.0, l1.GetB())
	l1.SetC(1)
	assert.Equal(t, 1.0, l1.GetC())

	l2 := NewSketchLine(1, 1, 0, 0)
	assert.True(t, l1.Is(l2))
	assert.False(t, l1.IsEquivalent(l2))
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		line *SketchLine
	}{
		{NewSketchLine(0, 5, 10, 3)},
		{NewSketchLine(1, 3.5, 1.04, 3.7)},
		{NewSketchLine(2, 8.2, 2.8, 9.1)},
	}
	for _, tt := range tests {
		slope := tt.line.GetSlope()
		point := tt.line.PointNearestOrigin()
		tt.line.Normalize()
		assert.Zero(t, utils.StandardFloatCompare(slope, tt.line.GetSlope()))
		assert.Equal(t, point, tt.line.PointNearestOrigin())
	}
}

func TestAngleToLine(t *testing.T) {
	l1 := NewSketchLine(0, -0.611735, -0.791063, 6.155367)
	l2 := NewSketchLine(1, -0.563309, 0.826247, -3.804226)

	a := l1.AngleToLine(l2)
	b := l2.AngleToLine(l1)

	var angle = 108 * math.Pi / 180
	if utils.StandardFloatCompare(a, -angle) != 0 {
		t.Errorf("Expected angle to be -108º (%f), got %f\n", angle, a)
	}
	if utils.StandardFloatCompare(b, angle) != 0 {
		t.Errorf("Expected angle to be -108º (%f), got %f\n", angle, b)
	}
}

func TestNearestPoint(t *testing.T) {
	tests := []struct {
		line    *SketchLine
		point   *SketchPoint
		nearest *SketchPoint
	}{
		{NewSketchLine(0, 0, 1, 0), NewSketchPoint(0, 0, 0), NewSketchPoint(0, 0, 0)},
		{NewSketchLine(0, 0, 1, 0), NewSketchPoint(0, 0, 1), NewSketchPoint(0, 0, 0)},
		{NewSketchLine(0, 0, 1, 0), NewSketchPoint(0, 1, 1), NewSketchPoint(0, 1, 0)},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.nearest, tt.line.NearestPoint(tt.point.X, tt.point.Y))
	}

}

func TestTranslateDistance(t *testing.T) {
	tests := []struct {
		line       *SketchLine
		distance   float64
		translated *SketchLine
	}{
		{NewSketchLine(0, 0, 1, 0), 4, NewSketchLine(0, 0, 1, -4)},
		{NewSketchLine(0, 0, 1, 0), -4, NewSketchLine(0, 0, 1, 4)},
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
	tests := []struct {
		line  *SketchLine
		angle float64
	}{
		{NewSketchLine(0, 1, 4, 7), math.Pi / 6},
		{NewSketchLine(0, 0.67, 1.455, 2.34), math.Pi / 2},
	}
	for _, tt := range tests {
		original := CopySketchElement(tt.line).AsLine()
		tt.line.Rotate(tt.angle)
		assert.InDelta(t, tt.angle, original.AngleToLine(tt.line), utils.StandardCompare)
		assert.InDelta(t, original.GetOriginDistance(), tt.line.GetOriginDistance(), utils.StandardCompare)
	}
}
