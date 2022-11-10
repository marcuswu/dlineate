package element

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestPointBasics(t *testing.T) {
	p1 := NewSketchPoint(0, 0, 1)

	assert.Equal(t, uint(0), p1.GetID())
	p1.SetID(1)
	assert.Equal(t, uint(1), p1.GetID())
	p1.X = 1
	assert.Equal(t, 1.0, p1.GetX())
	p1.Y = 2
	assert.Equal(t, 2.0, p1.GetY())

	p2 := NewSketchPoint(1, 1, 0)
	assert.True(t, p1.Is(p2))
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name      string
		l1        *SketchLine
		l2        *SketchLine
		intersect *SketchPoint
	}{
		{
			"1x + 2y -4 = 0 and 3x -1y + 2 = 0",
			NewSketchLine(0, 1, 2, -4),
			NewSketchLine(0, 3, -1, 2),
			NewSketchPoint(0, 0, 2),
		},
		{
			"1x + 2y + 1 = 0 and 2x + 3y + 5 = 0",
			NewSketchLine(0, 1, 2, 1),
			NewSketchLine(0, 2, 3, 5),
			NewSketchPoint(0, -7, 3),
		},
		{
			"0x + 1y + 3.83 = 0 and 1x + 0y + 0 = 0",
			NewSketchLine(0, 0, 1, 3.83),
			NewSketchLine(0, 1, 0, 0),
			NewSketchPoint(0, 0, -3.83),
		},
		{
			"1x + 0y + 0 = 0 and 0x + 1y + 3.83 = 0",
			NewSketchLine(0, 1, 0, 0),
			NewSketchLine(0, 0, 1, 3.83),
			NewSketchPoint(0, 0, -3.83),
		},
	}
	for _, tt := range tests {
		result := tt.l1.Intersection(tt.l2)
		assert.InDelta(t, tt.intersect.X, result.X, utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.intersect.Y, result.Y, utils.StandardCompare, tt.name)
	}
}

func TestGetSlope(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -4)
	var result = l1.GetSlope()
	if result != -0.5 {
		t.Error("Expected -0.5 slope, got ", result)
	}
}

func TestReverseTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -4)
	var p1 = NewSketchPoint(0, 1, 1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	l1.ReverseTranslateByElement(p1)
	v := &Vector{1, 2}
	v, _ = v.UnitVector()
	t.Log("after translate point nearest origin: ", l1.PointNearestOrigin())
	if utils.StandardFloatCompare(l1.GetA(), v.GetX()) != 0 ||
		utils.StandardFloatCompare(l1.GetB(), v.GetY()) != 0 ||
		utils.StandardFloatCompare(l1.GetC(), v.GetX()*-1) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ",", v.GetX()*-1, ") got ", l1)
	}
	l1 = NewSketchLine(0, 1, 2, -1.5)
	l2 := NewSketchLine(0, 1, 2, 0.3) // PointNearestOrigin = -0.06, -0.12
	p1 = NewSketchPoint(0, 0.3, 0.6)
	l1.ReverseTranslateByElement(p1)
	var pointNearOrigin = l1.PointNearestOrigin()
	if utils.StandardFloatCompare(pointNearOrigin.GetX(), 0) != 0 ||
		utils.StandardFloatCompare(pointNearOrigin.GetY(), 0) != 0 {
		t.Error("Expected Point near origin Point(0, 0), got ", pointNearOrigin)
	}
	desiredPointNearOrigin := l2.PointNearestOrigin()
	assert.Zero(t, utils.StandardFloatCompare(0.06, math.Abs(desiredPointNearOrigin.X)))
	assert.Zero(t, utils.StandardFloatCompare(0.12, math.Abs(desiredPointNearOrigin.Y)))
	l1.ReverseTranslateByElement(l2)
	pointNearOrigin = l1.PointNearestOrigin()
	assert.Zero(t, utils.StandardFloatCompare(math.Abs(desiredPointNearOrigin.X), pointNearOrigin.X))
	assert.Zero(t, utils.StandardFloatCompare(math.Abs(desiredPointNearOrigin.Y), pointNearOrigin.Y))

	p1.X = 0.94
	p1.Y = 1.38
	var p2 = NewSketchPoint(0, 1.3, 2.1)
	p1.ReverseTranslateByElement(l2)
	assert.InDelta(t, 1.0, p1.X, utils.StandardCompare)
	assert.InDelta(t, 1.5, p1.Y, utils.StandardCompare)
	p1.ReverseTranslateByElement(p2)
	assert.InDelta(t, -0.3, p1.X, utils.StandardCompare)
	assert.InDelta(t, -0.6, p1.Y, utils.StandardCompare)
}

func TestTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1.5)
	var l2 = NewSketchLine(0, 1, 2, 0.3) // PointNearestOrigin = -0.06, -0.12
	var p1 = NewSketchPoint(0, -0.3, -0.6)
	var p2 = NewSketchPoint(0, 1.3, 2.1)
	l1.TranslateByElement(p1)
	var pointNearOrigin = l1.PointNearestOrigin()
	if utils.StandardFloatCompare(pointNearOrigin.GetX(), 0) != 0 ||
		utils.StandardFloatCompare(pointNearOrigin.GetY(), 0) != 0 {
		t.Error("Expected Point near origin Point(0, 0), got ", pointNearOrigin)
	}

	desiredPointNearOrigin := l2.PointNearestOrigin()
	assert.Zero(t, utils.StandardFloatCompare(-0.06, desiredPointNearOrigin.X))
	assert.Zero(t, utils.StandardFloatCompare(-0.12, desiredPointNearOrigin.Y))
	l1.TranslateByElement(l2)
	pointNearOrigin = l1.PointNearestOrigin()
	assert.Zero(t, utils.StandardFloatCompare(desiredPointNearOrigin.X, pointNearOrigin.X))
	assert.Zero(t, utils.StandardFloatCompare(desiredPointNearOrigin.Y, pointNearOrigin.Y))

	p1.TranslateByElement(p2)
	assert.InDelta(t, 1.0, p1.X, utils.StandardCompare)
	assert.InDelta(t, 1.5, p1.Y, utils.StandardCompare)

	p1.TranslateByElement(l2)
	assert.InDelta(t, 0.94, p1.X, utils.StandardCompare)
	assert.InDelta(t, 1.38, p1.Y, utils.StandardCompare)
}

func TestTranslated(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	originalC := l1.GetC()
	nearest1 := l1.PointNearestOrigin()
	result := l1.Translated(1, 1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	nearest2 := l1.PointNearestOrigin()
	v := &Vector{1, 2}
	v, _ = v.UnitVector()
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardFloatCompare(result.GetC(), -1.7888543819998317) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ", -1.7888543819998317) got ", l1)
	}
	xDiff := nearest1.GetX() - nearest2.GetX()
	yDiff := nearest1.GetY() - nearest2.GetY()
	nearestDist := ((xDiff * xDiff) + (yDiff * yDiff))
	if nearestDist == 1 {
		t.Error("Expected nearestDist == 1, got ", nearestDist)
	}

	result = result.Translated(-1, -1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardFloatCompare(result.GetC(), originalC) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ",", originalC, "), got ", result)
	}
}

func TestTranslatedDistance(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	originalC := l1.GetC()
	v := &Vector{1, 2}
	v, _ = v.UnitVector()
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	t.Log("initial dist to origin: ", l1.distanceToPoint(0, 0))
	result := l1.TranslatedDistance(1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	t.Log("after dist to origin: ", result.distanceToPoint(0, 0))
	if utils.StandardFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardFloatCompare(result.GetC(), -1.4472135954999579) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ", -1.4472135954999579), got ", result)
	}

	result = result.TranslatedDistance(-1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardFloatCompare(result.GetC(), originalC) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ",", originalC, "), got ", result)
	}
}

