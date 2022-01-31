package dlineate

import (
	"errors"

	ic "github.com/marcuswu/dlineate/internal/constraint"
)

func AngleConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Angle
	constraint.resolved = true

	return constraint
}

func (s *Sketch) AddAngleConstraint(p1 *Element, p2 *Element, v float64) (*Constraint, error) {
	c := AngleConstraint(p1, p2)

	if p1.elementType != Line || p2.elementType != Line {
		return nil, errors.New("incorrect element types for angle constraint")
	}

	constraint := s.sketch.AddConstraint(ic.Angle, p1.elements[0], p2.elements[0], v)
	c.constraints = append(c.constraints, constraint)
	s.constraints = append(s.constraints, c)

	return c, nil
}
