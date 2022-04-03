package dlineate

func (e *Element) isLineOrArc() bool {
	return e.elementType == Line || e.elementType == Arc
}

func MidpointConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Midpoint
	constraint.state = Unresolved

	return constraint
}

/*
 * A midpoint is coincident AND half the distance away from one end.
 * Only applies to a line or an arc
 */
func (s *Sketch) AddMidpointConstraint(p1 *Element, p2 *Element) *Constraint {
	c := RatioConstraint(p1, p2)

	if p1.elementType != Point || p2.elementType != Point {
		return nil
	}
	if !p1.isLineOrArc() && !p2.isLineOrArc() {
		return nil
	}

	s.resolveMidpointConstraint(c)

	return c
}

func (s *Sketch) resolveMidpointConstraint(c *Constraint) bool {
	/*
	 * The line or arc must be fully constrained and solved first
	 */
	point := c.elements[0]
	other := c.elements[1]
	if c.elements[1].elementType == Point {
		point = c.elements[1]
		other = c.elements[0]
	}

	// Line tests
	dist, ok := s.resolveLineLength(other)
	if ok {
		// coincident with line
		constraint := s.addDistanceConstraint(other, point, 0)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		// distance from start
		constraint = s.addDistanceConstraint(other.children[0], point, dist/2.0)
		c.state = Resolved

		return c.state == Resolved
	}

	// Ensure start and end of arc is fully constrained and solved
	// calculate angle between lines formed from center to start and center to end
	// calculate line through center with half that angle
	// place midpoint at radius distance from center along calculated line

	return c.state == Resolved
}