func TestPointNearestOrigin(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.PointNearestOrigin()
	if utils.StandardFloatCompare(result.GetX(), 0.2) != 0 || utils.StandardFloatCompare(result.GetY(), 0.4) != 0 {
		t.Error("Expected Point(0.2, 0.4), got ", result)
	}
}

func TestGetOriginDistance(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.GetOriginDistance()
	if utils.StandardFloatCompare(result, -0.4472135954999579) != 0 {
		t.Error("Expected .4472135954999579, got ", result)
	}
}

func TestAngleTo(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 1, 1)
	result := l1.AngleTo(&Vector{1, 0})
	var fourtyFive = 45 * math.Pi / 180
	if utils.StandardFloatCompare(result, fourtyFive) != 0 {
		t.Errorf("Expected %f, got %f\n", fourtyFive, result)
	}
}

func TestRotated(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 1, 1)
	var fourtyFive = 45 * math.Pi / 180
	result := l1.Rotated(fourtyFive)
	if utils.StandardFloatCompare(result.GetA(), math.Sqrt(0)) != 0 {
		t.Errorf("Expected result.GetA() == 0, got %f\n", result.GetA())
	}
	if utils.StandardFloatCompare(result.GetB(), 1) != 0 {
		t.Errorf("Expected result.GetB() == √2, got %f\n", result.GetB())
	}
	if utils.StandardFloatCompare(result.GetOriginDistance(), l1.GetOriginDistance()) != 0 {
		t.Errorf("Expected result.GetOriginDistance() == %f, got %f\n", l1.GetOriginDistance(), result.GetOriginDistance())
	}
}

func TestDistanceTo(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	var p1 = NewSketchPoint(0, 1, 1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.DistanceTo(p1)
	if utils.StandardFloatCompare(result, 0.8944271909999159) != 0 {
		t.Error("Expected 0.8944271909999159, got ", result)
	}

	var l2 = NewSketchLine(0, 1, 2, -3.236067977)
	result = l1.DistanceTo(l2)
	if utils.StandardFloatCompare(result, 1) != 0 {
		t.Error("Expected 1, got ", result)
	}

	result = l1.SquareDistanceTo(p1)
	if utils.StandardFloatCompare(result, 0.8) != 0 {
		t.Error("Expected 0.8, got ", result)
	}

	result = l1.SquareDistanceTo(l2)
	if utils.StandardFloatCompare(result, 1) != 0 {
		t.Error("Expected 1, got ", result)
	}

	p2 := NewSketchPoint(0, 4, 5)
	result = p1.DistanceTo(p2)
	if utils.StandardFloatCompare(result, 5) != 0 {
		t.Error("Expected 5, got ", result)
	}

	result = p1.DistanceTo(l1)
	if utils.StandardFloatCompare(result, 0.8944271909999159) != 0 {
		t.Error("Expected 0.8944271909999159, got ", result)
	}

	result = p1.SquareDistanceTo(l1)
	if utils.StandardFloatCompare(result, 0.8) != 0 {
		t.Error("Expected 0.8, got ", result)
	}
}

func TestIs(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 2)
	p2 := NewSketchPoint(0, 1, 1)
	p3 := NewSketchPoint(1, 4, 5)

	if !p1.Is(p2) {
		t.Error("Expected p1 is p2, got ", false)
	}

	if p1.Is(p3) {
		t.Error("Expected p1 is not p3, got ", false)
	}
}

func TestElementTypeString(t *testing.T) {
	tests := []struct {
		elementType Type
		expected    string
	}{
		{Point, "Point"},
		{Line, "Line"},
		{7, "7"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.elementType.String())
	}
}

