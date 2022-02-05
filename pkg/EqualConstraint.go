package dlineate

func EqualConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Equal
	constraint.resolved = false

	return constraint
}

func (s *Sketch) AddEqualConstraint(p1 *Element, p2 *Element) *Constraint {
	c := EqualConstraint(p1, p2)

	if p1.elementType == Point || p2.elementType == Point {
		return nil
	}

	s.resolveEqualConstraint(c)

	return c
}

func (s *Sketch) resolveEqualConstraint(c *Constraint) bool {
	p1 := c.elements[0]
	p2 := c.elements[1]

	// First look to see if either element has a distance constraint
	dc, err := s.findConstraint(Distance, p1)
	if err == nil {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p2, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.resolved = true

		return c.resolved
	}

	dc, err = s.findConstraint(Distance, p2)
	if err == nil {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p1, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.resolved = true

		return c.resolved
	}

	// Determine if we can resolve this constraint indirectly
	/*
	  A Line is equal via its length
	  	* A line's length is constrained if
			* There is a distance constraint against the Line
	    	* If the start and end points of the Line are fully constrained
	  Circles and arcs are equal via their radii
	  	* A Circle or arc's radius is constrained if
			* There is a distance constraint against the Circle
			* The center is fully constrained and there is a coincident constraint to the Circle
		
		A Constraint may be:
			* Unresolved -- defined, but has a value dependent on other constraints being solved first
			* Resolved -- has a determined value (directly or via other solved constraints)
			* Solved -- is Resolved *and* the elements it relates meet the constraint criteria

		Need to know if one of the constraint elements is fully constrained by solved constraints
			Then, the equality itself can be resolved
	  
	 */
	if p1.elementType == 

	return c.resolved
}
