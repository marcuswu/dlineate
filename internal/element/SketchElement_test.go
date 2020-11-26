package element

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineate/utils"
)

func TestIntersection(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -4)
	var l2 = NewSketchLine(0, 3, -1, 2)
	var result = l1.Intersection(l2)
	if utils.StandardFloatCompare(result.GetX(), 0) != 0 || utils.StandardFloatCompare(result.GetY(), 2) != 0 {
		t.Error("Expected vector(0, 2), got", result)
	}

	l1 = NewSketchLine(0, 1, 2, 1)
	l2 = NewSketchLine(0, 2, 3, 5)
	result = l1.Intersection(l2)
	if utils.StandardFloatCompare(result.GetX(), -7) != 0 || utils.StandardFloatCompare(result.GetY(), 3) != 0 {
		t.Error("Expected vector(-7, 3), got", result)
	}

	l1 = NewSketchLine(0, 0, 1, 3.83)
	l2 = NewSketchLine(0, 1, 0, 0)
	result = l1.Intersection(l2)
	if utils.StandardFloatCompare(result.GetX(), 0) != 0 || utils.StandardFloatCompare(result.GetY(), -3.83) != 0 {
		t.Error("Expected vector(0, -3.83), got", result)
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
}

func TestTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1.5)
	var p1 = NewSketchPoint(0, -0.3, -0.6)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	t.Log("origin distance before translate: ", l1.GetOriginDistance())
	l1.TranslateByElement(p1)
	var pointNearOrigin = l1.PointNearestOrigin()
	t.Log("after translate point nearest origin: ", pointNearOrigin)
	t.Log("origin distance after translate: ", l1.GetOriginDistance())
	if utils.StandardFloatCompare(pointNearOrigin.GetX(), 0) != 0 ||
		utils.StandardFloatCompare(pointNearOrigin.GetY(), 0) != 0 {
		t.Error("Expected Point near origin Point(0, 0), got ", pointNearOrigin)
	}
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
	if utils.StandardFloatCompare(result, .4472135954999579) != 0 {
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
	if utils.StandardFloatCompare(result, -1) != 0 {
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