func TestConstraintLevelString(t *testing.T) {
	tests := []struct {
		level    ConstraintLevel
		expected string
	}{
		{OverConstrained, "over constrained"},
		{UnderConstrained, "under constrained"},
		{FullyConstrained, "fully constrained"},
		{7, "7"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.level.String())
	}
}

func TestCopyAndSort(t *testing.T) {
	elementList := List{
		NewSketchLine(1, 1, 0, 0),
		NewSketchPoint(0, 0, 1),
	}

	elementList = append(elementList, CopySketchElement(elementList[1]))
	elementList = append(elementList, CopySketchElement(elementList[0]))
	sort.Sort(elementList)

	sortOrder := []uint{0, 0, 1, 1}

	for i, tt := range elementList {
		assert.Equal(t, sortOrder[i], tt.GetID())

		str := tt.ToGraphViz(7)
		assert.True(t, strings.Contains(str, fmt.Sprintf("7-%d", tt.GetID())))
		assert.True(t, strings.Contains(str, fmt.Sprintf("label=%d", tt.GetID())))
	}
}

func TestVectorTo(t *testing.T) {
	tests := []struct {
		name     string
		e1       SketchElement
		e2       SketchElement
		expected *Vector
	}{
		{
			"Point to point VectorTo",
			NewSketchPoint(0, 1, 1),
			NewSketchPoint(0, 2, 3),
			&Vector{-1, -2},
		},
		{
			"Point to line VectorTo",
			NewSketchPoint(0, 1, 1),
			NewSketchLine(0, 1, 2, 1),
			&Vector{0.8, 1.6},
		},
		{
			"Line to point VectorTo",
			NewSketchLine(0, 1, 2, 1),
			NewSketchPoint(0, 1, 1),
			&Vector{-0.8, -1.6},
		},
		{
			"Line to line VectorTo",
			NewSketchLine(0, 1, 2, 1),
			NewSketchLine(0, 2, 3, 5),
			&Vector{-0.5692308, -0.7538462},
		},
	}
	for _, tt := range tests {
		v := tt.e1.VectorTo(tt.e2)
		assert.InDelta(t, tt.expected.X, v.X, utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.expected.Y, v.Y, utils.StandardCompare, tt.name)
	}
}

func TestAsPointAsLine(t *testing.T) {
	tests := []struct {
		el SketchElement
	}{
		{NewSketchPoint(0, 1, 1)},
		{NewSketchLine(0, 1, 2, 1)},
	}
	for _, tt := range tests {
		p := tt.el.AsPoint()
		l := tt.el.AsLine()
		if tt.el.GetType() == Point {
			assert.NotNil(t, p)
			assert.Nil(t, l)
		} else {
			assert.NotNil(t, l)
			assert.Nil(t, p)
		}
	}
}

func TestElementString(t *testing.T) {
	tests := []struct {
		el SketchElement
	}{
		{NewSketchPoint(0, 1, 1)},
		{NewSketchLine(0, 1, 2, 1)},
	}
	for _, tt := range tests {
		str := tt.el.String()
		if tt.el.GetType() == Point {
			assert.Contains(t, str, fmt.Sprintf("Point(%d)", 0))
			assert.Contains(t, str, fmt.Sprintf("(%f, %f)", tt.el.AsPoint().X, tt.el.AsPoint().Y))
		} else {
			assert.Contains(t, str, fmt.Sprintf("Line(%d)", 0))
			assert.Contains(t, str, fmt.Sprintf("%fx", tt.el.AsLine().a))
			assert.Contains(t, str, fmt.Sprintf("%fy", tt.el.AsLine().b))
			assert.Contains(t, str, fmt.Sprintf("%f = 0", tt.el.AsLine().c))
		}
	}
}

func TestSketchPointFromVector(t *testing.T) {
	v := Vector{1, 2}
	p := SketchPointFromVector(7, v)

	assert.Equal(t, uint(7), p.GetID())
	assert.Equal(t, float64(1), p.X)
	assert.Equal(t, float64(2), p.Y)
}
