package solver

import (
	"math"
	"math/big"
	"testing"

	"github.com/marcuswu/dlineate/internal/accessors"
	"github.com/marcuswu/dlineate/internal/constraint"
	el "github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestPointFromPoints(t *testing.T) {
	tests := []struct {
		name   string
		p1     *el.SketchPoint
		p1Dist *big.Float
		p2     *el.SketchPoint
		p2Dist *big.Float
		p3     *el.SketchPoint
		state  SolveState
	}{
		{
			"Test Nonconvergent",
			el.NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(1.0),
			el.NewSketchPoint(1, big.NewFloat(3), big.NewFloat(5)),
			big.NewFloat(3.0),
			el.NewSketchPoint(2, big.NewFloat(0), big.NewFloat(2)),
			NonConvergent,
		},
		{
			"Test 1",
			el.NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(1.0),
			el.NewSketchPoint(1, big.NewFloat(3), big.NewFloat(5)),
			big.NewFloat(5.0),
			el.NewSketchPoint(2, big.NewFloat(0), big.NewFloat(2)),
			Solved,
		},
		{
			"Test exact distance",
			el.NewSketchPoint(0, big.NewFloat(3), big.NewFloat(1)),
			big.NewFloat(1.0),
			el.NewSketchPoint(1, big.NewFloat(3), big.NewFloat(5)),
			big.NewFloat(3.0),
			el.NewSketchPoint(2, big.NewFloat(2), big.NewFloat(2)),
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
		var result big.Float
		result.Abs(tt.p1.DistanceTo(newP3))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&result, tt.p1Dist), tt.name)
		result.Abs(tt.p2.DistanceTo(newP3))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&result, tt.p2Dist), tt.name)
	}
}

func TestPointFromPointsExt(t *testing.T) {
	ea := accessors.NewElementRepository()
	var newP3 *el.SketchPoint = nil
	p1 := el.NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	p2 := el.NewSketchPoint(1, big.NewFloat(3), big.NewFloat(5))
	p3 := el.NewSketchPoint(2, big.NewFloat(0), big.NewFloat(2))
	ea.AddElement(p1)
	ea.AddElement(p2)
	ea.AddElement(p3)

	referenceP3, _ := GetPointFromPoints(p1, p2, p3, big.NewFloat(1), big.NewFloat(5))

	if utils.StandardBigFloatCompare(p1.DistanceTo(referenceP3), big.NewFloat(1)) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardBigFloatCompare(p2.DistanceTo(referenceP3), big.NewFloat(5)) != 0 {
		t.Error("Expected newP3 to have distance of 5 to p2, got ", p2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, p1.GetID(), p3.GetID(), big.NewFloat(1), false)

	c2 := constraint.NewConstraint(1, constraint.Distance, p2.GetID(), p3.GetID(), big.NewFloat(5), false)

	newP3, state := PointFromPoints(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3.String(), newP3.String())
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3.String(), newP3.String())
	}

	newP3, state = PointFromPoints(-1, ea, c2, c1)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3.String(), newP3.String())
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %s, newP3 %s\n", referenceP3.String(), newP3.String())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPoints(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPoints(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromPoints(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPoints(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPoints(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPoints(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromPointLine(t *testing.T) {
	p1 := el.NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	l2 := el.NewSketchLine(1, big.NewFloat(1), big.NewFloat(1), big.NewFloat(2*math.Sqrt(0.5)))
	p3 := el.NewSketchPoint(2, big.NewFloat(0), big.NewFloat(2))

	_, state := pointFromPointLine(p1, l2, p3, big.NewFloat(1), big.NewFloat(1))

	if state != NonConvergent {
		t.Error("Expected non-convergent state got ", state)
	}

	newP3, state := pointFromPointLine(p1, l2, p3, big.NewFloat(1), big.NewFloat(2))

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(p1.DistanceTo(newP3), big.NewFloat(1)) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got", p1.DistanceTo(newP3))
	}

	if utils.StandardBigFloatCompare(l2.DistanceTo(newP3), big.NewFloat(2)) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got", l2.DistanceTo(newP3))
	}

	p3 = el.NewSketchPoint(2, big.NewFloat(2), big.NewFloat(1))

	_, state = pointFromPointLine(p1, l2, p3, big.NewFloat(1), big.NewFloat(5))

	if state != NonConvergent {
		t.Error("Expected non convergent state got ", state)
	}
}

