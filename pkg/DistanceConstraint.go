package dlineation

import (
	"fmt"

	ic "github.com/marcuswu/dlineation/internal/constraint"
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
	case Axis:
		fallthrough
	case Line:
		if p2 == nil {
			fmt.Printf(
				"Adding distance constraint for line %d. Translating to distance constraint between points %d and %d\n",
				p1.element.GetID(),
				p1.children[0].element.GetID(),
				p1.children[1].element.GetID(),
			)
			return s.sketch.AddConstraint(ic.Distance, p1.children[0].element, p1.children[1].element, v)
		}
		isCircle := p2.elementType == Circle
		isArc := p2.elementType == Arc
		if isArc || isCircle {
			return s.addDistanceConstraint(p2, p1, v)
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
	fmt.Printf("AddDistanceConstraint: added constraint id %d\n", constraint.GetID())
	if constraint != nil {
		p1.constraints = append(p1.constraints, constraint)
		if p2 != nil {
			p2.constraints = append(p2.constraints, constraint)
		}
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	}

	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	if p2 != nil {
		s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)
	}
	return c
}
