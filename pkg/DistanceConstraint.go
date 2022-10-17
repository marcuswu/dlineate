package dlineation

import (
	"fmt"

	ic "github.com/marcuswu/dlineation/internal/constraint"
)

func DistanceConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	if p2 != nil {
		constraint.elements = append(constraint.elements, p2)
	}
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
			// If p2 is nil, we're setting the circle radius
			// This is more of a placeholder for being able to fulfill other constraints as there is no
			// element to constrain to a distance from the center
			//return s.sketch.AddConstraint(ic.Distance, p1.children[0].element, nil, v)
			// Add a constraint to pkg/Sketch (not translatable to internal solver)
			return nil
		}
		// r, ok := s.resolveCurveRadius(p1)
		// if !ok {
		// 	break nil
		// }
		// c = s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, r+v)
		return nil
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
			// Add a constraint to pkg/Sketch (not translatable to internal solver)
			// If p2 is nil, we're setting the arc radius, so distance to start or end works
			//return s.sketch.AddConstraint(ic.Distance, p1.element, p1.children[1].element, v)
			return nil
		}
		// If p2 is not nil, we need to know the arc's radius is constrained
		// r, ok := s.resolveCurveRadius(p1)
		// if !ok {
		// 	return nil
		// }
		// return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, r+v)
		return nil
	default:
		return s.sketch.AddConstraint(ic.Distance, p1.element, p2.element, v)
	}
}

func (s *Sketch) AddDistanceConstraint(p1 *Element, p2 *Element, v float64) *Constraint {
	c := DistanceConstraint(p1, p2)

	constraint := s.addDistanceConstraint(p1, p2, v)
	if constraint != nil {
		fmt.Printf("AddDistanceConstraint: added constraint id %d\n", constraint.GetID())
		p1.constraints = append(p1.constraints, constraint)
		if p2 != nil {
			p2.constraints = append(p2.constraints, constraint)
		}
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
	} else {
		c.dataValue = v
		c.state = Unresolved
	}

	// This is might be wrong unless p1.element is always in constraint c -- for a line this is not true
	// check to see how eToC is used!
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	if p2 != nil {
		s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)
	}

	s.resolveDistanceConstraint(c)

	return c
}

func (s *Sketch) resolveCurveDistance(e1 *Element, e2 *Element, c *Constraint) bool {
	if c.state == Resolved {
		return c.state == Resolved
	}
	if e1 == nil {
		return false
	}
	eRadius, ok := s.resolveCurveRadius(e1)
	if ok {
		fmt.Printf("RESOLVED curve radius with center point (%f, %f) to %f\n", e1.values[0], e1.values[1], eRadius)
	}
	if !ok {
		return false
	}

	constraint := s.addDistanceConstraint(e1, e2, eRadius+c.dataValue)
	fmt.Printf("resolveDistanceConstraint: added constraint id %d\n", constraint.GetID())
	e1.constraints = append(e1.constraints, constraint)
	c.constraints = append(c.constraints, constraint)
	s.constraints = append(s.constraints, c)
	c.state = Resolved

	return c.state == Resolved
}

func (s *Sketch) resolveDistanceConstraint(c *Constraint) bool {
	p1 := c.elements[0]
	var p2 *Element
	if len(c.elements) > 1 {
		p2 = c.elements[1]
	}

	if s.resolveCurveDistance(p1, p2, c) {
		return c.state == Resolved
	}

	if s.resolveCurveDistance(p2, p1, c) {
		return c.state == Resolved
	}

	return c.state == Resolved
}
