package solver

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineation/internal/constraint"
	el "github.com/marcuswu/dlineation/internal/element"
	"github.com/marcuswu/dlineation/utils"
	"github.com/stretchr/testify/assert"
)

func TestPointFromPoints(t *testing.T) {
	tests := []struct {
		name   string
		p1     *el.SketchPoint
		p1Dist float64
		p2     *el.SketchPoint
		p2Dist float64
		p3     *el.SketchPoint
		state  SolveState
	}{
		{
			"Test Nonconvergent",
			el.NewSketchPoint(0, 1, 1),
			1.0,
			el.NewSketchPoint(1, 3, 5),
			3.0,
			el.NewSketchPoint(2, 0, 2),
			NonConvergent,
		},
		{
			"Test 1",
			el.NewSketchPoint(0, 1, 1),
			1.0,
			el.NewSketchPoint(1, 3, 5),
			5.0,
			el.NewSketchPoint(2, 0, 2),
			Solved,
		},
		{
			"Test exact distance",
			el.NewSketchPoint(0, 3, 1),
			1.0,
			el.NewSketchPoint(1, 3, 5),
			3.0,
			el.NewSketchPoint(2, 2, 2),
			Solved,
		},
	}
	for _, tt := range tests {
		newP3, state := GetPointFromPoints(tt.p1, tt.p2, tt.p3, tt.p1Dist, tt.p2Dist)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == NonConvergent {
			continue
		}
		assert.Equal(t, tt.p3.GetID(), newP3.GetID(), tt.name)
		assert.InDelta(t, math.Abs(tt.p1.DistanceTo(newP3)), tt.p1Dist, utils.StandardCompare, tt.name)
		assert.InDelta(t, math.Abs(tt.p2.DistanceTo(newP3)), tt.p2Dist, utils.StandardCompare, tt.name)
	}
}

