package dlineate

func EqualConstraint(p1 *Element, p2 *Element) *Constraint {
	constraint := emptyConstraint()
	constraint.elements = append(constraint.elements, p1)
	constraint.elements = append(constraint.elements, p2)
	constraint.constraintType = Equal
	constraint.state = Unresolved

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
	/*
		 		Determine if we can resolve this constraint indirectly
				  A Line is equal via its length
				  	* A line's length is constrained if
						* There is a distance constraint against the Line (Test A)
						* There is a distance constraint between the start and end points (Test B)
				    	* If the start and end points of the Line are fully constrained (Test C)
				  Circles and arcs are equal via their radii
				  	* A Circle or arc's radius is constrained if
						* There is a distance constraint against the Circle (Test A)
						* The center is fully constrained and there is a coincident or distance (Test C)
							constraint between the curve and a fully constrained element

					A Constraint may be:
						* Unresolved -- defined, but has a value dependent on other constraints being solved first
						    (need a value to pass to the solver)
						* Resolved -- has a determined value (directly or via other solved constraints)
						* Solved -- is Resolved *and* the elements it relates meet the constraint criteria

					Need to know if one of the constraint elements is fully constrained by solved constraints
						Then, the equality itself can be resolved
	*/
	p1 := c.elements[0]
	p2 := c.elements[1]

	// Tests A & B
	dc, ok := s.getDistanceConstraint(p1)
	if ok {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p2, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	dc, ok = s.getDistanceConstraint(p2)
	if ok {
		v := dc.constraints[0].Value
		constraint := s.addDistanceConstraint(p1, nil, v)
		c.constraints = append(c.constraints, constraint)
		s.constraints = append(s.constraints, c)
		c.state = Resolved

		return c.state == Resolved
	}

	// Check to see if fully constrained elements can give us a length -- Test A for lines
	if p1.elementType == Line {
		startConstrained := s.isElementSolved(p1.children[0])
		endConstrained := s.isElementSolved(p1.children[1])
		if startConstrained && endConstrained {
			// resolve constraint setting p2's distance to the distance from p1 start to p1 end
			v := p1.children[0].element.AsPoint().DistanceTo(p1.children[1].element.AsPoint())
			constraint := s.addDistanceConstraint(p2, nil, v)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			c.state = Resolved

			return c.state == Resolved
		}
	}
	if p2.elementType == Line {
		startConstrained := s.isElementSolved(p2.children[0])
		endConstrained := s.isElementSolved(p2.children[1])
		if startConstrained && endConstrained {
			// resolve constraint setting p1's distance to the distance from p2 start to p2 end
			v := p2.children[0].element.AsPoint().DistanceTo(p2.children[1].element.AsPoint())
			constraint := s.addDistanceConstraint(p1, nil, v)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			c.state = Resolved

			return c.state == Resolved
		}
	}

	// Circles and Arcs with solved center and solved elements coincident or distance to the circle / arc
	if centerSolved := s.isElementSolved(p1.children[0]); (p1.elementType == Circle || p1.elementType == Arc) && centerSolved {
		p1All := s.findConstraints(p1)
		var other *Element = nil
		for _, p1c := range p1All {
			if p1c.constraintType != Distance && p1c.constraintType != Coincident {
				continue
			}
			other = p1c.elements[0]
			if other == p1 {
				other = p1c.elements[1]
			}
			if !s.isElementSolved(other) {
				continue
			}
			// Other & p1 have a distance constraint between them. dist(other, p1.center) - c.value is radius
			v := other.element.AsPoint().DistanceTo(p1.children[0].element.AsPoint())
			constraint := s.addDistanceConstraint(p2, nil, v-p1c.constraints[0].Value)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			c.state = Resolved

			return c.state == Resolved
		}
	}

	if centerSolved := s.isElementSolved(p2.children[0]); (p2.elementType == Circle || p2.elementType == Arc) && centerSolved {
		p2All := s.findConstraints(p2)
		var other *Element = nil
		for _, p2c := range p2All {
			if p2c.constraintType != Distance && p2c.constraintType != Coincident {
				continue
			}
			other = p2c.elements[0]
			if other == p2 {
				other = p2c.elements[1]
			}
			if !s.isElementSolved(other) {
				continue
			}
			// Other & p1 have a distance constraint between them. dist(other, p1.center) - c.value is radius
			v := other.element.AsPoint().DistanceTo(p2.children[0].element.AsPoint())
			constraint := s.addDistanceConstraint(p1, nil, v-p2c.constraints[0].Value)
			c.constraints = append(c.constraints, constraint)
			s.constraints = append(s.constraints, c)
			c.state = Resolved

			return c.state == Resolved
		}
	}

	return c.state == Resolved
}
