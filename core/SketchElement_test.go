package core

import (
	"testing"

	"github.com/marcuswu/dlineation/utils"
)

func TestIntersection(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -4)
	var l2 = NewSketchLine(0, 3, -1, 2)
	var result = l1.Intersection(l2)
	if result.GetX() != 0 && result.GetY() != 2 {
		t.Error("Expected vector(0, 2), got ", result)
	}

	l1 = NewSketchLine(0, 1, 2, 1)
	l2 = NewSketchLine(0, 2, 3, 5)
	result = l1.Intersection(l2)
	if result.GetX() != -7 && result.GetY() != 3 {
		t.Error("Expected vector(-7, 3), got ", result)
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
	t.Log("after translate point nearest origin: ", l1.PointNearestOrigin())
	if l1.GetA() != 1 || l1.GetB() != 2 || l1.GetC() != -1 {
		t.Error("Expected Line(1, 2, -1), got ", l1)
	}
}

func TestTranslateByElement(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	var p1 = NewSketchPoint(0, 1, 1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	l1.TranslateByElement(p1)
	t.Log("after translate point nearest origin: ", l1.PointNearestOrigin())
	if l1.GetA() != 1 || l1.GetB() != 2 || l1.GetC() != -4 {
		t.Error("Expected Line(1, 2, -4), got ", l1)
	}
}

func TestTranslated(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.Translated(1, 1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if result.GetA() != 1 || result.GetB() != 2 || result.GetC() != -4 {
		t.Error("Expected Line(1, 2, -4), got ", result)
	}

	result = result.Translated(-1, -1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if result.GetA() != 1 || result.GetB() != 2 || result.GetC() != -1 {
		t.Error("Expected Line(1, 2, -1), got ", result)
	}
}

func TestTranslateDistance(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.TranslateDistance(1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if result.GetA() != 1 || result.GetB() != 2 || utils.StandardFloatCompare(result.GetC(), -3.236067977) != 0 {
		t.Error("Expected Line(1, 2, -3.236067977), got ", result)
	}

	result = result.TranslateDistance(-1)
	t.Log("after translate point nearest origin: ", result.PointNearestOrigin())
	if result.GetA() != 1 || result.GetB() != 2 || utils.StandardFloatCompare(result.GetC(), -1) != 0 {
		t.Error("Expected Line(1, 2, -1), got ", result)
	}
}

func TestPointNearestOrigin(t *testing.T) {
	var l1 = NewSketchLine(0, 1, 2, -1)
	t.Log("before translate point nearest origin: ", l1.PointNearestOrigin())
	result := l1.PointNearestOrigin()
	if result.GetX() != 0.2 || result.GetY() != 0.4 {
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