func TestPointFromPointLineExt(t *testing.T) {
	ea := accessors.NewElementRepository()
	p1 := el.NewSketchPoint(0, big.NewFloat(1), big.NewFloat(1))
	l2 := el.NewSketchLine(1, big.NewFloat(1), big.NewFloat(1), big.NewFloat(2))
	p3 := el.NewSketchPoint(2, big.NewFloat(0), big.NewFloat(2))
	ea.AddElement(p1)
	ea.AddElement(l2)
	ea.AddElement(p3)

	referenceP3, _ := pointFromPointLine(p1, l2, p3, big.NewFloat(1), big.NewFloat(2.5))

	if utils.StandardBigFloatCompare(p1.DistanceTo(referenceP3), big.NewFloat(1)) != 0 {
		t.Error("Expected newP3 to have distance of 1 to p1, got ", p1.DistanceTo(referenceP3))
	}

	if utils.StandardBigFloatCompare(l2.DistanceTo(referenceP3), big.NewFloat(2.5)) != 0 {
		t.Error("Expected newP3 to have distance of 2.5 to p2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, p1.GetID(), p3.GetID(), big.NewFloat(1), false)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2.GetID(), p3.GetID(), big.NewFloat(2.5), false)

	newP3, state := PointFromPointLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromPointLine(-1, ea, c2, c1)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPointLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPointLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromPointLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPointLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromPointLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromPointLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestPointFromLineLine(t *testing.T) {
	tests := []struct {
		name   string
		l1     *el.SketchLine
		l1Dist *big.Float
		l2     *el.SketchLine
		l2Dist *big.Float
		p3     *el.SketchPoint
		state  SolveState
	}{
		{
			"Test Nonconvergent",
			el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(1), big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(1.0),
			el.NewSketchPoint(2, big.NewFloat(0.7), big.NewFloat(1)),
			NonConvergent,
		},
		{
			"Test Parallel",
			el.NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(-3)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(1), big.NewFloat(0)),
			big.NewFloat(2.0),
			el.NewSketchPoint(2, big.NewFloat(0.7), big.NewFloat(1.8)),
			Solved,
		},
		{
			"Test Intersect 1",
			el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(-1), big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(2.0),
			el.NewSketchPoint(2, big.NewFloat(0.7), big.NewFloat(1)),
			Solved,
		},
		{
			"Test Intersect 2",
			el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(-1), big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(2.0),
			el.NewSketchPoint(2, big.NewFloat(3), big.NewFloat(0)),
			Solved,
		},
		{
			"Test Intersect 3",
			el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(-1), big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(2.0),
			el.NewSketchPoint(2, big.NewFloat(-1), big.NewFloat(0)),
			Solved,
		},
		{
			"Test Intersect 4",
			el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1)),
			big.NewFloat(1.0),
			el.NewSketchLine(1, big.NewFloat(-1), big.NewFloat(1), big.NewFloat(1)),
			big.NewFloat(2.0),
			el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(-2)),
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
		var result big.Float
		result.Abs(tt.l1.DistanceTo(newP3))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&result, tt.l1Dist), tt.name)
		result.Abs(tt.l2.DistanceTo(newP3))
		assert.Equal(t, 0, utils.StandardBigFloatCompare(&result, tt.l2Dist), tt.name)
	}
}
func TestPointFromLineLineExt(t *testing.T) {
	ea := accessors.NewElementRepository()
	l1 := el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(-1))
	l2 := el.NewSketchLine(1, big.NewFloat(-1), big.NewFloat(1), big.NewFloat(1))
	p3 := el.NewSketchPoint(2, big.NewFloat(0.7), big.NewFloat(1))
	ea.AddElement(l1)
	ea.AddElement(l2)
	ea.AddElement(p3)

	referenceP3, _ := pointFromLineLine(l1, l2, p3, big.NewFloat(1), big.NewFloat(2))

	if utils.StandardBigFloatCompare(l1.DistanceTo(referenceP3), big.NewFloat(1)) != 0 {
		t.Error("Expected newP3 to have distance of 1 to l1, got ", l1.DistanceTo(referenceP3))
	}

	if utils.StandardBigFloatCompare(l2.DistanceTo(referenceP3), big.NewFloat(2)) != 0 {
		t.Error("Expected newP3 to have distance of 2 to l2, got ", l2.DistanceTo(referenceP3))
	}

	c1 := constraint.NewConstraint(0, constraint.Distance, l1.GetID(), p3.GetID(), big.NewFloat(1), false)

	c2 := constraint.NewConstraint(1, constraint.Distance, l2.GetID(), p3.GetID(), big.NewFloat(2), false)

	newP3, state := PointFromLineLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, state = PointFromLineLine(-1, ea, c2, c1)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetX(), newP3.GetX())
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %f, newP3 %f\n", referenceP3.GetY(), newP3.GetY())
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromLineLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromLineLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1
	c2.Element1, c2.Element2 = c2.Element2, c2.Element1

	newP3, state = PointFromLineLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromLineLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	c1.Element1, c1.Element2 = c1.Element2, c1.Element1

	newP3, state = PointFromLineLine(-1, ea, c1, c2)

	if state != Solved {
		t.Error("Expected solved state got ", state)
	}

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	newP3, _ = PointFromLineLine(-1, ea, c2, c1)

	if utils.StandardBigFloatCompare(newP3.GetX(), referenceP3.GetX()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}

	if utils.StandardBigFloatCompare(newP3.GetY(), referenceP3.GetY()) != 0 {
		t.Errorf("Expected newP3 to to be equivalent to the reference, got reference %v, newP3 %v\n", referenceP3, newP3)
	}
}

