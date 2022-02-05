package dlineate

import (
	c "github.com/marcuswu/dlineate/internal/constraint"
)

// Type of a Constraint(Distance or Angle)
type ConstraintType uint

// ElementType constants
const (
	Coincident ConstraintType = iota
	Distance
	Angle
	Perpendicular
	Parallel
	Tangent

	// Two pass constraints
	Equal
)

type Constraint struct {
	constraints    []*c.Constraint
	elements       []*Element
	constraintType ConstraintType
	resolved       bool
}

func emptyConstraint() *Constraint {
	ec := new(Constraint)
	ec.constraints = make([]*c.Constraint, 0, 1)
	ec.elements = make([]*Element, 0, 1)
	ec.resolved = false
	return &Constraint{}
}

/*

One Pass Constraints
-------------
Distance constraint -- line segment, between elements, radius
Coincident constraint -- points, point & line, point & curve, line & curve
Equal constraint -- 2nd pass constraint
Distance ratio constraint -- 2nd pass constraint
Angle -- two lines
Perpendicular -- two lines
Parallel -- two lines

Two Pass Constraints
-------------
Equal angle -- 2nd pass constraint
Midpoint -- 2nd pass constraint (equal distances to either end of the line)
Tangent -- line and curve
Symmetric -- TODO

*/
