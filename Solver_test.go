package dlineate

import (
	"bytes"
	"errors"
	"testing"

	"github.com/marcuswu/dlineate/internal/constraint"
	"github.com/marcuswu/dlineate/internal/element"
	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestResolveConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 1)
	l1 := s.AddLine(2, 3, 4, 5)
	l2 := s.AddLine(6, 7, 8, 9)
	a1 := s.AddArc(0, 1, 2, 3, 4, 5)

	angle, err := s.AddAngleConstraint(l1, l2, 30, true)
	assert.Nil(t, err, "There should be no error adding an Angle constraint")
	parallel, err := s.AddParallelConstraint(l1, l2)
	assert.Nil(t, err, "There should be no error adding an Parallel constraint")
	perpendicular, err := s.AddPerpendicularConstraint(l1, l2)
	assert.Nil(t, err, "There should be no error adding an Perpendicular constraint")
	tangent, err := s.AddTangentConstraint(a1, l2)
	assert.Nil(t, err, "There should be no error adding an Tangent constraint")

	tests := []struct {
		name         string
		constraint   *Constraint
		desiredState ConstraintState
	}{
		{"Unknown constraint", s.AddDistanceConstraint(l1, p1, 2), Unresolved},
		{"Resolved constraint", s.AddDistanceConstraint(l1, p1, 2), Resolved},
		{"Not found constraint", s.AddDistanceConstraint(l1, p1, 2), Resolved},
		{"Coincident constraint", s.AddCoincidentConstraint(l1, p1), Resolved},
		{"Distance constraint", s.AddDistanceConstraint(l1.Start(), l1.End(), 1), Resolved},
		{"Angle constraint", angle, Resolved},
		{"Parallel constraint", parallel, Resolved},
		{"Perpendicular constraint", perpendicular, Resolved},
		{"Ratio constraint", s.AddRatioConstraint(l1, a1, 2), Solved},
		{"Midpoint constraint", s.AddMidpointConstraint(p1, l1), Resolved},
		{"Tangent constraint", tangent, Unresolved},
	}
	for _, c := range s.constraints {
		c.state = Unresolved
	}

	tests[0].constraint.constraintType = 100
	tests[1].constraint.state = Resolved
	ic := constraint.NewConstraint(20, constraint.Distance, l1.Start().element.GetID(), p1.element.GetID(), 2, false)
	tests[2].constraint.constraints = append(tests[2].constraint.constraints, ic)

	s.resolveConstraints()
	assert.True(t, s.resolveConstraint(tests[1].constraint), "Resolved constraint should return resolved")
	assert.Equal(t, tests[1].desiredState, tests[1].constraint.state, "Resolved constraint should return resolved")

	for _, tt := range tests {
		assert.NotNil(t, tt.constraint, "%s: Ensure constraint was created", tt.name)
		assert.Equal(t, tt.desiredState, tt.constraint.state, "%s: Ensure desired constraint state", tt.name)
	}
}

func TestGetDistanceConstraint(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(2, 3, 4, 5)
	s.AddDistanceConstraint(l1.Start(), l1.End(), 1)

	c, ok := s.getDistanceConstraint(l1)

	assert.True(t, ok, "Should find a constraint")
	assert.Equal(t, c.constraintType, Distance, "Should find a Distance constraint")
	assert.Contains(t, c.elements, l1.Start(), "Constraint should contain l1.Start")
	assert.Contains(t, c.elements, l1.End(), "Constraint should contain l1.End")
}

func TestResolveLineLength(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(2, 3, 4, 5)
	l2 := s.AddLine(2, 3, 4, 5)
	c1 := s.AddDistanceConstraint(l1.Start(), l2, 1)
	c1.constraintType = 20
	s.AddDistanceConstraint(l1.Start(), l1.End(), 1)
	length, ok := s.resolveLineLength(l1)

	assert.True(t, ok, "Should find a length")
	assert.Equal(t, 1.0, length, "Should find the correct length")

	c0 := s.eToC[s.Origin.id][0]
	c1 = s.AddDistanceConstraint(l2.Start(), s.Origin, 0)
	c2 := s.AddDistanceConstraint(l2.End(), s.XAxis, 0)
	c3 := s.AddDistanceConstraint(l2.End(), l1.End(), 0)
	for _, c := range l2.constraints {
		c.Solved = true
	}
	c0.state = Solved
	for _, c := range c0.constraints {
		c.Solved = true
	}
	c1.state = Solved
	for _, c := range c1.constraints {
		c.Solved = true
	}
	c2.state = Solved
	for _, c := range c2.constraints {
		c.Solved = true
	}
	c3.state = Solved
	for _, c := range c3.constraints {
		c.Solved = true
	}

	length, ok = s.resolveLineLength(l2)

	assert.True(t, ok, "Should find a length")
	assert.InDelta(t, 2.828427, length, utils.StandardCompare, "Should find the correct length")
}

