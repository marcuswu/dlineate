package solver

import (
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
)

func TestPointFromPoints(t *testing.T) {
	p1 := el.NewSketchPoint(0, 1, 1)
	p2 := el.NewSketchPoint(1, 3, 5)
	p3 := el.NewSketchPoint(2, 0, 2)

	newP3, state := GetPointFromPoints(p1, p2, p3, 1, 3)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state = GetPointFromPoints(p1, p2, p3, 1, 5)

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

	p3 = el.NewSketchPoint(2, 2, 1)
	var newP32 *el.SketchPoint

	newP32, state = GetPointFromPoints(p1, p2, p3, 1, 5)

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
	p1 := el.NewSketchPoint(0, 1, 1)
	p2 := el.NewSketchPoint(1, 3, 5)
	p3 := el.NewSketchPoint(2, 0, 2)

	referenceP3, state := GetPointFromPoints(p1, p2, p3, 1, 5)

	if utils.StandardFloatCompare(p1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(p2.DistanceTo(referenceP3), 5) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p3, 1)

	c2 := constraint.NewConstraint(1, constraint.Distance, p2, p3, 5)

	state = PointFromPoints(c1, c2)
	newP3 := c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPoints(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromPoints(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPoints(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	state = PointFromPoints(c1, c2)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPoints(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromPoints(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPoints(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromPointLine(t *testing.T) {
	p1 := el.NewSketchPoint(0, 1, 1)
	l2 := el.NewSketchLine(1, 1, 1, 2)
	p3 := el.NewSketchPoint(2, 0, 2)

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
	if l2.GetA() != 1 || l2.GetB() != 1 || l2.GetC() != 2 {
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

	p3 = el.NewSketchPoint(2, 2, 1)

	_, state = pointFromPointLine(p1, l2, p3, 1, 5)

	if state != NonConvergent {
		t.Error("Expected non convergent state got ", state)
	}
}

func TestPointFromPointLineExt(t *testing.T) {
	p1 := el.NewSketchPoint(0, 1, 1)
	l2 := el.NewSketchLine(1, 1, 1, 2)
	p3 := el.NewSketchPoint(2, 0, 2)

	referenceP3, state := pointFromPointLine(p1, l2, p3, 1, 2.5)

	if utils.StandardFloatCompare(p1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(referenceP3), 2.5) != 0 {
		t.Error("Expected newP3 to have distance of 2.5 to p2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p3, 1)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2, p3, 2.5)

	state = PointFromPointLine(c1, c2)
	newP3 := c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPointLine(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromPointLine(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPointLine(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	state = PointFromPointLine(c1, c2)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPointLine(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromPointLine(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromPointLine(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromLineLine(t *testing.T) {
	l1 := el.NewSketchLine(0, 1, 1, -1)
	l2 := el.NewSketchLine(1, 1, 1, 1)
	p3 := el.NewSketchPoint(2, 0.7, 1)

	newP3, state := pointFromLineLine(l1, l2, p3, 1, 1)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	l2 = el.NewSketchLine(0, -1, 1, 1)
	newP3, state = pointFromLineLine(l1, l2, p3, 1, 2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if newP3.GetID() != 2 {
		t.Error("Expected newP3 to have id 2, got ", newP3.GetID())
	}

	// p1, p2, and p3 should remain the same
	if l1.GetA() != 1 || l1.GetB() != 1 || l1.GetC() != -1 {
		t.Error("Expected l1 to remain at 1, 1, -1 got: ", l1)
	}
	if l2.GetA() != -1 || l2.GetB() != 1 || l2.GetC() != 1 {
		t.Error("Expected l2 to remain at -1, 1, 1 got: ", l2)
	}
	if p3.GetX() != 0.7 || p3.GetY() != 1 {
		t.Error("Expected p3 to remain at 0.7, 1, got: ", p3)
	}

	if utils.StandardFloatCompare(l1.DistanceTo(newP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to l1, got", l1.DistanceTo(newP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(newP3), 2) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got", l2.DistanceTo(newP3))
	}
}
func TestPointFromLineLineExt(t *testing.T) {
	l1 := el.NewSketchLine(0, 1, 1, -1)
	l2 := el.NewSketchLine(1, -1, 1, 1)
	p3 := el.NewSketchPoint(2, 0.7, 1)

	referenceP3, state := pointFromLineLine(l1, l2, p3, 1, 2)

	if utils.StandardFloatCompare(l1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to l1, got ", l1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(referenceP3), 2) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, l1, p3, 1)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2, p3, 2)

	state = PointFromLineLine(c1, c2)
	newP3 := c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromLineLine(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromLineLine(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromLineLine(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	state = PointFromLineLine(c1, c2)
	newP3 = c1.Element2.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromLineLine(c2, c1)
	newP3 = c1.Element2.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	state = PointFromLineLine(c1, c2)
	newP3 = c1.Element1.(*el.SketchPoint)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	state = PointFromLineLine(c2, c1)
	newP3 = c1.Element1.(*el.SketchPoint)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}
