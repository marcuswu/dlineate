package dlineate

/*
 * Order matters for ratio constraints. p2's magnitude = p1's magnitude * constraint value
 */
func RatioConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Ratio
	constraint.state = Unresolved

	return constraint
}

func (s *Sketch) AddRatioConstraint(p1 *Element, p2 *Element, v float64) *Constraint {
	c := RatioConstraint(p1, p2)
	c.dataValue = v

	if p1.elementType == Point || p2.elementType == Point {
		return nil
	}
	s.eToC[p1.element.GetID()] = append(s.eToC[p1.element.GetID()], c)
	s.eToC[p2.element.GetID()] = append(s.eToC[p2.element.GetID()], c)

	s.resolveRatioConstraint(c)

	return c
}

func (s *Sketch) resolveRatioConstraint(c *Constraint) bool {
	p1 := c.elements[0]
	p2 := c.elements[1]

	// All line tests
	dist, ok := s.resolveLineLength(p1)
	if ok {
		constraint := s.addDistanceConstraint(p2, nil, dist*c.dataValue)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}
	dist, ok = s.resolveLineLength(p2)
	if ok {
		constraint := s.addDistanceConstraint(p1, nil, dist/c.dataValue)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	// Circles and Arcs with solved center and solved elements coincident or distance to the circle / arc
	p1Radius, ok := s.resolveCurveRadius(p1)
	if ok {
		constraint := s.addDistanceConstraint(p2, nil, p1Radius*c.dataValue)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	p2Radius, ok := s.resolveCurveRadius(p2)
	if ok {
		constraint := s.addDistanceConstraint(p2, nil, p2Radius/c.dataValue)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	return c.state == Resolved
}
