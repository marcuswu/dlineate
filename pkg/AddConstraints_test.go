package dlineation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAngleConstraint(t *testing.T) {
	s := NewSketch()
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

func TestMidpointConstraint(t *testing.T) {
	// s := NewSketch()
	// s.AddPoint(1, 0)
	// s.AddLine(1, 0, 5, 5)
	// c1 :=
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
}
