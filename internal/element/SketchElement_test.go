package element

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"testing"

	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestPointBasics(t *testing.T) {
	p1 := NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1))

	assert.Equal(t, uint(0), p1.GetID())
	p1.SetID(1)
	assert.Equal(t, uint(1), p1.GetID())
	p1.X.SetFloat64(1)
	assert.Equal(t, p1.GetX().Cmp(big.NewFloat(1)), 0)
	p1.Y.SetFloat64(2)
	assert.Equal(t, p1.GetY().Cmp(big.NewFloat(2)), 0)

	p2 := NewSketchPoint(1, big.NewFloat(1), big.NewFloat(0))
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
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-4)),
			NewSketchLine(0, big.NewFloat(3), big.NewFloat(-1), big.NewFloat(2)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(2)),
		},
		{
			"1x + 2y + 1 = 0 and 2x + 3y + 5 = 0",
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(1)),
			NewSketchLine(0, big.NewFloat(2), big.NewFloat(3), big.NewFloat(5)),
			NewSketchPoint(0, big.NewFloat(-7), big.NewFloat(3)),
		},
		{
			"0x + 1y + 3.83 = 0 and 1x + 0y + 0 = 0",
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(3.83)),
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(-3.83)),
		},
		{
			"1x + 0y + 0 = 0 and 0x + 1y + 3.83 = 0",
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0)),
			NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(3.83)),
			NewSketchPoint(0, big.NewFloat(0), big.NewFloat(-3.83)),
		},
	}
	for _, tt := range tests {
		result := tt.l1.Intersection(tt.l2)
		fmt.Printf("test %s expecting %s, have %s\n", tt.name, tt.intersect.String(), result.String())
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&tt.intersect.X, &result.X))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&tt.intersect.Y, &result.Y))
		// assert.InDelta(t, tt.intersect.X, result.X, utils.StandardCompare, tt.name)
		// assert.InDelta(t, tt.intersect.Y, result.Y, utils.StandardCompare, tt.name)
	}
}

func TestGetSlope(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-4))
	var result = l1.GetSlope()
	if result.Cmp(big.NewFloat(-0.5)) != 0 {
		t.Error("Expected -0.5 slope, got ", result)
	}
}

func TestReverseTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-4))
	var p1 = NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	l1.ReverseTranslateByElement(p1)
	v := &Vector{*big.NewFloat(1), *big.NewFloat(2)}
	v, _ = v.UnitVector()
	t.Log("after translate point nearest origin: ", l1.PointNearestOrigin())
	var negX big.Float
	negX.Neg(v.GetX())
	if utils.StandardBigFloatCompare(l1.GetA(), v.GetX()) != 0 ||
		utils.StandardBigFloatCompare(l1.GetB(), v.GetY()) != 0 ||
		utils.StandardBigFloatCompare(l1.GetC(), &negX) != 0 {
		t.Error("Expected Line(", v.GetX().String(), ",", v.GetY().String(), ",", negX.String(), ") got ", l1)
	}
	l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1.5))
	l2 := NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(0.3)) // PointNearestOrigin = -0.06, -0.12
	p1 = NewSketchPoint(0, big.NewFloat(0.3), big.NewFloat(0.6))
	l1.ReverseTranslateByElement(p1)
	var pointNearOrigin = l1.PointNearestOrigin()
	if utils.StandardBigFloatCompare(pointNearOrigin.GetX(), big.NewFloat(0)) != 0 ||
		utils.StandardBigFloatCompare(pointNearOrigin.GetY(), big.NewFloat(0)) != 0 {
		t.Error("Expected Point near origin Point(0, 0), got ", pointNearOrigin)
	}
	desiredPointNearOrigin := l2.PointNearestOrigin()
	var temp big.Float
	temp.Abs(&desiredPointNearOrigin.X)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(0.06), &temp))
	temp.Abs(&desiredPointNearOrigin.Y)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(0.12), &temp))
	l1.ReverseTranslateByElement(l2)
	pointNearOrigin = l1.PointNearestOrigin()
	assert.Equal(t, 0, utils.StandardBigFloatCompare(temp.Abs(&desiredPointNearOrigin.X), &pointNearOrigin.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(temp.Abs(&desiredPointNearOrigin.Y), &pointNearOrigin.Y))

	p1.X.SetFloat64(0.94)
	p1.Y.SetFloat64(1.38)
	var p2 = NewSketchPoint(0, big.NewFloat(1.3), big.NewFloat(2.1))
	p1.ReverseTranslateByElement(l2)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1.0), &p1.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1.5), &p1.Y))
	p1.ReverseTranslateByElement(p2)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(-0.3), &p1.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(-0.6), &p1.Y))
}

func TestTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1.5))
	var l2 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(0.3)) // PointNearestOrigin = -0.06, -0.12
	var p1 = NewSketchPoint(0, big.NewFloat(-0.3), big.NewFloat(-0.6))
	var p2 = NewSketchPoint(0, big.NewFloat(1.3), big.NewFloat(2.1))
	l1.TranslateByElement(p1)
	var pointNearOrigin = l1.PointNearestOrigin()
	if utils.StandardBigFloatCompare(pointNearOrigin.GetX(), big.NewFloat(0)) != 0 ||
		utils.StandardBigFloatCompare(pointNearOrigin.GetY(), big.NewFloat(0)) != 0 {
		t.Error("Expected Point near origin Point(0, 0), got ", pointNearOrigin)
	}

	desiredPointNearOrigin := l2.PointNearestOrigin()
	assert.Zero(t, utils.StandardBigFloatCompare(big.NewFloat(-0.06), &desiredPointNearOrigin.X))
	assert.Zero(t, utils.StandardBigFloatCompare(big.NewFloat(-0.12), &desiredPointNearOrigin.Y))
	l1.TranslateByElement(l2)
	pointNearOrigin = l1.PointNearestOrigin()
	assert.Zero(t, utils.StandardBigFloatCompare(&desiredPointNearOrigin.X, &pointNearOrigin.X))
	assert.Zero(t, utils.StandardBigFloatCompare(&desiredPointNearOrigin.Y, &pointNearOrigin.Y))

	p1.TranslateByElement(p2)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1.0), &p1.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1.5), &p1.Y))

	p1.TranslateByElement(l2)
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(0.94), &p1.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1.38), &p1.Y))
}

func TestTranslated(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1))
	originalC := l1.GetC()
	nearest1 := l1.PointNearestOrigin()
	result := l1.Translated(big.NewFloat(1), big.NewFloat(1))
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	nearest2 := result.PointNearestOrigin()
	v := &Vector{*big.NewFloat(1), *big.NewFloat(2)}
	v, _ = v.UnitVector()
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardBigFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardBigFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardBigFloatCompare(result.GetC(), big.NewFloat(-1.7888543819998317)) != 0 {
		t.Error("Expected Line(", v.GetX().String(), ",", v.GetY().String(), ", -1.7888543819998317) got ", l1)
	}
	var xDiff, yDiff, nearestDist big.Float
	xDiff.Sub(nearest1.GetX(), nearest2.GetX())
	yDiff.Sub(nearest1.GetY(), nearest2.GetY())
	fmt.Printf("x diff: %s, y diff: %s\n", xDiff.String(), yDiff.String())
	nearestDist.Add(xDiff.Mul(&xDiff, &xDiff), yDiff.Mul(&yDiff, &yDiff))
	if utils.StandardBigFloatCompare(&nearestDist, big.NewFloat(1.8)) != 0 {
		t.Error("Expected nearestDist == 1.8, got ", nearestDist.String())
	}

	result = result.Translated(big.NewFloat(-1), big.NewFloat(-1))
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardBigFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardBigFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardBigFloatCompare(result.GetC(), originalC) != 0 {
		t.Error("Expected Line(", v.GetX().String(), ",", v.GetY().String(), ",", originalC.String(), "), got ", result)
	}
}

