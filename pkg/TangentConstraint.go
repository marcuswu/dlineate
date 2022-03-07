package dlineate

import (
	"errors"
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

	// TODO: Look for curve radius. Can be defined as a Distance constraint or derived from a 2nd pass

	/*constraint := s.sketch.AddConstraint(ic.Distance, p1.elements[0], p2.elements[0], v)
	c.constraints = append(c.constraints, constraint)
	s.constraints = append(s.constraints, c)*/

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
