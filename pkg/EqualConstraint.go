package dlineate

func EqualConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Distance
	constraint.resolved = false

	return constraint
}

func (s *Sketch) AddEqualConstraint(p1 *Element, p2 *Element) *Constraint {
	c := EqualConstraint(p1, p2)

	if p1.elementType == Point || p2.elementType == Point {
		return nil
	}

	// First look to see if either element has a distance constraint
	dc, err := s.findConstraint(Distance, p1)
	if err == nil {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p2, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.resolved = true

		return c
	}

	dc, err = s.findConstraint(Distance, p2)
	if err == nil {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p1, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.resolved = true

		return c
	}

	// Determine if we can resolve this constraint indirectly

	return c
}