func TestSolveConstraints(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(0), big.NewFloat(1), big.NewFloat(-1.1)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1.1)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	c0 := constraint.NewConstraint(0, constraint.Distance, 1, 2, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 1, 0, big.NewFloat(0), false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 3, 4, big.NewFloat((70.0/180.0)*math.Pi), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(1), false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name     string
		c1       *constraint.Constraint
		c2       *constraint.Constraint
		solveFor el.SketchElement
		state    SolveState
	}{
		{"Test Solve For Point", c0, c1, el.NewSketchPoint(1, big.NewFloat(0.1), big.NewFloat(1)), Solved},
		{"Test Solve For Line", c2, c3, el.NewSketchLine(4, big.NewFloat(0.151089), big.NewFloat(0.988520), big.NewFloat(-0.139610)), Solved},
	}
	for _, tt := range tests {
		solved := SolveConstraints(-1, ea, tt.c1, tt.c2, tt.solveFor)
		assert.Equal(t, tt.state, solved, tt.name)
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
		assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
	}
}

func TestSolveDistanceConstraint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(1), big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(2), big.NewFloat(2), big.NewFloat(2)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(3, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(4, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(6, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(7, big.NewFloat(1), big.NewFloat(2)))
	ea.AddElement(el.NewSketchLine(8, big.NewFloat(0), big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchPoint(9, big.NewFloat(1), big.NewFloat(2)))
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 3, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 4, 5, big.NewFloat(0), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 6, 7, big.NewFloat(2), false)
	c4 := constraint.NewConstraint(4, constraint.Distance, 8, 9, big.NewFloat(2), false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	ca.AddConstraint(c4)

	tests := []struct {
		name    string
		c1      *constraint.Constraint
		desired *el.SketchPoint
		state   SolveState
	}{
		{"Angle constraint passed", c0, nil, NonConvergent},
		{"Coincident point values with constraint value > 0", c1, nil, NonConvergent},
		{"Coincident point values already solved", c2, el.NewSketchPoint(4, big.NewFloat(1), big.NewFloat(1)), Solved},
		{"Test 1", c3, el.NewSketchPoint(6, big.NewFloat(1), big.NewFloat(0)), Solved},
		{"Test 2", c4, el.NewSketchPoint(9, big.NewFloat(1), big.NewFloat(2)), Solved},
	}
	for _, tt := range tests {
		state := SolveDistanceConstraint(-1, ea, tt.c1)
		assert.Equal(t, tt.state, state, tt.name)
		if tt.state != Solved {
			continue
		}
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
		e, _ := ea.GetElement(-1, tt.c1.Element1)
		e2, _ := ea.GetElement(-1, tt.c1.Element2)
		newPoint := e.AsPoint()
		if newPoint == nil {
			e, _ := ea.GetElement(-1, tt.c1.Element2)
			newPoint = e.AsPoint()
		}
		assert.Equal(t, tt.desired.GetID(), newPoint.GetID(), tt.name)
		t.Logf("Desired: %s, Calculated: %s\n", tt.desired.String(), newPoint.String())
		t.Logf("Other: %s\n", e2.String())
		// for Test 1, we get 0, 0 instead of 1, 0
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetX(), newPoint.GetX()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetY(), newPoint.GetY()), tt.name)
	}
}