func TestResolveCurveRadius(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(3, 0)
	c1 := s.AddCircle(0, 0, 2)
	l1 := s.AddLine(3, 0, 3, 3)

	s.AddTangentConstraint(l1, c1)
	s.AddCoincidentConstraint(c1.Center(), s.Origin)
	for _, c := range s.Origin.constraints {
		c.Solved = true
	}
	s.AddDistanceConstraint(c1, p1, 0)
	constraint3 := s.AddCoincidentConstraint(s.XAxis, p1)
	constraint3.state = Solved
	for _, c := range constraint3.constraints {
		c.Solved = true
	}
	constraint4 := s.AddDistanceConstraint(s.Origin, p1, 3)
	constraint4.state = Solved
	for _, c := range constraint4.constraints {
		c.Solved = true
	}

	r, ok := s.resolveCurveRadius(c1)
	assert.True(t, ok, "Curve radius should be resolved")
	assert.Equal(t, 3.0, r, "Curve radius should be 3")
}

func TestSolve(t *testing.T) {
	s := NewSketch()
	c1 := s.AddCircle(0, 0, 3)
	s.AddCoincidentConstraint(c1.Center(), s.Origin)

	err := s.Solve()
	assert.NotNil(t, err, "Expected inconsistent constraints")
	assert.Equal(t, errors.New("failed to solve completely"), err, "Should not solve")

	s = NewSketch()
	c1 = s.AddCircle(0, 0, 3)
	l1 := s.AddLine(0, 0, 1, 1)
	s.AddTangentConstraint(c1, l1)
	s.AddCoincidentConstraint(l1.Start(), s.Origin)
	s.AddDistanceConstraint(l1, nil, 5)
	s.AddCoincidentConstraint(c1.Center(), l1.End())

	err = s.Solve()
	s.ConflictingConstraints()
	assert.NotNil(t, err, "Expected inconsistent constraints")
	assert.Equal(t, errors.New("failed to solve completely"), err, "Should not solve")

	s = NewSketch()

	// Add elements
	l1 = s.AddLine(0.0, 0.0, 3.13, 0.0)
	l2 := s.AddLine(3.13, 0.0, 5.14, 2.27)
	l3 := s.AddLine(5.14, 2.27, 2.28, 4.72)
	l4 := s.AddLine(2.28, 4.72, -1.04, 3.56)
	l5 := s.AddLine(-1.04, 3.56, 0.0, 0.0)

	// Add constraints
	// Bottom of pentagon starts at origin and aligns with x axis
	s.AddCoincidentConstraint(s.Origin, l1.Start())
	s.AddParallelConstraint(s.XAxis, l1)

	// line points are coincident
	s.AddCoincidentConstraint(l1.End(), l2.Start())
	s.AddCoincidentConstraint(l2.End(), l3.Start())
	s.AddCoincidentConstraint(l3.End(), l4.Start())
	s.AddCoincidentConstraint(l4.End(), l5.Start())
	s.AddCoincidentConstraint(l5.End(), l1.Start())

	// 108 degrees between lines (skip 2 to not over constrain)
	s.AddAngleConstraint(l2, l3, 108, true)
	s.AddAngleConstraint(l3, l4, 108, true)
	s.AddAngleConstraint(l4, l5, 108, true)

	// 4 unit length on lines (skip 1 to not over constrain)
	s.AddDistanceConstraint(l1, nil, 4.0)
	s.AddDistanceConstraint(l2, nil, 4.0)
	s.AddDistanceConstraint(l4, nil, 4.0)
	s.AddDistanceConstraint(l5, nil, 4.0)

	// Solve
	err = s.Solve()
	s.ConflictingConstraints()
	assert.Nil(t, err, "Expected no error")

	minx, miny, maxx, maxy := s.calculateRectangle(1.0)
	assert.InDelta(t, minx, -1.236068, utils.StandardCompare, "MinX")
	assert.InDelta(t, miny, 0, utils.StandardCompare, "MinY")
	assert.InDelta(t, maxx, 5.236068, utils.StandardCompare, "MaxX")
	assert.InDelta(t, maxy, 6.155367, utils.StandardCompare, "MaxY")

	var b bytes.Buffer
	s.AddArc(0, 0, -1, -0.1, -1, 0)
	c := s.AddCircle(0, 0, 2)
	s.MakeFixed(c)
	c.Center().element.SetConstraintLevel(element.OverConstrained)
	err = s.WriteImage(&b, 500, 200)
	assert.Nil(t, err, "Expect no error from WriteImage")
	assert.Contains(t, b.String(), "svg", "wrote an svg")
}
