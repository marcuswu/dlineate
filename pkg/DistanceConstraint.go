package dlineate

import (
	ic "github.com/marcuswu/dlineate/internal/constraint"
)

func DistanceConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Distance
	constraint.state = Resolved

	return constraint
}

func (s *Sketch) addDistanceConstraint(p1 *Element, p2 *Element, v float64) *ic.Constraint {
	switch p1.elementType {
	case Point:
		if p2.elementType != Point {
			return s.addDistanceConstraint(p2, p1, v)
		}

		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	case Circle:
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.children[0].element, nil, v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	case Line:
		isCircle := p2.elementType == Circle
		isArc := p2.elementType == Arc
		if isArc || isCircle {
			return s.addDistanceConstraint(p2, p1, v)
		}
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.children[0].element, p1.children[1].element, v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	case Arc:
		if p2 == nil {
			return s.sketch.AddConstraint(ic.Distance, p1.element, p1.children[0].element, v)
		}
		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	default:
		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	}
}

func (s *Sketch) AddDistanceConstraint(p1 *Element, p2 *Element, v float64) *Constraint {
	c := DistanceConstraint(p1, p2)

	constraint := s.addDistanceConstraint(p1, p2, v)
	if constraint != nil {
		p1.constraints = append(p1.constraints, constraint)
		p2.constraints = append(p2.constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	}

	return c
}