func TestPointResult(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.25), big.NewFloat(0)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.151089), big.NewFloat(0.988520), big.NewFloat(-0.139610)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Distance, 5, 3, big.NewFloat(1), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(1), false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchPoint
		state   SolveState
	}{
		{"Test PointFromPoints", c0, c1, el.NewSketchPoint(1, big.NewFloat(0.515383), big.NewFloat(0.125274)), Solved},
		{"Test PointFromLineLine", c2, c3, el.NewSketchPoint(5, big.NewFloat(0.745353), big.NewFloat(1.038922)), Solved},
	}
	for _, tt := range tests {
		newPoint, state := PointResult(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
		e, _ := tt.c1.Shared(tt.c2)
		shared, _ := ea.GetElement(-1, e)
		if shared == nil || tt.state == NonConvergent {
			continue
		}
		shared.AsPoint().X = newPoint.X
		shared.AsPoint().Y = newPoint.Y
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
		assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
		assert.Equal(t, tt.desired.GetID(), shared.GetID(), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetX(), shared.AsPoint().GetX()), tt.name)
		assert.Equal(t, 0, utils.StandardBigFloatCompare(tt.desired.GetY(), shared.AsPoint().GetY()), tt.name)
	}
}

func TestSolveForPoint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(0.151089), big.NewFloat(0.988520), big.NewFloat(-0.139610)))
	ea.AddElement(el.NewSketchLine(2, big.NewFloat(1), big.NewFloat(1), big.NewFloat(0)))
	c0 := constraint.NewConstraint(0, constraint.Angle, 2, 0, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Angle, 2, 1, big.NewFloat(1), false)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired *el.SketchPoint
		state   SolveState
	}{
		{"Test Nonconvergent", c0, c1, nil, NonConvergent},
	}
	for _, tt := range tests {
		state := SolveForPoint(-1, ea, tt.c1, tt.c2)
		assert.Equal(t, state, tt.state, tt.name)
	}
}