func TestPointFromPointsExt(t *testing.T) {
	var newP3 *el.SketchPoint = nil
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

	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p3, 1, false)

	c2 := constraint.NewConstraint(1, constraint.Distance, p2, p3, 5, false)

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
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

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPoints(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPoints(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromPointLine(t *testing.T) {
	p1 := el.NewSketchPoint(0, 1, 1)
	l2 := el.NewSketchLine(1, 1, 1, 2*math.Sqrt(0.5))
	p3 := el.NewSketchPoint(2, 0, 2)

	_, state := pointFromPointLine(p1, l2, p3, 1, 1)

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state := pointFromPointLine(p1, l2, p3, 1, 2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
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

	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p3, 1, false)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2, p3, 2.5, false)

	newP3, state := PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
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

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPointLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromLineLine(t *testing.T) {
	tests := []struct {
		name   string
		l1     *el.SketchLine
		l1Dist float64
		l2     *el.SketchLine
		l2Dist float64
		p3     *el.SketchPoint
		state  SolveState
	}{
		{
			"Test Nonconvergent",
			el.NewSketchLine(0, 1, 1, -1),
			1.0,
			el.NewSketchLine(1, 1, 1, 1),
			1.0,
			el.NewSketchPoint(2, 0.7, 1),
			NonConvergent,
		},
		{
			"Test Parallel",
			el.NewSketchLine(0, 0, 1, -1),
			1.0,
			el.NewSketchLine(1, 0, 1, 0),
			2.0,
			el.NewSketchPoint(2, 0.7, 1.8),
			Solved,
		},
		{
			"Test Intersect 1",
			el.NewSketchLine(0, 1, 1, -1),
			1.0,
			el.NewSketchLine(1, -1, 1, 1),
			2.0,
			el.NewSketchPoint(2, 0.7, 1),
			Solved,
		},
		{
			"Test Intersect 2",
			el.NewSketchLine(0, 1, 1, -1),
			1.0,
			el.NewSketchLine(1, -1, 1, 1),
			2.0,
			el.NewSketchPoint(2, 3, 0),
			Solved,
		},
		{
			"Test Intersect 3",
			el.NewSketchLine(0, 1, 1, -1),
			1.0,
			el.NewSketchLine(1, -1, 1, 1),
			2.0,
			el.NewSketchPoint(2, -1, 0),
			Solved,
		},
		{
			"Test Intersect 4",
			el.NewSketchLine(0, 1, 1, -1),
			1.0,
			el.NewSketchLine(1, -1, 1, 1),
			2.0,
			el.NewSketchPoint(2, 1, -2),
			Solved,
		},
	}
	for _, tt := range tests {
		newP3, state := pointFromLineLine(tt.l1, tt.l2, tt.p3, tt.l1Dist, tt.l2Dist)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == NonConvergent {
			continue
		}
		assert.Equal(t, tt.p3.GetID(), newP3.GetID(), tt.name)
		assert.InDelta(t, math.Abs(tt.l1.DistanceTo(newP3)), tt.l1Dist, utils.StandardCompare, tt.name)
		assert.InDelta(t, math.Abs(tt.l2.DistanceTo(newP3)), tt.l2Dist, utils.StandardCompare, tt.name)
	}
}
func TestPointFromLineLineExt(t *testing.T) {
	l1 := el.NewSketchLine(0, 1, 1, -1)
	l2 := el.NewSketchLine(1, -1, 1, 1)
	p3 := el.NewSketchPoint(2, 0.7, 1)

	referenceP3, _ := pointFromLineLine(l1, l2, p3, 1, 2)

	if utils.StandardFloatCompare(l1.DistanceTo(referenceP3), 1) != 0 {
		t.Error("Expected newP3 to have distance of 1 to l1, got ", l1.DistanceTo(referenceP3))
	}

	if utils.StandardFloatCompare(l2.DistanceTo(referenceP3), 2) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, l1, p3, 1, false)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2, p3, 2, false)

	newP3, state := PointFromLineLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromLineLine(c2, c1)

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

	newP3, state = PointFromLineLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromLineLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromLineLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromLineLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromLineLine(c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromLineLine(c2, c1)

	if utils.StandardFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestSolveConstraints(t *testing.T) {
	l1 := el.NewSketchLine(0, 0, 1, -1.1) // top line
	p1 := el.NewSketchPoint(2, 0.1, 1)    // top left
	p2 := el.NewSketchPoint(3, 1, 1.1)    // top right
	c1 := constraint.NewConstraint(0, constraint.Distance, p1, p2, 1, false)
	c2 := constraint.NewConstraint(1, constraint.Distance, p1, l1, 0, false)

	solved := SolveConstraints(c1, c2, p1)
	p1 = c1.Element1.(*el.SketchPoint)
	p2 = c1.Element2.(*el.SketchPoint)
	l1 = c2.Element2.(*el.SketchLine)

	if solved != Solved {
		t.Error("Expected solved state, got", solved)
	}

	if p1.DistanceTo(p2) != 1 {
		t.Error("Expected distance between p1 and p2 to be 1, found", p1.DistanceTo(p2))
	}

	if p1.DistanceTo(l1) != 0 {
		t.Error("Expected distance between p1 and l1 to be 0, found", p1.DistanceTo(l1))
	}
}

func TestSolveDistanceConstraint(t *testing.T) {
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		desired *el.SketchPoint
		state   SolveState
	}{
		{
			"Angle constraint passed",
			constraint.NewConstraint(0, constraint.Angle, el.NewSketchLine(0, 1, 1, 1), el.NewSketchLine(1, 2, 2, 2), 1, false),
			nil,
			NonConvergent,
		},
		{
			"Coincident point values with constraint value > 0",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchPoint(1, 1, 1), 1, false),
			nil,
			NonConvergent,
		},
		{
			"Coincident point values already solved",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchPoint(1, 1, 1), 0, false),
			el.NewSketchPoint(0, 1, 1),
			Solved,
		},
		{
			"Test 1",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchPoint(0, 1, 1), el.NewSketchPoint(1, 1, 2), 2, false),
			el.NewSketchPoint(0, 1, 0),
			Solved,
		},
		{
			"Test 2",
			constraint.NewConstraint(0, constraint.Distance, el.NewSketchLine(0, 0, 1, 1), el.NewSketchPoint(1, 1, 2), 2, false),
			el.NewSketchPoint(1, 1, 1),
			Solved,
		},
	}
	for _, tt := range tests {
		state := SolveDistanceConstraint(tt.c1)
		assert.Equal(t, tt.state, state, tt.name)
		if tt.state != Solved {
			continue
		}
		assert.True(t, tt.c1.IsMet(), tt.name)
		newPoint := tt.c1.Element1.AsPoint()
		if newPoint == nil {
			newPoint = tt.c1.Element2.AsPoint()
		}
		assert.Equal(t, tt.desired.GetID(), newPoint.GetID(), tt.name)
		assert.InDelta(t, tt.desired.GetX(), newPoint.GetX(), utils.StandardCompare, tt.name)
		assert.InDelta(t, tt.desired.GetY(), newPoint.GetY(), utils.StandardCompare, tt.name)
	}
}
