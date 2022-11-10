package dlineate

import (
	"math"
	"testing"

	"github.com/marcuswu/dlineate/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddAngleConstraint(t *testing.T) {
	UseLogger(utils.Logger)
	s := NewSketch()
	o := NewVector(0, 0, 0)
	xDir := NewVector(0, -1, 0)
	yDir := NewVector(1, 0, 0)
	wp := NewWorkPlane(o, xDir, yDir)
	s.SetWorkplane(wp)
	p1 := s.AddPoint(0, 0)
	l1 := s.AddLine(1, 0, 0, 1)
	c1, err := s.AddAngleConstraint(p1, l1, 45, false)

	assert.Nil(t, c1, "AngleConstraint should fail without two lines")
	assert.NotNil(t, err, "AngleConstraint should fail without two lines")

	l2 := s.AddLine(2, 0, 1, 0)
	c1, err = s.AddAngleConstraint(l1, l2, 40, false)

	assert.NotNil(t, c1, "AngleConstraint should be created")
	assert.Nil(t, err, "AngleConstraint should be created")
	assert.Equal(t, (40/180.0)*math.Pi, c1.constraints[0].Value, "Angle constraint is 40 degrees")

	c1, err = s.AddAngleConstraint(l1, l2, 40, true)

	assert.NotNil(t, c1, "AngleConstraint should be created")
	assert.Nil(t, err, "AngleConstraint should be created")
	assert.Equal(t, (140/180.0)*math.Pi, c1.constraints[0].Value, "Angle constraint is 140 degrees")
}

func TestAddCoincidentConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	l1 := s.AddLine(1, 1, 1, 2)
	// p2 := s.AddPoint(0, 0)
	c1 := s.AddCoincidentConstraint(p1, l1)
	assert.NotNil(t, c1, "Coincident points creates no constraint")
	assert.Equal(t, c1.constraintType, Coincident, "coincident constraint created")

	c2 := s.AddCoincidentConstraint(p1, s.Origin)

	assert.Equal(t, s.Origin.element.GetID(), p1.element.GetID(), "Coincident points share id")
	assert.Nil(t, c2, "Coincident points creates no constraint")

	assert.Equal(t, s.Origin.element.GetID(), c1.constraints[0].Element2.GetID(), "Updated id is reflected in previous constraints")
}

func TestAddDistanceConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(1, 0)

	c1 := s.AddDistanceConstraint(s.Origin, p1, 1)
	assert.NotNil(t, c1, "Distance constraint for points is created")
	assert.Equal(t, c1.state, Resolved, "Distance constraint is resolved")

	l1 := s.AddLine(0, 1, 1, 1)
	c2 := s.AddDistanceConstraint(l1, nil, 1)
	assert.Equal(t, c2.state, Resolved, "Line length constraint is resolved")

	c3 := s.AddDistanceConstraint(p1, l1, 1)
	assert.NotNil(t, c3, "Distance constraint for point and line is created")
	assert.Equal(t, c3.state, Resolved, "Distance constraint is resolved")

	circle1 := s.AddCircle(0, 1, 2)
	c4 := s.AddDistanceConstraint(l1, circle1, 0)
	assert.Equal(t, c4.state, Unresolved, "No constraints under the unresolved distance constraint")
	assert.Empty(t, c4.constraints, "No constraints under the unresolved distance constraint")

	c5 := s.AddDistanceConstraint(circle1, nil, 2)
	assert.Equal(t, c5.state, Resolved, "Circle radius constraint is resolved")

	ok := s.resolveCurveDistance(nil, nil, c4)
	assert.False(t, ok, "Circle curve distance is not resolved")

	s.resolveConstraint(c4)
	assert.Equal(t, c4.state, Resolved, "Setting circle radius resolves circle-line distance constraint")

	ok = s.resolveCurveDistance(circle1, nil, c5)
	assert.True(t, ok, "Circle curve distance is resolved")

	a1 := s.AddArc(0, 1, 2, 0, 0, 0)
	c6 := s.AddDistanceConstraint(a1, l1, 0)
	assert.Equal(t, c6.state, Unresolved, "No constraints under the unresolved distance constraint")
	assert.Empty(t, c6.constraints, "No constraints under the unresolved distance constraint")

	c7 := s.AddDistanceConstraint(a1, nil, 2)
	assert.Equal(t, c7.state, Resolved, "Arc radius constraint is resolved")

	ok = s.resolveCurveDistance(nil, nil, c6)
	assert.False(t, ok, "Arc curve distance is not resolved")

	s.resolveConstraint(c6)
	assert.Equal(t, c4.state, Resolved, "Setting arc radius resolves arc-line distance constraint")

	ok = s.resolveCurveDistance(circle1, nil, c7)
	assert.True(t, ok, "Arc curve distance is resolved")

	p1.elementType = 9
	c8 := s.AddDistanceConstraint(p1, l1, 1)
	assert.Equal(t, c8.state, Unresolved, "Distance cosntraint with unknown element type creates unresolved constraint")
}

func TestAddRatioConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(1, 0)
	l1 := s.AddLine(1, 0, 5, 5)
	c1 := s.AddRatioConstraint(p1, l1, 0.5)
	assert.Nil(t, c1, "Ratio constraint with a point is invalid")

	l2 := s.AddLine(2, 1, 3, 4)
	c2 := s.AddRatioConstraint(l1, l2, 0.5)
	assert.NotNil(t, c2, "Ratio constraint with lines is valid")
	assert.Equal(t, Unresolved, c2.state, "Ratio constraint should be unresolved without a line length")

	s.AddDistanceConstraint(l1, nil, 10)
	s.resolveConstraints()
	assert.Equal(t, Resolved, c2.state, "Ratio constraint should be resolved with a line length")

	c3 := s.AddRatioConstraint(l2, l1, 0.5)
	assert.NotNil(t, c3, "Ratio constraint with lines is valid")
	assert.Equal(t, Resolved, c3.state, "Ratio constraint should be resolved with a line length")

	l3 := s.AddLine(2, 1, 3, 4)
	curve1 := s.AddCircle(0, 1, 2)
	s.AddDistanceConstraint(curve1, nil, 2)
	c4 := s.AddRatioConstraint(l3, curve1, 0.5)
	assert.NotNil(t, c4, "Ratio constraint with lines is valid")
	assert.Equal(t, Resolved, c4.state, "Ratio constraint should be resolved with a line length")
	c5 := s.AddRatioConstraint(curve1, l3, 0.5)
	assert.NotNil(t, c5, "Ratio constraint with lines is valid")
	assert.Equal(t, Resolved, c5.state, "Ratio constraint should be resolved with a line length")
}

func TestAddEqualConstraint(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(1, 0, 1, 1)
	l2 := s.AddLine(1, 0, 5, 5)
	c1 := s.AddEqualConstraint(l1, l2)
	assert.NotNil(t, c1, "Equal constraint with lines is valid")
	assert.Equal(t, Unresolved, c1.state, "Equal constraint should be unresolved without a line length")
	c2 := s.AddDistanceConstraint(l1, nil, 5)
	s.resolveConstraints()
	assert.NotNil(t, c2, "Equal constraint with lines is valid")
	assert.Equal(t, Resolved, c1.state, "Equal constraint should be resolved with a line length")
	assert.Equal(t, Resolved, c2.state, "Equal constraint should be resolved with a line length")
}

func TestAddMidpointConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(5, 5)
	p2 := s.AddPoint(5, 5)
	c0 := s.AddMidpointConstraint(p1, p2)
	assert.Nil(t, c0, "Midpoint constraint with point and point is invalid")
	l1 := s.AddLine(1, 0, 1, 1)
	l2 := s.AddLine(1, 0, 1, 1)
	c0 = s.AddMidpointConstraint(l1, l2)
	assert.Nil(t, c0, "Midpoint constraint with line and line is invalid")
	c1 := s.AddMidpointConstraint(l1, p1)
	assert.NotNil(t, c1, "Midpoint constraint with point and line is valid")
	assert.Equal(t, Unresolved, c1.state, "Midpoint constraint should be unresolved without a line length")
	c2 := s.AddDistanceConstraint(l1, nil, 5)
	s.resolveConstraints()
	assert.NotNil(t, c2, "Midpoint constraint with point and line is valid")
	assert.Equal(t, Resolved, c1.state, "Midpoint constraint should be resolved with a line length")
	assert.Equal(t, Resolved, c2.state, "Midpoint constraint should be resolved with a line length")
	a1 := s.AddArc(0, 1, 2, 3, 4, 5)
	c3 := s.AddMidpointConstraint(a1, p1)
	s.resolveConstraints()
	assert.NotNil(t, c3, "Midpoint constraint with point and arc is valid")
	assert.Equal(t, Unresolved, c3.state, "Midpoint constraint should be unresolved without a fully constrained arc")
	c4 := s.AddDistanceConstraint(a1.Center(), s.Origin, 0) // This wouldn't really converge, but just for the test
	c5 := s.AddDistanceConstraint(a1.Start(), s.Origin, 0)
	c6 := s.AddDistanceConstraint(a1.End(), s.Origin, 0)
	c7 := s.AddDistanceConstraint(a1, nil, 2)
	c4.state = Solved
	c5.state = Solved
	c6.state = Solved
	c7.state = Solved
	s.resolveConstraints()
	for _, c := range c4.constraints {
		c.Solved = true
	}
	for _, c := range c5.constraints {
		c.Solved = true
	}
	for _, c := range c6.constraints {
		c.Solved = true
	}
	for _, c := range c7.constraints {
		c.Solved = true
	}
	for _, c := range a1.constraints {
		c.Solved = true
	}
	s.resolveConstraints()
	assert.Equal(t, Resolved, c3.state, "Midpoint constraint should be resolved with a fully constrained arc")
}

