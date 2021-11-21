package dlineate

import (
	el "github.com/marcuswu/dlineate/internal/element"
	c "github.com/marcuswu/dlineate/internal/constraint"
)

// Type of a Constraint(Distance or Angle)
type ConstraintType uint

// ElementType constants
const (
	Coincident ConstraintType = iota
	Distance
)
type Constraint struct {
	constraints []*c.Constraint
	elements []*el.SketchElement
	constraintType ConstraintType
}

func emptyConstraint() *Constraint {
	ec := new(Constraint)
	ec.constraints = make([]*c.Constraint, 0, 1)
	ec.elements = make([]*el.SketchElement, 0, 1)
	return &Constraint{}
}