func TestConstraintResult(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	ea.AddElement(el.NewSketchPoint(0, big.NewFloat(1.5), big.NewFloat(0.3)))
	ea.AddElement(el.NewSketchPoint(1, big.NewFloat(0.25), big.NewFloat(0)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(3, big.NewFloat(1.5), big.NewFloat(0.3), big.NewFloat(0.1)))
	ea.AddElement(el.NewSketchLine(4, big.NewFloat(0.3), big.NewFloat(1.5), big.NewFloat(-0.1)))
	ea.AddElement(el.NewSketchPoint(5, big.NewFloat(1), big.NewFloat(1)))
	c0 := constraint.NewConstraint(0, constraint.Distance, 0, 1, big.NewFloat(1), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 1, big.NewFloat(1), false)
	c2 := constraint.NewConstraint(2, constraint.Angle, 3, 4, big.NewFloat((70.0/180.0)*math.Pi), false)
	c3 := constraint.NewConstraint(3, constraint.Distance, 5, 4, big.NewFloat(1), false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	ca.AddConstraint(c2)
	ca.AddConstraint(c3)
	tests := []struct {
		name    string
		c1      *constraint.Constraint
		c2      *constraint.Constraint
		desired el.SketchElement
		state   SolveState
	}{
		{"Test Point Solve", c0, c1, el.NewSketchPoint(1, big.NewFloat(0.515383), big.NewFloat(0.125274)), Solved},
		{"Test Line Solve", c2, c3, el.NewSketchLine(4, big.NewFloat(0.151089), big.NewFloat(0.988520), big.NewFloat(-0.139610)), Solved},
	}
	for _, tt := range tests {
		result, state := ConstraintResult(-1, ea, tt.c1, tt.c2, tt.desired)
		assert.Equal(t, state, tt.state, tt.name)
		if tt.state == NonConvergent {
			continue
		}
		c1e, _ := ea.GetElement(-1, tt.desired.GetID())
		if c1p := c1e.AsPoint(); c1p != nil {
			c1p.X = result.AsPoint().X
			c1p.Y = result.AsPoint().Y
		}
		if c1l := c1e.AsLine(); c1l != nil {
			c1l.SetA(result.AsLine().GetA())
			c1l.SetB(result.AsLine().GetB())
			c1l.SetC(result.AsLine().GetC())
		}
		c2e, _ := ea.GetElement(-1, tt.desired.GetID())
		if c2p := c2e.AsPoint(); c2p != nil {
			c2p.X = result.AsPoint().X
			c2p.Y = result.AsPoint().Y
		}
		if c2l := c2e.AsLine(); c2l != nil {
			c2l.SetA(result.AsLine().GetA())
			c2l.SetB(result.AsLine().GetB())
			c2l.SetC(result.AsLine().GetC())
		}
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
		assert.True(t, ca.IsMet(tt.c2.GetID(), -1, ea), tt.name)
	}

}

func TestSolveConstraint(t *testing.T) {
	ea := accessors.NewElementRepository()
	ca := accessors.NewConstraintRepository()
	ea.AddElement(el.NewSketchLine(0, big.NewFloat(0.98), big.NewFloat(0), big.NewFloat(1)))
	ea.AddElement(el.NewSketchLine(1, big.NewFloat(0), big.NewFloat(0.98), big.NewFloat(0)))
	ea.AddElement(el.NewSketchPoint(2, big.NewFloat(1), big.NewFloat(0)))
	ea.AddElement(el.NewSketchPoint(3, big.NewFloat(1), big.NewFloat(1)))
	c0 := constraint.NewConstraint(0, constraint.Angle, 0, 1, big.NewFloat(math.Pi/2), false)
	c1 := constraint.NewConstraint(1, constraint.Distance, 2, 3, big.NewFloat(1.2), false)
	ca.AddConstraint(c0)
	ca.AddConstraint(c1)
	tests := []struct {
		name  string
		c1    *constraint.Constraint
		state SolveState
	}{
		{"Solve Angle Constraint", c0, Solved},
		{"Solve Distance Constraint", c1, Solved},
	}
	for _, tt := range tests {
		state := SolveConstraint(-1, ea, tt.c1)
		assert.Equal(t, tt.state, state, tt.name)
		assert.True(t, ca.IsMet(tt.c1.GetID(), -1, ea), tt.name)
	}
}