func TestAddParallelConstraint(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(0, 1, 2, 3)
	p1 := s.AddPoint(0, 0)
	l2 := s.AddLine(4, 5, 6, 7)
	c1, err := s.AddParallelConstraint(l1, p1)
	assert.Nil(t, c1, "Parallel constraint between point and line should error")
	assert.NotNil(t, err, "Parallel constraint between point and line should error")

	c1, err = s.AddParallelConstraint(l1, l2)
	assert.NotNil(t, c1, "Parallel constraint between line and line is valid")
	assert.Nil(t, err, "Parallel constraint between line and line is valid")
	assert.Equal(t, Resolved, c1.state, "Parallel constraint should be resolved")
}

func TestHorizontalConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	l1 := s.AddLine(0, 1, 2, 3)
	c1, err := s.AddHorizontalConstraint(p1)
	assert.Nil(t, c1, "Horizontal constraint on a point should error")
	assert.NotNil(t, err, "Horizontal constraint on a point should error")

	c1, err = s.AddHorizontalConstraint(l1)
	assert.NotNil(t, c1, "Horizontal constraint on a line is valid")
	assert.Nil(t, err, "Horizontal constraint on a line is valid")
	assert.Equal(t, Resolved, c1.state, "Horizontal constraint should be resolved")
}

func TestVerticalConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(0, 0)
	l1 := s.AddLine(0, 1, 2, 3)
	c1, err := s.AddVerticalConstraint(p1)
	assert.Nil(t, c1, "Vertical constraint on a point should error")
	assert.NotNil(t, err, "Vertical constraint on a point should error")

	c1, err = s.AddVerticalConstraint(l1)
	assert.NotNil(t, c1, "Vertical constraint on a line is valid")
	assert.Nil(t, err, "Vertical constraint on a line is valid")
	assert.Equal(t, Resolved, c1.state, "Vertical constraint should be resolved")
}

func TestAddPerpendicularConstraint(t *testing.T) {
	s := NewSketch()
	l1 := s.AddLine(0, 1, 2, 3)
	p1 := s.AddPoint(0, 0)
	l2 := s.AddLine(4, 5, 6, 7)
	c1, err := s.AddPerpendicularConstraint(l1, p1)
	assert.Nil(t, c1, "Perpendicular constraint between point and line should error")
	assert.NotNil(t, err, "Perpendicular constraint between point and line should error")

	c1, err = s.AddPerpendicularConstraint(l1, l2)
	assert.NotNil(t, c1, "Perpendicular constraint between line and line is valid")
	assert.Nil(t, err, "Perpendicular constraint between line and line is valid")
	assert.Equal(t, Resolved, c1.state, "Perpendicular constraint should be resolved")
}

func TestTangentConstraint(t *testing.T) {
	s := NewSketch()
	p1 := s.AddPoint(1, 1)
	l1 := s.AddLine(2, 3, 5, 8)
	a1 := s.AddArc(13, 21, 34, 55, 89, 144)
	c1, err := s.AddTangentConstraint(p1, l1)
	assert.Nil(t, c1, "Tangent constraint between point and line should error")
	assert.NotNil(t, err, "Tangent constraint between point and line should error")

	c1, err = s.AddTangentConstraint(l1, a1)
	assert.NotNil(t, c1, "Tangent constraint between line and arc is valid")
	assert.Nil(t, err, "Tangent constraint between line and arc is valid")
	assert.Equal(t, Unresolved, c1.state, "Tangent constraint should be unresolved")

	s.AddDistanceConstraint(a1, nil, 3)
	s.resolveConstraints()
	assert.Equal(t, Resolved, c1.state, "Tangent constraint should be resolved")

	c1, err = s.AddTangentConstraint(a1, l1)
	assert.NotNil(t, c1, "Tangent constraint between arc and line is valid")
	assert.Nil(t, err, "Tangent constraint between arc and line is valid")
	assert.Equal(t, Resolved, c1.state, "Tangent constraint should be resolved")
}
