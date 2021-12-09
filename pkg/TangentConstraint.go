package dlineate

import (
	"errors"

	ic "github.com/marcuswu/dlineate/internal/constraint"
)

func TangentConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Tangent

	return constraint
}

func (s *Sketch) AddTangentConstraint(p1 *Element, p2 *Element, v float64) (*Constraint, error) {
	c := TangentConstraint(p1, p2)

	l := p1
	curve := p2
	if l.elementType != Line {
		l = p2
		curve = p1
	}
	if l.elementType != Line || (curve.elementType != Circle && curve.elementType != Arc) {
		return nil, errors.New("Incorrect element types for tangent constraint")
	}

	constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	c.constraints = append(c.constraints, constraint)
	s.constraints = append(s.constraints, c)

	return c, nil
}
