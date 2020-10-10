package core

import (
	"testing"

	"github.com/marcuswu/dlineation/utils"
)

func TestPointFromPoints(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 1)
	p2 := NewSketchPoint(1, 3, 5)
	p3 := NewSketchPoint(2, 0, 2)

	newP3, state := pointFromPoints(p1, p2, p3, 1, 3)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state = pointFromPoints(p1, p2, p3, 1, 5)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP3.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	// p1, p2, and p3 should remain the same
	if p1.GetX() != 1 || p1.GetY() != 1 {
		t.Error("Expected p1 to remain at 1, 1, got: ", p1)
	}
	if p2.GetX() != 3 || p2.GetY() != 5 {
		t.Error("Expected p2 to remain at 3, 5, got: ", p2)
	}
	if p3.GetX() != 0 || p3.GetY() != 2 {
		t.Error("Expected p3 to remain at 0, 2, got: ", p3)
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(newP3), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(newP3))
	}

	p3 = NewSketchPoint(2, 2, 1)
	var newP32 SketchElement

	newP32, state = pointFromPoints(p1, p2, p3, 1, 5)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP32.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP32), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(newP32), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(newP3.GetX(), newP32.GetX()) == 0 {
		t.Error("Expected newP3 and newP32 to be different points, got ", newP3, newP32)
	}

	if utils.StandardFloatCompare(newP3.GetY(), newP32.GetY()) == 0 {
		t.Error("Expected newP3 and newP32 to be different points, got ", newP3, newP32)
	}
}

func TestPointFromPointsExt(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 1)
	p2 := NewSketchPoint(1, 3, 5)
	p3 := NewSketchPoint(2, 0, 2)

	referenceP3, state := pointFromPoints(p1, p2, p3, 1, 5)

	if utils.StandardFloatCompare(p1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(referenceP3), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(referenceP3))
	}

	c1 := Constraint{
		id:             0,
		constraintType: Distance,
		value:          1,
		element1:       p1,
		element2:       p3,
	}

	c2 := Constraint{
		id:             1,
		constraintType: Distance,
		value:          5,
		element1:       p2,
		element2:       p3,
	}

	newP3, state := PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.element1, c1.element2 = c1.element2, c1.element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	c1.element1, c1.element2 = c1.element2, c1.element1
	c2.element1, c2.element2 = c2.element2, c2.element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	c1.element1, c1.element2 = c1.element2, c1.element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}
}

func TestPointFromPointLine(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 1)
	l2 := NewSketchLine(1, 1, 1, 2)
	p3 := NewSketchPoint(2, 0, 2)

	newP3, state := pointFromPointLine(p1, l2, p3, 1, 1)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state = pointFromPointLine(p1, l2, p3, 1, 2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP3.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	// p1, p2, and p3 should remain the same
	if p1.GetX() != 1 || p1.GetY() != 1 {
		t.Error("Expected p1 to remain at 1, 1, got: ", p1)
	}
	if l2.GetX() != 1 || l2.GetY() != 1 || l2.GetC() != 2 {
		t.Error("Expected l2 to remain at 1, 1, 2 got: ", l2)
	}
	if p3.GetX() != 0 || p3.GetY() != 2 {
		t.Error("Expected p3 to remain at 0, 2, got: ", p3)
	}

	if utils.StandardFloatCompare(p1.DistanceTo(newP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got", p1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(newP3), 2) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got", l2.DistanceTo(newP3))
	}

	p3 = NewSketchPoint(2, 2, 1)

	_, state = pointFromPointLine(p1, l2, p3, 1, 5)

	if state != NonConvergent {
		t.Error("Expected non convergent state got ", state)
	}
}

func TestPointFromPointLineExt(t *testing.T) {
	p1 := NewSketchPoint(0, 1, 1)
	l2 := NewSketchLine(1, 1, 1, 2)
	p3 := NewSketchPoint(2, 0, 2)

	referenceP3, state := pointFromPointLine(p1, l2, p3, 1, 2.5)

	if utils.StandardFloatCompare(p1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(referenceP3), 2.5) != 0 {
		t.Error("Expected newP3 to have distance of 2.5 to p2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := Constraint{
		id:             0,
		constraintType: Distance,
		value:          1,
		element1:       p1,
		element2:       p3,
	}

	c2 := Constraint{
		id:             1,
		constraintType: Distance,
		value:          2.5,
		element1:       l2,
		element2:       p3,
	}

	newP3, state := PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.element1, c1.element2 = c1.element2, c1.element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	c1.element1, c1.element2 = c1.element2, c1.element1
	c2.element1, c2.element2 = c2.element2, c2.element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	c1.element1, c1.element2 = c1.element2, c1.element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3, newP3)
	}
}