func TestTranslatedDistance(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1))
	originalC := l1.GetC()
	v := &Vector{*big.NewFloat(1), *big.NewFloat(2)}
	v, _ = v.UnitVector()
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	t.Log("initial dist to origin: ", l1.distanceToPoint(big.NewFloat(0), big.NewFloat(0)))
	result := l1.TranslatedDistance(big.NewFloat(1))
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	t.Log("after dist to origin: ", result.distanceToPoint(big.NewFloat(0), big.NewFloat(0)))
	if utils.StandardBigFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardBigFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardBigFloatCompare(result.GetC(), big.NewFloat(-1.4472135954999579)) != 0 {
		t.Error("Expected Line(", v.GetX(), ",", v.GetY(), ", -1.4472135954999579), got ", result)
	}

	result = result.TranslatedDistance(big.NewFloat(-1))
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if utils.StandardBigFloatCompare(result.GetA(), v.GetX()) != 0 ||
		utils.StandardBigFloatCompare(result.GetB(), v.GetY()) != 0 ||
		utils.StandardBigFloatCompare(result.GetC(), originalC) != 0 {
		t.Error("Expected Line(", v.GetX().String(), ",", v.GetY().String(), ",", originalC.String(), "), got ", result)
	}
}

func TestPointNearestOrigin(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1))
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.PointNearestOrigin()
	if utils.StandardBigFloatCompare(result.GetX(), big.NewFloat(0.2)) != 0 ||
		utils.StandardBigFloatCompare(result.GetY(), big.NewFloat(0.4)) != 0 {
		t.Error("Expected Point(0.2, 0.4), got ", result)
	}
}

func TestGetOriginDistance(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1))
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.GetOriginDistance()
	if utils.StandardBigFloatCompare(result, big.NewFloat(0.4472135954999579)) != 0 {
		t.Error("Expected .4472135954999579, got ", result)
	}
}

func TestAngleTo(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	result := l1.AngleTo(&Vector{*big.NewFloat(1), *big.NewFloat(0)})
	// var fourtyFive = 45 * math.Pi / 180
	var fourtyFive = big.NewFloat(math.Pi)
	fourtyFive.Quo(fourtyFive, big.NewFloat(180))
	fourtyFive.Mul(fourtyFive, big.NewFloat(45))
	if utils.StandardBigFloatCompare(result, fourtyFive) != 0 {
		t.Errorf("Expected %s, got %s\n", fourtyFive.String(), result.String())
	}
}

func TestRotated(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(1))
	// var fourtyFive = 45 * math.Pi / 180
	var fourtyFive = big.NewFloat(math.Pi)
	fourtyFive.Quo(fourtyFive, big.NewFloat(180))
	fourtyFive.Mul(fourtyFive, big.NewFloat(45))
	result := l1.Rotated(fourtyFive)
	if utils.StandardBigFloatCompare(result.GetA(), big.NewFloat(0)) != 0 {
		t.Errorf("Expected result.GetA() == 0, got %s\n", result.GetA().String())
	}
	if utils.StandardBigFloatCompare(result.GetB(), big.NewFloat(1)) != 0 {
		t.Errorf("Expected result.GetB() == 1, got %s\n", result.GetB().String())
	}
	if utils.StandardBigFloatCompare(result.GetOriginDistance(), l1.GetOriginDistance()) != 0 {
		t.Errorf("Expected result.GetOriginDistance() == %s, got %s\n", l1.GetOriginDistance().String(), result.GetOriginDistance().String())
	}
}

