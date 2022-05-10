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
	constraint.state = Resolved

	return constraint
}

func (s *Sketch) AddAngleConstraint(p1 *Element, p2 *Element, v float64) (*Constraint, error) {
	c := AngleConstraint(p1, p2)

	if p1.elementType != Line || p1.elementType != Axis || p2.elementType != Line || p2.elementType != Axis {
		return nil, errors.New("incorrect element types for angle constraint")
	}

	constraint := s.sketch.AddConstraint(ic.Angle, p1.element, p2.element, v)
	p1.constraints = append(p1.constraints, constraint)
	p2.constraints = append(p2.constraints, constraint)
	c.constraints = append(c.constraints, constraint)
	s.constraints = append(s.constraints, c)
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)

	return c, nil
}
