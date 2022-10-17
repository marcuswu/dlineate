package dlineation

import (
	c "github.com/marcuswu/dlineation/internal/constraint"
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
	Ratio
	Midpoint
)

type ConstraintState uint

const (
	Unresolved ConstraintState = iota
	Resolved
	Solved
)

type Constraint struct {
	constraints    []*c.Constraint
	elements       []*Element
	constraintType ConstraintType
	state          ConstraintState
	dataValue      float64
}

func emptyConstraint() *Constraint {
	ec := new(Constraint)
	ec.constraints = make([]*c.Constraint, 0, 1)
	ec.elements = make([]*Element, 0, 1)
	ec.state = Unresolved
	return ec
}

func (c *Constraint) checkSolved() {
	solved := true
	if len(c.constraints) == 0 {
		solved = false
	}
	for _, constraint := range c.constraints {
		solved = solved && constraint.Solved
	}
	if solved {
		c.state = Solved
	}
}

/*

One Pass Constraints
-------------
Distance constraint -- line segment, between elements, radius
Coincident constraint -- points, point & line, point & curve, line & curve
Angle -- two lines
Perpendicular -- two lines
Parallel -- two lines

Two Pass Constraints
-------------
Equal constraint -- 2nd pass constraint
Distance ratio constraint -- 2nd pass constraint
Midpoint -- 2nd pass constraint (equal distances to either end of the line or arc)
Tangent -- line and curve
Symmetric -- TODO

*/
