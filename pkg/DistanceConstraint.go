package dlineate

import (
	ic "github.com/marcuswu/dlineate/internal/constraint"
)

func DistanceConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Distance

	return constraint
}

func (s *Sketch) AddDistanceConstraint(p1 *Element, p2 *Element, v float64) *Constraint {
	c := DistanceConstraint(p1, p2)

	// Add constraints to sketch
	switch p1.elementType {
	case Point:
		if p2.elementType != Point {
			return s.AddDistanceConstraint(p2, p1, v)
		}

		constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	case Circle:
		if p2 == nil {
			constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], nil, v)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			return c
		}
		constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	case Line:
		isCircle := p2.elementType == Circle
		isArc := p2.elementType == Arc
		if isArc || isCircle {
			return s.AddDistanceConstraint(p2, p1, v)
		}
		if p2 == nil {
			constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[1], p1.elements[2], v)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			return c
		}
		constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	case Arc:
		if p2 == nil {
			constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p1.elements[1], v)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			return c
		}
		constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	default:
		constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	}

	return c
}
