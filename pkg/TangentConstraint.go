package dlineation

import (
	"errors"
	"fmt"
)

func TangentConstraint(p1 *Element, p2 *Element, p3 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.elements = append(constraint.elements, p3)
	constraint.constraintType = Tangent
	constraint.state = Unresolved

	return constraint
}

func (s *Sketch) AddTangentConstraint(p1 *Element, p2 *Element, p3 *Element) (*Constraint, error) {
	var line, point, curve, err = orderParams(p1, p2, p3)

	if err != nil {
		return nil, err
	}

	c := TangentConstraint(line, point, curve)
	fmt.Printf("for element %d adding ratio constraint to elements %d and %d\n", p1.element.GetID(), p2.element.GetID(), p3.element.GetID())
	fmt.Printf("for element %d adding ratio constraint to elements %d and %d\n", p2.element.GetID(), p1.element.GetID(), p3.element.GetID())
	fmt.Printf("for element %d adding ratio constraint to elements %d and %d\n", p3.element.GetID(), p1.element.GetID(), p2.element.GetID())
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)
	s.eToC[p3.element.GetID()] = append(s.eToC[p3.element.GetID()], c)

	s.resolveTangentConstraint(c)

	return c, nil
}

func orderParams(p1 *Element, p2 *Element, p3 *Element) (*Element, *Element, *Element, error) {
	var line, point, curve *Element

	switch Line {
	case p1.elementType:
		line = p1
	case p2.elementType:
		line = p2
	default:
		line = p3
	}

	switch Point {
	case p1.elementType:
		point = p1
	case p2.elementType:
		point = p2
	default:
		point = p3
	}

	switch true {
	case p1.elementType == Circle || p1.elementType == Arc:
		curve = p1
	case p2.elementType == Circle || p2.elementType == Arc:
		curve = p2
	default:
		curve = p3
	}

	if line == point || line == curve || point == curve {
		return p1, p2, p3, errors.New("incorrect element types for tangent constraint")
	}

	return line, point, curve, nil
}

func (s *Sketch) resolveTangentConstraint(c *Constraint) bool {
	radius, ok := s.resolveCurveRadius(c.elements[2])
	if ok {
		constraint := s.addDistanceConstraint(c.elements[0], c.elements[2], radius)
		fmt.Printf("resolveTangentConstraint: added constraint id %d\n", constraint.GetID())
		c.elements[0].constraints = append(c.elements[0].constraints, constraint)
		c.elements[2].constraints = append(c.elements[2].constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		constraint = s.addDistanceConstraint(c.elements[1], c.elements[2], radius)
		fmt.Printf("resolveTangentConstraint: added constraint id %d\n", constraint.GetID())
		c.elements[1].constraints = append(c.elements[1].constraints, constraint)
		c.elements[2].constraints = append(c.elements[2].constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		constraint = s.addDistanceConstraint(c.elements[1], c.elements[0], 0)
		fmt.Printf("resolveTangentConstraint: added constraint id %d\n", constraint.GetID())
		c.elements[1].constraints = append(c.elements[1].constraints, constraint)
		c.elements[0].constraints = append(c.elements[0].constraints, constraint)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved
	}

	return c.state == Resolved
}