func TestDistanceTo(t *testing.T) {
	var l1 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-1))
	var p1 = NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.DistanceTo(p1)
	if utils.StandardBigFloatCompare(result, big.NewFloat(0.8944271909999159)) != 0 {
		t.Error("Expected 0.8944271909999159, got ", result)
	}

	var l2 = NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(-3.236067977))
	result = l1.DistanceTo(l2)
	if utils.StandardBigFloatCompare(result, big.NewFloat(1)) != 0 {
		t.Error("Expected 1, got ", result)
	}

	result = l1.SquareDistanceTo(p1)
	if utils.StandardBigFloatCompare(result, big.NewFloat(0.8)) != 0 {
		t.Error("Expected 0.8, got ", result)
	}

	result = l1.SquareDistanceTo(l2)
	if utils.StandardBigFloatCompare(result, big.NewFloat(1)) != 0 {
		t.Error("Expected 1, got ", result)
	}

	p2 := NewSketchPoint(0, big.NewFloat(4), big.NewFloat(5))
	result = p1.DistanceTo(p2)
	if utils.StandardBigFloatCompare(result, big.NewFloat(5)) != 0 {
		t.Error("Expected 5, got ", result)
	}

	result = p1.DistanceTo(l1)
	if utils.StandardBigFloatCompare(result, big.NewFloat(0.8944271909999159)) != 0 {
		t.Error("Expected 0.8944271909999159, got ", result)
	}

	result = p1.SquareDistanceTo(l1)
	if utils.StandardBigFloatCompare(result, big.NewFloat(0.8)) != 0 {
		t.Error("Expected 0.8, got ", result)
	}
}

func TestIs(t *testing.T) {
	p1 := NewSketchPoint(0, big.NewFloat(1), big.NewFloat(2))
	p2 := NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	p3 := NewSketchPoint(1, big.NewFloat(4), big.NewFloat(5))

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
		NewSketchLine(1, big.NewFloat(1), big.NewFloat(0), big.NewFloat(0)),
		NewSketchPoint(0, big.NewFloat(0), big.NewFloat(1)),
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
			NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			NewSketchPoint(0, big.NewFloat(2), big.NewFloat(3)),
			&Vector{*big.NewFloat(-1), *big.NewFloat(-2)},
		},
		{
			"Point to line VectorTo",
			NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(1)),
			&Vector{*big.NewFloat(0.8), *big.NewFloat(1.6)},
		},
		{
			"Line to point VectorTo",
			NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(1)),
			NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			&Vector{*big.NewFloat(-0.8), *big.NewFloat(-1.6)},
		},
		{
			"Line to line VectorTo",
			NewSketchLine(0, big.NewFloat(1.0), big.NewFloat(2.0), big.NewFloat(1.0)),
			NewSketchLine(0, big.NewFloat(2.0), big.NewFloat(3.0), big.NewFloat(5.0)),
			&Vector{*big.NewFloat(-0.5692307692), *big.NewFloat(-0.7538461538)},
		},
	}
	for _, tt := range tests {
		t.Logf("e1: %s, e2: %s\n", tt.e1.String(), tt.e2.String())
		v := tt.e1.VectorTo(tt.e2)
		t.Logf("expecting %s, found %s\n", tt.expected.String(), v.String())
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&tt.expected.X, &v.X), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&tt.expected.Y, &v.Y), tt.name)
	}
}

func TestAsPointAsLine(t *testing.T) {
	tests := []struct {
		el SketchElement
	}{
		{NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))},
		{NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(1))},
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
		{NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))},
		{NewSketchLine(0, big.NewFloat(1), big.NewFloat(2), big.NewFloat(1))},
	}
	for _, tt := range tests {
		str := tt.el.String()
		if tt.el.GetType() == Point {
			assert.Contains(t, str, fmt.Sprintf("Point(%d)", 0))
			assert.Contains(t, str, fmt.Sprintf("(%s, %s)", tt.el.AsPoint().X.String(), tt.el.AsPoint().Y.String()))
		} else {
			assert.Contains(t, str, fmt.Sprintf("Line(%d)", 0))
			assert.Contains(t, str, fmt.Sprintf("%sx", tt.el.AsLine().a.String()))
			assert.Contains(t, str, fmt.Sprintf("%sy", tt.el.AsLine().b.String()))
			assert.Contains(t, str, fmt.Sprintf("%s = 0", tt.el.AsLine().c.String()))
		}
	}
}

func TestSketchPointFromVector(t *testing.T) {
	v := Vector{*big.NewFloat(1), *big.NewFloat(2)}
	p := SketchPointFromVector(7, v)

	assert.Equal(t, uint(7), p.GetID())
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(1), &p.X))
	assert.Equal(t, 0, utils.StandardBigFloatCompare(big.NewFloat(2), &p.Y))
}
