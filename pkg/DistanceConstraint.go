package dlineate

import (
	ic "github.com/marcuswu/dlineate/internal/constraint"
)

func DistanceConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Distance
	constraint.resolved = true

	return constraint
}

func (s *Sketch) addDistanceConstraint(p1 *Element, p2 *Element, v float64) *ic.Constraint {
	switch p1.elementType {
	case Point:
		if p2.elementType != Point {
			return s.addDistanceConstraint(p2, p1, v)
		}

		return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	case Circle:
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.elements[0], nil, v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	case Line:
		isCircle := p2.elementType == Circle
		isArc := p2.elementType == Arc
		if isArc || isCircle {
			return s.addDistanceConstraint(p2, p1, v)
		}
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.elements[1], p1.elements[2], v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	case Arc:
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p1.elements[1], v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	default:
		return s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	}
}

func (s *Sketch) AddDistanceConstraint(p1 *Element, p2 *Element, v float64) *Constraint {
	c := DistanceConstraint(p1, p2)

	constraint := s.addDistanceConstraint(p1, p2, v)
	if constraint != nil {
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	}

	return c
}
